package main

import (
	log "github.com/Sirupsen/logrus"
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

func newSendableQueue() *sendableQueue {
	return &sendableQueue{
		q:           []common.Sendable{},
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

// return true if there is stuff in the queue
func (q *sendableQueue) hasReplayValue() bool {
	q.RLock()
	defer q.RUnlock()
	return len(q.q) > 0
}

func (q *sendableQueue) startReplay() chan common.Sendable {
	c := make(chan common.Sendable)
	go func(c chan common.Sendable) {
		// for each message in the queue
		q.RLock()
		for index, msg := range q.q {
			select {
			case c <- msg:
			case <-q.stop:
				break
			}
			q.replayIndex = index
		}
		close(c)
		q.RUnlock()
	}(c)
	return c
}

func (q *sendableQueue) stopReplay() {
	q.stop <- true
}
