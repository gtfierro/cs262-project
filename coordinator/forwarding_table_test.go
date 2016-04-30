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
var publishMessage1 = &common.BrokerPublishMessage{
	MessageIDStruct: common.MessageIDStruct{1}, UUID: "pub0",
	Metadata: make(map[string]interface{}), Value: "1",
}
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

	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Return(nil)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0")

	bm.AssertExpectations(t)
}

func TestSingleSubscribeSingleMatchingPublisher(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	publishMessage1.Metadata["Building"] = "Soda"
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffPlusOne).Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Return(nil)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0")
	ft.HandlePublish(publishMessage1, "brokerid1")

	bm.AssertExpectations(t)
}

func TestSingleQueryMultipleSubscribe(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	publishMessage1.Metadata["Building"] = "Soda"
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffBlank).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0")
	ft.HandleSubscription(queryStr, "43", "brokerid0")
	ft.HandleSubscription(queryStr, "44", "brokerid1")
	ft.HandlePublish(publishMessage1, "brokerid1")
	delete(publishMessage1.Metadata, "Building")
	ft.HandlePublish(publishMessage1, "brokerid1")

	bm.AssertExpectations(t)
}

func TestSingleQueryMultipleSubscribeMetadataChange(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffBlank).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffMinusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffMinusOne).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid0")
	ft.HandleSubscription(queryStr, "43", "brokerid0")
	ft.HandleSubscription(queryStr, "44", "brokerid1")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid1")
	publishMessage1.Metadata["Room"] = "372"
	ft.HandlePublish(publishMessage1, "brokerid1")
	publishMessage1.Metadata["Building"] = "Not Soda"
	ft.HandlePublish(publishMessage1, "brokerid1")

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

	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffMinusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), expectedDiffBlankNewQuery).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscription(queryStr, "42", "brokerid1")
	ft.HandlePublisherTermination("pub0", "brokerid0")
	ft.HandleSubscription(queryStr+" and Foo != 'Bar'", "43", "brokerid2")
	ft.HandleSubscription(queryStr, "44", "brokerid0")

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

	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffBlank).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), expectedDiffBlank).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker2Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid2"), expectedDiffPub1).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker2Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub1")
	})).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)
	bm.On("GetBrokerInfo", common.UUID("brokerid2")).Return(&broker2Info)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1")
	ft.HandleSubscription(queryStr, "43", "brokerid2")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscriberTermination("42", "brokerid1")
	var publishMessage2 = &common.BrokerPublishMessage{
		MessageIDStruct: common.MessageIDStruct{2},
		UUID:            "pub1", Metadata: make(map[string]interface{}), Value: "2",
	}
	publishMessage2.Metadata["Building"] = "Soda"
	publishMessage2.Metadata["Room"] = "500"
	ft.HandlePublish(publishMessage2, "brokerid0")

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

	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffBlank).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPub1).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub1")
	})).Once().Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1")
	ft.HandleSubscription(queryStr, "43", "brokerid1")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscriberTermination("42", "brokerid1")
	var publishMessage2 = &common.BrokerPublishMessage{
		MessageIDStruct: common.MessageIDStruct{2},
		UUID:            "pub1", Metadata: make(map[string]interface{}), Value: "2",
	}
	publishMessage2.Metadata["Building"] = "Soda"
	publishMessage2.Metadata["Room"] = "500"
	ft.HandlePublish(publishMessage2, "brokerid0")

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

	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffBlank).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffBadQuery).Once().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid1"), expectedDiffPlusOne).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Twice().Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Once().Return(nil)

	bm.On("GetBrokerInfo", common.UUID("brokerid1")).Return(&broker1Info)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "42", "brokerid1")
	ft.HandleSubscription("Building = 'Cory'", "43", "brokerid1")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscriberTermination("42", "brokerid1")
	ft.HandleSubscription(queryStr, "42", "brokerid1")

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

	var publishMessage2 = &common.BrokerPublishMessage{
		MessageIDStruct: common.MessageIDStruct{2},
		UUID:            "pub1", Metadata: make(map[string]interface{}), Value: "2",
	}

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
	pid1 := common.UUID("pub1")
	pid0 := common.UUID("pub0")

	bm.On("GetBrokerInfo", bid1).Return(&broker1Info)
	bm.On("GetBrokerInfo", bid0).Return(&broker0Info)
	// initial setup
	bm.On("SendToBroker", bid0, expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pid0
	})).Once().Return(nil)
	bm.On("SendToBroker", bid0, expectedDiffBlankNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid1, expectedDiffPlusOne).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pid0
	})).Once().Return(nil)
	bm.On("SendToBroker", bid0, expectedDiffPlusOneNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pid1
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, expectedDiffPlusOneNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pid1
	})).Once().Return(nil)
	// Then death, cancel forward routes
	bm.On("SendToBroker", bid1, expectedDiffMinusOneNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pid1
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.CancelForwardRequest) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pid0
	})).Once().Return(nil)
	// pub1 and c1 connect to brokerid1; c0 doesn't
	bm.On("SendToBroker", bid1, expectedDiffPlusOneNewQuery).Twice().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pid1
	})).Once().Return(nil)
	// brokerid0 comes alive again, cancel old forwarding, set up new forwarding for c0
	bm.On("SendToBroker", bid1, expectedDiffMinusOneNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ClientTerminationRequest) bool {
		return msg.ClientIDs[0] == cid1
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.PublisherTerminationRequest) bool {
		return msg.PublisherIDs[0] == pid1
	})).Once().Return(nil)
	bm.On("SendToBroker", bid1, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == pid0
	})).Once().Return(nil)
	// set up new forwarding for c1 and pub1
	bm.On("SendToBroker", bid1, expectedDiffPlusOneNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker1Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pid1
	})).Once().Return(nil)
	bm.On("SendToBroker", bid0, expectedDiffPlusOneNewQuery).Once().Return(nil)
	bm.On("SendToBroker", bid0, mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == newQuery &&
			msg.PublisherList[0] == pid1
	})).Once().Return(nil)

	ft := NewForwardingTable(mongo, bm, brokerDeathChan, brokerLiveChan, brokerReassignChan)
	go ft.monitorInboundChannels()
	publishMessage1.Metadata["Building"] = "Soda"

	ft.HandlePublish(publishMessage1, "brokerid1")
	ft.HandleSubscription(queryStr, "c0", "brokerid0")
	ft.HandleSubscription(newQuery, "c1", "brokerid0")
	ft.HandleSubscription(queryStr, "c2", "brokerid1")
	publishMessage2.Metadata["Building"] = "Cory"
	ft.HandlePublish(publishMessage2, "brokerid0")
	ft.HandleSubscription(newQuery, "c3", "brokerid1")

	brokerDeathChan <- &bid0
	brokerReassignChan <- &BrokerReassignment{IsPublisher: false, UUID: &cid1, HomeBrokerID: &bid0}
	brokerReassignChan <- &BrokerReassignment{IsPublisher: true, UUID: &pid1, HomeBrokerID: &bid0}
	time.Sleep(50 * time.Millisecond)

	// connect to failover broker
	ft.HandlePublish(publishMessage2, "brokerid1")
	ft.HandleSubscription(newQuery, "c1", "brokerid1")

	brokerLiveChan <- &bid0
	time.Sleep(50 * time.Millisecond)

	// connect back home
	ft.HandlePublish(publishMessage2, "brokerid0")
	ft.HandleSubscription(newQuery, "c1", "brokerid0")

	bm.AssertExpectations(t)
}
