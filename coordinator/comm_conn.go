package main

import (
	log "github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
	Close()
}

// Communication Connection
type ReplicaCommConn struct {
	etcdManager       *EtcdManager
	leaderService     *LeaderService
	heartbeatInterval time.Duration
	logPrefix         string
	highestRev        int64
	revLock           sync.Mutex
	watchChan         chan *etcdc.WatchResponse
	messageBuffer     chan *etcdc.Event
	heartbeatChan     chan bool
	leaderChan        chan bool
	closeChan         chan bool
}

// Communication Connection
type LeaderCommConn struct {
	etcdManager   *EtcdManager
	leaderService *LeaderService
	logPrefix     string
	tcpConn       *net.TCPConn
	reader        *msgp.Reader
	writer        *msgp.Writer
}

// Has a single event to be received; passes Send back to the parent
type SingleEventCommConn struct {
	parentComm *ReplicaCommConn
	event      common.Sendable
}

func NewReplicaCommConn(etcdMgr *EtcdManager, leaderService *LeaderService,
	logPrefix string, startingRev int64, heartbeatInterval time.Duration) *ReplicaCommConn {
	rcc := new(ReplicaCommConn)
	rcc.etcdManager = etcdMgr
	rcc.leaderService = leaderService
	rcc.logPrefix = logPrefix
	rcc.highestRev = startingRev
	rcc.heartbeatInterval = heartbeatInterval
	rcc.messageBuffer = make(chan *etcdc.Event, 1000)
	rcc.heartbeatChan = make(chan bool)
	rcc.leaderChan = leaderService.WaitForLeadership()

	watchStartKey, err := rcc.etcdManager.GetHighestKeyAtRev(logPrefix+"rcvd/", startingRev)
	if err != nil {
		return
	}
	rcc.watchChan = rcc.etcdManager.WatchFromKeyAtRevision(logPrefix+"rcvd/", watchStartKey)

	return rcc
}

// Simulate heartbeats from the broker
func (rcc *ReplicaCommConn) sendHeartbeats() {
	for {
		select {
		case <-rcc.closeChan:
			return
		case <-rcc.leaderChan:
			return
		case <-time.After(rcc.heartbeatInterval):
			if rcc.logPrefix != GeneralLog {
				// Send heartbeats to brokers only; ignore for general
				rcc.heartbeatChan <- true
			}
		}
	}
}

func (rcc *ReplicaCommConn) nudgeHighestRev(rev int64) {
	rcc.revLock.Lock()
	defer rcc.revLock.Unlock()
	if rev > rcc.highestRev {
		rcc.highestRev = rev
	}
}

func (rcc *ReplicaCommConn) ReceiveMessage() (msg common.Sendable, err error) {
	// TODO should have logic here that if the watchChan closes, we reopen it
	// TODO also need to GC the log here:
	//   every so often when reading, write to some node within the prefix
	//   what key you've read up to. also read the other replica's last read key.
	//  then you can delete everything up to the lower of the two
	for {
		select {
		case <-rcc.closeChan:
			return nil, io.EOF
		case <-rcc.leaderChan:
			return nil, io.EOF
		case <-rcc.heartbeatChan:
			return &common.HeartbeatMessage{}, nil
		case event := <-rcc.messageBuffer:
			if event.Type != mvccpb.PUT || !event.IsCreate() {
				log.WithFields(log.Fields{
					"eventType": event.Type, "key": event.Kv.Key,
				}).Warn("Non put+create event found in the general receive log!")
				continue
			}
			msg, err := common.MessageFromBytes(event.Kv.Value)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err, "key": event.Kv.Key,
				}).Error("Error while unmarshalling bytes from general rcvd log at the given key")
				continue
			}
			return msg, nil
		case watchResp := <-rcc.watchChan:
			rcc.nudgeHighestRev(watchResp.Header.Revision)
			for _, event := range watchResp.Events {
				rcc.messageBuffer <- event
			}
		}
	}
}

func (rcc *ReplicaCommConn) Send(msg common.Sendable) error {
	select {
	case <-rcc.closeChan:
		return nil, io.EOF
	case <-rcc.leaderChan:
		return nil, io.EOF
	}
	// not leader; can safely ignore this message
}

func (rcc *ReplicaCommConn) Close() {
	close(rcc.closeChan)
}

func (rcc *ReplicaCommConn) GetBrokerConn(brokerID common.UUID) *ReplicaCommConn {
	rcc.revLock.Lock()
	rev := rcc.highestRev
	rcc.revLock.Unlock()
	return NewReplicaCommConn(rcc.etcdManager, rcc.leaderService, GetBrokerLogPrefix(brokerID),
		rev, rcc.heartbeatInterval)
}

func NewLeaderCommConn(etcdMgr *EtcdManager, leaderService *LeaderService, logPrefix string, tcpConn *net.TCPConn) *LeaderCommConn {
	lcc := new(LeaderCommConn)
	lcc.etcdManager = etcdMgr
	lcc.leaderService = leaderService
	lcc.logPrefix = logPrefix
	lcc.tcpConn = tcpConn
	lcc.reader = msgp.NewReader(tcpConn)
	lcc.writer = msgp.NewWriter(tcpConn)
	return lcc
}

func (lcc *LeaderCommConn) ReceiveMessage() (msg common.Sendable, err error) {
	msg, err = common.MessageFromDecoderMsgp(lcc.reader)
	if err != nil {
		log.WithField("error", err).Error("Error receiving message!")
		return
	}
	if ack, ok := msg.(*common.AcknowledgeMessage); ok {
		// We store ACKs in the send log so it's easier to see which messages were acked
		err = lcc.etcdManager.WriteToLog(lcc.logPrefix+"sent/", ack)
	} else if _, ok := msg.(*common.HeartbeatMessage); ok {
		// Do nothing - we don't want to log Heartbeats
	} else {
		err = lcc.etcdManager.WriteToLog(lcc.logPrefix+"rcvd/", msg)
	}
	return
}

func (lcc *LeaderCommConn) Send(msg common.Sendable) error {
	// TODO sender needs to GC its send log at some point
	// should have some sort of interaction with the ACKs
	err := msg.Encode(lcc.writer)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "message": msg,
		}).Error("Error sending message!")
		return err
	}
	if _, ok := msg.(*common.AcknowledgeMessage); ok {
		// Do nothing - don't want to log outbound acks (the broker can always resend)
	} else if _, ok := msg.(*common.RequestHeartbeatMessage); ok {
		// Don't want to log these
	} else if _, ok := msg.(*common.BrokerAssignmentMessage); ok {
		// No need to log these; if the client/pub doesn't receive it, it will ask again
	}
	err = lcc.etcdManager.WriteToLog(lcc.logPrefix+"sent/", msg)
	return err
}

func (lcc *LeaderCommConn) Close() {
	lcc.tcpConn.Close()
}

func (lcc *LeaderCommConn) GetBrokerConn(brokerID common.UUID) *LeaderCommConn {
	return NewLeaderCommConn(lcc.etcdManager, lcc.leaderService, GetBrokerLogPrefix(brokerID),
		lcc.tcpConn)
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

func (secc *SingleEventCommConn) GetBrokerConn(brokerID common.UUID) *ReplicaCommConn {
	return secc.parentComm.GetBrokerConn(brokerID)
}
