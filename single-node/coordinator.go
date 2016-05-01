package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"sync"
	"time"
)

type Coordinator struct {
	address            *net.TCPAddr
	conn               *net.TCPConn
	clientside_address string
	coordside_address  string
	sendL              sync.Mutex
	broker             *RemoteBroker
	brokerID           common.UUID
	encoder            *msgp.Writer

	// the number of seconds to wait before retrying to
	// contact coordinator server
	retryTime int
	// the maximum interval to increase to between attempts to contact
	// the coordinator server
	retryTimeMax int
	// handles outstanding messages that need an ACK
	requests *outstandingManager
	queue    *sendableQueue
}

func ConnectCoordinator(config common.ServerConfig, s *Server) *Coordinator {
	var err error

	c := &Coordinator{broker: s.broker.(*RemoteBroker),
		retryTime:          1,
		retryTimeMax:       60,
		requests:           newOutstandingManager(),
		queue:              newSendableQueue(),
		clientside_address: fmt.Sprintf("%s:%d", config.Host, config.Port),
		brokerID:           config.BrokerID,
	}

	coordinatorAddress := fmt.Sprintf("%s:%d", config.CoordinatorHost, config.CoordinatorPort)
	c.address, err = net.ResolveTCPAddr("tcp", coordinatorAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"addr": c.address, "error": err,
		}).Fatal("Could not resolve the generated TCP address")
	}
	// Dial a connection to the Coordinator server
	c.rebuildConnection()
	go c.startBeating()

	return c
}

func (c *Coordinator) rebuildConnection() {
	var err error
	c.sendL.Lock()

	c.conn, err = net.DialTCP("tcp", nil, c.address)
	for err != nil {
		log.WithFields(log.Fields{
			"err": err, "server": c.address, "retry": c.retryTime,
		}).Error("Failed to contact coordinator server. Retrying")
		time.Sleep(time.Duration(c.retryTime) * time.Second)
		// increase retry window by factor of 2
		if c.retryTime*2 < c.retryTimeMax {
			c.retryTime *= 2
		} else {
			c.retryTime = c.retryTimeMax
		}
		// Dial a connection to the Coordinator server
		c.conn, err = net.DialTCP("tcp", nil, c.address)
	}
	c.sendL.Unlock()
	// if we were successful, reset the wait timer
	c.retryTime = 1
	c.encoder = msgp.NewWriter(c.conn)
	// save the address we're using to connect
	c.coordside_address = c.conn.LocalAddr().String()

	// when we come online, send the BrokerConnectMessage to inform the coordinator
	// server where it should send clients
	// TODO should do something else for the BrokerID since we want it to persist after restarts
	bcm := &common.BrokerConnectMessage{BrokerInfo: common.BrokerInfo{
		BrokerID:         c.brokerID,
		ClientBrokerAddr: c.clientside_address,
		CoordBrokerAddr:  c.coordside_address,
	}, MessageIDStruct: common.GetMessageIDStruct()}
	bcm.Encode(c.encoder)
	// do the actual sending
	c.encoder.Flush()

	// send a heartbeat as well
	c.sendHeartbeat()
	// before we send, we want to setup the ping/pong service
	go c.handleStateMachine()

	//TODO: does this actually work as expected?
	// do replay
	if c.queue.hasReplayValue() {
		replay := c.queue.startReplay()
		for msg := range replay {
			c.send(msg)
		}
	}
}

// This method handles the bookkeeping messages from the coordinator server
func (c *Coordinator) handleStateMachine() {
	reader := msgp.NewReader(c.conn)
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		// when the connection with the coordinator breaks, buffer
		// all outgoing messages
		if err == io.EOF {
			c.rebuildConnection()
			log.Warn("Coordinator is no longer reachable!")
			break
		}
		if err != nil {
			log.WithFields(log.Fields{
				"brokerID": c.brokerID, "message": msg, "error": err, "coordinator": c.address,
			}).Warn("Could not decode incoming message from coordinator")
		}
		// handle incoming message types
		switch m := msg.(type) {
		case *common.RequestHeartbeatMessage:
			log.Info("Received heartbeat from coordinator")
			c.sendHeartbeat()
		case *common.ForwardRequestMessage:
			log.Infof("Received forward request message %v", m)
			c.broker.AddForwardingEntries(m)
			c.ack(m.GetID())
		case *common.BrokerSubscriptionDiffMessage:
			log.Infof("Subscription Diff message %v", m)
			c.broker.ForwardSubscriptionDiffs(m)
			c.broker.UpdateForwardingEntries(m)
			c.ack(m.GetID())
		case *common.CancelForwardRequest:
			c.broker.RemoveForwardingEntry(m)
			c.ack(m.GetID())
		case *common.AcknowledgeMessage:
			c.requests.GotMessage(m)
		case *common.BrokerDeathMessage:
			log.Infof("Got Broker Death Message %v", m)
			//TODO: handle broker death
			c.ack(m.GetID())
		case *common.ClientTerminationRequest:
			c.broker.killClients(m)
			c.ack(m.GetID())
		case *common.PublisherTerminationRequest:
			c.broker.killPublishers(m)
			c.ack(m.GetID())
		case common.Message:
			log.Infof("Got a %T message %v", m, m)
			c.requests.GotMessage(m)
		default:
			log.WithFields(log.Fields{
				"message": m, "coordinator": c.address,
			}).Warn("I don't know what to do with this")
		}
	}
}

func (c *Coordinator) send(m common.Sendable) {
	c.sendL.Lock()
	defer c.sendL.Unlock()
	if err := m.Encode(c.encoder); err != nil {
		log.WithFields(log.Fields{
			"error": err, "coordinator": c.address, "message": m,
		}).Error("Could not encode message to coordinator")
		return
	}
	if err := c.encoder.Flush(); err != nil {
		log.WithFields(log.Fields{
			"error": err, "coordinator": c.address, "message": m,
		}).Error("Could not send message to coordinator")
		// buffer!
		c.queue.append(m)
	}
}

func (c *Coordinator) ack(id common.MessageIDType) {
	c.send(&common.AcknowledgeMessage{MessageID: id})
}

func (c *Coordinator) sendHeartbeat() {
	log.WithFields(log.Fields{
		"coordinator": c.address,
	}).Debug("Sending hearbeat")
	hb := &common.HeartbeatMessage{}
	c.send(hb)
}

func (c *Coordinator) startBeating() {
	tick := time.NewTicker(5 * time.Second)
	for range tick.C {
		c.sendHeartbeat()
	}
}

// if we receive a subscription and we are *not* using local evaluation,
// then we wrap it up in a BrokerQueryMessage and forward it to the coordinator
//   type BrokerQueryMessage struct {
//   	QueryMessage string
//   	ClientAddr   string
//   }
func (c *Coordinator) forwardSubscription(query string, clientID common.UUID, client net.Conn) {
	bqm := &common.BrokerQueryMessage{
		Query: query,
		UUID:  clientID,
	}
	bqm.MessageID = common.GetMessageID()
	c.send(bqm)
	response, _ := c.requests.WaitForMessage(bqm.GetID())
	switch m := response.(type) {
	case *common.AcknowledgeMessage:
		log.Debugf("Response %v", m)
	case *common.CancelForwardRequest:
		log.Debugf("Got cancel forward request %v", m)
	}
}

// this forwards a publish message from a local producer to the coordinator and receives
// a BrokerSubscriptionDiffMessage in response
func (c *Coordinator) forwardPublish(msg *common.PublishMessage) {
	var bpm = new(common.BrokerPublishMessage)
	bpm.FromPublishMessage(msg)
	c.send(bpm)
	log.Debug("Waiting for publish response")
	response, _ := c.requests.WaitForMessage(bpm.GetID())
	log.Debugf("Got response for pub %v", response)
	if _, ok := response.(*common.AcknowledgeMessage); !ok {
		log.WithFields(log.Fields{
			"response": response,
		}).Error("Did not get a proper acknowledgement")
	}
}

func (c *Coordinator) terminateClient(client *Client) {
	var ctm = &common.ClientTerminationMessage{
		ClientID: client.ID,
	}
	ctm.MessageID = common.GetMessageID()
	c.send(ctm)
	response, _ := c.requests.WaitForMessage(ctm.GetID())
	switch response.(type) {
	case *common.AcknowledgeMessage:
	default:
		log.Error("Did not get an ACK!")
	}
}

func (c *Coordinator) terminatePublisher(prod *Producer) {
	var ptm = &common.PublisherTerminationMessage{
		PublisherID: prod.ID,
	}
	ptm.MessageID = common.GetMessageID()
	c.send(ptm)
	response, _ := c.requests.WaitForMessage(ptm.GetID())
	switch response.(type) {
	case *common.AcknowledgeMessage:
	default:
		log.Error("Did not get an ACK!")
	}
}
