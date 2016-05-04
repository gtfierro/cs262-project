package main

import (
	log "github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gtfierro/cs262-project/common"
	"sync"
	"time"
)

const LeaderKey = "leader"

type LeaderService interface {
	MaintainLeaderLease()
	WatchForLeadershipChange()
	CancelWatch()
	IsLeader() bool
	AttemptToBecomeLeader() (bool, error)
	GetLeadershipChangeRevision() int64
	WaitForLeadership() chan bool
	WaitForNonleadership() chan bool
}

type LeaderServiceImpl struct {
	isLeader          bool
	leaderChangeRev   int64
	leaderLease       etcdc.LeaseID
	leaderLock        sync.RWMutex
	leaderWaitChans   []chan bool
	unleaderWaitChans []chan bool
	etcdConn          *EtcdConnection
	stop              chan bool
	waitGroup         sync.WaitGroup
	ipswitcher        IPSwitcher
}

func NewLeaderService(etcdConn *EtcdConnection, timeout time.Duration, ipswitcher IPSwitcher) *LeaderServiceImpl {
	cs := new(LeaderServiceImpl)
	cs.etcdConn = etcdConn
	cs.leaderChangeRev = -1
	cs.leaderWaitChans = []chan bool{}
	cs.unleaderWaitChans = []chan bool{}
	cs.stop = make(chan bool, 1)
	cs.ipswitcher = ipswitcher
	return cs
}

// Doesn't return. Watches for a lack of a leader and if so
// attempts to become the new leader
func (cs *LeaderServiceImpl) WatchForLeadershipChange() {
	cs.waitGroup.Add(1)
	defer cs.waitGroup.Done()
	var watchResp etcdc.WatchResponse
	watchChan := cs.etcdConn.watcher.Watch(cs.etcdConn.GetCtx(), LeaderKey)
	for {
		select {
		case <-cs.stop:
			return
		case watchResp = <-watchChan:
		}
		if common.IsChanClosed(cs.stop) {
			return
		}
		if watchResp.Canceled {
			watchChan = cs.etcdConn.watcher.Watch(cs.etcdConn.GetCtx(), LeaderKey)
		}
		for _, event := range watchResp.Events {
			if event.Type == mvccpb.DELETE && string(event.Kv.Key) == LeaderKey { // Currently no leader!
				log.WithFields(log.Fields{
					"isCreate": event.IsCreate(), "isModify": event.IsModify(), "version": event.Kv.Version,
				}).Debug("WatchForLeadershipChange detected a deletion event!")
				cs.AttemptToBecomeLeader()
			}
		}
	}
}

func (cs *LeaderServiceImpl) CancelWatch() {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	for _, waitChan := range cs.unleaderWaitChans {
		close(waitChan)
	}
	for _, waitChan := range cs.leaderWaitChans {
		close(waitChan)
	}
	close(cs.stop)
	cs.waitGroup.Wait()
}

// Maintain the leadership lease; doesn't return
func (cs *LeaderServiceImpl) MaintainLeaderLease() {
	var waitChan chan bool
	cs.waitGroup.Add(1)
	defer cs.waitGroup.Done()
	for {
		// If we're not a leader, just wait... nothing to be done here
		waitChan = cs.WaitForLeadership()
		select {
		case <-cs.stop:
			return
		case <-waitChan:
			if common.IsChanClosed(cs.stop) {
				return
			}
		}
		waitChan = cs.WaitForNonleadership()

		select {
		case <-cs.stop:
			return
		case <-waitChan:
			if common.IsChanClosed(cs.stop) {
				return
			} else {
				continue
			}
		case <-time.After(3 * time.Second): // to maintain lease
			if common.IsChanClosed(cs.stop) {
				return
			}
			cs.leaderLock.RLock()
			_, err := cs.etcdConn.client.KeepAliveOnce(cs.etcdConn.GetCtx(), cs.leaderLease)
			cs.leaderLock.RUnlock()
			if err == rpctypes.ErrLeaseNotFound {
				log.Info("Lost leadership! Lease expired.")
				// Lost our lease; we are no longer the leader
				cs.AttemptToBecomeLeader()
			} else if err != nil {
				log.WithField("error", err).Error("Error while attempting to renew lease for leader key")
			}
		}
	}
}

func (cs *LeaderServiceImpl) GetLeadershipChangeRevision() int64 {
	cs.leaderLock.RLock()
	defer cs.leaderLock.RUnlock()
	return cs.leaderChangeRev
}

// return true iff we became the leader which will happen only if there is
// currently no leader
func (cs *LeaderServiceImpl) AttemptToBecomeLeader() (bool, error) {
	var (
		changeRev int64
		isLeader  bool
	)
	txn := cs.etcdConn.kv.Txn(cs.etcdConn.GetCtx())
	cmp := etcdc.Compare(etcdc.Version(LeaderKey), "=", 0)
	leaseResp, err := cs.etcdConn.client.Grant(cs.etcdConn.GetCtx(), 5)
	if err != nil {
		log.WithField("error", err).Error("Error while attempting to get a lease for a leader key!")
		return false, err
	}
	putKeyOp := etcdc.OpPut(LeaderKey, "", etcdc.WithLease(leaseResp.ID))
	getKeyOp := etcdc.OpGet(LeaderKey)
	txnResp, err := txn.If(cmp).Then(putKeyOp).Else(getKeyOp).Commit()
	if err != nil {
		return false, err
	}
	for _, resp := range txnResp.Responses {
		if r, ok := resp.Response.(*etcdserverpb.ResponseUnion_ResponseRange); ok && !txnResp.Succeeded {
			isLeader = false
			changeRev = r.ResponseRange.Kvs[0].ModRevision
		} else if r, ok := resp.Response.(*etcdserverpb.ResponseUnion_ResponsePut); ok && txnResp.Succeeded {
			isLeader = true
			changeRev = r.ResponsePut.GetHeader().Revision
			cs.leaderLock.Lock()
			cs.leaderLease = leaseResp.ID
			cs.leaderLock.Unlock()

			// Acquire the IP
			err := cs.ipswitcher.AcquireIP()
			if err != nil {
				log.WithField("Error", err).Error("Came leader but could not acquire IP!")
			} else {
				log.Warn("Successfully became leader and got IP!")
			}
		}
	}
	log.WithField("isLeader", isLeader).Info("Attempted to become leader")
	cs.setLeader(isLeader, changeRev)
	return isLeader, nil
}

func (cs *LeaderServiceImpl) IsLeader() bool {
	cs.leaderLock.RLock()
	defer cs.leaderLock.RUnlock()
	return cs.isLeader
}

func (cs *LeaderServiceImpl) setLeader(isLeader bool, changeRev int64) {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	cs.leaderChangeRev = changeRev
	cs.isLeader = isLeader
	if isLeader {
		for _, c := range cs.leaderWaitChans {
			close(c)
		}
		cs.leaderWaitChans = []chan bool{}
	} else {
		for _, c := range cs.unleaderWaitChans {
			close(c)
		}
		cs.unleaderWaitChans = []chan bool{}
	}
}

// Returns a channel which will be closed when this is leader
func (cs *LeaderServiceImpl) WaitForLeadership() chan bool {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	// if these are not buffered channels, then sending on the channel
	// can block indefinitely and deadlock -- GTF
	c := make(chan bool, 1)
	if cs.isLeader {
		close(c)
		return c
	} else {
		cs.leaderWaitChans = append(cs.leaderWaitChans, c)
		return c
	}
}

// Returns a channel which will be closed when this is nonleader
func (cs *LeaderServiceImpl) WaitForNonleadership() chan bool {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	// if these are not buffered channels, then sending on the channel
	// can block indefinitely and deadlock -- GTF
	c := make(chan bool, 1)
	if !cs.isLeader {
		c <- true
		return c
	} else {
		cs.unleaderWaitChans = append(cs.unleaderWaitChans, c)
		return c
	}
}

type DummyLeaderService struct{}

func (dls *DummyLeaderService) MaintainLeaderLease()      {}
func (dls *DummyLeaderService) WatchForLeadershipChange() {}
func (dls *DummyLeaderService) CancelWatch()              {}
func (dls *DummyLeaderService) IsLeader() bool {
	return true
}
func (dls *DummyLeaderService) AttemptToBecomeLeader() (bool, error) {
	return true, nil
}
func (dls *DummyLeaderService) GetLeadershipChangeRevision() int64 {
	return -1
}
func (dls *DummyLeaderService) WaitForLeadership() chan bool {
	c := make(chan bool)
	close(c)
	return c
}
func (dls *DummyLeaderService) WaitForNonleadership() chan bool {
	return make(chan bool)
}
