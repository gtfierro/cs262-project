package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"time"
)

type Coordinator struct {
	address  *net.TCPAddr
	conn     *net.TCPConn
	broker   Broker
	brokerID common.UUID
	encoder  *msgp.Writer
	// the number of seconds to wait before retrying to
	// contact coordinator server
	retryTime int
	// the maximum interval to increase to between attempts to contact
	// the coordinator server
	retryTimeMax int
	requests     *outstandingManager
}

func ConnectCoordinator(config common.ServerConfig, s *Server) *Coordinator {
	var err error

	c := &Coordinator{broker: s.broker,
		retryTime:    1,
		retryTimeMax: 60,
		requests:     newOutstandingManager(),
	}

	coordinatorAddress := fmt.Sprintf("%s:%d", config.CoordinatorHost, config.CoordinatorPort)
	c.address, err = net.ResolveTCPAddr("tcp", coordinatorAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"addr": c.address, "error": err,
		}).Fatal("Could not resolve the generated TCP address")
	}
	// Dial a connection to the Coordinator server
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
	// if we were successful, reset the wait timer
	c.retryTime = 1
	c.encoder = msgp.NewWriter(c.conn)

	// when we come online, send the BrokerConnectMessage to inform the coordinator
	// server where it should send clients
	// TODO should do something else for the BrokerID since we want it to persist after restarts
	bcm := &common.BrokerConnectMessage{BrokerInfo: common.BrokerInfo{
		BrokerID:   c.brokerID,
		BrokerAddr: c.address.String(),
	}, MessageIDStruct: common.GetMessageIDStruct()}
	bcm.Encode(c.encoder)
	// do the actual sending
	c.encoder.Flush()
	// send a heartbeat as well
	c.sendHeartbeat()
	// before we send, we want to setup the ping/pong service
	go c.handleStateMachine()
	go c.startBeating()

	return c
}

// This method handles the bookkeeping messages from the coordinator server
func (c *Coordinator) handleStateMachine() {
	reader := msgp.NewReader(c.conn)
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		//TODO: when the connection with the coordinator breaks,
		//WHAT DO WE DO?!
		if err == io.EOF {
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
		case common.Message:
			log.Infof("Got a message %v", m)
			c.requests.GotMessage(m)
		default:
			log.WithFields(log.Fields{
				"message": m, "coordinator": c.address,
			}).Warn("I don't know what to do with this")
		}
	}
}

func (c *Coordinator) sendHeartbeat() {
	log.WithFields(log.Fields{
		"coordinator": c.address,
	}).Debug("Sending hearbeat")
	hb := &common.HeartbeatMessage{}
	hb.Encode(c.encoder)
	if err := c.encoder.Flush(); err != nil {
		log.WithFields(log.Fields{
			"error": err, "coordinator": c.address,
		}).Error("Could not send heartbeat to coordinator")
	}
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
	bqm.Encode(c.encoder)
	if err := c.encoder.Flush(); err != nil {
		log.WithFields(log.Fields{
			"query": query, "client": client.RemoteAddr(), "error": err, "coordinator": c.address,
		}).Error("Could not forward query to coordinator")
	}
	response, _ := c.requests.WaitForMessage(bqm.GetID())
	log.Debugf("Response %v", response.(*common.BrokerSubscriptionDiffMessage))
}

// this forwards a publish message from a local producer to the coordinator and receives
// a BrokerSubscriptionDiffMessage in response
func (c *Coordinator) forwardPublish(msg *common.PublishMessage) *common.ForwardRequestMessage {
	var bpm common.BrokerPublishMessage
	bpm.FromPublishMessage(msg)
	bpm.Encode(c.encoder)
	if err := c.encoder.Flush(); err != nil {
		log.WithFields(log.Fields{
			"query": msg, "error": err, "coordinator": c.address,
		}).Error("Could not forward query to coordinator")
	}
	log.Debug("Waiting for publish response")
	response, _ := c.requests.WaitForMessage(bpm.GetID())
	return response.(*common.ForwardRequestMessage)
}
