package main

import (
	"github.com/Sirupsen/logrus"
)

// Handles the subscriptions, updates, forwarding
type Broker struct {
	metadata *MetadataStore
}

//TODO: config for broker?
func NewBroker(metadata *MetadataStore) {
	return &Broker{metadata: metadata}
}

// Evaluates the query and establishes the forwarding decisions.
// Returns the client
func (b *Broker) NewSubscription(query string, conn *net.Conn) *Client {
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
}
