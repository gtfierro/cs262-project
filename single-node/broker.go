package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
	"sync"
)

var emptyList = []common.UUID{}

// Handles the subscriptions, updates, forwarding
type Broker struct {
	metadata *MetadataStore

	// map queries to clients
	subscriber_lock sync.RWMutex
	subscribers     map[string]clientList

	// map of producer ids to queries
	forwarding_lock sync.RWMutex
	forwarding      map[common.UUID]*queryList

	// index of producers
	producers_lock sync.RWMutex
	producers      map[common.UUID]*Producer

	// map query string to query struct
	query_lock sync.RWMutex
	queries    map[string]*Query

	// dead client notification
	killClient chan *Client
}

//TODO: config for broker?
func NewBroker(metadata *MetadataStore) *Broker {
	b := &Broker{
		metadata:    metadata,
		subscribers: make(map[string]clientList),
		forwarding:  make(map[common.UUID]*queryList),
		producers:   make(map[common.UUID]*Producer),
		queries:     make(map[string]*Query),
		killClient:  make(chan *Client),
	}

	go func(b *Broker) {
		for deadClient := range b.killClient {
			log.WithFields(log.Fields{
				"client": deadClient,
			}).Info("Removing dead client")
			b.subscriber_lock.Lock()
			for _, cl := range b.subscribers {
				cl.removeClient(deadClient)
			}
			b.subscriber_lock.Unlock()
		}
	}(b)

	return b
}

// safely adds entry to map[query][]Client map
func (b *Broker) mapQueryToClient(query string, c *Client) {
	b.subscriber_lock.Lock()
	if list, found := b.subscribers[query]; found {
		list.addClient(c)
	} else {
		// otherwise, create a new list with us in it
		b.subscribers[query] = clientList{c}
	}
	b.subscriber_lock.Unlock()
}

func (b *Broker) updateForwardingTable(query *Query) {
	b.forwarding_lock.Lock()
	query.RLock()
Loop:
	for producerID, _ := range query.MatchingProducers {
		if list, found := b.forwarding[producerID]; found {
			// check if we are already in the list
			for _, q2 := range list.queries {
				if q2 == query.Query {
					continue Loop
				}
			}
			log.WithFields(log.Fields{
				"producerID": producerID, "query": query.Query, "list": list,
			}).Debug("Adding query to existing query list")
			b.forwarding[producerID].addQuery(query.Query)
		} else {
			log.WithFields(log.Fields{
				"producerID": producerID, "query": query.Query,
			}).Debug("Adding query to NEW query list")
			b.forwarding[producerID] = &queryList{queries: []string{query.Query}}
		}
	}
	log.Debugf("forwarding table %v", b.forwarding)
	query.RUnlock()
	b.forwarding_lock.Unlock()
}

// we have the list of new and removed UUIDs for a query,
// so we update the forwarding table to match that
func (b *Broker) updateForwardingDiffs(query *Query, added, removed []common.UUID) {
	var (
		list  *queryList
		found bool
	)
	if len(removed) > 0 {
		b.forwarding_lock.Lock()
		for _, rm_uuid := range removed {
			if list, found = b.forwarding[rm_uuid]; !found {
				// no subscribers for this uuid
				continue
			}
			for _, tmp_query := range list.queries {
				if tmp_query == query.Query {
					list.removeQuery(query.Query)
					continue
				}
			}
		}
		b.forwarding_lock.Unlock()
	}
	b.updateForwardingTable(query)
}

func (b *Broker) ForwardMessage(m *common.PublishMessage) {
	var (
		matchingQueries *queryList
		found           bool
	)
	log.Debugf("forwarding msg? %v", m)
	b.forwarding_lock.RLock()
	// return if we can't find anyone to forward to
	matchingQueries, found = b.forwarding[m.UUID]
	b.forwarding_lock.RUnlock()

	if !found || matchingQueries.empty() {
		log.Debugf("no forwarding targets")
		return
	}

	var clientList []*Client
	for _, query := range matchingQueries.queries {
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

func (b *Broker) SendSubscriptionDiffs(query string, added, removed []common.UUID) {
	// if we don't do this, then empty lists show up as None
	// when we pack them
	if len(added) == 0 {
		added = emptyList
	}
	if len(removed) == 0 {
		removed = emptyList
	}
	if len(added) == 0 && len(removed) == 0 {
		return
	}
	msg := common.SubscriptionDiffMessage{"New": added, "Del": removed}
	// send to subscribers
	b.subscriber_lock.RLock()
	subscribers := b.subscribers[query]
	b.subscriber_lock.RUnlock()
	for _, sub := range subscribers {
		go sub.Send(&msg)
	}
}

// Evaluates the query and establishes the forwarding decisions.
// Returns the client
func (b *Broker) NewSubscription(querystring string, conn net.Conn) *Client {
	var (
		query *Query
		found bool
		err   error
	)
	b.query_lock.RLock()
	query, found = b.queries[querystring]
	b.query_lock.RUnlock()

	if !found {
		// parse it!
		queryAST := Parse(querystring)
		query, err = b.metadata.Query(queryAST)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err, "query": querystring,
			}).Error("Error evaluating mongo query")
		}

		// add to our map of queries
		b.query_lock.Lock()
		b.queries[querystring] = query
		b.query_lock.Unlock()

		log.WithFields(log.Fields{
			"query": querystring, "results": query.MatchingProducers,
		}).Debug("Evaluated query")
	}

	c := NewClient(querystring, &conn, b.killClient)

	// set up forwarding for all initial producers
	b.updateForwardingTable(query)
	b.mapQueryToClient(querystring, c)
	msg := common.MatchingProducersMessage(query.MatchingProducers)
	c.Send(&msg)

	return c
}

func (b *Broker) HandleProducer(msg *common.PublishMessage, dec *msgp.Reader, conn net.Conn) {
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
		log.WithFields(log.Fields{
			"message": msg, "error": err,
		}).Error("Could not save metadata")
		conn.Close()
		return
	}

	// find the producer
	b.producers_lock.RLock()
	p, found = b.producers[msg.UUID]
	b.producers_lock.RUnlock()
	if !found {
		p = NewProducer(msg.UUID, dec)
	}
	b.RemapProducer(p, msg)

	// queue first message to be sent
	b.ForwardMessage(msg)

	go func(p *Producer) {
		for p.C != nil {
			select {
			case <-p.stop:
				return
			case msg := <-p.C:
				msg.L.RLock()
				if len(msg.Metadata) > 0 {
					err = b.metadata.Save(msg)
					b.RemapProducer(p, msg)
				}
				b.ForwardMessage(msg)
				msg.L.RUnlock()
			}
		}
	}(p)
}

// 1. Firstly, have a method that given a producer (which is an implicit pointer to its
//    metadata) reevaluates all queries.

// when we receive a new producer, find related queries and update the forwarding
// tables that match. If newMetadata is nil, it will just reevaluate using current
// producer metadata across *all* queries, else it can use metadata in the provided
// Message to filter which queries to reevaluate
func (b *Broker) RemapProducer(p *Producer, newMetadata *common.PublishMessage) {
	//TODO: make a list of queries to update so that its unique THEN actually reevaluate
	//      to get added/removed lists. Once we have added/removed lists we use
	//      that to update the forwarding table.
	//      Question: Do i save the loop over to make OLD until after the forwarding table
	//      is updated? NO this should be encapsulated?
	var queriesToReevaluate = make(map[string]*Query)
	// TODO: remove this TRUE statement when we implement key-informed reevaluation
	if newMetadata == nil || true { // reevaluate ALL queries
		b.subscriber_lock.RLock()
		for querystring, _ := range b.subscribers {
			if _, found := queriesToReevaluate[querystring]; !found {
				b.query_lock.RLock()
				queriesToReevaluate[querystring] = b.queries[querystring]
				b.query_lock.RUnlock()
			}
		}
		b.subscriber_lock.RUnlock()

		for _, query := range queriesToReevaluate {
			added, removed := b.metadata.Reevaluate(query)
			log.WithFields(log.Fields{
				"query": query.Query, "added": added, "removed": removed,
			}).Info("Reevaluated query")
			b.updateForwardingDiffs(query, added, removed)
			b.SendSubscriptionDiffs(query.Query, added, removed)
		}
	}
}
