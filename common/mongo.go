package common

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
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

var selectID = bson.M{"uuid": 1}

func NewMetadataStore(c *Config) *MetadataStore {
	var (
		m   = new(MetadataStore)
		err error
	)

	address := fmt.Sprintf("%s:%d", c.Mongo.Host, c.Mongo.Port)

	m.session, err = mgo.Dial(address)
	if err != nil {
		log.WithFields(log.Fields{
			"address": address, "error": err.Error(),
		}).Fatal("Could not connect to MongoDB")
		// exits
	}

	log.WithFields(log.Fields{
		"address": address,
	}).Info("Connected to MongoDB")

	m.db = m.session.DB(c.Mongo.Database)
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
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Could not create index on metadata.uuid")
	}

	return m
}

func (ms *MetadataStore) Save(publisherID *UUID, metadata map[string]interface{}) error {
	if publisherID == nil {
		return errors.New("PublisherID is null")
	}

	if len(metadata) == 0 { // nothing to save
		log.WithField("UUID", *publisherID).Debug("No message metadata to save")
		return nil
	}

	_, err := ms.metadata.Upsert(bson.M{"uuid": *publisherID}, bson.M{"$set": bson.M(metadata)})
	return err
}

func (ms *MetadataStore) RemovePublisher(uuid UUID) error {
	return ms.metadata.Remove(bson.M{"uuid": uuid})
}

func (ms *MetadataStore) Query(node rootNode) (*Query, error) {
	query := node.Tree.MongoQuery()
	log.WithFields(log.Fields{
		"query": query,
	}).Debug("Running mongo query")

	var q = NewQuery(node.String, node.Keys, node.Tree)

	iter := ms.metadata.Find(query).Select(selectID).Iter()
	var r map[string]string
	for iter.Next(&r) {
		uuid := UUID(r["uuid"])
		q.MatchingProducers[uuid] = ProdStateNew
	}
	err := iter.Close()
	return q, err
}

func (ms *MetadataStore) Reevaluate(query *Query) (added, removed []UUID) {
	iter := ms.metadata.Find(query.Mongo).Select(selectID).Iter()
	var r map[string]string
	// mark new UUIDs
	query.Lock()
	for iter.Next(&r) {
		uuid := UUID(r["uuid"])
		if _, found := query.MatchingProducers[uuid]; found {
			query.MatchingProducers[uuid] = ProdStateSame
		} else {
			query.MatchingProducers[uuid] = ProdStateNew
			added = append(added, uuid)
		}
	}
	// remove old ones
	for uuid, status := range query.MatchingProducers {
		if status == ProdStateOld {
			removed = append(removed, uuid)
			delete(query.MatchingProducers, uuid)
		}
	}

	// prepare statuses
	for uuid, _ := range query.MatchingProducers {
		query.MatchingProducers[uuid] = ProdStateOld
	}
	query.Unlock()
	return
}

// Obviously use carefully - in place only for testing!
func (ms *MetadataStore) DropDatabase() {
	ms.db.DropDatabase()
}

type Query struct {
	QueryString       string
	Keys              []string
	MatchingProducers map[UUID]ProducerState
	Mongo             bson.M
	sync.RWMutex
}

func NewQuery(query string, keys []string, root Node) *Query {
	return &Query{
		QueryString:       query,
		Keys:              keys,
		MatchingProducers: make(map[UUID]ProducerState),
		Mongo:             root.MongoQuery(),
	}
}

// updates internal list of qualified streams. Returns the lists of added and removed UUIDs
func (q *Query) changeUUIDs(uuids []UUID) (added, removed []UUID) {
	// mark the UUIDs that are new
	q.Lock()
	for _, uuid := range uuids {
		if _, found := q.MatchingProducers[uuid]; found {
			q.MatchingProducers[uuid] = ProdStateSame
		} else {
			q.MatchingProducers[uuid] = ProdStateNew
			added = append(added, uuid)
		}
	}

	// remove the old ones
	for uuid, status := range q.MatchingProducers {
		if status == ProdStateOld {
			removed = append(removed, uuid)
			delete(q.MatchingProducers, uuid)
		}
	}

	for uuid, _ := range q.MatchingProducers {
		q.MatchingProducers[uuid] = ProdStateOld
	}
	q.Unlock()
	return
}
