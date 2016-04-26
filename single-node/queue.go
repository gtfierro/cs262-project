package main

import (
	"github.com/gtfierro/cs262-project/common"
	"sync"
)

// This buffers messages and replays them in a channel on command
type sendableQueue struct {
	q           []common.Sendable
	c           chan common.Sendable
	replayIndex int
	sync.RWMutex
}

func newSendableQueue(size int) *sendableQueue {
	return &sendableQueue{
		q:           make([]common.Sendable, size),
		c:           make(chan common.Sendable),
		replayIndex: 0,
	}
}

func (q *sendableQueue) append(msg common.Sendable) {
	q.Lock()
	q.q = append(q.q, msg)
	q.Unlock()
}

func (q *sendableQueue) replay() chan common.Sendable {
	q.RLock()
	// for each message in the queue
	for index, msg := range q.q {
		q.replayIndex = index

	}
	q.RUnlock()
}
