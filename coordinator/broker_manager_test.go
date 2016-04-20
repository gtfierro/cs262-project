package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func TestBrokerDeath(t *testing.T) {
	assert := require.New(t)

	deathChan := make(chan *common.UUID, 5)
	liveChan := make(chan *common.UUID, 5)
	clock := common.NewFakeClock(time.Now())

	expectMsg1 := make(chan common.Sendable, 5)
	expectMsg2 := make(chan common.Sendable, 5)
	expectMsg2a := make(chan common.Sendable, 5)
	expectMsg3 := make(chan common.Sendable, 5)
	respMsg1 := make(chan common.Sendable, 5)
	respMsg2 := make(chan common.Sendable, 5)
	respMsg2a := make(chan common.Sendable, 5)
	respMsg3 := make(chan common.Sendable, 5)

	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:56000")
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	go fakeBroker(tcpAddr, expectMsg1, respMsg1)
	conn1, _ := listener.AcceptTCP()
	uuid1 := common.UUID("1")
	broker1 := common.BrokerInfo{BrokerID: uuid1, BrokerAddr: "0.0.0.0:0001"}

	go fakeBroker(tcpAddr, expectMsg2, respMsg2)
	conn2, _ := listener.AcceptTCP()
	uuid2 := common.UUID("2")
	broker2 := common.BrokerInfo{BrokerID: uuid2, BrokerAddr: "0.0.0.0:0002"}

	go fakeBroker(tcpAddr, expectMsg3, respMsg3)
	conn3, _ := listener.AcceptTCP()
	uuid3 := common.UUID("3")
	broker3 := common.BrokerInfo{BrokerID: uuid3, BrokerAddr: "0.0.0.0:0003"}

	bm := NewBrokerManager(10, deathChan, liveChan, nil, clock)

	defer func() {
		bm.TerminateBroker(uuid1)
		bm.TerminateBroker(uuid2)
		bm.TerminateBroker(uuid3)
		listener.Close()
		time.Sleep(50 * time.Millisecond) // Brief pause to let TCP close
	}()

	bm.ConnectBroker(&broker1, conn1)
	bm.ConnectBroker(&broker2, conn2)
	bm.ConnectBroker(&broker3, conn3)

	for i := 0; i < 3; i++ {
		<-liveChan // Clear out the old live channel
	}

	time.Sleep(50 * time.Millisecond) // Give time for heartbeat threads to get current time

	clock.AdvanceNowTime(22 * time.Second) // Force to send out heartbeat requests
	respMsg1 <- new(common.HeartbeatMessage)
	common.AssertStrEqual(assert, new(common.RequestHeartbeatMessage), <-expectMsg1)
	respMsg3 <- new(common.HeartbeatMessage)
	common.AssertStrEqual(assert, new(common.RequestHeartbeatMessage), <-expectMsg3)

	time.Sleep(50 * time.Millisecond) // Give time for heartbeat threads to get current time

	clock.AdvanceNowTime(12 * time.Second) // Second broker should be considered dead
	assert.Equal(&uuid2, <-deathChan)

	assert.Equal(uuid1, bm.GetLiveBroker().BrokerID)
	assert.Equal(uuid3, bm.GetLiveBroker().BrokerID)

	assert.False(bm.IsBrokerAlive(uuid2))
	close(respMsg2)

	go fakeBroker(tcpAddr, expectMsg2a, respMsg2a)
	conn2a, _ := listener.AcceptTCP()
	bm.ConnectBroker(&broker2, conn2a)

	assert.True(bm.IsBrokerAlive(uuid2))
	assert.Equal(&uuid2, <-liveChan)
	ids := []*common.UUID{&bm.GetLiveBroker().BrokerID, &bm.GetLiveBroker().BrokerID, &bm.GetLiveBroker().BrokerID}
	assert.Contains(ids, &uuid2)
}
