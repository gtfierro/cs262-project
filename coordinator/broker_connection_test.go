package main

import (
	"fmt"
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	common.SetupTestLogging()
	os.Exit(m.Run())
}

func fakeBroker(assert *require.Assertions, coordAddr *net.TCPAddr, expectedMsgs, responses chan common.Sendable) {
	conn, _ := net.DialTCP("tcp", nil, coordAddr)
	reader := msgp.NewReader(conn)
	writer := msgp.NewWriter(conn)
	for {
		msg, _ := common.MessageFromDecoderMsgp(reader)
		expectedMsgs <- msg
		responseMsg, ok := <-responses
		if !ok {
			return
		}
		if responseMsg != nil {
			responseMsg.Encode(writer)
			writer.Flush()
		}
	}
}

func setup(assert *require.Assertions, expectedMsgs, responseMsgs chan common.Sendable) (*BrokerConnection, chan common.Sendable,
	*common.FakeClock, *net.TCPConn, *net.TCPListener) {
	clock := common.NewFakeClock(time.Now())
	msgRcvChan := make(chan common.Sendable, 5)
	msgHandler := func(msg common.Sendable) { msgRcvChan <- msg }
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:56000")
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	go fakeBroker(assert, tcpAddr, expectedMsgs, responseMsgs)
	conn, _ := listener.AcceptTCP()
	broker := Broker{address: "127.0.0.1:56001"}
	bc := NewBrokerConnection(&broker, msgHandler, 5*time.Second, clock)
	return bc, msgRcvChan, clock, conn, listener
}

func TestSendAndReceive(t *testing.T) {
	assert := require.New(t)

	expectedMsgs := make(chan common.Sendable, 3)
	responseMsgs := make(chan common.Sendable, 3)
	bc, msgRcvChan, clock, conn, listener := setup(assert, expectedMsgs, responseMsgs)
	defer func() {
		// Terminate and wait for close
		bc.Terminate()
		bc.WaitForCleanup()
		listener.Close()
		time.Sleep(50 * time.Millisecond) // Brief pause to let TCP close
	}()

	bc.StartAsynchronously(conn)
	smsg1 := common.BrokerPublishMessage{"", make(map[string]interface{}), nil, sync.RWMutex{}}
	rmsg1 := common.BrokerTerminateMessage{}
	smsg2 := common.ForwardRequestMessage{common.MessageIDStruct{MessageID: 1}, nil, "", ""}
	rmsg2 := common.AcknowledgeMessage{MessageID: 1}
	smsg3 := common.ForwardRequestMessage{common.MessageIDStruct{MessageID: 2}, nil, "", ""}
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
	common.AssertStrEqual(assert, &rmsg1, <-msgRcvChan) // BrokerTerminateMessage should be here
	common.AssertChanEmpty(assert, msgRcvChan)
	assert.Len(bc.outstandingMessages, 0) // should be no more - we ACKed both
	clock.AdvanceNowTime(7 * time.Second)
	common.AssertChanEmpty(assert, bc.messageSendBuffer)
	close(responseMsgs)
}

func TestEnsureDelivery(t *testing.T) {
	assert := require.New(t)

	expectedMsgs := make(chan common.Sendable)
	responseMsgs := make(chan common.Sendable, 3)
	bc, msgRcvChan, clock, conn, listener := setup(assert, expectedMsgs, responseMsgs)
	defer func() {
		bc.Terminate()
		bc.WaitForCleanup()
		listener.Close()
		time.Sleep(50 * time.Millisecond) // Brief pause to let TCP close
	}()

	bc.StartAsynchronously(conn)

	smsg := common.ForwardRequestMessage{common.MessageIDStruct{MessageID: 42}, nil, "", ""}
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

	common.AssertChanEmpty(assert, expectedMsgs)
	common.AssertChanEmpty(assert, bc.messageSendBuffer)
	common.AssertChanEmpty(assert, msgRcvChan)
}
