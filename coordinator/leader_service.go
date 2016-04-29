package main

import (
	log "github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
	"time"
)

const LeaderKey = "leader"

type LeaderService struct {
	isLeader          bool
	leaderChangeRev   int64
	coordAddr         string
	leaderLease       etcdc.LeaseID
	leaderLock        sync.RWMutex
	leaderWaitChans   []chan bool
	unleaderWaitChans []chan bool
	etcdConn          *EtcdConnection
}

func NewLeaderService(etcdConn *EtcdConnection, coordAddr string, timeout time.Duration) *LeaderService {
	cs := new(LeaderService)
	cs.etcdConn = etcdConn
	cs.coordAddr = coordAddr
	cs.leaderChangeRev = -1
	cs.leaderWaitChans = []chan bool{}
	cs.unleaderWaitChans = []chan bool{}
	return cs
}

// Doesn't return. Watches for a lack of a leader and if so
// attempts to become the new leader
// TODO should restart the watch if it dies
func (cs *LeaderService) WatchForLeadershipChange() {
	watchChan := cs.etcdConn.watcher.Watch(cs.etcdConn.GetCtx(), LeaderKey)
	for {
		watchResp := <-watchChan
		for _, event := range watchResp.Events {
			if event.Type == mvccpb.DELETE { // Currently no leader!
				cs.AttemptToBecomeLeader()
			}
		}
	}
}

// Maintain the leadership lease; doesn't return
func (cs *LeaderService) MaintainLeaderLease() {
	for {
		// If we're not a leader, just wait... nothing to be done here
		<-cs.WaitForLeadership()

		select {
		case <-cs.WaitForNonleadership():
			continue
		case <-time.After(3 * time.Second): // to maintain lease
			cs.leaderLock.RLock()
			_, err := cs.etcdConn.client.KeepAliveOnce(cs.etcdConn.GetCtx(), cs.leaderLease)
			cs.leaderLock.RUnlock()
			if err == rpctypes.ErrLeaseNotFound {
				// Lost our lease; we are no longer the leader
				cs.AttemptToBecomeLeader()
			} else if err != nil {
				log.WithField("error", err).Error("Error while attempting to renew lease for leader key")
			}
		}
	}
}

func (cs *LeaderService) GetLeadershipChangeRevision() int64 {
	cs.leaderLock.RLock()
	defer cs.leaderLock.RUnlock()
	return cs.leaderChangeRev
}

// return true iff we became the leader which will happen only if there is
// currently no leader
func (cs *LeaderService) AttemptToBecomeLeader() (bool, error) {
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
	putKeyOp := etcdc.OpPut(LeaderKey, cs.coordAddr, etcdc.WithLease(leaseResp.ID))
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
		}
	}
	cs.SetLeader(isLeader, changeRev)
	return isLeader, nil
}

func (cs *LeaderService) IsLeader() bool {
	cs.leaderLock.RLock()
	defer cs.leaderLock.RUnlock()
	return cs.isLeader
}

func (cs *LeaderService) SetLeader(isLeader bool, changeRev int64) {
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
func (cs *LeaderService) WaitForLeadership() chan bool {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	c := make(chan bool)
	if cs.isLeader {
		c <- true
		return c
	} else {
		cs.leaderWaitChans = append(cs.leaderWaitChans, c)
		return c
	}
}

// Returns a channel which will be closed when this is nonleader
func (cs *LeaderService) WaitForNonleadership() chan bool {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	c := make(chan bool)
	if !cs.isLeader {
		c <- true
		return c
	} else {
		cs.unleaderWaitChans = append(cs.unleaderWaitChans, c)
		return c
	}
}
