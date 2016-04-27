package main

import (
	"container/list"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"net"
	"sync"
	"time"
)

type MessageHandler func(*MessageFromBroker)

type BrokerManager interface {
	ConnectBroker(brokerInfo *common.BrokerInfo, conn *net.TCPConn) (err error)
	GetBrokerAddr(brokerID common.UUID) *string
	GetBrokerInfo(brokerID common.UUID) *common.BrokerInfo
	IsBrokerAlive(brokerID common.UUID) bool
	GetLiveBroker() *Broker
	TerminateBroker(brokerID common.UUID)
	SendToBroker(brokerID common.UUID, message common.Sendable) (err error)
	BroadcastToBrokers(message common.Sendable)
	//RequestHeartbeat(brokerID common.UUID) (err error)
	HandlePubClientRemapping(msg *common.BrokerRequestMessage) (*common.BrokerAssignmentMessage, error)
}

type MessageFromBroker struct {
	message common.Sendable
	broker  *Broker
}

type BrokerManagerImpl struct {
	brokerMap          map[common.UUID]*Broker
	brokerAddrMap      map[string]*Broker // Map of Broker contact address -> broker
	mapLock            sync.RWMutex
	heartbeatInterval  int // seconds
	clock              common.Clock
	deadBrokerQueue    *list.List
	liveBrokerQueue    *list.List
	queueLock          sync.Mutex
	internalDeathChan  chan *Broker
	BrokerDeathChan    chan *common.UUID        // Written to when a broker dies
	BrokerLiveChan     chan *common.UUID        // Written to when a broker comes alive
	BrokerReassignChan chan *BrokerReassignment // Written to when a client or pub is reassigned
	MessageBuffer      chan *MessageFromBroker  // Buffers incoming messages meant for other systems
}

func NewBrokerManager(heartbeatInterval int, brokerDeathChan, brokerLiveChan chan *common.UUID,
	messageBuffer chan *MessageFromBroker, brokerReassignChan chan *BrokerReassignment, clock common.Clock) *BrokerManagerImpl {
	bm := new(BrokerManagerImpl)
	bm.brokerMap = make(map[common.UUID]*Broker)
	bm.brokerAddrMap = make(map[string]*Broker)
	bm.mapLock = sync.RWMutex{}
	bm.heartbeatInterval = heartbeatInterval
	bm.clock = clock
	bm.deadBrokerQueue = list.New()
	bm.liveBrokerQueue = list.New()
	bm.queueLock = sync.Mutex{}
	bm.internalDeathChan = make(chan *Broker, 10)

	bm.MessageBuffer = messageBuffer
	bm.BrokerDeathChan = brokerDeathChan
	bm.BrokerLiveChan = brokerLiveChan
	bm.BrokerReassignChan = brokerReassignChan
	go bm.monitorDeathChan()
	return bm
}

// If broker already exists in the mapping, this should be a reconnection
//   so simply update with the new connection and restart
// Otherwise, create a new Broker, put it into the map, and start it
func (bm *BrokerManagerImpl) ConnectBroker(brokerInfo *common.BrokerInfo, conn *net.TCPConn) (err error) {
	bm.mapLock.RLock()
	brokerConn, ok := bm.brokerMap[brokerInfo.BrokerID]
	bm.mapLock.RUnlock()
	if ok {
		brokerConn.WaitForCleanup()
		bm.queueLock.Lock()
		brokerConn.RemoveFromList(bm.deadBrokerQueue)
		bm.queueLock.Unlock()
		bm.mapLock.RLock()
		//for _, brok := range bm.brokerMap {
		// TODO I think this will be handled within the forwarding table now
		//	// Terminate any clients which migrated so that they will come back
		//	brok.TerminateRemotePubClients(&brokerInfo.BrokerID)
		//}
		bm.mapLock.RUnlock()
	} else {
		messageHandler := bm.createMessageHandler(brokerInfo.BrokerID)
		brokerConn = NewBroker(brokerInfo, messageHandler, time.Duration(bm.heartbeatInterval)*time.Second,
			bm.clock, bm.internalDeathChan)
		bm.mapLock.Lock()
		bm.brokerMap[brokerInfo.BrokerID] = brokerConn
		bm.brokerAddrMap[brokerInfo.BrokerAddr] = brokerConn
		bm.mapLock.Unlock()
	}
	brokerConn.StartAsynchronously(conn)
	bm.queueLock.Lock()
	brokerConn.PushToList(bm.liveBrokerQueue)
	bm.queueLock.Unlock()
	bm.BrokerLiveChan <- &brokerInfo.BrokerID
	return
}

func (bm *BrokerManagerImpl) GetBrokerAddr(brokerID common.UUID) *string {
	bm.mapLock.RLock()
	bconn, ok := bm.brokerMap[brokerID]
	bm.mapLock.RUnlock()
	if !ok {
		log.WithField("brokerID", brokerID).Warn("Attempted to get address of a nonexistent broker")
		return nil
	}
	return &bconn.BrokerAddr
}

func (bm *BrokerManagerImpl) GetBrokerInfo(brokerID common.UUID) *common.BrokerInfo {
	bm.mapLock.RLock()
	bconn, ok := bm.brokerMap[brokerID]
	bm.mapLock.RUnlock()
	if !ok {
		log.WithField("brokerID", brokerID).Warn("Attempted to get info of a nonexistent broker")
		return nil
	}
	return &bconn.BrokerInfo
}

func (bm *BrokerManagerImpl) IsBrokerAlive(brokerID common.UUID) bool {
	bm.mapLock.RLock()
	bconn, ok := bm.brokerMap[brokerID]
	bm.mapLock.RUnlock()
	if !ok {
		log.WithField("brokerID", brokerID).Warn("Attempted to check liveness of a nonexistent broker")
		return false
	}
	return bconn.IsAlive()
}

// Get an arbitrary live broker; for redirecting clients/publishers with a dead main broker
// Will return nil if none are available
func (bm *BrokerManagerImpl) GetLiveBroker() *Broker {
	bm.queueLock.Lock()
	defer bm.queueLock.Unlock()
	tempDeadList := list.New()
	// Store until the end so that we can push the dead ones back onto
	// the live list; we don't want to mess up the routines that will
	// be moving them later
	defer func() {
		for tempDeadList.Len() > 0 {
			e := tempDeadList.Front()
			b := e.Value.(*Broker)
			b.RemoveFromList(tempDeadList)
			b.PushToList(bm.liveBrokerQueue)
		}
	}()
	for {
		elem := bm.liveBrokerQueue.Front()
		if elem == nil {
			return nil
		}
		bconn := elem.Value.(*Broker)
		bconn.RemoveFromList(bm.liveBrokerQueue)
		if bconn.IsAlive() { // Double check liveness to be sure
			bconn.PushToList(bm.liveBrokerQueue)
			return bconn
		} else {
			bconn.PushToList(tempDeadList)
		}
	}
}

// Terminate the Broker and remove from our map
func (bm *BrokerManagerImpl) TerminateBroker(brokerID common.UUID) {
	bm.mapLock.RLock()
	bconn, ok := bm.brokerMap[brokerID]
	bm.mapLock.RUnlock()
	if !ok {
		log.WithField("brokerID", brokerID).Warn("Attempted to Terminate a nonexistent broker")
		return
	}
	bconn.Terminate()
	bm.mapLock.Lock()
	delete(bm.brokerMap, brokerID)
	delete(bm.brokerAddrMap, bconn.BrokerAddr)
	bm.mapLock.Unlock()
}

// Asynchronously send a message to the given broker
func (bm *BrokerManagerImpl) SendToBroker(brokerID common.UUID, message common.Sendable) (err error) {
	bm.mapLock.RLock()
	defer bm.mapLock.RUnlock()
	brokerConn, ok := bm.brokerMap[brokerID]
	if !ok {
		log.WithField("brokerID", brokerID).Warn("Attempted to send message to nonexistent broker")
		return errors.New("Broker was unable to be found")
	}
	brokerConn.Send(message)
	return
}

// Asynchronously send a message to all currently living brokers
// Automatically adjusts the MessageID of the messages to be distinct
// if the message contains it
func (bm *BrokerManagerImpl) BroadcastToBrokers(message common.Sendable) {
	bm.mapLock.RLock()
	defer bm.mapLock.RUnlock()
	switch msg := message.(type) {
	case common.SendableWithID:
		for _, brokerConn := range bm.brokerMap {
			msg.SetID(common.GetMessageID())
			brokerConn.Send(msg)
		}
	case common.Sendable:
		for _, brokerConn := range bm.brokerMap {
			brokerConn.Send(message)
		}
	}
}

// First confirm the lack of liveness of the home broker
// If the home broker is still alive, simply redirect the client back there and do nothing else
// If it is confirmed dead, pick a random live broker and redirect the client there
func (bm *BrokerManagerImpl) HandlePubClientRemapping(msg *common.BrokerRequestMessage) (*common.BrokerAssignmentMessage, error) {
	var (
		homeBroker *Broker
		newBroker  *Broker
		brokerDead bool
		found      bool
	)
	bm.mapLock.RLock()
	if homeBroker, found = bm.brokerAddrMap[msg.LocalBrokerAddr]; !found {
		return nil, errors.New("Local broker could not be found")
	}
	bm.mapLock.RUnlock()

	if !homeBroker.IsAlive() {
		brokerDead = true
	} else {
		brokerDead = !homeBroker.RequestHeartbeatAndWait()
		homeBroker.WaitForCleanup()
	}

	if brokerDead { // actions to take if broker determined dead
		newBroker = bm.GetLiveBroker()
		//if msg.IsPublisher {
		//	newBroker.AddRemotePublisher(common.UUID(msg.UUID), &homeBroker.BrokerID)
		//} else {
		//	newBroker.AddRemoteClient(msg.UUID, &homeBroker.BrokerID)
		//}
		bm.BrokerReassignChan <- &BrokerReassignment{
			HomeBrokerID: &homeBroker.BrokerID,
			IsPublisher:  msg.IsPublisher,
			UUID:         &msg.UUID,
		}
	} else {
		newBroker = homeBroker // no real redirect necessary
	}
	return &common.BrokerAssignmentMessage{
		BrokerInfo: common.BrokerInfo{
			BrokerID:   newBroker.BrokerID,
			BrokerAddr: newBroker.BrokerAddr,
		},
	}, nil
}

// Send a request for a heartbeat to broker; should be sent if contacted by a client
// saying that they couldn't contact the broker
// TODO this shouldn't be necessary?
//func (bm *BrokerManagerImpl) RequestHeartbeat(brokerID common.UUID) (err error) {
//	bm.mapLock.RLock()
//	defer bm.mapLock.RUnlock()
//	brokerConn, ok := bm.brokerMap[brokerID]
//	if !ok {
//		log.WithField("brokerID", brokerID).Warn("Attempted to send heartbeat request to nonexistent broker")
//		return errors.New("Broker was unable to be found")
//	}
//	brokerConn.RequestHeartbeat()
//	return
//}

func (bm *BrokerManagerImpl) createMessageHandler(brokerID common.UUID) MessageHandler {
	return func(brokerMessage *MessageFromBroker) {
		message := brokerMessage.message
		switch msg := message.(type) {
		case *common.BrokerConnectMessage:
			log.WithFields(log.Fields{
				"connBrokerID": brokerID, "newBrokerID": msg.BrokerID, "newBrokerAddr": msg.BrokerAddr,
			}).Warn("Received a BrokerConnectMessage over an existing broker connection")
		case *common.BrokerTerminateMessage:
			log.WithField("brokerID", brokerID).Info("BrokerManagerImpl terminating broker")
			brokerMessage.broker.Send(&common.AcknowledgeMessage{MessageID: msg.MessageID})
			bm.mapLock.Lock()
			delete(bm.brokerMap, brokerMessage.broker.BrokerID)
			delete(bm.brokerAddrMap, brokerMessage.broker.BrokerAddr)
			bm.mapLock.Unlock()
			brokerMessage.broker.Terminate()
		// Messages below this take some action in the BrokerManager but also forward beyond
		// TODO cleanup once I'm sure this isn't needed
		//case *common.ClientTerminationMessage:
		//	brokerMessage.broker.RemoveClient(msg.ClientID)
		//	bm.MessageBuffer <- brokerMessage
		//case *common.PublisherTerminationMessage:
		//	brokerMessage.broker.RemovePublisher(msg.PublisherID)
		//	bm.MessageBuffer <- brokerMessage
		//case *common.BrokerPublishMessage:
		//	brokerMessage.broker.AddLocalPublisher(msg.UUID)
		//	bm.MessageBuffer <- brokerMessage
		//case *common.BrokerQueryMessage:
		//	brokerMessage.broker.AddLocalClient(msg.UUID)
		//	bm.MessageBuffer <- brokerMessage
		default:
			bm.MessageBuffer <- brokerMessage
		}
	}
}

// Monitor Broker death, moving the dying broker to the correct
// queue. Also forward the broker on to the outward Death chan
func (bm *BrokerManagerImpl) monitorDeathChan() {
	for {
		deadBroker, ok := <-bm.internalDeathChan
		log.WithField("broker", deadBroker).Debug("Broker determined as dead!")
		if !ok {
			return
		}
		bm.queueLock.Lock()
		deadBroker.RemoveFromList(bm.liveBrokerQueue)
		deadBroker.PushToList(bm.deadBrokerQueue)
		bm.queueLock.Unlock()

		//deaths := PubClientDeaths{
		//	BrokerID:        &deadBroker.BrokerID,
		//	PublisherDeaths: deadBroker.GetAndClearPublishers(),
		//	ClientDeaths:    deadBroker.GetAndClearClients(),
		//}
		//if len(deaths.PublisherDeaths) > 0 || len(deaths.ClientDeaths) > 0 {
		//	bm.pubClientDeathChan <- &deaths
		//}

		bm.BrokerDeathChan <- &deadBroker.BrokerID // Forward on to outward death chan
		bm.BroadcastToBrokers(&common.BrokerDeathMessage{BrokerInfo: deadBroker.BrokerInfo})
	}
}
