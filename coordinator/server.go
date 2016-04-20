package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
)

type Server struct {
	address       *net.TCPAddr
	listener      *net.TCPListener
	metadata      *common.MetadataStore
	fwdTable      *ForwardingTable
	brokerManager BrokerManager

	messageBuffer   chan *MessageFromBroker
	brokerDeathChan chan *common.UUID
	brokerLiveChan  chan *common.UUID

	stop    chan bool
	stopped bool
}

func NewServer(config *common.Config) *Server {
	var (
		address string
		err     error
		s       = &Server{}
	)
	if config.Coordinator.Global {
		address = fmt.Sprintf("0.0.0.0:%d", config.Coordinator.Port)
	} else {
		address = fmt.Sprintf(":%d", config.Coordinator.Port)
	}

	// parse the config into an address
	s.address, err = net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.WithFields(log.Fields{
			"port": config.Coordinator.Port, "global": config.Coordinator.Global, "error": err.Error(),
		}).Fatal("Could not resolve the generated TCP address")
	}

	// listen on the address
	s.listener, err = net.ListenTCP("tcp", s.address)
	if err != nil {
		log.WithFields(log.Fields{
			"address": s.address, "error": err.Error(),
		}).Fatal("Could not listen on the provided address")
	}

	s.metadata = common.NewMetadataStore(config)
	s.brokerDeathChan = make(chan *common.UUID)
	s.brokerLiveChan = make(chan *common.UUID)
	s.messageBuffer = make(chan *MessageFromBroker, 50)
	s.brokerManager = NewBrokerManager(config.Coordinator.HeartbeatInterval, s.brokerDeathChan,
		s.brokerLiveChan, s.messageBuffer, new(common.RealClock))
	s.fwdTable = NewForwardingTable(s.metadata, s.brokerManager)
	s.stop = make(chan bool)
	s.stopped = false

	return s
}

func (s *Server) handleMessage(brokerMessage *MessageFromBroker) {
	brokerID := brokerMessage.broker.BrokerID
	switch msg := brokerMessage.message.(type) {
	case *common.PublishMessage:
		s.fwdTable.HandlePublish(msg, brokerID)
	case *common.BrokerQueryMessage:
		s.fwdTable.HandleSubscription(msg.QueryMessage, msg.ClientAddr, brokerID)
	case *common.PublisherTerminationMessage:
		s.fwdTable.HandlePublisherDeath(msg.PublisherID, brokerID)
	case *common.ClientTerminationMessage:
		s.fwdTable.HandleSubscriberDeath(msg.ClientAddr, brokerID)
	default:
		log.WithFields(log.Fields{
			"message": msg, "messageType": common.GetMessageType(msg), "brokerID": brokerID,
		}).Warn("Received unexpected message from a broker")
	}
}

func (s *Server) dispatch(conn *net.TCPConn) {
	reader := msgp.NewReader(conn)
	writer := msgp.NewWriter(conn)
	msg, err := common.MessageFromDecoderMsgp(reader)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "tcpAddr": conn.RemoteAddr(),
		}).Error("Error decoding message from connection")
		conn.Close()
		return
	}
	switch m := msg.(type) {
	case *common.BrokerConnectMessage:
		err = s.brokerManager.ConnectBroker(&m.BrokerInfo, conn)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err, "brokerInfo": m.BrokerInfo, "tcpAddr": conn.RemoteAddr(),
			}).Error("Error while connecting to broker")
		}
		ack := &common.AcknowledgeMessage{MessageID: m.MessageID}
		ack.Encode(writer)
	case *common.BrokerRequestMessage:
		// TODO
	default:
		log.WithFields(log.Fields{
			"tcpAddr": conn.RemoteAddr(), "message": msg, "messageType": common.GetMessageType(msg),
		}).Warn("Received unexpected message type over a new connection")
	}
}

func (s *Server) startBrokerMessageHandler() {
	go func() {
		for {
			go s.handleMessage(<-s.messageBuffer)
		}
	}()
}

func (s *Server) listenAndDispatch() {
	var (
		conn *net.TCPConn
		err  error
	)
	log.WithFields(log.Fields{
		"address": s.address,
	}).Info("Coordinator listening for requests!")

	// loop on the TCP connection and hand new connections to the dispatcher
	for {
		conn, err = s.listener.AcceptTCP()
		if err != nil {
			//if s.closed {
			//	return // exit
			//}
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error accepting connection")
		}
		go s.dispatch(conn)
	}
}
