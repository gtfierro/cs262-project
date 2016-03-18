package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// This struct handles all communication with the Mongo database that provides
// metadata storage and query capabilities. The "schema" of the metadata
// collection is simple. Each document is flat (just k-v pairs) and corresponds
// to a producer. The producer UUIDv4 is stored in the primary key _id field,
// and the rest of the document is just the key/value pairs of metadata
type MetadataStore struct {
	session  *mgo.Session
	db       *mgo.Database
	metadata *mgo.Collection
}

func NewMetadataStore(c *Config) *MetadataStore {
	var (
		m   = new(MetadataStore)
		err error
	)

	address := fmt.Sprintf("%s:%d", *c.Mongo.Host, *c.Mongo.Port)

	m.session, err = mgo.Dial(address)
	if err != nil {
		log.WithFields(logrus.Fields{
			"address": address, "error": err.Error(),
		}).Fatal("Could not connect to MongoDB")
		// exits
	}

	log.WithFields(logrus.Fields{
		"address": address,
	}).Info("Connected to MongoDB")

	m.db = m.session.DB("broker")
	m.metadata = m.db.C("metadata")

	index := mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	err = m.metadata.EnsureIndex(index)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("Could not create index on metadata.uuid")
	}

	return m
}

func (ms *MetadataStore) Save(msg *Message) error {
	if msg == nil {
		return errors.New("Message is null")
	}

	if len(msg.Metadata) == 0 { // nothing to save
		return nil
	}

	_, err := ms.metadata.UpsertId(msg.UUID, bson.M{"$set": bson.M(msg.Metadata)})
	return err
}
