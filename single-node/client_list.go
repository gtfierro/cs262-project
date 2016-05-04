package main

import (
	"github.com/gtfierro/cs262-project/common"
	"sync"
)

type clientList struct {
	List []*Client
	sync.RWMutex
}

func (sl *clientList) addClient(sub *Client) {
	sl.Lock()
	defer sl.Unlock()
	for _, oldSub := range sl.List {
		if oldSub != nil && oldSub.ID == sub.ID {
			return
		}
	}
	for i, oldSub := range sl.List {
		if oldSub == nil {
			sl.List[i] = sub
			return
		}
	}

	sl.List = append(sl.List, sub)
}

// TODO: something is funky here. if a client disconnects/reconnects very quickly,
// we can get a race condition where they disconnect, the broker fails to forward
// to them, then at the same time the new client comes in, then the broker "deletes"
// the old client?
func (sl *clientList) removeClient(sub *Client) {
	sl.Lock()
	defer sl.Unlock()
	for i, oldSub := range sl.List {
		if oldSub == sub {
			sl.List[i] = sl.List[len(sl.List)-1]
			sl.List[len(sl.List)-1] = nil
			sl.List = sl.List[:len(sl.List)-1]
			// above: this is supposedly a better delete method for slices because
			// it avoids leaving references to "deleted" objects at the end of the list
			// below is the "normal" one
			//*sl = append((*sl)[:i], (*sl)[i+1:]...)
			return
		}
	}
}

func (sl *clientList) sendToList(msg common.Sendable) {
	sl.RLock()
	for _, c := range sl.List {
		if c == nil {
			continue
		}
		c.Send(msg)
	}
	sl.RUnlock()
}

// removes the client from the list that matches the UUID
func (sl *clientList) removeByUUID(uuid common.UUID) {
	sl.Lock()
	for i, oldSub := range sl.List {
		if oldSub.ID == uuid {
			sl.List[i] = sl.List[len(sl.List)-1]
			sl.List[len(sl.List)-1] = nil
			sl.List = sl.List[:len(sl.List)-1]
		}
	}
	sl.Unlock()
}

type brokerClientList struct {
	List []*BrokerClient
	sync.RWMutex
}

func (sl *brokerClientList) addBrokerClient(sub *BrokerClient) {
	sl.Lock()
	defer sl.Unlock()
	for _, oldSub := range sl.List {
		if oldSub != nil && oldSub.ID == sub.ID {
			return
		}
	}
	for i, oldSub := range sl.List {
		if oldSub == nil {
			sl.List[i] = sub
			return
		}
	}

	sl.List = append(sl.List, sub)
}

// TODO: something is funky here. if a client disconnects/reconnects very quickly,
// we can get a race condition where they disconnect, the broker fails to forward
// to them, then at the same time the new client comes in, then the broker "deletes"
// the old client?
func (sl *brokerClientList) removeBrokerClient(sub *BrokerClient) {
	sl.Lock()
	defer sl.Unlock()
	for i, oldSub := range sl.List {
		if oldSub == sub {
			sl.List[i] = sl.List[len(sl.List)-1]
			sl.List[len(sl.List)-1] = nil
			sl.List = sl.List[:len(sl.List)-1]
			// above: this is supposedly a better delete method for slices because
			// it avoids leaving references to "deleted" objects at the end of the list
			// below is the "normal" one
			//*sl = append((*sl)[:i], (*sl)[i+1:]...)
			return
		}
	}
}

func (sl *brokerClientList) sendToList(msg common.Sendable) {
	sl.RLock()
	for _, c := range sl.List {
		if c == nil {
			continue
		}
		c.Send(msg)
	}
	sl.RUnlock()
}

// removes the client from the list that matches the UUID
func (sl *brokerClientList) removeByUUID(uuid common.UUID) {
	sl.Lock()
	for i, oldSub := range sl.List {
		if oldSub.ID == uuid {
			sl.List[i] = sl.List[len(sl.List)-1]
			sl.List[len(sl.List)-1] = nil
			sl.List = sl.List[:len(sl.List)-1]
		}
	}
	sl.Unlock()
}
