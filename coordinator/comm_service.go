package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"sync"
)

type CommService struct {
	isLeader          bool
	leaderLock        sync.RWMutex
	leaderWaitChans   []chan bool
	unleaderWaitChans []chan bool
	etcdManager       *EtcdManager
}

func NewCommService(isLeader bool, etcdManager *EtcdManager) *CommService {
	cs := new(CommService)
	cs.isLeader = isLeader
	cs.etcdManager = etcdManager
	cs.leaderWaitChans = []chan bool{}
	cs.unleaderWaitChans = []chan bool{}
	return cs
}

func (cs *CommService) SendAndFlush(msg common.Sendable, writer msgp.Writer) error {
	if cs.IsLeader() {
		err := msg.Encode(writer)
		if err != nil {
			return err
		}
		err = writer.Flush()
		return err
	} else {
		// TODO
		return nil
	}
}

func (cs *CommService) ReceiveMessage(reader msgp.Reader) (common.Sendable, error) {
	if cs.IsLeader() {
		msg, err := common.MessageFromDecoderMsgp(reader)
		return msg, err
	} else {
		return nil, nil // TODO
	}
}

func (cs *CommService) IsLeader() bool {
	cs.leaderLock.RLock()
	defer cs.leaderLock.RUnlock()
	return cs.isLeader
}

func (cs *CommService) SetLeader(isLeader bool) {
	cs.leaderLock.Lock()
	defer cs.leaderLock.Unlock()
	cs.isLeader = isLeader
	if isLeader {
		for _, c := range cs.leaderWaitChans {
			c <- true
		}
		cs.leaderWaitChans = []chan bool{}
	} else {
		for _, c := range cs.unleaderWaitChans {
			c <- true
		}
		cs.unleaderWaitChans = []chan bool{}
	}
}

// Returns a channel which will be sent to when this is leader
func (cs *CommService) WaitForLeadership() chan bool {
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

// Returns a channel which will be sent to when this is nonleader
func (cs *CommService) WaitForNonleadership() chan bool {
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
