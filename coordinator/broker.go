package main

import (
	"container/list"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"io"
	"sync"
	"time"
)

type Broker struct {
	common.BrokerInfo
	alive                  bool
	aliveLock              sync.Mutex
	aliveCond              *sync.Cond
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
	bc.aliveLock.Lock()
	bc.alive = true
	bc.terminating = make(chan bool)
	bc.aliveLock.Unlock()
	done := make(chan bool)
	wg := new(sync.WaitGroup)
	receiveFinishChannel := make(chan bool)
	go bc.receiveLoop(commConn, done, receiveFinishChannel)
	go bc.sendLoop(commConn, done, wg)
	go bc.monitorHeartbeats(done, wg)
	go bc.cleanupWhenDone(commConn, done, receiveFinishChannel, wg)
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
	wg.Add(1)
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
func (bc *Broker) cleanupWhenDone(commConn CommConn, done, rcvLoopDone chan bool, wg *sync.WaitGroup) {
	// Simply wait for a done or terminating signal
	select {
	case <-done:
	case <-bc.terminating:
	}
	log.WithField("broker", bc.BrokerID).Debug("Broker cleanup; waiting on the workGroup")
	wg.Wait()        // Wait for heartbeat monitoring and sendLoop
	commConn.Close() // Close connection to force receiveLoop to finish
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
func (bc *Broker) receiveLoop(commConn CommConn, done chan bool, loopFinished chan bool) {
	log.WithField("brokerID", bc.BrokerID).Info("Beginning to receive messages from broker")
	defer func() { loopFinished <- true }()
	for {
		msg, err := commConn.ReceiveMessage()
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

func (bc *Broker) sendLoop(commConn CommConn, done chan bool, wg *sync.WaitGroup) {
	var err error
	wg.Add(1)
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
				}).Debug("Sending message to broker...")
				err = commConn.Send(m)
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

func (bc *Broker) sendEnsureDelivery(commConn CommConn, message common.SendableWithID,
	done chan bool) (err error) {
	log.WithFields(log.Fields{
		"brokerID": bc.BrokerID, "message": message, "messageType": common.GetMessageType(message),
	}).Debug("Sending message to broker with ensured delivery...")
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
