package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"sync"
	"time"
)

type BrokerConnection struct {
	broker                 *Broker
	outstandingMessages    map[common.MessageIDType]common.SendableWithID // messageID -> message
	outMessageLock         sync.RWMutex
	messageHandler         MessageHandler
	messageSendBuffer      chan common.Sendable
	requestHeartbeatBuffer chan bool
	heartbeatBuffer        chan bool
	alive                  bool
	aliveLock              sync.Mutex
	aliveCond              *sync.Cond
	heartbeatInterval      time.Duration
	terminating            chan bool // send on this channel when broker is shutting down
	clock                  common.Clock
}

func NewBrokerConnection(broker *Broker, messageHandler MessageHandler, heartbeatInterval time.Duration,
	clock common.Clock) *BrokerConnection {
	bc := new(BrokerConnection)
	bc.broker = broker
	bc.terminating = make(chan bool)
	bc.messageSendBuffer = make(chan common.Sendable, 20) // TODO buffer size?
	bc.requestHeartbeatBuffer = make(chan bool, 2)
	bc.heartbeatBuffer = make(chan bool, 5)
	bc.outstandingMessages = make(map[common.MessageIDType]common.SendableWithID)
	bc.aliveLock = sync.Mutex{}
	bc.aliveCond = sync.NewCond(&bc.aliveLock)
	bc.outMessageLock = sync.RWMutex{}
	bc.messageHandler = messageHandler
	bc.heartbeatInterval = heartbeatInterval
	bc.clock = clock
	return bc
}

// Start up communication with this broker, asynchronously
// Calling this on an existing BrokerConnection before calling WaitForCleanup
// will result in undefined behavior
func (bc *BrokerConnection) StartAsynchronously(conn *net.TCPConn) {
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

// Instruct this BrokerConnection to terminate itself
func (bc *BrokerConnection) Terminate() {
	close(bc.terminating)
}

// Wait until all of this BrokerConnection's threads and connections have been cleaned
func (bc *BrokerConnection) WaitForCleanup() {
	bc.aliveLock.Lock()
	for bc.alive == true {
		bc.aliveCond.Wait()
	}
	bc.aliveLock.Unlock()
}

// Asynchonously a message to this Broker
func (bc *BrokerConnection) Send(msg common.Sendable) {
	bc.messageSendBuffer <- msg
}

// Request a heartbeat from this broker
func (bc *BrokerConnection) RequestHeartbeat() {
	bc.requestHeartbeatBuffer <- true
}

// Monitor heartbeats from the broker and determine death if necessary
func (bc *BrokerConnection) monitorHeartbeats(done chan bool, wg *sync.WaitGroup) {
	lastHeartbeat := time.Now()
	heartbeatRequested := false
	hbRequest := common.RequestHeartbeatMessage{}
	defer wg.Done()
	for {
		select {
		case <-bc.clock.After(bc.heartbeatInterval):
			elapsed := bc.clock.Now().Sub(lastHeartbeat)
			if elapsed >= 3*bc.heartbeatInterval { // determine that it's dead
				log.WithFields(log.Fields{
					"broker": *bc.broker, "heartbeatInterval": bc.heartbeatInterval, "secSinceResponse": elapsed.Seconds(),
				}).Warn("Broker has not responded; determining death")
				done <- true
			} else if elapsed >= 2*bc.heartbeatInterval && !heartbeatRequested {
				log.WithFields(log.Fields{
					"broker": *bc.broker, "heartbeatInterval": bc.heartbeatInterval, "secSinceResponse": elapsed.Seconds(),
				}).Warn("Broker has not responded; sending HeartbeatRequest")
				bc.messageSendBuffer <- &hbRequest
			} else if elapsed >= 1*bc.heartbeatInterval && heartbeatRequested {
				log.WithFields(log.Fields{
					"broker": *bc.broker, "heartbeatInterval": bc.heartbeatInterval, "secSinceResponse": elapsed.Seconds(),
				}).Warn("Broker has not responded after requesting a heartbeat; determining death")
				done <- true
			}
		case <-bc.requestHeartbeatBuffer:
			heartbeatRequested = true
		case <-bc.heartbeatBuffer:
			lastHeartbeat = bc.clock.Now()
		case <-done:
			return
		case <-bc.terminating:
			return
		}
	}
}

// Wait for a signal on done or bc.terminating, and then close conn
func (bc *BrokerConnection) cleanupWhenDone(conn *net.TCPConn, done, rcvLoopDone chan bool, wg *sync.WaitGroup) {
	// Simply wait for a done or terminating signal
	select {
	case <-done:
	case <-bc.terminating:
	}
	wg.Wait()    // Wait for heartbeat monitoring and sendLoop
	conn.Close() // Close connection to force receiveLoop to finish
	<-rcvLoopDone
	bc.aliveLock.Lock()
	bc.alive = false
	bc.aliveCond.Broadcast()
	bc.aliveLock.Unlock()
}

// Watch for incoming messages from the broker. Internally handles acknowledgments and heartbeats;
// other messages are provided to bc.messageHandler.
// Watches for messages until bc.alive is false or an EOF is reached
func (bc *BrokerConnection) receiveLoop(conn *net.TCPConn, done chan bool, loopFinished chan bool) {
	reader := msgp.NewReader(conn)
	log.WithField("broker", *bc.broker).Info("Beginning to receive messages from broker")
	defer func() { loopFinished <- true }()
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		select {
		case <-done:
			log.WithField("broker", *bc.broker).Warn("No longer receiving messages from broker (death)")
			return
		case <-bc.terminating:
			log.WithField("broker", *bc.broker).Info("No longer receiving messages from broker (termination)")
			return
		default: // fall through
		}
		if err == nil {
			log.WithFields(log.Fields{
				"broker": bc.broker, "message": msg, "messageType": common.GetMessageType(msg),
			}).Debug("Received message from broker")
			switch m := msg.(type) {
			case *common.AcknowledgeMessage:
				bc.handleAck(m)
			case *common.HeartbeatMessage:
				bc.heartbeatBuffer <- true
			default:
				go bc.messageHandler(conn, msg)
			}
		} else {
			if err == io.EOF {
				close(done)
				log.WithField("broker", *bc.broker).Warn("No longer receiving messages from broker; EOF reached")
				return // connection closed
			}
			log.WithFields(log.Fields{
				"broker": *bc.broker, "error": err,
			}).Warn("Error while reading message from broker")
			bc.clock.Sleep(time.Second) // wait before retry
		}
	}
}

func (bc *BrokerConnection) sendLoop(conn *net.TCPConn, done chan bool, wg *sync.WaitGroup) {
	var err error
	writer := msgp.NewWriter(conn)
	log.WithField("broker", *bc.broker).Info("Beginning the send loop to broker")
	defer wg.Done()
	for {
		select {
		case msg := <-bc.messageSendBuffer:
			switch m := msg.(type) {
			case common.SendableWithID:
				err = bc.sendEnsureDelivery(writer, m, done)
			case common.Sendable:
				log.WithFields(log.Fields{
					"broker": bc.broker, "message": m, "messageType": common.GetMessageType(m),
				}).Debug("Sending message to broker...")
				err = bc.sendInternal(writer, m)
			}
			if err != nil {
				if err == io.EOF {
					bc.messageSendBuffer <- msg
					close(done)
					log.WithField("broker", bc.broker).Warn("No longer sending to broker; EOF reached")
				} else {
					bc.messageSendBuffer <- msg
					log.WithFields(log.Fields{
						"broker": bc.broker, "error": err, "message": msg, "messageType": common.GetMessageType(msg),
					}).Error("Error sending message to broker")
				}
			}
		case <-done:
			log.WithField("broker", *bc.broker).Warn("Exiting the send loop to broker (death)")
			return
		case <-bc.terminating:
			log.WithField("broker", *bc.broker).Info("Exiting the send loop to broker (termination)")
			return
		}
	}
}

func (bc *BrokerConnection) sendEnsureDelivery(writer *msgp.Writer, message common.SendableWithID,
	done chan bool) (err error) {
	log.WithFields(log.Fields{
		"broker": bc.broker, "message": message, "messageType": common.GetMessageType(message),
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
				"broker": bc.broker, "message": msg, "messageType": common.GetMessageType(msg),
			}).Info("Send with ensured delivery has not been acked; placing back onto send buf")
			bc.messageSendBuffer <- msg // put back onto the queue to be sent again
		}
	}()
	return
}

func (bc *BrokerConnection) sendInternal(writer *msgp.Writer, message common.Sendable) (err error) {
	err = message.Encode(writer)
	writer.Flush()
	return
}

// Handles an acknowledgment by removing the MessageID from the outstanding
// messages map
func (bc *BrokerConnection) handleAck(ack *common.AcknowledgeMessage) {
	log.WithFields(log.Fields{
		"broker": bc.broker, "ackMessage": ack,
	}).Debug("Received ack from broker for message")
	bc.outMessageLock.Lock()
	defer bc.outMessageLock.Unlock()
	delete(bc.outstandingMessages, ack.MessageID)
}
