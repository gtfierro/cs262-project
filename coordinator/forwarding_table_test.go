package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/mock"
	"testing"
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
var broker0Info = common.BrokerInfo{BrokerID: "brokerid0", BrokerAddr: "127.0.0.1:60007"}
var broker1Info = common.BrokerInfo{BrokerID: "brokerid1", BrokerAddr: "127.0.0.1:60008"}
var broker2Info = common.BrokerInfo{BrokerID: "brokerid2", BrokerAddr: "127.0.0.1:60009"}

func TestSingleSubscription(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(MockBrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Return(nil)

	ft := NewForwardingTable(mongo, bm, nil, nil, nil)
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid0")

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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid0")
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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid0")
	ft.HandleSubscription(queryStr, "127.0.0.1:4243", "brokerid0")
	ft.HandleSubscription(queryStr, "127.0.0.1:4244", "brokerid1")
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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid0")
	ft.HandleSubscription(queryStr, "127.0.0.1:4243", "brokerid0")
	ft.HandleSubscription(queryStr, "127.0.0.1:4244", "brokerid1")
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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid1")
	ft.HandlePublisherDeath("pub0", "brokerid0")
	ft.HandleSubscription(queryStr+" and Foo != 'Bar'", "127.0.0.1:4243", "brokerid2")
	ft.HandleSubscription(queryStr, "127.0.0.1:4244", "brokerid0")

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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid1")
	ft.HandleSubscription(queryStr, "127.0.0.1:4243", "brokerid2")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscriberDeath("127.0.0.1:4242", "brokerid1")
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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid1")
	ft.HandleSubscription(queryStr, "127.0.0.1:4243", "brokerid1")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscriberDeath("127.0.0.1:4242", "brokerid1")
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
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid1")
	ft.HandleSubscription("Building = 'Cory'", "127.0.0.1:4243", "brokerid1")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid0")
	ft.HandleSubscriberDeath("127.0.0.1:4242", "brokerid1")
	ft.HandleSubscription(queryStr, "127.0.0.1:4242", "brokerid1")

	bm.AssertExpectations(t)
}
