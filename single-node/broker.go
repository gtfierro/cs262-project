package main

import (
	"github.com/Sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
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
	forwarding_lock sync.RWMutex
	forwarding      map[UUID][]string

	// index of producers
	producers_lock sync.RWMutex
	producers      map[UUID]*Producer
}

//TODO: config for broker?
func NewBroker(metadata *MetadataStore) *Broker {
	return &Broker{
		metadata:    metadata,
		subscribers: make(map[string][]Client),
		forwarding:  make(map[UUID][]string),
		producers:   make(map[UUID]*Producer),
	}
}

// safely adds entry to map[query][]Client map
func (b *Broker) mapQueryToClient(query string, c *Client) {
	b.subscriber_lock.Lock()
	if list, found := b.subscribers[query]; found {
		// check if we are already in the list
		for _, c2 := range list {
			if c2 == *c {
				b.subscriber_lock.Unlock()
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
	b.forwarding_lock.Lock()
	for _, producerID := range producerIDs {
		if list, found := b.forwarding[producerID]; found {
			// check if we are already in the list
			for _, q2 := range list {
				if q2 == query {
					b.forwarding_lock.Unlock()
					return
				}
			}
			b.forwarding[producerID] = append(list, query)
		} else {
			b.forwarding[producerID] = []string{query}
		}
	}
	log.Debugf("forwarding table %v", b.forwarding)
	b.forwarding_lock.Unlock()
}

func (b *Broker) ForwardMessage(m *Message) {
	var (
		matchingQueries []string
		found           bool
	)
	log.Debugf("forwarding msg? %v", m)
	b.forwarding_lock.RLock()
	// return if we can't find anyone to forward to
	matchingQueries, found = b.forwarding[m.UUID]
	b.forwarding_lock.RUnlock()

	log.Debug("forward to: %v", matchingQueries)
	if !found || len(matchingQueries) == 0 {
		log.Debugf("no forwarding targets")
		return
	}

	var clientList []Client
	for _, query := range matchingQueries {
		b.subscriber_lock.RLock()
		clientList, found = b.subscribers[query]
		b.subscriber_lock.RUnlock()
		if !found || len(clientList) == 0 {
			log.Debugf("found no clients")
			break
		}
		for _, client := range clientList {
			go client.Send(m)
		}
	}
}

// Evaluates the query and establishes the forwarding decisions.
// Returns the client
func (b *Broker) NewSubscription(query string, conn net.Conn) *Client {
	// parse it!
	queryAST := Parse(query)
	producerIDs, err := b.metadata.Query(queryAST)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err, "query": query,
		}).Error("Error evaluating mongo query")
	}

	log.WithFields(logrus.Fields{
		"query": query, "results": producerIDs,
	}).Debug("Evaluated query")

	c := NewClient(query, &conn)
	go c.dosend()

	// set up forwarding for all initial producers
	b.producerMatchesQuery(query, producerIDs...)
	b.mapQueryToClient(query, c)
	c.Send(producerIDs)

	return c
}

func (b *Broker) HandleProducer(msg *Message, dec *msgpack.Decoder, conn net.Conn) {
	// use uuid to find old producer or create new one
	// add producer.C to a list of channels to select from
	// when we receive a message from a producer, save the
	// metadata and evaluate it, then forward it.
	// decode the incoming message
	var (
		err   error
		found bool
		p     *Producer
	)

	// save the metadata
	err = b.metadata.Save(msg)
	if err != nil {
		log.WithFields(logrus.Fields{
			"message": msg, "error": err,
		}).Error("Could not save metadata")
		conn.Close()
		return
	}

	// find the producer
	b.producers_lock.RLock()
	p, found = b.producers[msg.UUID]
	b.producers_lock.RUnlock()
	if found {
		err = b.metadata.Save(msg)
		return
	}
	p = NewProducer(msg.UUID, dec)

	// queue first message to be sent
	b.ForwardMessage(msg)

	go p.dorecv()

	go func(p *Producer) {
		for p.C != nil {
			select {
			case msg := <-p.C:
				log.Debugf("read msg %v", msg)
				err = b.metadata.Save(msg)
				b.ForwardMessage(msg)
			}
		}
	}(p)
}
