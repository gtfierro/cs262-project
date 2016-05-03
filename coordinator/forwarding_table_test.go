package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var queryStr = "Building = 'Soda'"
var expectedDiffBlank = &common.BrokerSubscriptionDiffMessage{
	NewPublishers: []common.UUID{}, DelPublishers: []common.UUID{}, Query: queryStr,
}
var expectedDiffPlusOne = &common.BrokerSubscriptionDiffMessage{
	NewPublishers: []common.UUID{"pub0"}, DelPublishers: []common.UUID{}, Query: queryStr,
}
var expectedDiffMinusOne = &common.BrokerSubscriptionDiffMessage{
	NewPublishers: []common.UUID{}, DelPublishers: []common.UUID{"pub0"}, Query: queryStr,
}

func getBrokerDiffMatcher(expectedDiff *common.BrokerSubscriptionDiffMessage) interface{} {
	return func(msg *common.BrokerSubscriptionDiffMessage) bool {
		if len(msg.NewPublishers) != len(expectedDiff.NewPublishers) ||
			len(msg.DelPublishers) != len(expectedDiff.DelPublishers) {
			return false
		}
		for idx, pub := range msg.NewPublishers {
			if expectedDiff.NewPublishers[idx] != pub {
				return false
			}
		}
		for idx, pub := range msg.DelPublishers {
			if expectedDiff.DelPublishers[idx] != pub {
				return false
			}
		}
		return msg.Query == expectedDiff.Query
	}
}

var metadata = make(map[string]interface{})
var metadata2 = make(map[string]interface{})
var pub0id = common.UUID("pub0")
var pub1id = common.UUID("pub1")
var broker0Info = common.BrokerInfo{BrokerID: "brokerid0", CoordBrokerAddr: "127.0.0.1:60007"}
var broker1Info = common.BrokerInfo{BrokerID: "brokerid1", CoordBrokerAddr: "127.0.0.1:60008"}
var broker2Info = common.BrokerInfo{BrokerID: "brokerid2", CoordBrokerAddr: "127.0.0.1:60009"}

func LoadMongo() *common.MetadataStore {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	mongo := common.NewMetadataStore(config)
	mongo.DropDatabase()
	return mongo
}

func TestSingleSubscription(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Return(nil)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0", nil)

	bm.AssertExpectations(t)
}

func TestSingleSubscribeSingleMatchingPublisher(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	metadata["Building"] = "Soda"
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Return(nil)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0", nil)
	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)

	bm.AssertExpectations(t)
}

func TestSingleQueryMultipleSubscribe(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	metadata["Building"] = "Soda"
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0", nil)
	ft.HandleSubscription(queryStr, "43", "brokerid0", nil)
	ft.HandleSubscription(queryStr, "44", "brokerid1", nil)
	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)
	delete(metadata, "Building")
	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)

	bm.AssertExpectations(t)
}

func TestSingleQueryMultipleSubscribeMetadataChange(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffMinusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffMinusOne))).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0", nil)
	ft.HandleSubscription(queryStr, "43", "brokerid0", nil)
	ft.HandleSubscription(queryStr, "44", "brokerid1", nil)
	metadata["Building"] = "Soda"
	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)
	metadata["Room"] = "372"
	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)
	metadata["Building"] = "Not Soda"
	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)

	bm.AssertExpectations(t)
}

func TestPublisherDeath(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	expectedDiffBlankNewQuery := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{}, DelPublishers: []common.UUID{},
		Query: queryStr + " and Foo != 'Bar'",
	}

	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffMinusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlankNewQuery))).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	metadata["Building"] = "Soda"
	ft.HandlePublish(pub0id, metadata, "brokerid0", nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1", nil)
	ft.HandlePublisherTermination("pub0", "brokerid0")
	ft.HandleSubscription(queryStr+" and Foo != 'Bar'", "43", "brokerid2", nil)
	ft.HandleSubscription(queryStr, "44", "brokerid0", nil)

	bm.AssertExpectations(t)
}

func TestSubscriberDeathSharedQuery(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	expectedDiffPub1 := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{"pub1"}, DelPublishers: []common.UUID{},
		Query: queryStr,
	}

	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker2Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPub1))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker2Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)
	bm.On("GetBrokerInfo", common.UUID("brokerid2")).Return(&broker2Info)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1", nil)
	ft.HandleSubscription(queryStr, "43", "brokerid2", nil)
	metadata["Building"] = "Soda"
	ft.HandlePublish(pub0id, metadata, "brokerid0", nil)
	ft.HandleSubscriberTermination("42", "brokerid1")
	metadata2["Building"] = "Soda"
	metadata2["Room"] = "500"
	ft.HandlePublish(pub1id, metadata2, "brokerid0", nil)

	bm.AssertExpectations(t)
}

func TestSubscriberDeathSharedQuerySharedBroker(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	expectedDiffPub1 := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{"pub1"}, DelPublishers: []common.UUID{},
		Query: queryStr,
	}

	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPub1))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1", nil)
	ft.HandleSubscription(queryStr, "43", "brokerid1", nil)
	metadata["Building"] = "Soda"
	ft.HandlePublish(pub0id, metadata, "brokerid0", nil)
	ft.HandleSubscriberTermination("42", "brokerid1")

	metadata2["Building"] = "Soda"
	metadata2["Room"] = "500"
	ft.HandlePublish(pub1id, metadata2, "brokerid0", nil)

	bm.AssertExpectations(t)
}

func TestSingleSubscriberQueryDeath(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	expectedDiffBadQuery := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{}, DelPublishers: []common.UUID{},
		Query: "Building = 'Cory'",
	}

	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlank))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBadQuery))).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)

	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1", nil)
	ft.HandleSubscription("Building = 'Cory'", "43", "brokerid1", nil)
	metadata["Building"] = "Soda"
	ft.HandlePublish(pub0id, metadata, "brokerid0", nil)
	ft.HandleSubscriberTermination("42", "brokerid1")
	ft.HandleSubscription(queryStr, "42", "brokerid1", nil)

	bm.AssertExpectations(t)
}

func TestBrokerFailureFailover(t *testing.T) {
	// brokerid0 and brokerid1. two clients & a publisher connect to brokerid0
	//                          two client and one publisher at brokerid1
	// brokerid0 dies, one client and the publisher fail over to brokerid1 but one client never does
	// brokerid0 comes back online, they migrate back to brokerid0, client is auto-reactivated
	mongo := LoadMongo()
	defer mongo.DropDatabase()

	bm := new(MockBrokerManager)
	brokerDeathChan := make(chan *common.UUID, 2)
	brokerLiveChan := make(chan *common.UUID, 2)
	brokerReassignChan := make(chan *BrokerReassignment, 2)

	newQuery := "Building = 'Cory'"
	expectedDiffBlankNewQuery := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{}, DelPublishers: []common.UUID{},
		Query: newQuery,
	}
	expectedDiffPlusOneNewQuery := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{"pub1"}, DelPublishers: []common.UUID{},
		Query: newQuery,
	}
	expectedDiffMinusOneNewQuery := &common.BrokerSubscriptionDiffMessage{
		NewPublishers: []common.UUID{}, DelPublishers: []common.UUID{"pub1"},
		Query: newQuery,
	}

	bid0 := common.UUID("brokerid0")
	bid1 := common.UUID("brokerid1")
	cid1 := common.UUID("c1")

	bm.On("GetBrokerInfo", bid1).Return(&broker1Info)
	bm.On("GetBrokerInfo", bid0).Return(&broker0Info)
	// initial setup
	bm.On("SendToBroker", bid0, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffBlankNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOne))).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOneNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOneNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	// Then death, cancel forward routes
	bm.On("SendToBroker", bid1, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffMinusOneNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	// pub1 and c1 connect to brokerid1; c0 doesn't
	bm.On("SendToBroker", bid1, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOneNewQuery))).Twice().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	// brokerid0 comes alive again, cancel old forwarding, set up new forwarding for c0
	bm.On("SendToBroker", bid1, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffMinusOneNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ClientTerminationRequest) bool {
		return msg.ClientIDs[0] == cid1
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.PublisherTerminationRequest) bool {
		return msg.PublisherIDs[0] == pub1id
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pub0id
	})).Once().Return(nil)
	// set up new forwarding for c1 and pub1
	bm.On("SendToBroker", bid1, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOneNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(getBrokerDiffMatcher(expectedDiffPlusOneNewQuery))).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pub1id
	})).Once().Return(nil)

	ft := NewForwardingTable(mongo, bm, new(DummyEtcdManager), brokerDeathChan, brokerLiveChan, brokerReassignChan)
	go ft.monitorInboundChannels()
	metadata["Building"] = "Soda"

	ft.HandlePublish(pub0id, metadata, "brokerid1", nil)
	ft.HandleSubscription(queryStr, "c0", "brokerid0", nil)
	ft.HandleSubscription(newQuery, "c1", "brokerid0", nil)
	ft.HandleSubscription(queryStr, "c2", "brokerid1", nil)
	metadata2["Building"] = "Cory"
	ft.HandlePublish(pub1id, metadata2, "brokerid0", nil)
	ft.HandleSubscription(newQuery, "c3", "brokerid1", nil)

	brokerDeathChan <- &bid0
	brokerReassignChan <- &BrokerReassignment{IsPublisher: false, UUID: &cid1, HomeBrokerID: &bid0}
	brokerReassignChan <- &BrokerReassignment{IsPublisher: true, UUID: &pub1id, HomeBrokerID: &bid0}
	time.Sleep(50 * time.Millisecond)

	// connect to failover broker
	ft.HandlePublish(pub1id, metadata2, "brokerid1", nil)
	ft.HandleSubscription(newQuery, "c1", "brokerid1", nil)

	brokerLiveChan <- &bid0
	time.Sleep(50 * time.Millisecond)

	// connect back home
	ft.HandlePublish(pub1id, metadata2, "brokerid0", nil)
	ft.HandleSubscription(newQuery, "c1", "brokerid0", nil)

	bm.AssertExpectations(t)
}
