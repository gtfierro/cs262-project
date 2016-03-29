package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"cs262-project/common"
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

func NewMetadataStore(c *common.Config) *MetadataStore {
	var (
		m   = new(MetadataStore)
		err error
	)

	address := fmt.Sprintf("%s:%d", *c.Mongo.Host, *c.Mongo.Port)

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
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Could not create index on metadata.uuid")
	}

	return m
}

func (ms *MetadataStore) Save(msg *common.PublishMessage) error {
	if msg == nil {
		return errors.New("Message is null")
	}

	if len(msg.Metadata) == 0 { // nothing to save
		log.WithFields(log.Fields{
			"UUID": msg.UUID, "value": msg.Value,
		}).Debug("No message metadata to save")
		return nil
	}

	_, err := ms.metadata.Upsert(bson.M{"uuid": msg.UUID}, bson.M{"$set": bson.M(msg.Metadata)})
	return err
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
		uuid := common.UUID(r["uuid"])
		q.MatchingProducers[uuid] = common.ProdStateNew
	}
	err := iter.Close()
	return q, err
}

func (ms *MetadataStore) Reevaluate(query *Query) (added, removed []common.UUID) {
	//TODO: cache the mongo query generation
	iter := ms.metadata.Find(query.Mongo).Select(selectID).Iter()
	var r map[string]string
	// mark new UUIDs
	query.Lock()
	for iter.Next(&r) {
		uuid := common.UUID(r["uuid"])
		if _, found := query.MatchingProducers[uuid]; found {
			query.MatchingProducers[uuid] = common.ProdStateSame
		} else {
			query.MatchingProducers[uuid] = common.ProdStateNew
			added = append(added, uuid)
		}
	}
	// remove old ones
	for uuid, status := range query.MatchingProducers {
		if status == common.ProdStateOld {
			removed = append(removed, uuid)
			delete(query.MatchingProducers, uuid)
		}
	}

	// prepare statuses
	for uuid, _ := range query.MatchingProducers {
		query.MatchingProducers[uuid] = common.ProdStateOld
	}
	query.Unlock()
	return
}

// TODO: change Query to return a Query struct, modeled after
// giles2 cqbs.go
type Query struct {
	Query             string
	Keys              []string
	MatchingProducers map[common.UUID]common.ProducerState
	Clients           *clientList
	Mongo             bson.M
	sync.RWMutex
}

func NewQuery(query string, keys []string, root Node) *Query {
	return &Query{
		Query:             query,
		Keys:              keys,
		MatchingProducers: make(map[common.UUID]common.ProducerState),
		Clients:           new(clientList),
		Mongo:             root.MongoQuery(),
	}
}

// updates internal list of qualified streams. Returns the lists of added and removed UUIDs
func (q *Query) changeUUIDs(uuids []common.UUID) (added, removed []common.UUID) {
	// mark the UUIDs that are new
	q.Lock()
	for _, uuid := range uuids {
		if _, found := q.MatchingProducers[uuid]; found {
			q.MatchingProducers[uuid] = common.ProdStateSame
		} else {
			q.MatchingProducers[uuid] = common.ProdStateNew
			added = append(added, uuid)
		}
	}

	// remove the old ones
	for uuid, status := range q.MatchingProducers {
		if status == common.ProdStateOld {
			removed = append(removed, uuid)
			delete(q.MatchingProducers, uuid)
		}
	}

	for uuid, _ := range q.MatchingProducers {
		q.MatchingProducers[uuid] = common.ProdStateOld
	}
	q.Unlock()
	return
}

// TODO: more efficient implementation?
type clientList []*Client

func (sl *clientList) addClient(sub *Client) {
	for _, oldSub := range *sl {
		if oldSub == sub {
			return
		}
	}

	*sl = append(*sl, sub)
}

func (sl *clientList) removeClient(sub *Client) {
	for i, oldSub := range *sl {
		if oldSub == sub {
			*sl = append((*sl)[:i], (*sl)[i+1:]...)
			return
		}
	}
}

type queryList []string

func (ql *queryList) addQuery(q string) {
	for _, oldSub := range *ql {
		if oldSub == q {
			return
		}
	}

	*ql = append(*ql, q)
}

func (ql *queryList) removeQuery(q string) {
	for i, oldSub := range *ql {
		if oldSub == q {
			*ql = append((*ql)[:i], (*ql)[i+1:]...)
			return
		}
	}
}
