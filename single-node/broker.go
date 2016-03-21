package main

import (
	"github.com/Sirupsen/logrus"
	"net"
	"sync"
)

// Handles the subscriptions, updates, forwarding
type Broker struct {
	metadata *MetadataStore

	// map queries to clients
	subscriber_lock sync.RWMutex
	subscribers     map[string][]Client

	// map of producer ids to queries
	producers_lock sync.RWMutex
	producers      map[UUID][]string
}

//TODO: config for broker?
func NewBroker(metadata *MetadataStore) *Broker {
	return &Broker{
		metadata:    metadata,
		subscribers: make(map[string][]Client),
		producers:   make(map[UUID][]string),
	}
}

// safely adds entry to map[query][]Client map
func (b *Broker) mapQueryToClient(query string, c *Client) {
	b.subscriber_lock.Lock()
	if list, found := b.subscribers[query]; found {
		// check if we are already in the list
		for _, c2 := range list {
			if c2 == *c {
				return
			}
		}
		b.subscribers[query] = append(list, *c)
	} else {
		b.subscribers[query] = []Client{*c}
	}
	b.subscriber_lock.Unlock()
}

// for any producer ID, we want to be able to quickly find which queries
// it maps to. From the query, we can easily find the list of clients
func (b *Broker) producerMatchesQuery(query string, producerIDs ...UUID) {
	b.producers_lock.Lock()
	for _, producerID := range producerIDs {
		if list, found := b.producers[producerID]; found {
			// check if we are already in the list
			for _, q2 := range list {
				if q2 == query {
					return
				}
			}
			b.producers[producerID] = append(list, query)
		} else {
			b.producers[producerID] = []string{query}
		}
	}
	b.producers_lock.Unlock()
}

func (b *Broker) ForwardMessage(m *Message) {
	var (
		matchingQueries []string
		found           bool
	)
	b.producers_lock.RLock()
	// return if we can't find anyone to forward to
	matchingQueries, found = b.producers[m.UUID]
	b.producers_lock.RUnlock()

	if !found || len(matchingQueries) == 0 {
		b.producers_lock.RUnlock()
		return
	}

	var clientList []Client
	for _, query := range matchingQueries {
		b.subscriber_lock.RLock()
		clientList, found = b.subscribers[query]
		b.subscriber_lock.RUnlock()
		if !found || len(clientList) == 0 {
			break
		}
		for _, client := range clientList {
			client.Send(m)
		}
	}
}

// Evaluates the query and establishes the forwarding decisions.
// Returns the client
func (b *Broker) NewSubscription(query string, conn net.Conn) *Client {
	// parse it!
	node := Parse(query)
	producerIDs, err := b.metadata.Query(node)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err, "query": query,
		}).Error("Error evaluating mongo query")
	}

	log.WithFields(logrus.Fields{
		"query": query, "results": producerIDs,
	}).Debug("Evaluated query")
	//TODO: put this into a client struct, evaluate it and return initial results, and
	// 		establish which publishers are going to be forwarding
	c := NewClient(query, &conn)
	go c.dosend()

	// set up forwarding for all initial producers
	b.producerMatchesQuery(query, producerIDs...)
	b.mapQueryToClient(query, c)

	return c
}
