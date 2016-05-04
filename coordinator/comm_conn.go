package main

import (
	log "github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"sync"
	"time"
)

type CommConn interface {
	Send(msg common.Sendable) error
	ReceiveMessage() (msg common.Sendable, err error)
	GetBrokerConn(brokerID common.UUID) CommConn
	GetPendingMessages() map[common.MessageIDType]common.SendableWithID
	Close()
}

// Communication Connection
type ReplicaCommConn struct {
	etcdManager          EtcdManager
	leaderService        LeaderService
	heartbeatInterval    time.Duration
	idOrGeneral          string
	revLock              sync.Mutex
	watchChan            chan *etcdc.WatchResponse
	messageBuffer        chan common.Sendable
	pendingMessageBuffer map[common.MessageIDType]common.SendableWithID
	heartbeatChan        chan bool
	leaderStatusChanged  bool
	leaderStatusLock     sync.RWMutex
	closeChan            chan bool
}

// Communication Connection
type LeaderCommConn struct {
	etcdManager   EtcdManager
	leaderService LeaderService
	idOrGeneral   string
	tcpConn       *net.TCPConn
	reader        *msgp.Reader
	writer        *msgp.Writer
}

// Has a single event to be received; passes Send back to the parent
type SingleEventCommConn struct {
	parentComm *ReplicaCommConn
	event      common.Sendable
}

func NewReplicaCommConn(etcdMgr EtcdManager, leaderService LeaderService,
	idOrGeneral string, heartbeatInterval time.Duration) *ReplicaCommConn {
	rcc := new(ReplicaCommConn)
	rcc.etcdManager = etcdMgr
	rcc.leaderService = leaderService
	rcc.idOrGeneral = idOrGeneral
	rcc.heartbeatInterval = heartbeatInterval
	rcc.messageBuffer = make(chan common.Sendable, 20)
	rcc.pendingMessageBuffer = make(map[common.MessageIDType]common.SendableWithID)
	rcc.heartbeatChan = make(chan bool)
	rcc.closeChan = make(chan bool, 1)

	go func() {
		<-leaderService.WaitForLeadership()
		rcc.leaderStatusLock.Lock()
		rcc.leaderStatusChanged = true
		rcc.leaderStatusLock.Unlock()
	}()

	etcdMgr.RegisterLogHandler(idOrGeneral, rcc.logHandler)

	// Send heartbeats to brokers only; ignore for general
	if rcc.idOrGeneral != GeneralSuffix {
		go rcc.sendHeartbeats()
	}

	return rcc
}

func (rcc *ReplicaCommConn) GetPendingMessages() map[common.MessageIDType]common.SendableWithID {
	msgs := rcc.pendingMessageBuffer
	rcc.pendingMessageBuffer = make(map[common.MessageIDType]common.SendableWithID)
	return msgs
}

func (rcc *ReplicaCommConn) logHandler(msg common.Sendable, isSend bool) {
	if isSend {
		switch m := msg.(type) {
		case *common.LeaderChangeMessage:
			rcc.leaderStatusLock.RLock()
			if rcc.leaderStatusChanged && !common.IsChanClosed(rcc.closeChan) {
				close(rcc.closeChan)
			}
			rcc.leaderStatusLock.RUnlock()
		case *common.AcknowledgeMessage:
			delete(rcc.pendingMessageBuffer, m.MessageID)
		case common.SendableWithID:
			rcc.pendingMessageBuffer[m.GetID()] = m
		default:
			// ignore
		}
	} else {
		rcc.messageBuffer <- msg
	}
}

// Simulate heartbeats from the broker
func (rcc *ReplicaCommConn) sendHeartbeats() {
	for {
		select {
		case <-rcc.closeChan:
			return
		case <-time.After(rcc.heartbeatInterval):
			rcc.heartbeatChan <- true
		}
	}
}

func (rcc *ReplicaCommConn) ReceiveMessage() (msg common.Sendable, err error) {
	for {
		select {
		case msg := <-rcc.messageBuffer:
			log.WithFields(log.Fields{
				"message": msg, "messageType": common.GetMessageType(msg),
			}).Debug("Receiving message on replica")
			return msg, nil
		case <-rcc.closeChan:
			return nil, io.EOF
		case <-rcc.heartbeatChan:
			return &common.HeartbeatMessage{}, nil
		}
	}
}

func (rcc *ReplicaCommConn) Send(msg common.Sendable) error {
	select {
	case <-rcc.closeChan:
		return io.EOF
	default:
	}
	// not leader; can safely ignore this message except for ACKing if necessary
	if withID, ok := msg.(common.SendableWithID); ok {
		rcc.messageBuffer <- &common.AcknowledgeMessage{MessageID: withID.GetID()}
	}
	return nil
}

func (rcc *ReplicaCommConn) Close() {
	if !common.IsChanClosed(rcc.closeChan) {
		close(rcc.closeChan)
	}
	rcc.etcdManager.UnregisterLogHandler(rcc.idOrGeneral)
}

func (rcc *ReplicaCommConn) GetBrokerConn(brokerID common.UUID) CommConn {
	return NewReplicaCommConn(rcc.etcdManager, rcc.leaderService, string(brokerID), rcc.heartbeatInterval)
}

func NewLeaderCommConn(etcdMgr EtcdManager, leaderService LeaderService, idOrGeneral string, tcpConn *net.TCPConn) *LeaderCommConn {
	lcc := new(LeaderCommConn)
	lcc.etcdManager = etcdMgr
	lcc.leaderService = leaderService
	lcc.idOrGeneral = idOrGeneral
	lcc.tcpConn = tcpConn
	lcc.reader = msgp.NewReader(tcpConn)
	lcc.writer = msgp.NewWriter(tcpConn)

	nonleaderChan := leaderService.WaitForNonleadership()
	go func() {
		<-nonleaderChan
		lcc.Close() // Close connection if we're not the leader
	}()

	return lcc
}

func (lcc *LeaderCommConn) GetPendingMessages() map[common.MessageIDType]common.SendableWithID {
	return make(map[common.MessageIDType]common.SendableWithID)
}

func (lcc *LeaderCommConn) ReceiveMessage() (msg common.Sendable, err error) {
	msg, err = common.MessageFromDecoderMsgp(lcc.reader)
	if err != nil {
		log.WithField("error", err).Error("Error receiving message!")
		return
	} else {
		log.WithFields(log.Fields{
			"msg": msg, "messageType": common.GetMessageType(msg),
		}).Debug("Received message on leader")
	}
	if ack, ok := msg.(*common.AcknowledgeMessage); ok {
		// We store ACKs in the send log so it's easier to see which messages were acked
		err = lcc.etcdManager.WriteToLog(lcc.idOrGeneral, true, ack)
	} else if _, ok := msg.(*common.HeartbeatMessage); ok {
		// Do nothing - we don't want to log Heartbeats
	} else {
		err = lcc.etcdManager.WriteToLog(lcc.idOrGeneral, false, msg)
	}
	return
}

func (lcc *LeaderCommConn) Send(msg common.Sendable) (err error) {
	// TODO sender needs to GC its send log at some point
	// should have some sort of interaction with the ACKs
	if err = msg.Encode(lcc.writer); err != nil {
		log.WithFields(log.Fields{
			"error": err, "message": msg,
		}).Error("Error sending message!")
		return
	}
	if err = lcc.writer.Flush(); err != nil {
		log.WithFields(log.Fields{
			"error": err, "message": msg,
		}).Warn("Error sending message!")
		return
	}
	if _, ok := msg.(*common.AcknowledgeMessage); ok {
		// Do nothing - don't want to log outbound acks (the broker can always resend)
	} else if _, ok := msg.(*common.RequestHeartbeatMessage); ok {
		// Don't want to log these
	} else if _, ok := msg.(*common.BrokerAssignmentMessage); ok {
		// No need to log these; if the client/pub doesn't receive it, it will ask again
	} else {
		err = lcc.etcdManager.WriteToLog(lcc.idOrGeneral, true, msg)
	}
	return
}

func (lcc *LeaderCommConn) Close() {
	lcc.tcpConn.Close()
}

func (lcc *LeaderCommConn) GetBrokerConn(brokerID common.UUID) CommConn {
	return NewLeaderCommConn(lcc.etcdManager, lcc.leaderService, string(brokerID), lcc.tcpConn)
}

func NewSingleEventCommConn(parent *ReplicaCommConn, event common.Sendable) *SingleEventCommConn {
	secc := new(SingleEventCommConn)
	secc.parentComm = parent
	secc.event = event
	return secc
}

func (secc *SingleEventCommConn) Send(msg common.Sendable) error {
	return secc.parentComm.Send(msg)
}

func (secc *SingleEventCommConn) ReceiveMessage() (common.Sendable, error) {
	if secc.event == nil {
		return nil, io.EOF
	}
	event := secc.event
	secc.event = nil
	return event, nil
}

func (secc *SingleEventCommConn) Close() {
	// No-op
}

func (secc *SingleEventCommConn) GetBrokerConn(brokerID common.UUID) CommConn {
	return secc.parentComm.GetBrokerConn(brokerID)
}

func (secc *SingleEventCommConn) GetPendingMessages() map[common.MessageIDType]common.SendableWithID {
	return make(map[common.MessageIDType]common.SendableWithID)
}
