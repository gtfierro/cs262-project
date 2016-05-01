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

	// background loop to wait for clients to die so that we can
	// garbage collect them from our internal control structures.
	go func(b *RemoteBroker) {
		for deadClient := range b.killClient {
			log.WithFields(log.Fields{
				"client": deadClient,
			}).Info("Removing dead client")
			b.subscriber_lock.Lock()
			cl := b.subscribers[deadClient.query]
			(&cl).removeClient(deadClient)
			b.subscribers[deadClient.query] = cl
			b.subscriber_lock.Unlock()
			b.coordinator.terminateClient(deadClient)
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
	b.coordinator.forwardPublish(msg)
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
				b.cleanupProducer(p)
				return
			case msg := <-p.C:
				msg.L.RLock()
				// if we have metadata, then subscriptions could change
				if len(msg.Metadata) > 0 {
					// this blocks until the acknowledgement has been sent
					b.coordinator.forwardPublish(msg)
				}
				log.Debugf("Forwarding msg to clients: %v", msg)
				b.ForwardMessage(msg)
				msg.L.RUnlock()
			}
		}
	}(p)
}

// Removes the producer UUID from the forwarding and producers maps
func (b *RemoteBroker) cleanupProducer(p *Producer) {
	b.producers_lock.Lock()
	b.forwarding_lock.Lock()
	delete(b.producers, p.ID)
	delete(b.forwarding, p.ID)
	b.forwarding_lock.Unlock()
	b.producers_lock.Unlock()
	b.coordinator.terminatePublisher(p)
}

func (b *RemoteBroker) NewSubscription(query string, clientID common.UUID, conn net.Conn) *Client {
	// create the local client
	c := NewClient(query, clientID, &conn, b.killClient)
	// map the local client to the query
	b.mapQueryToClient(query, c)
	// forward our subscription information to the coordinator
	b.coordinator.forwardSubscription(query, clientID, conn)
	return c
}

func (b *RemoteBroker) ForwardMessage(msg *common.PublishMessage) {
	var (
		matchingQueries *queryList
		found           bool
	)
	b.forwarding_lock.RLock()
	// return if we can't find anyone to forward to
	matchingQueries, found = b.forwarding[msg.UUID]
	b.forwarding_lock.RUnlock()

	if !found || matchingQueries.empty() {
		log.Debugf("no forwarding targets")
		return
	}

	var deliveries clientList
	for _, query := range matchingQueries.queries {
		b.subscriber_lock.RLock()
		deliveries, found = b.subscribers[query]
		b.subscriber_lock.RUnlock()
		if !found || len(deliveries.List) == 0 {
			log.Debugf("found no clients")
			break
		}
		log.Debugf("found clients %v", deliveries)
		deliveries.sendToList(msg)
	}
}

// This method is called to add forwarding entries for remote brokers.
// The received structure has this form:
//	// list of publishers whose messages should be forwarded
//	PublisherList []UUID
//	// the destination broker
//	BrokerInfo
//	// the query string which defines this forward request
//	Query string
func (b *RemoteBroker) AddForwardingEntries(msg *common.ForwardRequestMessage) {
	//TODO: figure out what to do with the error here
	client, err := ClientFromBrokerString(b.coordinator.brokerID, msg.Query, msg.ClientBrokerAddr, b.killClient)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "address": msg.ClientBrokerAddr,
		}).Error("Could not create broker client")
		return
	}
	// for this list of publishers, setup forwarding to the client
	b.remapPublishers(msg.PublisherList, msg.Query)
	b.mapQueryToClient(msg.Query, client)
}

// The coordinator may send us a CancelForwardRequest which contains:
// 	 MessageIDStruct
// 	 // list of publishers whose messages should be cancelled
// 	 PublisherList []UUID
// 	 // the query that has been cancelled
// 	 Query string
// 	 // Necessary so you know who you need to stop forwarding to, since you
// 	 // may be forwarding to multiple brokers for the same query
// 	 BrokerInfo
// We want to find the client that corresponds to the broker that we want to cancel
// forwarding to and kill it.
//TODO implement this
func (b *RemoteBroker) RemoveForwardingEntry(msg *common.CancelForwardRequest) {

}

//	NewPublishers []UUID
//	DelPublishers []UUID
//	Query         string
func (b *RemoteBroker) UpdateForwardingEntries(msg *common.BrokerSubscriptionDiffMessage) {
	var (
		list  *queryList
		found bool
	)
	// TODO this will only activate when a query no longer applies to a publisher due to
	// metadata changes - what about when a query should no longer exist due to no more clients?
	b.forwarding_lock.Lock()
	if len(msg.DelPublishers) > 0 {
		for _, rm_uuid := range msg.DelPublishers {
			if list, found = b.forwarding[rm_uuid]; !found {
				// no subscribers for this uuid
				continue
			}
			for _, tmp_query := range list.queries {
				if tmp_query == msg.Query {
					list.removeQuery(msg.Query)
					continue
				}
			}
		}
	}
	b.forwarding_lock.Unlock()
	b.remapPublishers(msg.NewPublishers, msg.Query)
}

func (b *RemoteBroker) remapPublishers(new []common.UUID, query string) {
	var (
		list  *queryList
		found bool
	)
	if len(new) == 0 {
		return
	}
	b.forwarding_lock.Lock()
	for _, add_uuid := range new {
		if list, found = b.forwarding[add_uuid]; !found {
			b.forwarding[add_uuid] = &queryList{queries: []string{query}}
			log.WithFields(log.Fields{
				"producerID": add_uuid, "query": query, "list": list,
			}).Debug("Adding query to new query list")
		} else {
			list.addQuery(query)
			b.forwarding[add_uuid] = list
			log.WithFields(log.Fields{
				"producerID": add_uuid, "query": query, "list": list,
			}).Debug("Adding query to existing query list")
		}
	}
	b.forwarding_lock.Unlock()
}

func (b *RemoteBroker) ForwardSubscriptionDiffs(msg *common.BrokerSubscriptionDiffMessage) {
	if len(msg.NewPublishers) == 0 && len(msg.DelPublishers) == 0 {
		return
	}
	sdm := common.SubscriptionDiffMessage{"New": msg.NewPublishers, "Del": msg.DelPublishers}
	b.subscriber_lock.RLock()
	subscribers := b.subscribers[msg.Query]
	b.subscriber_lock.RUnlock()
	go subscribers.sendToList(&sdm)
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

func (b *RemoteBroker) killClients(m *common.ClientTerminationRequest) {
	b.subscriber_lock.Lock()
	for _, clients := range b.subscribers {
		for _, uuid := range m.ClientIDs {
			clients.removeByUUID(uuid)
		}
	}
	b.subscriber_lock.Unlock()
}
