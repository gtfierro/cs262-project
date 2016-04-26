package main

import (
	"container/list"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"sync"
	"time"
)

type Broker struct {
	common.BrokerInfo
	outstandingMessages    map[common.MessageIDType]common.SendableWithID // messageID -> message
	outMessageLock         sync.RWMutex
	messageHandler         MessageHandler
	messageSendBuffer      chan common.Sendable
	requestHeartbeatBuffer chan chan bool
	heartbeatBuffer        chan bool
	alive                  bool
	aliveLock              sync.Mutex
	aliveCond              *sync.Cond
	heartbeatInterval      time.Duration
	terminating            chan bool // send on this channel when broker is shutting down permanently
	clock                  common.Clock
	deathChannel           chan *Broker
	listElem               *list.Element
	// TODO ETK combine the client and pub map?
	remoteClientMap    map[common.UUID]*common.UUID // ClientID -> HomeBrokerID
	remotePublisherMap map[common.UUID]*common.UUID // PublisherID -> HomeBrokerID
	localClientMap     map[common.UUID]struct{}     // ClientID -> HomeBrokerID
	localPublisherMap  map[common.UUID]struct{}     // PublisherID -> HomeBrokerID
	pubClientMapLock   sync.RWMutex
}

func NewBroker(broker *common.BrokerInfo, messageHandler MessageHandler,
	heartbeatInterval time.Duration, clock common.Clock, deathChannel chan *Broker) *Broker {
	bc := new(Broker)
	bc.BrokerInfo = *broker
	bc.terminating = make(chan bool)
	bc.messageSendBuffer = make(chan common.Sendable, 20) // TODO buffer size?
	bc.requestHeartbeatBuffer = make(chan chan bool, 20)
	bc.heartbeatBuffer = make(chan bool, 5)
	bc.outstandingMessages = make(map[common.MessageIDType]common.SendableWithID)
	bc.aliveLock = sync.Mutex{}
	bc.aliveCond = sync.NewCond(&bc.aliveLock)
	bc.outMessageLock = sync.RWMutex{}
	bc.messageHandler = messageHandler
	bc.heartbeatInterval = heartbeatInterval
	bc.clock = clock
	bc.deathChannel = deathChannel
	bc.remoteClientMap = make(map[common.UUID]*common.UUID)
	bc.remotePublisherMap = make(map[common.UUID]*common.UUID)
	bc.localClientMap = make(map[common.UUID]struct{})
	bc.localPublisherMap = make(map[common.UUID]struct{})
	return bc
}

func (bc *Broker) PushToList(l *list.List) {
	bc.listElem = l.PushBack(bc)
}

func (bc *Broker) RemoveFromList(l *list.List) {
	l.Remove(bc.listElem)
}

// Start up communication with this broker, asynchronously
// Calling this on an existing Broker before calling WaitForCleanup
// will result in undefined behavior
func (bc *Broker) StartAsynchronously(conn *net.TCPConn) {
	bc.aliveLock.Lock()
	bc.alive = true
	bc.terminating = make(chan bool)
	bc.aliveLock.Unlock()
	done := make(chan bool)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	receiveFinishChannel := make(chan bool)
	go bc.receiveLoop(conn, done, receiveFinishChannel)
	go bc.sendLoop(conn, done, wg)
	go bc.monitorHeartbeats(done, wg)
	go bc.cleanupWhenDone(conn, done, receiveFinishChannel, wg)
}

// Instruct this Broker to terminate itself
func (bc *Broker) Terminate() {
	close(bc.terminating)
}

// Wait until all of this Broker's threads and connections have been cleaned
func (bc *Broker) WaitForCleanup() {
	bc.aliveLock.Lock()
	for bc.alive == true {
		bc.aliveCond.Wait()
	}
	bc.aliveLock.Unlock()
}

// Asynchonously a message to this Broker
func (bc *Broker) Send(msg common.Sendable) {
	bc.messageSendBuffer <- msg
}

// Request a heartbeat from this broker and wait until it is determined whether
// or not the broker is alive (a heartbeat is received or it times out)
// Returns true if the broker is alive, or else false
func (bc *Broker) RequestHeartbeatAndWait() bool {
	bc.messageSendBuffer <- &common.RequestHeartbeatMessage{}
	respChan := make(chan bool)
	bc.requestHeartbeatBuffer <- respChan
	return <-respChan
}

// Returns true iff currently alive
func (bc *Broker) IsAlive() (alive bool) {
	bc.aliveLock.Lock()
	alive = bc.alive
	bc.aliveLock.Unlock()
	return
}

// Called LocalPublisher because the PublishMessage came from
// this broker, but it may actually be remote - if it is,
// and it's already in the remote map, no need to add it
func (bc *Broker) AddLocalPublisher(publisherID common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()

	if _, found := bc.remotePublisherMap[publisherID]; !found {
		bc.localPublisherMap[publisherID] = struct{}{}
	}
}

// Called LocalClient because the QueryMessage came from
// this broker, but it may actually be remote - if it is,
// and it's already in the remote map, no need to add it
func (bc *Broker) AddLocalClient(clientID common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()

	if _, found := bc.localClientMap[clientID]; !found {
		bc.localClientMap[clientID] = struct{}{}
	}
}

func (bc *Broker) RemovePublisher(publisherID common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()

	delete(bc.localPublisherMap, publisherID)
	delete(bc.remotePublisherMap, publisherID)
}

func (bc *Broker) RemoveClient(clientID common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()

	delete(bc.localClientMap, clientID)
	delete(bc.remoteClientMap, clientID)
}

func (bc *Broker) GetAndClearPublishers() map[common.UUID]struct{} {
	bc.pubClientMapLock.Lock()
	remoteMap := bc.remotePublisherMap
	localMap := bc.localPublisherMap
	bc.remotePublisherMap = make(map[common.UUID]*common.UUID)
	bc.localPublisherMap = make(map[common.UUID]struct{})
	bc.pubClientMapLock.Unlock()

	for publisherID, _ := range remoteMap {
		localMap[publisherID] = struct{}{}
	}
	return localMap
}

func (bc *Broker) GetAndClearClients() map[common.UUID]struct{} {
	bc.pubClientMapLock.Lock()
	remoteMap := bc.remoteClientMap
	localMap := bc.localClientMap
	bc.remoteClientMap = make(map[common.UUID]*common.UUID)
	bc.localClientMap = make(map[common.UUID]struct{})
	bc.pubClientMapLock.Unlock()

	for clientID, _ := range remoteMap {
		localMap[clientID] = struct{}{}
	}
	return localMap
}

func (bc *Broker) AddRemotePublisher(publisherID common.UUID, homeBrokerID *common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()

	bc.remotePublisherMap[publisherID] = homeBrokerID
}

func (bc *Broker) AddRemoteClient(clientID common.UUID, homeBrokerID *common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()

	bc.remoteClientMap[clientID] = homeBrokerID
}

// Send a termination message to the broker instructing it to
// terminate its connection with all clients and publishers whose
// home broker is homeBrokerID
func (bc *Broker) TerminateRemotePubClients(homeBrokerID *common.UUID) {
	bc.pubClientMapLock.Lock()
	defer bc.pubClientMapLock.Unlock()
	clientIDs := make([]common.UUID, 0, 50)
	for clientAddr, brokerID := range bc.remoteClientMap {
		if *brokerID == *homeBrokerID {
			if cap(clientIDs) == len(clientIDs) {
				clientIDsTmp := clientIDs
				clientIDs = make([]common.UUID, len(clientIDs), cap(clientIDs)*2)
				copy(clientIDs, clientIDsTmp)
			}
			clientIDs = append(clientIDs, clientAddr)
			delete(bc.remoteClientMap, clientAddr)
		}
	}
	publisherIDs := make([]common.UUID, 0, 50)
	for publisherID, brokerID := range bc.remotePublisherMap {
		if *brokerID == *homeBrokerID {
			if cap(publisherIDs) == len(publisherIDs) {
				publisherIDsTmp := publisherIDs
				publisherIDs = make([]common.UUID, len(publisherIDs), cap(publisherIDs)*2)
				copy(publisherIDs, publisherIDsTmp)
			}
			publisherIDs = append(publisherIDs, publisherID)
			delete(bc.remotePublisherMap, publisherID)
		}
	}
	if len(clientIDs) > 0 {
		bc.Send(&common.ClientTerminationRequest{
			MessageIDStruct: common.GetMessageIDStruct(),
			ClientIDs:       clientIDs,
		})
	}
	if len(publisherIDs) > 0 {
		bc.Send(&common.PublisherTerminationRequest{
			MessageIDStruct: common.GetMessageIDStruct(),
			PublisherIDs:    publisherIDs,
		})
	}
}

// Monitor heartbeats from the broker and determine death if necessary
func (bc *Broker) monitorHeartbeats(done chan bool, wg *sync.WaitGroup) {
	lastHeartbeat := time.Now()
	nextHeartbeat := lastHeartbeat.Add(bc.heartbeatInterval)
	heartbeatRequested := false
	outstandingRespChans := make(map[chan bool]struct{})
	hbRequest := common.RequestHeartbeatMessage{}
	defer wg.Done()
	for {
		log.WithFields(log.Fields{
			"broker": bc.BrokerID, "heartbeatInterval": bc.heartbeatInterval,
		}).Debug("About to sleep while waiting for a heartbeat from broker")
		select {
		case <-bc.clock.After(nextHeartbeat.Sub(lastHeartbeat)):
			elapsed := bc.clock.Now().Sub(lastHeartbeat)
			if elapsed >= 3*bc.heartbeatInterval { // determine that it's dead
				log.WithFields(log.Fields{
					"broker": bc.BrokerID, "heartbeatInterval": bc.heartbeatInterval, "secSinceResponse": elapsed.Seconds(),
				}).Warn("Broker has not responded; determining death")
				for respChan, _ := range outstandingRespChans {
					respChan <- false // Broker is dead
				}
				close(done)
			} else if elapsed >= 2*bc.heartbeatInterval && !heartbeatRequested {
				log.WithFields(log.Fields{
					"brokerID": bc.BrokerID, "heartbeatInterval": bc.heartbeatInterval, "secSinceResponse": elapsed.Seconds(),
				}).Warn("Broker has not responded; sending HeartbeatRequest")
				bc.messageSendBuffer <- &hbRequest
			} else if elapsed >= 1*bc.heartbeatInterval && heartbeatRequested {
				log.WithFields(log.Fields{
					"brokerID": bc.BrokerID, "heartbeatInterval": bc.heartbeatInterval, "secSinceResponse": elapsed.Seconds(),
				}).Warn("Broker has not responded after requesting a heartbeat; determining death")
				for respChan, _ := range outstandingRespChans {
					respChan <- false // Broker is dead
				}
				close(done)
			}
		case respChan := <-bc.requestHeartbeatBuffer:
			// TODO maybe ignore it and just return true on the channel if it's only been
			// e.g. 1/10 of an heartbeatInterval since the last heartbeat to avoid flooding?
			if !heartbeatRequested { // only do anything if we haven't already requested one
				heartbeatRequested = true
				nextHeartbeat = bc.clock.Now().Add(bc.heartbeatInterval)
			}
			outstandingRespChans[respChan] = struct{}{}
		case <-bc.heartbeatBuffer:
			lastHeartbeat = bc.clock.Now()
			nextHeartbeat = lastHeartbeat.Add(bc.heartbeatInterval)
			heartbeatRequested = false
			for respChan, _ := range outstandingRespChans {
				respChan <- true // Broker is still alive
			}
		case <-done:
			return
		case <-bc.terminating:
			return
		}
	}
}

// Wait for a signal on done or bc.terminating, and then close conn
func (bc *Broker) cleanupWhenDone(conn *net.TCPConn, done, rcvLoopDone chan bool, wg *sync.WaitGroup) {
	// Simply wait for a done or terminating signal
	select {
	case <-done:
	case <-bc.terminating:
	}
	log.WithField("broker", bc.BrokerID).Debug("Broker cleanup; waiting on the workGroup")
	wg.Wait()    // Wait for heartbeat monitoring and sendLoop
	conn.Close() // Close connection to force receiveLoop to finish
	log.WithField("broker", bc.BrokerID).Debug("Broker cleanup; waiting on the rcvLoop")
	<-rcvLoopDone
	bc.deathChannel <- bc
	bc.aliveLock.Lock()
	bc.alive = false
	bc.aliveCond.Broadcast()
	bc.aliveLock.Unlock()
}

// Watch for incoming messages from the broker. Internally handles acknowledgments and heartbeats;
// other messages are provided to bc.messageHandler.
// Watches for messages until bc.alive is false or an EOF is reached
func (bc *Broker) receiveLoop(conn *net.TCPConn, done chan bool, loopFinished chan bool) {
	reader := msgp.NewReader(conn)
	log.WithField("brokerID", bc.BrokerID).Info("Beginning to receive messages from broker")
	defer func() { loopFinished <- true }()
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		select {
		case <-done:
			log.WithField("brokerID", bc.BrokerID).Warn("No longer receiving messages from broker (death)")
			return
		case <-bc.terminating:
			log.WithField("brokerID", bc.BrokerID).Info("No longer receiving messages from broker (termination)")
			return
		default: // fall through
		}
		if err == nil {
			log.WithFields(log.Fields{
				"brokerID": bc.BrokerID, "message": msg, "messageType": common.GetMessageType(msg),
			}).Debug("Received message from broker")
			switch m := msg.(type) {
			case *common.AcknowledgeMessage:
				bc.handleAck(m)
			case *common.HeartbeatMessage:
				bc.heartbeatBuffer <- true
			default:
				go bc.messageHandler(&MessageFromBroker{msg, bc})
			}
		} else {
			if err == io.EOF {
				close(done)
				log.WithField("brokerID", bc.BrokerID).Warn("No longer receiving messages from broker; EOF reached")
				return // connection closed
			}
			log.WithFields(log.Fields{
				"brokerID": bc.BrokerID, "error": err,
			}).Warn("Error while reading message from broker")
			bc.clock.Sleep(time.Second) // wait before retry
		}
	}
}

func (bc *Broker) sendLoop(conn *net.TCPConn, done chan bool, wg *sync.WaitGroup) {
	var err error
	writer := msgp.NewWriter(conn)
	log.WithField("brokerID", bc.BrokerID).Info("Beginning the send loop to broker")
	defer wg.Done()
	for {
		select {
		case msg := <-bc.messageSendBuffer:
			switch m := msg.(type) {
			case common.SendableWithID:
				err = bc.sendEnsureDelivery(writer, m, done)
			case common.Sendable:
				log.WithFields(log.Fields{
					"brokerID": bc.BrokerID, "message": m, "messageType": common.GetMessageType(m),
				}).Debug("Sending message to broker...")
				err = bc.sendInternal(writer, m)
			}
			if err != nil {
				if err == io.EOF {
					bc.messageSendBuffer <- msg
					close(done)
					log.WithField("brokerID", bc.BrokerID).Warn("No longer sending to broker; EOF reached")
				} else {
					bc.messageSendBuffer <- msg
					log.WithFields(log.Fields{
						"brokerID": bc.BrokerID, "error": err, "message": msg, "messageType": common.GetMessageType(msg),
					}).Error("Error sending message to broker")
				}
			}
		case <-done:
			log.WithField("brokerID", bc.BrokerID).Warn("Exiting the send loop to broker (death)")
			return
		case <-bc.terminating:
			log.WithField("brokerID", bc.BrokerID).Info("Exiting the send loop to broker (termination)")
			return
		}
	}
}

func (bc *Broker) sendEnsureDelivery(writer *msgp.Writer, message common.SendableWithID,
	done chan bool) (err error) {
	log.WithFields(log.Fields{
		"brokerID": bc.BrokerID, "message": message, "messageType": common.GetMessageType(message),
	}).Debug("Sending message to broker with ensured delivery...")
	err = bc.sendInternal(writer, message)
	if err != nil {
		return err
	}
	bc.outMessageLock.Lock()
	bc.outstandingMessages[message.GetID()] = message
	bc.outMessageLock.Unlock()
	go func() {
		select {
		case <-bc.terminating:
			return
		case <-done:
		case <-bc.clock.After(bc.heartbeatInterval):
		}
		bc.outMessageLock.RLock()
		msg, ok := bc.outstandingMessages[message.GetID()]
		bc.outMessageLock.RUnlock()
		if ok {
			log.WithFields(log.Fields{
				"brokerID": bc.BrokerID, "message": msg, "messageType": common.GetMessageType(msg),
			}).Info("Send with ensured delivery has not been acked; placing back onto send buf")
			bc.messageSendBuffer <- msg // put back onto the queue to be sent again
		}
	}()
	return
}

func (bc *Broker) sendInternal(writer *msgp.Writer, message common.Sendable) (err error) {
	err = message.Encode(writer)
	writer.Flush()
	return
}

// Handles an acknowledgment by removing the MessageID from the outstanding
// messages map
func (bc *Broker) handleAck(ack *common.AcknowledgeMessage) {
	log.WithFields(log.Fields{
		"brokerID": bc.BrokerID, "ackMessage": ack,
	}).Debug("Received ack from broker for message")
	bc.outMessageLock.Lock()
	defer bc.outMessageLock.Unlock()
	delete(bc.outstandingMessages, ack.MessageID)
}
