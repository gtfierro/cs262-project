package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	addrString    string
	address       *net.TCPAddr
	metadata      *common.MetadataStore
	fwdTable      *ForwardingTable
	brokerManager BrokerManager
	leaderService LeaderService
	etcdManager   EtcdManager

	heartbeatInterval time.Duration

	messageBuffer      chan *MessageFromBroker
	brokerDeathChan    chan *common.UUID
	brokerLiveChan     chan *common.UUID
	brokerReassignChan chan *BrokerReassignment

	waitGroup sync.WaitGroup
	stop      chan bool
	stopped   bool
}

func NewServer(config *common.Config) *Server {
	var (
		err error
		s   = &Server{}
	)
	if config.Coordinator.Global {
		s.addrString = fmt.Sprintf("0.0.0.0:%d", config.Coordinator.Port)
	} else {
		s.addrString = fmt.Sprintf(":%d", config.Coordinator.Port)
	}

	// parse the config into an address
	s.address, err = net.ResolveTCPAddr("tcp", s.addrString)
	if err != nil {
		log.WithFields(log.Fields{
			"port": config.Coordinator.Port, "global": config.Coordinator.Global, "error": err.Error(),
		}).Fatal("Could not resolve the generated TCP address")
	}

	s.heartbeatInterval = time.Duration(config.Coordinator.HeartbeatInterval) * time.Second

	s.metadata = common.NewMetadataStore(config)
	s.brokerDeathChan = make(chan *common.UUID, 10)
	s.brokerLiveChan = make(chan *common.UUID, 10)
	s.brokerReassignChan = make(chan *BrokerReassignment, 500)
	s.messageBuffer = make(chan *MessageFromBroker, 50)

	var ipswitcher IPSwitcher
	if config.Coordinator.UseAWSIPSwitcher {
		ipswitcher, err = NewAWSIPSwitcher(config.Coordinator.InstanceId, config.Coordinator.Region, config.Coordinator.ElasticIP)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal("Could not create AWS IP Switcher")
		}
	} else {
		ipswitcher = &DummyIPSwitcher{}
	}

	if config.Coordinator.UseEtcd {
		etcdConn := NewEtcdConnection(strings.Split(config.Coordinator.EtcdAddresses, ","))
		s.leaderService = NewLeaderService(etcdConn, 5*time.Second, ipswitcher)
		s.etcdManager = NewEtcdManager(etcdConn, s.leaderService, config.Coordinator.CoordinatorCount,
			config.Coordinator.GCFreq, 5*time.Second, 1000, config.Coordinator.EnableContinuousCheckpointing)
	} else {
		s.leaderService = &DummyLeaderService{}
		s.etcdManager = &DummyEtcdManager{}
	}
	s.brokerManager = NewBrokerManager(s.etcdManager, s.heartbeatInterval, s.brokerDeathChan,
		s.brokerLiveChan, s.messageBuffer, s.brokerReassignChan, new(common.RealClock))
	s.fwdTable = NewForwardingTable(s.metadata, s.brokerManager, s.etcdManager, s.brokerDeathChan, s.brokerLiveChan, s.brokerReassignChan)
	go s.fwdTable.monitorInboundChannels()
	s.stop = make(chan bool, 1)
	s.stopped = false

	return s
}

func (s *Server) Shutdown() {
	close(s.stop)
	s.etcdManager.CancelWatch()
	s.leaderService.CancelWatch()
	s.waitGroup.Wait()
	s.stopped = true
}

// Starts the necessary goroutines and then listens; does not return
func (s *Server) Run() {
	logStartKey := s.rebuildIfNecessary()
	go s.handleLeadership()
	go s.handleBrokerMessages()
	go s.monitorLog(logStartKey)
	go s.monitorGeneralConnections()
	s.listenAndDispatch()
}

// Check the log: if there is more than one entry (the initial leader entry),
// then start the rebuild process
// Returns the key at which the log should start being monitored
func (s *Server) rebuildIfNecessary() string {
	s.metadata.DropDatabase() // Clear out DB; it's transient and only for running queries
	logBelowThreshold, logKey, logRev := s.etcdManager.GetLogStatus()
	if !logBelowThreshold {
		s.brokerManager.RebuildFromEtcd(logRev)
		s.fwdTable.RebuildFromEtcd(logRev)
	}
	return logKey
}

// Doesn't return
func (s *Server) handleLeadership() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	_, err := s.leaderService.AttemptToBecomeLeader()
	if err != nil {
		log.WithField("error", err).Error("Error while attempting to become the initial leader")
	}
	go s.leaderService.WatchForLeadershipChange()
	s.leaderService.MaintainLeaderLease()
}

// Won't return
func (s *Server) monitorLog(startKey string) {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	endKey := startKey
	for {
		// If we're a leader, just wait... nothing to be done here
		select {
		case <-s.stop:
		case <-s.leaderService.WaitForNonleadership():
		}
		if common.IsChanClosed(s.stop) {
			return
		}

		endKey = s.etcdManager.WatchLog(endKey)
	}
}

func (s *Server) monitorGeneralConnections() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	for {
		// If we're a leader, just wait... nothing to be done here
		select {
		case <-s.stop:
		case <-s.leaderService.WaitForNonleadership():
		}
		if common.IsChanClosed(s.stop) {
			return
		}

		commConn := NewReplicaCommConn(s.etcdManager, s.leaderService, GeneralSuffix, s.heartbeatInterval)
		go func() {
			select {
			case <-s.stop:
			case <-s.leaderService.WaitForLeadership():
			}
			commConn.Close()
		}()
		// Now we're definitely not a leader, set up a watch for the leader's events
		for {
			msg, err := commConn.ReceiveMessage()
			if err == nil {
				go s.dispatch(NewSingleEventCommConn(commConn, msg), LogPrefix+"/"+GeneralSuffix)
			} else {
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
	log.WithFields(log.Fields{
		"msg": msg, "messageType": common.GetMessageType(msg), "address": address,
	}).Info("Received a message")
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
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.stop:
			return
		case msg := <-s.messageBuffer:
			go s.handleMessage(msg)
		}
	}
}

func (s *Server) listenAndDispatch() {
	var (
		listener *net.TCPListener
		conn     *net.TCPConn
		err      error
	)
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	log.WithFields(log.Fields{
		"address": s.address,
	}).Info("Coordinator listening for requests!")

	// loop on the TCP connection and hand new connections to the dispatcher
LeaderLoop:
	for {
		waitChan := s.leaderService.WaitForLeadership()
		select {
		case <-waitChan:
		case <-s.stop:
		}
		if common.IsChanClosed(s.stop) {
			return
		}
		// listen on the address
		listener, err = net.ListenTCP("tcp", s.address)
		if err != nil {
			log.WithFields(log.Fields{
				"address": s.address, "error": err.Error(),
			}).Fatal("Could not listen on the provided address")
			return
		}
		go func() {
			waitChan := s.leaderService.WaitForNonleadership()
			select {
			case <-s.stop:
				listener.Close()
			case <-waitChan:
				listener.Close()
			}
		}()
	ListenLoop:
		for {
			conn, err = listener.AcceptTCP()
			if err != nil {
				if common.IsChanClosed(s.stop) {
					return
				} else {
					log.WithField("error", err.Error()).Error("Error accepting connection")
					continue ListenLoop
				}
			}
			if !s.leaderService.IsLeader() {
				log.WithField("address", conn.RemoteAddr()).Info("Rejecting inbound connection because leadership is not held")
				conn.Close() // Reject connections when not the leader
				listener.Close()
				continue LeaderLoop
			} else {
				log.WithField("address", conn.RemoteAddr()).Info("Accepting inbound connection on leader")
				commConn := NewLeaderCommConn(s.etcdManager, s.leaderService, GeneralSuffix, conn)
				go s.dispatch(commConn, fmt.Sprintf("%v", conn.RemoteAddr()))
			}
		}
	}
}
