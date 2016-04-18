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

type MessageHandler func(common.Sendable)

type BrokerManager struct {
	brokerMap         map[common.UUID]*Broker
	mapLock           sync.RWMutex
	heartbeatInterval int // seconds
	clock             common.Clock
	deadBrokerQueue   *list.List
	liveBrokerQueue   *list.List
	queueLock         sync.Mutex
	internalDeathChan chan *Broker
	BrokerDeathChan   chan *common.UUID // Written to when a broker dies
	BrokerLiveChan    chan *common.UUID // Written to when a broker comes alive
}

func NewBrokerManager(heartbeatInterval int, brokerDeathChan, brokerLiveChan chan *common.UUID,
	clock common.Clock) *BrokerManager {
	bm := new(BrokerManager)
	bm.brokerMap = make(map[common.UUID]*Broker)
	bm.mapLock = sync.RWMutex{}
	bm.heartbeatInterval = heartbeatInterval
	bm.clock = clock
	bm.deadBrokerQueue = list.New()
	bm.liveBrokerQueue = list.New()
	bm.queueLock = sync.Mutex{}
	bm.internalDeathChan = make(chan *Broker, 10)
	bm.BrokerDeathChan = brokerDeathChan
	bm.BrokerLiveChan = brokerLiveChan
	go bm.monitorDeathChan()
	return bm
}

// If broker already exists in the mapping, this should be a reconnection
//   so simply update with the new connection and restart
// Otherwise, create a new Broker, put it into the map, and start it
func (bm *BrokerManager) ConnectBroker(brokerInfo *common.BrokerInfo, conn *net.TCPConn) (err error) {
	bm.mapLock.RLock()
	brokerConn, ok := bm.brokerMap[brokerInfo.BrokerID]
	bm.mapLock.RUnlock()
	if ok {
		brokerConn.WaitForCleanup()
		bm.queueLock.Lock()
		bm.deadBrokerQueue.Remove(brokerConn.listElem)
		bm.queueLock.Unlock()
	} else {
		messageHandler := bm.createMessageHandler(brokerInfo.BrokerID)
		brokerConn = NewBroker(brokerInfo, messageHandler,
			time.Duration(bm.heartbeatInterval)*time.Second, bm.clock, bm.internalDeathChan)
		bm.mapLock.Lock()
		bm.brokerMap[brokerInfo.BrokerID] = brokerConn
		bm.mapLock.Unlock()
	}
	brokerConn.StartAsynchronously(conn)
	bm.queueLock.Lock()
	brokerConn.listElem = bm.liveBrokerQueue.PushBack(brokerConn)
	bm.queueLock.Unlock()
	bm.BrokerLiveChan <- &brokerInfo.BrokerID
	return
}

func (bm *BrokerManager) IsBrokerAlive(brokerID common.UUID) bool {
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
func (bm *BrokerManager) GetLiveBroker() *common.BrokerInfo {
	bm.queueLock.Lock()
	defer bm.queueLock.Unlock()
	tempDeadList := list.New()
	// Store until the end so that we can push the dead ones back onto
	// the live list; we don't want to mess up the routines that will
	// be moving them later
	defer bm.liveBrokerQueue.PushBackList(tempDeadList)
	for {
		elem := bm.liveBrokerQueue.Front()
		if elem == nil {
			return nil
		}
		bconn := bm.liveBrokerQueue.Remove(elem).(*Broker)
		if bconn.IsAlive() { // Double check liveness to be sure
			bconn.listElem = bm.liveBrokerQueue.PushBack(bconn)
			return &bconn.BrokerInfo
		} else {
			tempDeadList.PushBack(bconn)
		}
	}
}

// Terminate the Broker and remove from our map
func (bm *BrokerManager) TerminateBroker(brokerID common.UUID) {
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
	bm.mapLock.Unlock()
}

// Asynchronously send a message to the given broker
func (bm *BrokerManager) SendToBroker(brokerID common.UUID, message common.Sendable) (err error) {
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

// Asynchronously send a message to all currently living  brokers
func (bm *BrokerManager) BroadcastToBrokers(message common.Sendable) {
	bm.mapLock.RLock()
	defer bm.mapLock.RUnlock()
	for _, brokerConn := range bm.brokerMap {
		brokerConn.Send(message)
	}
}

// Send a request for a heartbeat to broker; should be sent if contacted by a client
// saying that they couldn't contact the broker
func (bm *BrokerManager) RequestHeartbeat(brokerID common.UUID) (err error) {
	bm.mapLock.RLock()
	defer bm.mapLock.RUnlock()
	brokerConn, ok := bm.brokerMap[brokerID]
	if !ok {
		log.WithField("brokerID", brokerID).Warn("Attempted to send heartbeat request to nonexistent broker")
		return errors.New("Broker was unable to be found")
	}
	brokerConn.RequestHeartbeat()
	return
}

func (bm *BrokerManager) createMessageHandler(brokerID common.UUID) MessageHandler {
	return func(message common.Sendable) {
		switch msg := message.(type) {
		case *common.PublishMessage:
			// TODO forward on to the subsystem for this
		case *common.QueryMessage:
			// TODO forward on to the subsystem for this
		case *common.ClientTerminationMessage:
		// TODO communicate to the ClientTrackerService
		case *common.PublisherTerminationMessage:
		// TODO same as above
		case *common.BrokerConnectMessage:
			log.WithFields(log.Fields{
				"connBrokerID": brokerID, "newBrokerID": msg.BrokerID, "newBrokerAddr": msg.BrokerAddr,
			}).Warn("Received a BrokerConnectMessage over an existing broker connection")
		case *common.BrokerTerminateMessage:
			log.WithField("brokerID", brokerID).Info("BrokerManager terminating broker")
			bm.mapLock.RLock()
			brokerConn, ok := bm.brokerMap[brokerID]
			bm.mapLock.RUnlock()
			if !ok {
				log.WithField("brokerID", brokerID).Fatal("Received BrokerTerminateMessage but can't find the brokerID")
			}
			brokerConn.Send(&common.AcknowledgeMessage{MessageID: msg.MessageID})
			bm.mapLock.Lock()
			delete(bm.brokerMap, brokerConn.BrokerID)
			bm.mapLock.Unlock()
			brokerConn.Terminate()
		default:
			log.WithFields(log.Fields{
				"brokerID": brokerID, "message": msg, "messageType": common.GetMessageType(msg),
			}).Warn("Coordinator received an unexpected message!")
		}
	}
}

// Monitor Broker death, moving the dying broker to the correct
// queue. Also forward the broker on to the outward Death chan
func (bm *BrokerManager) monitorDeathChan() {
	for {
		deadBroker, ok := <-bm.internalDeathChan
		log.WithField("broker", deadBroker).Debug("Broker determined as dead!")
		if !ok {
			return
		}
		bm.queueLock.Lock()
		bm.liveBrokerQueue.Remove(deadBroker.listElem)
		deadBroker.listElem = bm.deadBrokerQueue.PushBack(deadBroker)
		bm.queueLock.Unlock()
		bm.BrokerDeathChan <- &deadBroker.BrokerID // Forward on to outward death chan
	}
}
