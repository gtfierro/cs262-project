package main

import (
	"sync"
)

type LeaderService struct {
	isLeader          bool
	leaderLock        sync.RWMutex
	leaderWaitChans   []chan bool
	unleaderWaitChans []chan bool
	etcdConn          *EtcdConnection
}

func NewLeaderService(etcdConn *EtcdConnection) *LeaderService {
	cs := new(LeaderService)
	cs.isLeader = false
	cs.etcdConn = etcdConn
	cs.leaderWaitChans = []chan bool{}
	cs.unleaderWaitChans = []chan bool{}
	// TODO need to set up a watch to monitor leadership changes
	return cs
}

func (cs *LeaderService) GetLeadershipChangeRevision() int64 {
	// Get the revision of the last leadership change
	// TODO
}

func (cs *LeaderService) AttemptToBecomeInitialLeader() bool {
	// TODO do a CAS transaction checking if a leader exists
}

func (cs *LeaderService) AttemptToBecomeLeader() bool {
	// TODO do a CAS transaction checking if someone else already leadered up
}

func (cs *LeaderService) IsLeader() bool {
	cs.leaderLock.RLock()
	defer cs.leaderLock.RUnlock()
	return cs.isLeader
}

func (cs *LeaderService) SetLeader(isLeader bool) {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
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
