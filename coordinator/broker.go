package main

import (
	"container/list"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"sync"
	"time"
)

type Broker struct {
	common.BrokerInfo
	alive        bool
	doneChan     chan bool
	waitGroup    *sync.WaitGroup
	rcvWaitGroup *sync.WaitGroup
	aliveLock    sync.Mutex
	aliveCond    *sync.Cond

	outstandingMessages    map[common.MessageIDType]common.SendableWithID // messageID -> message
	outMessageLock         sync.RWMutex
	messageHandler         MessageHandler
	messageSendBuffer      chan common.Sendable
	requestHeartbeatBuffer chan chan bool
	heartbeatBuffer        chan bool
	heartbeatInterval      time.Duration
	terminating            chan bool // send on this channel when broker is shutting down permanently
	clock                  common.Clock
	deathChannel           chan *Broker
	listElem               *list.Element
	pubClientMapLock       sync.RWMutex
}

func NewBroker(broker *common.BrokerInfo, messageHandler MessageHandler,
	heartbeatInterval time.Duration, clock common.Clock, deathChannel chan *Broker) *Broker {
	bc := new(Broker)
	bc.BrokerInfo = *broker
	bc.terminating = make(chan bool)
	bc.messageSendBuffer = make(chan common.Sendable, 200) // TODO buffer size?
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
func (bc *Broker) StartAsynchronously(commConn CommConn) {
	var wasAlive bool
	bc.aliveLock.Lock()
	if !bc.alive {
		wasAlive = false
		bc.doneChan = make(chan bool, 1)
		bc.waitGroup = new(sync.WaitGroup)
		bc.rcvWaitGroup = new(sync.WaitGroup)
	} else {
		// transient connection issue; didn't determine death of broker
		wasAlive = true
	}
	done, wg, receiveWg := bc.doneChan, bc.waitGroup, bc.rcvWaitGroup
	bc.alive = true
	bc.terminating = make(chan bool)
	bc.aliveLock.Unlock()
	receiveWg.Add(1)
	go bc.receiveLoop(commConn, done, receiveWg)
	wg.Add(1)
	go bc.sendLoop(commConn, done, wg)
	if !wasAlive {
		wg.Add(1)
		go bc.monitorHeartbeats(done, wg)
		go bc.cleanupWhenDone(done, receiveWg, wg)
	}
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
func (bc *Broker) IsAlive() bool {
	bc.aliveLock.Lock()
	defer bc.aliveLock.Unlock()
	return bc.alive
}

// Monitor heartbeats from the broker and determine death if necessary
func (bc *Broker) monitorHeartbeats(done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	lastHeartbeat := time.Now()
	nextHeartbeat := lastHeartbeat.Add(bc.heartbeatInterval)
	heartbeatRequested := false
	outstandingRespChans := make(map[chan bool]struct{})
	hbRequest := common.RequestHeartbeatMessage{}
	for {
		log.WithFields(log.Fields{
			"broker": bc.BrokerID, "heartbeatInterval": bc.heartbeatInterval,
		}).Debug("About to sleep while waiting for a heartbeat from broker")
		select {
		case <-done:
			return
		case <-bc.terminating:
			return
		default: // continue on
		}
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
				closeIfNotClosed(done)
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
				closeIfNotClosed(done)
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
func (bc *Broker) cleanupWhenDone(done chan bool, receiveWg, wg *sync.WaitGroup) {
	// Simply wait for a done or terminating signal
	select {
	case <-done:
	case <-bc.terminating:
	}
	log.WithField("broker", bc.BrokerID).Debug("Broker cleanup; waiting on the workGroup")
	wg.Wait() // Wait for heartbeat monitoring and sendLoop
	//commConn.Close() // Close connection to force receiveLoop to finish
	log.WithField("broker", bc.BrokerID).Debug("Broker cleanup; waiting on the rcvLoop")
	receiveWg.Wait()
	bc.deathChannel <- bc
	bc.aliveLock.Lock()
	bc.alive = false
	bc.aliveCond.Broadcast()
	bc.aliveLock.Unlock()
}

// Watch for incoming messages from the broker. Internally handles acknowledgments and heartbeats;
// other messages are provided to bc.messageHandler.
// Watches for messages until bc.alive is false or an EOF is reached
func (bc *Broker) receiveLoop(commConn CommConn, done chan bool, wg *sync.WaitGroup) {
	log.WithField("brokerID", bc.BrokerID).Info("Beginning to receive messages from broker")
	defer wg.Done()
	for {
		msg, err := commConn.ReceiveMessage()
		select {
		case <-done:
			commConn.Close()
			log.WithField("brokerID", bc.BrokerID).Warn("No longer receiving messages from broker (death)")
			return
		case <-bc.terminating:
			log.WithField("brokerID", bc.BrokerID).Info("No longer receiving messages from broker (termination)")
			return
		default: // fall through
		}
		if err == nil {
			switch m := msg.(type) {
			case *common.AcknowledgeMessage:
				log.WithField("brokerID", bc.BrokerID).WithField("ID", m.MessageID).Debug("Received ACK from Broker for messageID")
				bc.handleAck(m)
			case *common.HeartbeatMessage:
				log.WithField("brokerID", bc.BrokerID).Debug("Received heartbeat from Broker")
				bc.heartbeatBuffer <- true
			default:
				log.WithFields(log.Fields{
					"brokerID": bc.BrokerID, "message": msg, "messageType": common.GetMessageType(msg),
				}).Info("Received message from broker")
				go bc.messageHandler(&MessageFromBroker{msg, bc})
			}
		} else {
			//closeIfNotClosed(done)
			commConn.Close()
			log.WithFields(log.Fields{
				"brokerID": bc.BrokerID, "error": err,
			}).Warn("Temporarily no longer receiving messages from broker; error encountered")
			return // connection closed
		}
	}
}

func (bc *Broker) sendLoop(commConn CommConn, done chan bool, wg *sync.WaitGroup) {
	var err error
	log.WithField("brokerID", bc.BrokerID).Info("Beginning the send loop to broker")
	defer wg.Done()
	for {
		select {
		case msg := <-bc.messageSendBuffer:
			switch m := msg.(type) {
			case common.SendableWithID:
				err = bc.sendEnsureDelivery(commConn, m, done)
			case common.Sendable:
				log.WithFields(log.Fields{
					"brokerID": bc.BrokerID, "message": m, "messageType": common.GetMessageType(m),
				}).Info("Sending message to broker...")
				err = commConn.Send(m)
			}
			if err != nil {
				bc.messageSendBuffer <- msg
				//closeIfNotClosed(done)
				commConn.Close()
				log.WithFields(log.Fields{
					"brokerID": bc.BrokerID, "error": err,
				}).Warn("Temporarily not sending to broker; error encountered")
				return
			}
		case <-done:
			commConn.Close()
			log.WithField("brokerID", bc.BrokerID).Warn("Exiting the send loop to broker (death)")
			for _, pendingMsg := range commConn.GetPendingMessages() {
				bc.messageSendBuffer <- pendingMsg // Store for later
			}
			return
		case <-bc.terminating:
			log.WithField("brokerID", bc.BrokerID).Info("Exiting the send loop to broker (termination)")
			return
		}
	}
}

func (bc *Broker) sendEnsureDelivery(commConn CommConn, message common.SendableWithID,
	done chan bool) (err error) {
	log.WithFields(log.Fields{
		"brokerID": bc.BrokerID, "message": message, "messageType": common.GetMessageType(message),
	}).Info("Sending message to broker with ensured delivery...")
	err = commConn.Send(message)
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

func closeIfNotClosed(channel chan bool) {
	select {
	case _, ok := <-channel:
		if ok {
			close(channel)
		}
	default:
		close(channel)
	}
}
