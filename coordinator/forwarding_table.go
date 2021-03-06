package main

import (
	"container/list"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"sync"
)

var emptyList = []common.UUID{}

const UNASSIGNED = common.UUID("__BROKER_ID_UNASSIGNED__")

// Handles the subscriptions, updates, forwarding routes
type ForwardingTable struct {
	metadata      *common.MetadataStore
	brokerManager BrokerManager
	etcdManager   EtcdManager

	// map of client addresses to clients
	clientLock sync.RWMutex
	clientMap  map[common.UUID]*Client

	// map of publisher ids to queries
	pubQueryLock sync.Mutex
	pubQueryMap  map[common.UUID]*queryList

	// map of queries to publisher ids
	queryPubLock sync.Mutex
	queryPubMap  map[*ForwardedQuery]map[common.UUID]struct{}

	// index of publishers
	publisherLock sync.RWMutex
	publisherMap  map[common.UUID]*Publisher

	// map query string to query struct
	queryLock sync.RWMutex
	queryMap  map[string]*ForwardedQuery

	// map metadata key to list of queries involving that key
	keyLock sync.RWMutex
	keyMap  map[string]*queryList

	brokerDeathChan    chan *common.UUID
	brokerLiveChan     chan *common.UUID
	brokerReassignChan chan *BrokerReassignment
}

type BrokerReassignment struct {
	IsPublisher  bool
	UUID         *common.UUID
	HomeBrokerID *common.UUID
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
			"clientID": client.ClientID, "query": client.QueryString, "brokerID": *brokerID,
		}).Debug("Not yet forwarding to this broker for this query; creating new subscribe list")
		clientList = list.New()
		fq.CurrentBrokerIDs[*brokerID] = clientList
		client.subscriberListElem = clientList.PushBack(client)
		return false
	} else {
		log.WithFields(log.Fields{
			"clientID": client.ClientID, "query": client.QueryString, "brokerID": brokerID,
		}).Debug("Already forwarding to this broker for this query; adding to subscribe list")
		if client.subscriberListElem != nil {
			// we already knew about this client and it should be in the list so remove it if it is
			clientList.Remove(client.subscriberListElem)
		}
		client.subscriberListElem = clientList.PushBack(client)
		return true
	}
}

func NewForwardingTable(metadata *common.MetadataStore, brokerMgr BrokerManager, etcdMgr EtcdManager,
	brokerDeathChan, brokerLiveChan chan *common.UUID, brokerReassignChan chan *BrokerReassignment) *ForwardingTable {
	return &ForwardingTable{
		metadata:           metadata,
		brokerManager:      brokerMgr,
		etcdManager:        etcdMgr,
		clientMap:          make(map[common.UUID]*Client),
		pubQueryMap:        make(map[common.UUID]*queryList),
		queryPubMap:        make(map[*ForwardedQuery]map[common.UUID]struct{}),
		publisherMap:       make(map[common.UUID]*Publisher),
		queryMap:           make(map[string]*ForwardedQuery),
		keyMap:             make(map[string]*queryList),
		brokerDeathChan:    brokerDeathChan,
		brokerLiveChan:     brokerLiveChan,
		brokerReassignChan: brokerReassignChan,
	}
}

func (ft *ForwardingTable) RebuildFromEtcd(upToRev int64) (err error) {
	err = ft.etcdManager.IterateOverAllEntities(ClientEntity, upToRev, ft.rebuildClientState)
	if err != nil {
		log.WithField("error", err).Fatal("Error while iterating over clients when rebuilding")
		return
	}
	err = ft.etcdManager.IterateOverAllEntities(PublisherEntity, upToRev, ft.rebuildPublisherState)
	if err != nil {
		log.WithField("error", err).Fatal("Error while iterating over publishers when rebuilding")
		return
	}
	return
}

func (ft *ForwardingTable) rebuildClientState(entity EtcdSerializable) {
	client := entity.(*Client)
	ft.HandleSubscription(client.QueryString, client.ClientID, client.CurrentBrokerID, &client.HomeBrokerID)
}

func (ft *ForwardingTable) rebuildPublisherState(entity EtcdSerializable) {
	publisher := entity.(*Publisher)
	ft.HandlePublish(publisher.PublisherID, publisher.Metadata, publisher.CurrentBrokerID, &publisher.HomeBrokerID)
}

func (ft *ForwardingTable) monitorInboundChannels() {
	for {
		select {
		case deadID := <-ft.brokerDeathChan:
			ft.HandleBrokerDeath(deadID)
		case liveID := <-ft.brokerLiveChan:
			ft.HandleBrokerLife(liveID)
		case brokerReassign := <-ft.brokerReassignChan:
			ft.HandleBrokerReassignment(brokerReassign)
		}
	}
}

func (ft *ForwardingTable) HandleBrokerDeath(deadID *common.UUID) {
	// when a broker dies, mark its publishers/clients as currently inactive
	log.WithField("brokerID", *deadID).Info("Forwarding table handling the death of broker")
	// TODO don't think we want this; we want to cancel the routes once the client *reconnects*
	//ft.clientLock.Lock()
	//for _, client := range ft.clientMap {
	//	if client.CurrentBrokerID == *deadID {
	//		ft.CancelSubscriberForwarding(client) // clear out forwarding routes for this client
	//		client.CurrentBrokerID = UNASSIGNED   // mark as currently homeless
	//		ft.etcdManager.UpdateEntity(client)
	//	}
	//}
	//ft.clientLock.Unlock()
	//ft.publisherLock.Lock()
	//for _, pub := range ft.publisherMap {
	//	if pub.CurrentBrokerID == *deadID {
	//		ft.CancelPublisherForwarding(pub.PublisherID)
	//		pub.CurrentBrokerID = UNASSIGNED // mark as currently homeless
	//		ft.etcdManager.UpdateEntity(pub)
	//	}
	//}
	//ft.publisherLock.Unlock()
}

func (ft *ForwardingTable) HandleBrokerReassignment(brokerReassign *BrokerReassignment) {
	// If the publisher/client being reassigned doesn't exist yet, add it to our mappings so that later
	// we will know what the home broker was supposed to be
	log.WithFields(log.Fields{
		"isPublisher": brokerReassign.IsPublisher, "uuid": *brokerReassign.UUID, "homeBroker": *brokerReassign.HomeBrokerID,
	}).Info("Forwarding table marking a client/publisher as reassigned")
	if brokerReassign.IsPublisher {
		ft.publisherLock.Lock()
		if pub, found := ft.publisherMap[*brokerReassign.UUID]; !found {
			pub = &Publisher{
				PublisherID:     *brokerReassign.UUID,
				HomeBrokerID:    *brokerReassign.HomeBrokerID,
				CurrentBrokerID: UNASSIGNED, // will be set when the publisher actually connects
				Metadata:        nil,        // will be set when the publisher actually connects
			}
			ft.publisherMap[*brokerReassign.UUID] = pub
			ft.etcdManager.UpdateEntity(pub)
		} else {
			if pub.HomeBrokerID != *brokerReassign.HomeBrokerID { // Check just in case it was set wrong somehow
				pub.HomeBrokerID = *brokerReassign.HomeBrokerID
				ft.etcdManager.UpdateEntity(pub)
			}
		}
		ft.publisherLock.Unlock()
	} else {
		ft.clientLock.Lock()
		if client, found := ft.clientMap[*brokerReassign.UUID]; !found {
			client = &Client{
				ClientID:        *brokerReassign.UUID,
				HomeBrokerID:    *brokerReassign.HomeBrokerID,
				CurrentBrokerID: UNASSIGNED, // will be set when the client actually connects
				QueryString:     "",         // will be set when the client actually connects
				query:           nil,        // will be set when the client actually connects
			}
			ft.clientMap[*brokerReassign.UUID] = client
			ft.etcdManager.UpdateEntity(client)
		} else {
			if client.HomeBrokerID != *brokerReassign.HomeBrokerID { // Check just in case it was set wrong somehow
				client.HomeBrokerID = *brokerReassign.HomeBrokerID
				ft.etcdManager.UpdateEntity(client)
			}
		}
		ft.clientLock.Unlock()
	}
}

func (ft *ForwardingTable) HandleBrokerLife(liveID *common.UUID) {
	// When a broker comes alive, we need to search for any clients/publishers that are currently
	// assigned to the wrong broker and ask that broker to terminate the connection
	// Also, any publishers or clients which did not yet get assigned a new broker should
	// be automatically reactivated
	log.WithField("liveID", *liveID).Info("Handling a broker coming alive")
	ft.clientLock.Lock()
	brokerClientMap := make(map[common.UUID][]common.UUID)
	for _, client := range ft.clientMap {
		if client.HomeBrokerID == *liveID && client.CurrentBrokerID != *liveID {
			if client.CurrentBrokerID == UNASSIGNED {
				client.CurrentBrokerID = client.HomeBrokerID
				//TODO: nil pointer here
				if client.query == nil {
					log.Warn("CLIENT HAS EMPTY QUERY - We haven't received a subscription for it yet")
				} else {
					fq := ft.getOrEvaluateQuery(client.query.QueryString)
					ft.activateClient(client, fq, &client.CurrentBrokerID)
				}
			} else {
				ft.CancelSubscriberForwarding(client) // cancel old routes
				if ids, found := brokerClientMap[client.CurrentBrokerID]; found {
					brokerClientMap[client.CurrentBrokerID] = append(ids, client.ClientID)
				} else {
					brokerClientMap[client.CurrentBrokerID] = []common.UUID{client.ClientID}
				}
				client.CurrentBrokerID = UNASSIGNED // mark as inactive for now
			}
			ft.etcdManager.UpdateEntity(client)
		}
	}
	ft.clientLock.Unlock()
	ft.publisherLock.Lock()
	brokerPublisherMap := make(map[common.UUID][]common.UUID)
	publishersToCancel := []*common.UUID{}
	for _, publisher := range ft.publisherMap {
		if publisher.HomeBrokerID == *liveID && publisher.CurrentBrokerID != *liveID {
			if publisher.CurrentBrokerID == UNASSIGNED {
				publisher.CurrentBrokerID = publisher.HomeBrokerID
				candidateQueries := ft.gatherCandidateQueries(&publisher.PublisherID, publisher.Metadata)
				ft.reevaluateQueries(candidateQueries, false, true)
			} else {
				publishersToCancel = append(publishersToCancel, &publisher.PublisherID)
				if ids, found := brokerPublisherMap[publisher.CurrentBrokerID]; found {
					brokerPublisherMap[publisher.CurrentBrokerID] = append(ids, publisher.PublisherID)
				} else {
					brokerPublisherMap[publisher.CurrentBrokerID] = []common.UUID{publisher.PublisherID}
				}
				publisher.CurrentBrokerID = UNASSIGNED // mark as inactive for now
			}
			ft.etcdManager.UpdateEntity(publisher)
		}
	}
	ft.publisherLock.Unlock()
	for _, pubID := range publishersToCancel {
		ft.CancelPublisherForwarding(*pubID) // cancel old routes
	}
	for brokerID, clients := range brokerClientMap {
		ft.brokerManager.SendToBroker(brokerID, &common.ClientTerminationRequest{
			MessageIDStruct: common.GetMessageIDStruct(),
			ClientIDs:       clients,
		})
	}
	for brokerID, publishers := range brokerPublisherMap {
		ft.brokerManager.SendToBroker(brokerID, &common.PublisherTerminationRequest{
			MessageIDStruct: common.GetMessageIDStruct(),
			PublisherIDs:    publishers,
		})
	}
}

// homeBrokerID can be specified if it is known and different from brokerID; otherwise leave as nil
func (ft *ForwardingTable) HandleSubscription(queryStr string, clientID common.UUID, brokerID common.UUID,
	homeBrokerID *common.UUID) {
	// check for matching client; if one doesn't exist, create one and add to mapping
	//                            otherwise, update its current broker
	// check if a query object already exists, if not create it and add to mapping
	//       if it does, add client to the list in the appropriate spot
	// if query didn't exist, evaluate it now
	// set up any additional routings as a result
	// respond with a BrokerSubscriptionDiffMessage
	var (
		client     *Client
		homeBroker *common.UUID
		found      bool
	)
	fq := ft.getOrEvaluateQuery(queryStr)

	// Create new client; add to client mapping
	ft.clientLock.Lock()
	if client, found = ft.clientMap[clientID]; found {
		if client.CurrentBrokerID == UNASSIGNED {
			client.CurrentBrokerID = brokerID
			client.query = fq
			log.WithFields(log.Fields{
				"homeBrokerID": client.HomeBrokerID, "clientID": clientID, "brokerID": brokerID, "query": client.QueryString,
			}).Info("New previously inactive client connected. Updating to new broker ID")
		} else if client.CurrentBrokerID != brokerID {
			log.WithFields(log.Fields{
				"homeBrokerID": client.HomeBrokerID, "clientID": clientID, "brokerID": brokerID, "query": client.QueryString,
			}).Warn("Client connected at new broker without going inactive first...")
			ft.CancelSubscriberForwarding(client) // Cancel old routes before making new ones
			client.CurrentBrokerID = brokerID
			client.query = fq
		}
	} else {
		if homeBrokerID == nil {
			homeBroker = &brokerID
		} else {
			homeBroker = homeBrokerID
		}
		client = &Client{
			ClientID:        clientID,
			CurrentBrokerID: brokerID,
			HomeBrokerID:    *homeBroker, // should be connected to home broker
			QueryString:     queryStr,
			query:           fq,
		}
		ft.clientMap[clientID] = client
		log.WithFields(log.Fields{
			"clientID": clientID, "brokerID": client.CurrentBrokerID, "query": queryStr,
		}).Info("New client connected")
	}
	ft.etcdManager.UpdateEntity(client)
	ft.clientLock.Unlock()

	ft.activateClient(client, fq, &brokerID)

	// Respond with SubscriptionDiff
	respMsg := &common.BrokerSubscriptionDiffMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		DelPublishers:   emptyList,
		Query:           queryStr,
	}
	respMsg.FromProducerState(fq.MatchingProducers)

	ft.brokerManager.SendToBroker(brokerID, respMsg)
}

func (ft *ForwardingTable) activateClient(client *Client, fq *ForwardedQuery, brokerID *common.UUID) {
	// Add new client to the appropriate query's list of clients
	// Creating forwarding routes as necessary
	existingBroker := fq.addClient(client, brokerID)
	if !existingBroker {
		ft.addForwardingRoutes(fq, brokerID, false)
	}
}

// Called anytime a publisher sends metadata (new publisher or MD update)
// homeBrokerID can be specified if it is known and different from brokerID; otherwise leave as nil
func (ft *ForwardingTable) HandlePublish(publisherID common.UUID, metadata map[string]interface{},
	brokerID common.UUID, homeBrokerID *common.UUID) {
	var (
		publisher  *Publisher
		homeBroker *common.UUID
		found      bool
	)
	if len(metadata) == 0 {
		// Ignore
		log.WithFields(log.Fields{
			"publisherID": publisherID, "brokerID": brokerID,
		}).Warn("Received a PublishMessage with no Metadata...")
		return
	}

	// save the metadata
	err := ft.metadata.Save(&publisherID, metadata)
	if err != nil {
		log.WithFields(log.Fields{
			"publisherID": publisherID, "metadata": metadata, "error": err,
		}).Error("Could not save metadata")
		return // TODO any other action?
	}

	// Add publisher to mappings if necessary
	ft.publisherLock.Lock()
	if publisher, found = ft.publisherMap[publisherID]; !found {
		if homeBrokerID == nil {
			homeBroker = &brokerID
		} else {
			homeBroker = homeBrokerID
		}
		publisher = &Publisher{PublisherID: publisherID, CurrentBrokerID: brokerID,
			HomeBrokerID: *homeBroker, Metadata: metadata}
		ft.publisherMap[publisherID] = publisher
	} else {
		if publisher.CurrentBrokerID == UNASSIGNED {
			publisher.CurrentBrokerID = brokerID // publisher newly regarded as alive
			publisher.Metadata = metadata
		} else if publisher.CurrentBrokerID != brokerID {
			// publisher somehow switched brokers without going inactive between...
			log.WithFields(log.Fields{
				"publisherID": publisher.PublisherID, "homeBrokerID": publisher.HomeBrokerID,
				"currentBrokerID": publisher.CurrentBrokerID, "newBrokerID": brokerID,
			}).Warn("Publisher switched brokers without going inactive first")
			ft.CancelPublisherForwarding(publisher.PublisherID) // cancel old forwarding before setting up new
			publisher.CurrentBrokerID = brokerID
			publisher.Metadata = metadata
		}
		// Otherwise just a metadata change
	}
	ft.etcdManager.UpdateEntity(publisher)
	ft.publisherLock.Unlock()

	candidateQueries := ft.gatherCandidateQueries(&publisherID, metadata)
	ft.reevaluateQueries(candidateQueries, true, false)
}

// Cancel all forwarding routes relevant to this client
func (ft *ForwardingTable) CancelSubscriberForwarding(client *Client) {
	// find out which queries the client was involved in
	// for each, if this client was the only one from brokerID, remove any forward mappings to that query
	// if this client was the only one from any broker, remove the query from our mappings
	var (
		clientList    *list.List
		oldPublishers map[common.UUID]struct{}
		found         bool
	)
	log.WithFields(log.Fields{
		"clientID": client.ClientID, "homeBrokerID": client.HomeBrokerID, "currentBrokerID": client.CurrentBrokerID,
	}).Info("Cancelling all forwarding relevant to client")
	fq := client.query
	fq.Lock()
	if clientList, found = fq.CurrentBrokerIDs[client.CurrentBrokerID]; !found {
		log.WithFields(log.Fields{
			"query": fq.QueryString, "brokerID": client.CurrentBrokerID, "clientID": client.ClientID,
		}).Warn("Broker attempted to cancel client forwarding but not found under current broker ID")
		return
	}
	clientList.Remove(client.subscriberListElem)
	remainingClientsFromBroker := clientList.Len()
	if remainingClientsFromBroker == 0 {
		delete(fq.CurrentBrokerIDs, client.CurrentBrokerID)
	}
	remainingBrokers := len(fq.CurrentBrokerIDs)
	fq.Unlock()
	if remainingClientsFromBroker > 0 {
		return
	}

	ft.pubQueryLock.Lock()
	ft.queryPubLock.Lock()
	if oldPublishers, found = ft.queryPubMap[fq]; !found {
		log.WithField("query", fq.QueryString).Info("No existing publishers found for query while removing it")
		ft.queryPubLock.Unlock()
		ft.pubQueryLock.Unlock()
		return // nothing to be done
	}
	if remainingBrokers == 0 {
		delete(ft.queryPubMap, fq)
	}

	cancelForwardRoutes := make(map[common.UUID]map[common.UUID][]common.UUID)
	ft.publisherLock.RLock()
	if remainingBrokers == 0 {
		for publisherID, _ := range oldPublishers {
			if queries, found := ft.pubQueryMap[publisherID]; found {
				queries.removeQuery(fq)
			}
		}
	}
	ft.pubQueryLock.Unlock()
	// Determine which forwarding routes must be cancelled
	for publisherID, _ := range oldPublishers {
		if publisher, found := ft.publisherMap[publisherID]; found && publisher.CurrentBrokerID != UNASSIGNED {
			addPublishRouteToMap(cancelForwardRoutes, &publisher.CurrentBrokerID, &client.CurrentBrokerID, &publisherID)
		}
	}
	ft.publisherLock.RUnlock()
	ft.queryPubLock.Unlock()

	if len(cancelForwardRoutes) > 0 {
		log.WithFields(log.Fields{
			"clientID": client.ClientID, "homeBrokerID": client.HomeBrokerID, "currentBrokerID": client.CurrentBrokerID,
			"query": fq.QueryString, "routesToCancel": cancelForwardRoutes,
		}).Debug("Sending forward cancellation requests")
		ft.sendForwardingChanges(cancelForwardRoutes, fq.QueryString, true)
	}
}

// Delete client from mappings and cancel all of its forwarding routes
func (ft *ForwardingTable) HandleSubscriberTermination(clientID common.UUID, brokerID common.UUID) {
	var (
		client *Client
		found  bool
	)
	ft.clientLock.Lock()
	if client, found = ft.clientMap[clientID]; !found {
		log.WithFields(log.Fields{
			"clientID": clientID, "brokerID": brokerID,
		}).Warn("Attempted to handle termination of a subscriber that did not exist")
		return
	}
	delete(ft.clientMap, clientID)
	ft.clientLock.Unlock()

	ft.etcdManager.DeleteEntity(client)
	ft.CancelSubscriberForwarding(client)
}

// Cancel forwarding routes relevant to this publisher
func (ft *ForwardingTable) CancelPublisherForwarding(publisherID common.UUID) {
	log.WithField("publisherID", publisherID).Debug("Cancelling all forwarding for publisher")
	ft.pubQueryLock.Lock()
	ft.queryPubLock.Lock()
	defer ft.pubQueryLock.Unlock()
	defer ft.queryPubLock.Unlock()
	if queries, found := ft.pubQueryMap[publisherID]; found {
		delete(ft.pubQueryMap, publisherID)
		for _, query := range queries.queries {
			if query == nil {
				log.Warn("QUERY WAS NULL")
				continue
			}
			if publishers, found := ft.queryPubMap[query]; found {
				delete(publishers, publisherID)
			}
			query.Lock()
			delete(query.MatchingProducers, publisherID)
			for brokerID, _ := range query.CurrentBrokerIDs {
				ft.brokerManager.SendToBroker(brokerID, &common.BrokerSubscriptionDiffMessage{
					MessageIDStruct: common.GetMessageIDStruct(),
					NewPublishers:   emptyList,
					DelPublishers:   []common.UUID{publisherID},
					Query:           query.QueryString,
				})
			}
			query.Unlock()
		}
	}
}

func (ft *ForwardingTable) HandlePublisherTermination(publisherID common.UUID, brokerID common.UUID) {
	var (
		publisher *Publisher
		found     bool
	)
	// simply remove publisher from local mappings and save its metadata as blank
	// send BrokerSubscriptionDiffMessage to any currently mapped
	err := ft.metadata.RemovePublisher(publisherID)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "publisherID": publisherID,
		}).Error("Error while removing publisher from metadata store")
	}

	ft.publisherLock.Lock()
	if publisher, found = ft.publisherMap[publisherID]; !found {
		log.WithFields(log.Fields{
			"brokerID": brokerID, "publisherID": publisherID,
		}).Warn("Attempted to terminate a non-existent publisher")
	}
	delete(ft.publisherMap, publisherID) // delete is a no-op if the ID doesn't exist
	ft.publisherLock.Unlock()

	if found {
		ft.etcdManager.DeleteEntity(publisher)
	}
	ft.CancelPublisherForwarding(publisherID)
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
			MessageIDStruct: common.GetMessageIDStruct(),
			NewPublishers:   added,
			DelPublishers:   removed,
			Query:           query.QueryString,
		}
		ft.brokerManager.SendToBroker(brokerID, msg)
	}
	query.RUnlock()
}

// Updates the forwarding table for query, including sending routing messages to brokers
// Does not remove old publishing routes!
// If brokerID is not nil, treats brokerID as newly introduced (i.e., not yet forwarding)
// for this query
func (ft *ForwardingTable) addForwardingRoutes(fq *ForwardedQuery, newBrokerID *common.UUID, pubLockHeld bool) {
	log.WithFields(log.Fields{
		"query": fq.QueryString, "MatchingProducers": fq.MatchingProducers,
	}).Debug("Adding forward routes for query with matching producers")
	ft.pubQueryLock.Lock()
	ft.queryPubLock.Lock()
	if !pubLockHeld {
		ft.publisherLock.RLock()
	}

	newPublishRoutes := make(map[common.UUID]map[common.UUID][]common.UUID) // SourceBrokerID -> DestBrokerID -> PublisherID (all map to fq)

	//FIXME: fq is null here when we restart the coordinator
	for publisherID, _ := range fq.MatchingProducers {
		var publishBrokerID common.UUID
		publishBroker, found := ft.publisherMap[publisherID]
		if found && publishBroker != nil {
			publishBrokerID = publishBroker.CurrentBrokerID
		}
		if publishBrokerID == UNASSIGNED {
			// this publisher isn't currently active
			log.WithFields(log.Fields{
				"publisherID": publisherID, "query": fq.QueryString,
			}).Info("Not creating a forwarding route because publisherBrokerID is unassigned")
			continue
		}
		if forwardedQueries, found := ft.pubQueryMap[publisherID]; found {
			// check if we are already in the list
			queryFound := forwardedQueries.containsQuery(fq)
			if queryFound && newBrokerID == nil {
				continue // no new broker and the query exists so nothing to be done for this pub
			} else if queryFound && newBrokerID != nil {
				// query was found but there's a new broker, need to set up forward to new broker
				addPublishRouteToMap(newPublishRoutes, &publishBrokerID, newBrokerID, &publisherID)
				continue
			} else { // we're an entirely new query for this publisher
				log.WithFields(log.Fields{
					"publisherID": publisherID, "query": fq.QueryString, "list": forwardedQueries,
				}).Debug("Adding query to existing query list")
				forwardedQueries.addQuery(fq)
			}
		} else {
			log.WithFields(log.Fields{
				"publisherID": publisherID, "query": fq.QueryString,
			}).Debug("Adding query to NEW query list")
			ft.pubQueryMap[publisherID] = &queryList{queries: []*ForwardedQuery{fq}, nonnilInArray: 1}
		}
		if qpm, found := ft.queryPubMap[fq]; found {
			qpm[publisherID] = struct{}{}
		} else {
			ft.queryPubMap[fq] = make(map[common.UUID]struct{})
			ft.queryPubMap[fq][publisherID] = struct{}{}
		}
		log.WithFields(log.Fields{
			"publisherID": publisherID, "query": fq.QueryString,
			"sourceBrokerID": publishBrokerID, "destinations": fq.CurrentBrokerIDs,
		}).Info("Submitting a new forwarding route from publisherID to query at destinations")
		for destBroker, _ := range fq.CurrentBrokerIDs {
			addPublishRouteToMap(newPublishRoutes, &publishBrokerID, &destBroker, &publisherID)
		}
	}
	ft.pubQueryLock.Unlock()
	ft.queryPubLock.Unlock()
	if !pubLockHeld {
		ft.publisherLock.RUnlock()
	}
	//log.Debugf("forwarding table %v", ft.forwarding)
	logtmp := log.WithField("query", fq.QueryString)
	if newBrokerID != nil {
		logtmp = logtmp.WithField("brokerID", *newBrokerID)
	}
	if len(newPublishRoutes) == 0 {
		logtmp.Debug("Not creating any new forwarding routes while processing updateForwardingTable")
	} else {
		logtmp = logtmp.WithField("newRoutes", newPublishRoutes)
		logtmp.Debug("Creating new forwarding routes while processing updateForwardingTable")
		ft.sendForwardingChanges(newPublishRoutes, fq.QueryString, false)
	}
}

func (ft *ForwardingTable) cancelForwardingRoutes(fq *ForwardedQuery, removed []common.UUID, hasPubLock bool) {
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
	if !hasPubLock {
		ft.publisherLock.RLock()
	}
	for _, rm_uuid := range removed {
		if queries, found = ft.pubQueryMap[rm_uuid]; !found {
			continue // no subscribers for this uuid
		}
		publisherBrokerID := ft.publisherMap[rm_uuid].CurrentBrokerID
		if publisherBrokerID == UNASSIGNED {
			log.WithFields(log.Fields{
				"queries": queries, "publisherID": rm_uuid,
			}).Warn("Found queries in pubQueryMap for an inactive publisher!")
			continue
		}
		if queries.containsQuery(fq) {
			if publishers, found := ft.queryPubMap[fq]; found {
				delete(publishers, rm_uuid)
			} else {
				log.WithFields(log.Fields{
					"publisherID": rm_uuid, "query": fq.QueryString,
				}).Error("Found a pub->query mapping without a corresponding query->pub mapping")
			}
			queries.removeQuery(fq)
			fq.RLock()
			for destBrokerID, _ := range fq.CurrentBrokerIDs {
				addPublishRouteToMap(cancelPublishRoutes, &publisherBrokerID, &destBrokerID, &rm_uuid)
			}
			fq.RUnlock()
		}
	}
	if !hasPubLock {
		ft.publisherLock.RUnlock()
	}
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
		if queryAST.Tree == nil {
			log.WithField("query", queryStr).Error("Error parsing query; queryAST.Tree was nil!")
		}
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
				ft.keyMap[key] = potentialQueries
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

func (ft *ForwardingTable) gatherCandidateQueries(publisherID *common.UUID, metadata map[string]interface{}) map[*ForwardedQuery]struct{} {
	candidateQueries := make(map[*ForwardedQuery]struct{})
	ft.keyLock.RLock()
	// loop through each of the metadata keys
	log.WithFields(log.Fields{
		"producerID": publisherID, "metadata": metadata,
	}).Debug("Gathering candidate queries based on publisher's metadata")
	for key, _ := range metadata {
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

// hasPubLock should be true if the goroutine calling this already has the publisherLock read lock held
func (ft *ForwardingTable) reevaluateQueries(candidateQueries map[*ForwardedQuery]struct{}, sendDiffs bool, hasPubLock bool) {
	for query, _ := range candidateQueries {
		added, removed := ft.metadata.Reevaluate(&query.Query)
		log.WithFields(log.Fields{
			"query": query.QueryString, "added": added, "removed": removed,
		}).Info("Reevaluated query")
		ft.addForwardingRoutes(query, nil, hasPubLock)
		ft.cancelForwardingRoutes(query, removed, hasPubLock)
		if sendDiffs {
			ft.SendSubscriptionDiffs(query, added, removed)
		}
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
