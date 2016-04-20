package main

import (
	"container/list"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"sync"
)

var emptyList = []common.UUID{}

// Handles the subscriptions, updates, forwarding routes
type ForwardingTable struct {
	metadata *common.MetadataStore

	// map of client addresses to clients
	clientLock sync.RWMutex
	clientMap  map[string]*Client

	// map of publisher ids to queries
	pubQueryLock sync.RWMutex
	pubQueryMap  map[common.UUID]*queryList

	// map of queries to publisher ids
	queryPubLock sync.RWMutex
	queryPubMap  map[*ForwardedQuery]map[*common.UUID]struct{}

	// index of publishers
	publisherLock sync.RWMutex
	publisherMap  map[common.UUID]*Publisher

	// map query string to query struct
	queryLock sync.RWMutex
	queryMap  map[string]*ForwardedQuery

	// map metadata key to list of queries involving that key
	keyLock sync.RWMutex
	keyMap  map[string]*queryList

	brokerManager BrokerManager
}

type Publisher struct {
	PublisherID     *common.UUID
	CurrentBrokerID *common.UUID
}

type Client struct {
	sync.RWMutex
	CurrentBrokerID    *common.UUID
	subscriberListElem *list.Element
	query              *ForwardedQuery
}

type ForwardedQuery struct {
	common.Query
	// NOTE: Must hold the lock on query when using this map
	CurrentBrokerIDs map[common.UUID]*list.List // BrokerID -> Client mapping
}

func (fq *ForwardedQuery) addClient(client *Client, brokerID *common.UUID) (existingBroker bool) {
	var clientList *list.List
	fq.Lock()
	defer fq.Unlock()
	if clientList, existingBroker = fq.CurrentBrokerIDs[*brokerID]; !existingBroker {
		log.WithFields(log.Fields{
			"client": client, "query": fq.QueryString, "brokerID": brokerID,
		}).Debug("Not yet forwarding to this client for this query; creating new subscribe list")
		clientList = list.New()
		fq.CurrentBrokerIDs[*brokerID] = clientList
		client.subscriberListElem = clientList.PushBack(client)
		return false
	} else {
		log.WithFields(log.Fields{
			"client": client, "query": fq.QueryString, "brokerID": brokerID,
		}).Debug("Already forwarding to this broker for this query; adding to subscribe list")
		client.subscriberListElem = clientList.PushBack(client)
		return true
	}
}

//TODO: config for broker?
func NewForwardingTable(metadata *common.MetadataStore, brokerManager BrokerManager) *ForwardingTable {
	return &ForwardingTable{
		metadata:      metadata,
		clientMap:     make(map[string]*Client),
		pubQueryMap:   make(map[common.UUID]*queryList),
		queryPubMap:   make(map[*ForwardedQuery]map[*common.UUID]struct{}),
		publisherMap:  make(map[common.UUID]*Publisher),
		queryMap:      make(map[string]*ForwardedQuery),
		keyMap:        make(map[string]*queryList),
		brokerManager: brokerManager,
	}
}

func (ft *ForwardingTable) HandleSubscription(queryStr string, clientAddr string, brokerID common.UUID) {
	// create a new client object and add to mapping
	// check if a query object already exists, if not create it and add to mapping
	//       if it does, add client to the list in the appropriate spot
	// if query didn't exist, evaluate it now
	// set up any additional routings as a result
	// respond with a BrokerSubscriptionDiffMessage
	fq := ft.getOrEvaluateQuery(queryStr)

	// Create new client; add to client mapping
	ft.clientLock.Lock()
	if c, found := ft.clientMap[clientAddr]; found {
		log.WithFields(log.Fields{
			"current_client": c, "client_addr": clientAddr,
		}).Warn("Got new client but already exists in client table for that address; replacing")
	}
	newClient := &Client{
		CurrentBrokerID: &brokerID,
		query:           fq,
	}
	ft.clientMap[clientAddr] = newClient
	ft.clientLock.Unlock()

	// Add new client to the appropriate query's list of clients
	// Creating forwarding routes as necessary
	existingBroker := fq.addClient(newClient, &brokerID)
	if !existingBroker {
		ft.addForwardingRoutes(fq, &brokerID)
	}

	// Respond with SubscriptionDiff
	respMsg := &common.BrokerSubscriptionDiffMessage{
		DelPublishers: emptyList,
		Query:         queryStr,
	}
	respMsg.FromProducerState(fq.MatchingProducers)

	ft.brokerManager.SendToBroker(brokerID, respMsg)
}

// Called anytime a publisher sends metadata (new publisher or MD update)
func (ft *ForwardingTable) HandlePublish(msg *common.PublishMessage, brokerID common.UUID) {
	if len(msg.Metadata) == 0 {
		// Ignore
		log.WithFields(log.Fields{
			"msg": msg, "brokerID": brokerID,
		}).Warn("Received a PublishMessage with no Metadata...")
		return
	}

	// save the metadata
	err := ft.metadata.Save(msg)
	if err != nil {
		log.WithFields(log.Fields{
			"message": msg, "error": err,
		}).Error("Could not save metadata")
		return // TODO any other action?
	}

	// Add publisher to mappings if necessary
	ft.publisherLock.Lock()
	publisher, found := ft.publisherMap[msg.UUID]
	if !found {
		publisher = &Publisher{PublisherID: &msg.UUID, CurrentBrokerID: &brokerID}
		ft.publisherMap[msg.UUID] = publisher
	}
	ft.publisherLock.Unlock()

	candidateQueries := ft.gatherCandidateQueries(msg)
	ft.reevaluateQueries(candidateQueries)
}

func (ft *ForwardingTable) HandleSubscriberDeath(clientAddr string, brokerID common.UUID) {
	// find out which queries the client was involved in
	// for each, if this client was the only one from brokerID, remove any forward mappings to that query
	// if this client was the only one from any broker, remove the query from our mappings
	var (
		client        *Client
		clientList    *list.List
		oldPublishers map[*common.UUID]struct{}
		found         bool
	)
	ft.clientLock.Lock()
	if client, found = ft.clientMap[clientAddr]; !found {
		log.WithFields(log.Fields{
			"clientAddr": clientAddr, "brokerID": brokerID,
		}).Warn("Attempted to handle death of a subscriber that did not exist")
		return
	}
	delete(ft.clientMap, clientAddr)
	ft.clientLock.Unlock()

	fq := client.query
	fq.Lock()
	if clientList, found = fq.CurrentBrokerIDs[brokerID]; !found {
		log.WithFields(log.Fields{
			"query": fq.QueryString, "brokerID": brokerID, "clientAddr": clientAddr,
		}).Warn("Broker attempted to terminate client but not found under that broker")
		return
	}
	clientList.Remove(client.subscriberListElem)
	remainingClientsFromBroker := clientList.Len()
	if remainingClientsFromBroker == 0 {
		delete(fq.CurrentBrokerIDs, brokerID)
	}
	remainingBrokers := len(fq.CurrentBrokerIDs)
	fq.Unlock()
	if remainingClientsFromBroker > 0 {
		return
	}

	ft.queryPubLock.Lock()
	if oldPublishers, found = ft.queryPubMap[fq]; !found {
		log.WithField("query", fq.QueryString).Info("No existing publishers found for query while removing it")
		ft.queryPubLock.Unlock()
		return // nothing to be done
	}
	if remainingBrokers == 0 {
		delete(ft.queryPubMap, fq)
	}

	cancelForwardRoutes := make(map[common.UUID]map[common.UUID][]common.UUID)
	ft.publisherLock.RLock()
	if remainingBrokers == 0 {
		ft.pubQueryLock.Lock()
		for publisherID, _ := range oldPublishers {
			if queries, found := ft.pubQueryMap[*publisherID]; found {
				queries.removeQuery(fq)
			}
		}
		ft.pubQueryLock.Unlock()
	}
	// Determine which forwarding routes must be cancelled
	for publisherID, _ := range oldPublishers {
		publisherBrokerID := ft.publisherMap[*publisherID].CurrentBrokerID
		addPublishRouteToMap(cancelForwardRoutes, publisherBrokerID, &brokerID, publisherID)
	}
	ft.publisherLock.RUnlock()
	ft.queryPubLock.Unlock()

	if len(cancelForwardRoutes) > 0 {
		ft.sendForwardingChanges(cancelForwardRoutes, fq.QueryString, true)
	}
}

func (ft *ForwardingTable) HandlePublisherDeath(publisherID common.UUID, brokerID common.UUID) {
	// simply remove publisher from local mappings and save its metadata as blank
	// send BrokerSubscriptionDiffMessage to any currently mapped
	err := ft.metadata.RemovePublisher(publisherID)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "publisherID": publisherID,
		}).Error("Error while removing publisher from metadata store")
	}

	ft.publisherLock.Lock()
	if _, found := ft.publisherMap[publisherID]; !found {
		log.WithFields(log.Fields{
			"brokerID": brokerID, "publisherID": publisherID,
		}).Warn("Attempted to terminate a non-existent publisher")
	}
	delete(ft.publisherMap, publisherID)
	ft.publisherLock.Unlock()

	ft.pubQueryLock.Lock()
	ft.queryPubLock.Lock()
	defer ft.pubQueryLock.Unlock()
	defer ft.queryPubLock.Unlock()
	if queries, found := ft.pubQueryMap[publisherID]; found {
		delete(ft.pubQueryMap, publisherID)
		for _, query := range queries.queries {
			if publishers, found := ft.queryPubMap[query]; found {
				delete(publishers, &publisherID)
			}
			query.RLock()
			delete(query.MatchingProducers, publisherID)
			for brokerID, _ := range query.CurrentBrokerIDs {
				ft.brokerManager.SendToBroker(brokerID, &common.BrokerSubscriptionDiffMessage{
					NewPublishers: emptyList,
					DelPublishers: []common.UUID{publisherID},
					Query:         query.QueryString,
				})
			}
			query.RUnlock()
		}
	}
}

func (ft *ForwardingTable) SendSubscriptionDiffs(query *ForwardedQuery, added, removed []common.UUID) {
	// if we don't do this, then empty lists show up as None
	// when we pack them
	if len(added) == 0 && len(removed) == 0 {
		return
	}
	if len(added) == 0 {
		added = emptyList
	}
	if len(removed) == 0 {
		removed = emptyList
	}
	query.RLock()
	for brokerID, _ := range query.CurrentBrokerIDs {
		msg := &common.BrokerSubscriptionDiffMessage{
			NewPublishers: added,
			DelPublishers: removed,
			Query:         query.QueryString,
		}
		ft.brokerManager.SendToBroker(brokerID, msg)
	}
	query.RUnlock()
}

// Updates the forwarding table for query, including sending routing messages to brokers
// Does not remove old publishing routes!
// If brokerID is not nil, treats brokerID as newly introduced (i.e., not yet forwarding)
// for this query
func (ft *ForwardingTable) addForwardingRoutes(fq *ForwardedQuery, newBrokerID *common.UUID) {
	ft.pubQueryLock.Lock()
	defer ft.pubQueryLock.Unlock()
	ft.queryPubLock.Lock()
	defer ft.queryPubLock.Unlock()
	ft.publisherLock.RLock()
	defer ft.publisherLock.RUnlock()

	newPublishRoutes := make(map[common.UUID]map[common.UUID][]common.UUID) // SourceBrokerID -> DestBrokerID -> PublisherID (all map to fq)

PublisherLoop:
	for publisherID, _ := range fq.MatchingProducers {
		if forwardedQueries, found := ft.pubQueryMap[publisherID]; found {
			// check if we are already in the list
			queryFound := forwardedQueries.containsQuery(fq)
			if queryFound && newBrokerID == nil {
				continue PublisherLoop // no new broker and the query exists so nothing to be done for this pub
			} else if queryFound && newBrokerID != nil {
				// query was found but there's a new broker, need to set up forward to new broker
				publishBrokerID := ft.publisherMap[publisherID].CurrentBrokerID
				addPublishRouteToMap(newPublishRoutes, publishBrokerID, newBrokerID, &publisherID)
				continue PublisherLoop
			} else { // we're an entirely new query for this publisher
				log.WithFields(log.Fields{
					"publisherID": publisherID, "query": fq.QueryString, "list": forwardedQueries,
				}).Debug("Adding query to existing query list")
				forwardedQueries.addQuery(fq)
			}
		} else {
			log.WithFields(log.Fields{
				"publisherID": publisherID, "query": fq.Query,
			}).Debug("Adding query to NEW query list")
			ft.pubQueryMap[publisherID] = &queryList{queries: []*ForwardedQuery{fq}, nonnilInArray: 1}
		}
		if qpm, found := ft.queryPubMap[fq]; found {
			qpm[&publisherID] = struct{}{}
		} else {
			ft.queryPubMap[fq] = make(map[*common.UUID]struct{})
			ft.queryPubMap[fq][&publisherID] = struct{}{}
		}
		publishBrokerID := ft.publisherMap[publisherID].CurrentBrokerID
		log.WithFields(log.Fields{
			"publisherID": publisherID, "query": fq.QueryString,
			"sourceBrokerID": publishBrokerID, "destinations": fq.CurrentBrokerIDs,
		}).Debug("Submitting a new forwarding route from publisherID to query at destinations")
		for destBroker, _ := range fq.CurrentBrokerIDs {
			addPublishRouteToMap(newPublishRoutes, publishBrokerID, &destBroker, &publisherID)
		}
	}
	//log.Debugf("forwarding table %v", ft.forwarding)
	if len(newPublishRoutes) == 0 {
		log.WithFields(log.Fields{
			"brokerID": newBrokerID, "query": fq.QueryString,
		}).Debug("Not creating any new forwarding routes while processing updateForwardingTable")
	} else {
		log.WithFields(log.Fields{
			"brokerID": newBrokerID, "query": fq.QueryString, "newRoutes": newPublishRoutes,
		}).Debug("Creating new forwarding routes while processing updateForwardingTable")
		ft.sendForwardingChanges(newPublishRoutes, fq.QueryString, false)
	}
}

func (ft *ForwardingTable) cancelForwardingRoutes(fq *ForwardedQuery, removed []common.UUID) {
	var (
		queries *queryList
		found   bool
	)
	if len(removed) == 0 {
		return
	}
	cancelPublishRoutes := make(map[common.UUID]map[common.UUID][]common.UUID) // SourceBrokerID -> DestBrokerID -> PublisherIDs
	ft.pubQueryLock.Lock()
	ft.queryPubLock.Lock()
	ft.publisherLock.RLock()
	for _, rm_uuid := range removed {
		if queries, found = ft.pubQueryMap[rm_uuid]; !found {
			continue // no subscribers for this uuid
		}
		if queries.containsQuery(fq) {
			if publishers, found := ft.queryPubMap[fq]; found {
				delete(publishers, &rm_uuid)
			} else {
				log.WithFields(log.Fields{
					"publisherID": rm_uuid, "query": fq.QueryString,
				}).Error("Found a pub->query mapping without a corresponding query->pub mapping")
			}
			queries.removeQuery(fq)
			publisherBrokerID := ft.publisherMap[rm_uuid].CurrentBrokerID
			fq.RLock()
			for destBrokerID, _ := range fq.CurrentBrokerIDs {
				addPublishRouteToMap(cancelPublishRoutes, publisherBrokerID, &destBrokerID, &rm_uuid)
			}
			fq.RUnlock()
		}
	}
	ft.publisherLock.RUnlock()
	ft.queryPubLock.Unlock()
	ft.pubQueryLock.Unlock()
	if len(cancelPublishRoutes) == 0 {
		log.WithFields(log.Fields{
			"removedUUIDs": removed, "query": fq,
		}).Debug("Not cancelling any forwarding routes while processing")
	} else {
		log.WithFields(log.Fields{
			"removedUUIDs": removed, "query": fq, "cancelRoutes": cancelPublishRoutes,
		}).Debug("Cancelling forwarding routes while processing")
		ft.sendForwardingChanges(cancelPublishRoutes, fq.QueryString, true)
	}
}

// Returns the query, and true iff this query was newly evaluated
func (ft *ForwardingTable) getOrEvaluateQuery(queryStr string) *ForwardedQuery {
	ft.queryLock.Lock()
	defer ft.queryLock.Unlock()
	if fq, found := ft.queryMap[queryStr]; !found {
		// New query; parse and add
		queryAST := common.Parse(queryStr)
		query, err := ft.metadata.Query(queryAST)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err, "query": queryStr,
			}).Error("Error evaluating mongo query")
		} else {
			log.WithFields(log.Fields{
				"query": queryStr, "results": query.MatchingProducers,
			}).Debug("Evaluated query")
		}

		fq = &ForwardedQuery{
			Query:            *query,
			CurrentBrokerIDs: make(map[common.UUID]*list.List),
		}
		ft.queryMap[queryStr] = fq

		// add the mapping of key -> query
		ft.keyLock.Lock()
		for _, key := range query.Keys {
			if potentialQueries, found := ft.keyMap[key]; found {
				potentialQueries.addQuery(fq)
			} else {
				ft.keyMap[key] = &queryList{queries: []*ForwardedQuery{fq}, nonnilInArray: 1}
			}
		}
		ft.keyLock.Unlock()

		return fq
	} else {
		return fq
	}
}

func (ft *ForwardingTable) gatherCandidateQueries(msg *common.PublishMessage) map[*ForwardedQuery]struct{} {
	candidateQueries := make(map[*ForwardedQuery]struct{})
	ft.keyLock.RLock()
	// loop through each of the metadata keys
	log.WithFields(log.Fields{
		"producerID": msg.UUID, "metadata": msg.Metadata,
	}).Debug("Gathering candidate queries based on incoming PublishMessage")
	for key, _ := range msg.Metadata {
		// pull out the list of affected queries
		queries, found := ft.keyMap[key]
		log.Debugf("For key %v found queries %v", key, queries)
		if !found { // if there are no affected queries, go on to the next key
			continue
		}
		// for each query in the found list
		for _, query := range queries.queries {
			candidateQueries[query] = struct{}{}
		}
	}
	ft.keyLock.RUnlock()
	return candidateQueries
}

func (ft *ForwardingTable) reevaluateQueries(candidateQueries map[*ForwardedQuery]struct{}) {
	for query, _ := range candidateQueries {
		added, removed := ft.metadata.Reevaluate(&query.Query)
		log.WithFields(log.Fields{
			"query": query.Query, "added": added, "removed": removed,
		}).Info("Reevaluated query")
		ft.addForwardingRoutes(query, nil)
		ft.cancelForwardingRoutes(query, removed)
		ft.SendSubscriptionDiffs(query, added, removed)
	}
}

func (ft *ForwardingTable) sendForwardingChanges(routesToChange map[common.UUID]map[common.UUID][]common.UUID,
	queryString string, requestCancel bool) {
	var msg common.Sendable
	for publishBrokerID, destBrokerMap := range routesToChange {
		for destBrokerID, publisherList := range destBrokerMap {
			if requestCancel {
				msg = &common.CancelForwardRequest{
					MessageIDStruct: common.GetMessageIDStruct(),
					PublisherList:   publisherList,
					Query:           queryString,
					BrokerInfo:      *ft.brokerManager.GetBrokerInfo(destBrokerID),
				}
			} else {
				msg = &common.ForwardRequestMessage{
					MessageIDStruct: common.GetMessageIDStruct(),
					PublisherList:   publisherList,
					Query:           queryString,
					BrokerInfo:      *ft.brokerManager.GetBrokerInfo(destBrokerID),
				}
			}
			ft.brokerManager.SendToBroker(publishBrokerID, msg)
		}
	}
}

func addPublishRouteToMap(publishRoutes map[common.UUID]map[common.UUID][]common.UUID,
	publishBrokerID, destBrokerID, publisherID *common.UUID) {
	if destBrokerMap, found := publishRoutes[*publishBrokerID]; found {
		if publisherList, found := destBrokerMap[*destBrokerID]; found {
			destBrokerMap[*destBrokerID] = append(publisherList, *publisherID)
		} else {
			destBrokerMap[*destBrokerID] = []common.UUID{*publisherID}
		}
	} else {
		destBrokerMap = make(map[common.UUID][]common.UUID)
		publishRoutes[*publishBrokerID] = destBrokerMap
		destBrokerMap[*destBrokerID] = []common.UUID{*publisherID}
	}
}
