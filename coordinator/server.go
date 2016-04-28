package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"io"
	"net"
	"time"
)

type Server struct {
	address       *net.TCPAddr
	listener      *net.TCPListener
	metadata      *common.MetadataStore
	fwdTable      *ForwardingTable
	brokerManager BrokerManager
	leaderService *LeaderService
	etcdManager   *EtcdManager

	heartbeatInterval time.Duration

	messageBuffer      chan *MessageFromBroker
	brokerDeathChan    chan *common.UUID
	brokerLiveChan     chan *common.UUID
	brokerReassignChan chan *BrokerReassignment

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

	s.heartbeatInterval = time.Duration(config.Coordinator.HeartbeatInterval) * time.Second

	s.metadata = common.NewMetadataStore(config)
	s.brokerDeathChan = make(chan *common.UUID, 10)
	s.brokerLiveChan = make(chan *common.UUID, 10)
	s.brokerReassignChan = make(chan *BrokerReassignment, 500)
	s.messageBuffer = make(chan *MessageFromBroker, 50)

	// TODO move to real config
	s.etcdManager = NewEtcdManager([]string{"127.0.0.1:2377"}, 2*time.Second, 1000)
	s.leaderService = NewLeaderService(s.etcdManager)
	s.brokerManager = NewBrokerManager(s.etcdManager, s.heartbeatInterval, s.brokerDeathChan,
		s.brokerLiveChan, s.messageBuffer, s.brokerReassignChan, new(common.RealClock))
	s.fwdTable = NewForwardingTable(s.metadata, s.brokerManager, s.etcdManager, s.brokerDeathChan, s.brokerLiveChan, s.brokerReassignChan)
	go s.fwdTable.monitorInboundChannels()
	s.stop = make(chan bool)
	s.stopped = false

	return s
}

func (s *Server) handleLeadership() {
	s.leaderService.AttemptToBecomeInitialLeader()
	for {
		if s.leaderService.IsLeader() {
			// TODO continually update etcd lease to demonstrate liveness
		} else {
			// TODO continually watch etcd for the possibility that the current leader is down
		}
	}
}

func (s *Server) monitorGeneralLog() {
	for {
		// If we're a leader, just wait... nothing to be done here
		<-s.leaderService.WaitForNonleadership()

		watchStartRev := s.leaderService.GetLeadershipChangeRevision()
		commConn := NewReplicaCommConn(s.etcdManager, s.leaderService, GeneralLog, watchStartRev, s.heartbeatInterval)
		go func() {
			<-s.leaderService.WaitForLeadership()
			commConn.Close()
		}()
		// Now we're definitely not a leader, set up a watch for the leader's events
		for {
			msg, err := commConn.ReceiveMessage()
			if err == nil {
				go s.dispatch(NewSingleEventCommConn(commConn, msg), GeneralLog)
			} else if err == io.EOF {
				break // continue outer loop since we're no longer leader
			}
		}
	}
}

func (s *Server) handleMessage(brokerMessage *MessageFromBroker) {
	brokerID := brokerMessage.broker.BrokerID
	switch msg := brokerMessage.message.(type) {
	case *common.BrokerPublishMessage:
		s.fwdTable.HandlePublish(msg.UUID, msg.Metadata, brokerID, nil)
		brokerMessage.broker.Send(&common.AcknowledgeMessage{msg.MessageID})
	case *common.BrokerQueryMessage:
		s.fwdTable.HandleSubscription(msg.Query, msg.UUID, brokerID, nil)
		brokerMessage.broker.Send(&common.AcknowledgeMessage{msg.MessageID})
	case *common.PublisherTerminationMessage:
		s.fwdTable.HandlePublisherTermination(msg.PublisherID, brokerID)
		brokerMessage.broker.Send(&common.AcknowledgeMessage{msg.MessageID})
	case *common.ClientTerminationMessage:
		s.fwdTable.HandleSubscriberTermination(msg.ClientID, brokerID)
		brokerMessage.broker.Send(&common.AcknowledgeMessage{msg.MessageID})
	default:
		log.WithFields(log.Fields{
			"message": msg, "messageType": common.GetMessageType(msg), "brokerID": brokerID,
		}).Warn("Received unexpected message from a broker")
	}
}

func (s *Server) dispatch(commConn CommConn, address string) {
	msg, err := commConn.ReceiveMessage()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "tcpAddr": address,
		}).Error("Error decoding message from connection")
		commConn.Close()
		return
	}
	switch m := msg.(type) {
	case *common.BrokerConnectMessage:
		err = s.brokerManager.ConnectBroker(&m.BrokerInfo, commConn.GetBrokerConn(m.BrokerID))
		if err != nil {
			log.WithFields(log.Fields{
				"error": err, "brokerInfo": m.BrokerInfo, "tcpAddr": address,
			}).Error("Error while connecting to broker")
		}
		ack := &common.AcknowledgeMessage{MessageID: m.MessageID}
		commConn.Send(ack)
	case *common.BrokerRequestMessage:
		if resp, err := s.brokerManager.HandlePubClientRemapping(m); err == nil {
			commConn.Send(resp)
		} else {
			log.WithFields(log.Fields{
				"requestMessage": msg, "error": err,
			}).Error("Publisher/client requested remapping but failed")
		}
	default:
		log.WithFields(log.Fields{
			"tcpAddr": address, "message": msg, "messageType": common.GetMessageType(msg),
		}).Warn("Received unexpected message type over a new connection")
	}
}

func (s *Server) handleBrokerMessages() {
	for {
		go s.handleMessage(<-s.messageBuffer)
	}
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
		if !s.leaderService.IsLeader() {
			conn.Close() // Reject connections when not the leader
		} else {
			commConn := NewLeaderCommConn(s.etcdManager, s.leaderService, GeneralLog, conn)
			go s.dispatch(commConn, fmt.Sprintf("%v", conn.RemoteAddr()))
		}
	}
}
