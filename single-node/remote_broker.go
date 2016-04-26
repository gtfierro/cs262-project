package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
	"sync"
)

type RemoteBroker struct {
	metadata    *common.MetadataStore
	coordinator *Coordinator

	// map queries to clients
	subscriber_lock sync.RWMutex
	subscribers     map[string]clientList

	// map of producer ids to queries
	forwarding_lock sync.RWMutex
	forwarding      map[common.UUID]*queryList

	// index of producers
	producers_lock sync.RWMutex
	producers      map[common.UUID]*Producer

	// dead client notification
	killClient chan *Client
}

func NewRemoteBroker(metadata *common.MetadataStore, coordinator *Coordinator) *RemoteBroker {
	b := &RemoteBroker{
		metadata:    metadata,
		coordinator: coordinator,
		subscribers: make(map[string]clientList),
		forwarding:  make(map[common.UUID]*queryList),
		producers:   make(map[common.UUID]*Producer),
		killClient:  make(chan *Client),
	}

	go func(b *RemoteBroker) {
		for deadClient := range b.killClient {
			log.WithFields(log.Fields{
				"client": deadClient,
			}).Info("Removing dead client")
			b.subscriber_lock.Lock()
			// TODO this seems inefficient - only one query will actually map
			// to this client, so why not just store that pointer in the client?
			for _, cl := range b.subscribers {
				cl.removeClient(deadClient)
			}
			b.subscriber_lock.Unlock()
		}
	}(b)

	return b
}

func (b *RemoteBroker) HandleProducer(msg *common.PublishMessage, dec *msgp.Reader, conn net.Conn) {
	var (
		found bool
		p     *Producer
	)
	// forward the new message to the coordinator
	diff := b.coordinator.forwardPublish(msg)
	log.Debugf("got diff %v", diff)
	// now wait for the response from the coordinator to tell us what changed
	// The PROBLEM here is that a) we only have a single connection to the broker and we need
	// to multiplex many publishers among it. How do I know that I have received the subscription
	// diff that corresponds to my message?

	b.producers_lock.RLock()
	p, found = b.producers[msg.UUID]
	b.producers_lock.RUnlock()
	if !found {
		p = NewProducer(msg.UUID, dec)
	}

	b.ForwardMessage(msg)

	go func(p *Producer) {
		for p.C != nil {
			select {
			case <-p.stop:
				//TODO: implement
				//b.cleanupProducer(p)
				return
			case msg := <-p.C:
				msg.L.RLock()
				// if doing local reevaluation then we save metadata
				// and handle differences here
				diff := b.coordinator.forwardPublish(msg)
				log.Debugf("GOT DIFF %v", diff)
				if len(msg.Metadata) > 0 {
					//err = b.metadata.Save(msg)
					//b.RemapProducer(p, msg)
				}
				//b.ForwardMessage(msg)
				msg.L.RUnlock()
			}
		}
	}(p)
}

func (b *RemoteBroker) NewSubscription(query string, clientID common.UUID, conn net.Conn) *Client {
	// create the local client
	c := NewClient(query, &conn, b.killClient)
	// map the local client to the query
	b.mapQueryToClient(query, c)
	// forward our subscription information to the coordinator
	b.coordinator.forwardSubscription(query, clientID, conn)
	return c
}

func (b *RemoteBroker) ForwardMessage(msg *common.PublishMessage) {
}

// safely adds entry to map[query][]Client map
func (b *RemoteBroker) mapQueryToClient(query string, c *Client) {
	b.subscriber_lock.Lock()
	if list, found := b.subscribers[query]; found {
		list.addClient(c)
		b.subscribers[query] = list
	} else {
		// otherwise, create a new list with us in it
		b.subscribers[query] = clientList{List: []*Client{c}}
	}
	b.subscriber_lock.Unlock()
}
