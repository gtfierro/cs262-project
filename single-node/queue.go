package main

import (
	"github.com/gtfierro/cs262-project/common"
	"sync"
)

// This buffers messages and replays them in a channel on command
type sendableQueue struct {
	q           []common.Sendable
	c           chan common.Sendable
	stop        chan bool
	replayIndex int
	sync.RWMutex
}

func newSendableQueue(size int) *sendableQueue {
	return &sendableQueue{
		q:           make([]common.Sendable, size),
		c:           make(chan common.Sendable),
		stop:        make(chan bool),
		replayIndex: 0,
	}
}

func (q *sendableQueue) append(msg common.Sendable) {
	log.Warnf("Queuing message %v", msg)
	q.Lock()
	q.q = append(q.q, msg)
	q.Unlock()
}

func (q *sendableQueue) startReplay() chan common.Sendable {
	go func() {
		// for each message in the queue
		q.RLock()
		for index, msg := range q.q {
			select {
			case q.c <- msg:
			case <-q.stop:
				break
			}
			q.replayIndex = index
		}
		q.RUnlock()
	}()
	return q.c
}

func (q *sendableQueue) stopReplay() {
	q.stop <- true
}
