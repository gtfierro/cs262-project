package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/gtfierro/cs262-project/coordinator/mocks"
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
var publishMessage1 = &common.PublishMessage{
	UUID: "pub0", Metadata: make(map[string]interface{}), Value: "1",
}

func TestSingleSubscription(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(mocks.BrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Return(nil)

	ft := NewForwardingTable(mongo, bm)
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4242", "brokerid0")

	bm.AssertExpectations(t)
}

func TestSingleSubscribeSingleMatchingPublisher(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(mocks.BrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	publishMessage1.Metadata["Building"] = "Soda"
	broker0Info := common.BrokerInfo{BrokerID: "brokerid0", BrokerAddr: "127.0.0.1:60007"}
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffBlank).Return(nil)
	bm.On("SendToBroker", common.UUID("brokerid0"), expectedDiffPlusOne).Return(nil)
	bm.On("GetBrokerInfo", common.UUID("brokerid0")).Return(&broker0Info)
	bm.On("SendToBroker", common.UUID("brokerid1"), mock.MatchedBy(func(msg *common.ForwardRequestMessage) bool {
		return msg.BrokerInfo == broker0Info && msg.Query == queryStr &&
			msg.PublisherList[0] == common.UUID("pub0")
	})).Return(nil)

	ft := NewForwardingTable(mongo, bm)
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4242", "brokerid0")
	ft.HandlePublish(publishMessage1, "brokerid1")

	bm.AssertExpectations(t)
}

func TestSingleQueryMultipleSubscribe(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(mocks.BrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	publishMessage1.Metadata["Building"] = "Soda"
	broker0Info := common.BrokerInfo{BrokerID: "brokerid0", BrokerAddr: "127.0.0.1:60007"}
	broker1Info := common.BrokerInfo{BrokerID: "brokerid1", BrokerAddr: "127.0.0.1:60008"}
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

	ft := NewForwardingTable(mongo, bm)
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4242", "brokerid0")
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4243", "brokerid0")
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4244", "brokerid1")
	ft.HandlePublish(publishMessage1, "brokerid1")
	delete(publishMessage1.Metadata, "Building")
	ft.HandlePublish(publishMessage1, "brokerid1")

	bm.AssertExpectations(t)
}

func TestSingleQueryMultipleSubscribeMetadataChange(t *testing.T) {
	config, _ := common.LoadConfig("config.ini")
	config.Mongo.Database = "test_coordinator"

	bm := new(mocks.BrokerManager)
	mongo := common.NewMetadataStore(config)
	defer mongo.DropDatabase()

	broker0Info := common.BrokerInfo{BrokerID: "brokerid0", BrokerAddr: "127.0.0.1:60007"}
	broker1Info := common.BrokerInfo{BrokerID: "brokerid1", BrokerAddr: "127.0.0.1:60008"}
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

	ft := NewForwardingTable(mongo, bm)
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4242", "brokerid0")
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4243", "brokerid0")
	ft.HandleSubscription("Building = 'Soda'", "127.0.0.1:4244", "brokerid1")
	publishMessage1.Metadata["Building"] = "Soda"
	ft.HandlePublish(publishMessage1, "brokerid1")
	publishMessage1.Metadata["Room"] = "372"
	ft.HandlePublish(publishMessage1, "brokerid1")
	publishMessage1.Metadata["Building"] = "Not Soda"
	ft.HandlePublish(publishMessage1, "brokerid1")

	bm.AssertExpectations(t)
}
