package main

import (
	"fmt"
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func setupBroker(expectedMsgs, responseMsgs chan common.Sendable, brokerDeath chan bool) (*Broker, chan *MessageFromBroker,
	*common.FakeClock, *net.TCPConn, *net.TCPListener) {
	clock := common.NewFakeClock(time.Now())
	msgRcvChan := make(chan *MessageFromBroker, 5)
	deathChan := make(chan *Broker, 5)
	msgHandler := func(msg *MessageFromBroker) { msgRcvChan <- msg }
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:56000")
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	go fakeBroker(tcpAddr, expectedMsgs, responseMsgs, brokerDeath)
	conn, _ := listener.AcceptTCP()
	broker := common.BrokerInfo{BrokerID: "42", BrokerAddr: "0.0.0.0:0000"}
	bc := NewBroker(&broker, msgHandler, 5*time.Second, clock, deathChan)
	return bc, msgRcvChan, clock, conn, listener
}

func TestSendAndReceive(t *testing.T) {
	assert := require.New(t)

	expectedMsgs := make(chan common.Sendable, 3)
	responseMsgs := make(chan common.Sendable, 3)
	brokerDeath := make(chan bool)
	bc, msgRcvChan, clock, conn, listener := setupBroker(expectedMsgs, responseMsgs, brokerDeath)
	defer func() {
		// Terminate and wait for close
		bc.Terminate()
		bc.WaitForCleanup()
		listener.Close()
		time.Sleep(50 * time.Millisecond) // Brief pause to let TCP close
	}()

	bc.StartAsynchronously(NewPassthroughCommConn(conn))
	smsg1 := common.BrokerAssignmentMessage{common.BrokerInfo{common.UUID("5"), ""}}
	rmsg1 := common.BrokerTerminateMessage{}
	smsg2 := common.ForwardRequestMessage{common.MessageIDStruct{MessageID: 1}, nil, common.BrokerInfo{}, ""}
	rmsg2 := common.AcknowledgeMessage{MessageID: 1}
	smsg3 := common.ForwardRequestMessage{common.MessageIDStruct{MessageID: 2}, nil, common.BrokerInfo{}, ""}
	rmsg3 := common.AcknowledgeMessage{MessageID: 2}

	responseMsgs <- &rmsg1

	bc.Send(&smsg1)
	bc.Send(&smsg2)
	bc.Send(&smsg3)

	responseMsgs <- &rmsg3
	responseMsgs <- &rmsg2

	common.AssertStrEqual(assert, &smsg1, <-expectedMsgs)
	common.AssertStrEqual(assert, &smsg2, <-expectedMsgs)
	common.AssertStrEqual(assert, &smsg3, <-expectedMsgs)
	common.AssertStrEqual(assert, &rmsg1, (<-msgRcvChan).message) // BrokerTerminateMessage should be here
	conn.CloseRead()
	close(responseMsgs)
	<-brokerDeath // wait for it to exit
	AssertMFBChanEmpty(assert, msgRcvChan)
	assert.Len(bc.outstandingMessages, 0) // should be no more - we ACKed both
	clock.AdvanceNowTime(7 * time.Second)
	common.AssertSendableChanEmpty(assert, bc.messageSendBuffer)
}

func TestEnsureDelivery(t *testing.T) {
	assert := require.New(t)

	expectedMsgs := make(chan common.Sendable)
	responseMsgs := make(chan common.Sendable, 3)
	brokerDeath := make(chan bool)
	bc, msgRcvChan, clock, conn, listener := setupBroker(expectedMsgs, responseMsgs, brokerDeath)
	defer func() {
		bc.Terminate()
		bc.WaitForCleanup()
		listener.Close()
		time.Sleep(50 * time.Millisecond) // Brief pause to let TCP close
	}()

	bc.StartAsynchronously(NewPassthroughCommConn(conn))

	smsg := common.ForwardRequestMessage{common.MessageIDStruct{MessageID: 42}, nil, common.BrokerInfo{}, ""}
	ackmsg := common.AcknowledgeMessage{MessageID: 42}
	wrongAckmsg := common.AcknowledgeMessage{MessageID: 1}
	heartbeat := common.HeartbeatMessage{}
	reqheartbeat := common.RequestHeartbeatMessage{}
	rmsgTerm := common.BrokerTerminateMessage{}

	responseMsgs <- &wrongAckmsg
	responseMsgs <- &wrongAckmsg
	responseMsgs <- &heartbeat

	bc.Send(&smsg)

	common.AssertStrEqual(assert, &smsg, <-expectedMsgs)
	clock.AdvanceNowTime(6 * time.Second)
	common.AssertStrEqual(assert, &smsg, <-expectedMsgs)
	clock.AdvanceNowTime(6 * time.Second)
	msgStrs := fmt.Sprintf("%v %v", <-expectedMsgs, <-expectedMsgs)
	assert.True(msgStrs == fmt.Sprintf("%v %v", &smsg, &reqheartbeat) ||
		msgStrs == fmt.Sprintf("%v %v", &reqheartbeat, &smsg))

	responseMsgs <- &ackmsg
	responseMsgs <- &rmsgTerm

	clock.AdvanceNowTime(2 * time.Second)

	bc.Send(&common.PublishMessage{}) // Just send something without an ID to continue the broker loop
	<-expectedMsgs

	<-msgRcvChan
	close(responseMsgs)
	conn.CloseRead()
	<-brokerDeath // wait for it to exit
	common.AssertSendableChanEmpty(assert, expectedMsgs)
	common.AssertSendableChanEmpty(assert, bc.messageSendBuffer)
	AssertMFBChanEmpty(assert, msgRcvChan)
}
