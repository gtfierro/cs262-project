package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"sync/atomic"
)

type CommService struct {
	isLeader    *int32 // 0 for false, 1 for true - int32 so we can use atomics
	etcdManager *EtcdManager
}

func NewCommService(isLeader bool, etcdManager *EtcdManager) *CommService {
	cs := new(CommService)
	var leaderVal int32
	if isLeader {
		leaderVal = 1
	} else {
		leaderVal = 0
	}
	cs.isLeader = &leaderVal
	cs.etcdManager = etcdManager
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
	val := atomic.LoadInt32(cs.isLeader)
	return val == 1
}

func (cs *CommService) SetLeader(isLeader bool) {
	var val int32
	if isLeader {
		val = 1
	} else {
		val = 0
	}
	atomic.StoreInt32(cs.isLeader, val)
}
