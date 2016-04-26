package main

import (
	"github.com/gtfierro/cs262-project/common"
	"sync"
	"sync/atomic"
)

type outstandingMessage struct {
	Cond *sync.Cond
	V    atomic.Value
}

func newOutstandingMessage() *outstandingMessage {
	return &outstandingMessage{Cond: &sync.Cond{L: &sync.Mutex{}}}
}

type outstandingManager struct {
	sync.RWMutex
	outstanding map[common.MessageIDType]*outstandingMessage
}

func newOutstandingManager() *outstandingManager {
	return &outstandingManager{outstanding: make(map[common.MessageIDType]*outstandingMessage)}
}

func (om *outstandingManager) WaitForMessage(id common.MessageIDType) (common.Message, error) {
	// first, check if the message has already arrived
	om.Lock()
	if o, found := om.outstanding[id]; found {
		// if it has, delete the reference and return the message
		delete(om.outstanding, id)
		om.Unlock()
		return o.V.Load().(common.Message), nil
	}
	// if it hasn't, create a new placeholder
	o := newOutstandingMessage()
	om.outstanding[id] = o
	om.Unlock()

	// here we wait on the condition being triggered, which is done
	// when the message finally arrives
	o.Cond.L.Lock()
	for o.V.Load() == nil {
		o.Cond.Wait()
	}
	o.Cond.L.Unlock()

	// delete the reference from the table
	om.Lock()
	delete(om.outstanding, id)
	om.Unlock()

	// finally, return the message
	return o.V.Load().(common.Message), nil
}

func (om *outstandingManager) GotMessage(m common.Message) {
	om.Lock()
	if o, found := om.outstanding[m.GetID()]; found {
		o.V.Store(m)
		delete(om.outstanding, m.GetID())
		o.Cond.Broadcast()
	} else {
		o := newOutstandingMessage()
		o.V.Store(m)
		om.outstanding[m.GetID()] = o
	}
	om.Unlock()
}
