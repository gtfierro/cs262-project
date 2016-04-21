package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"
	"net"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	common.SetupTestLogging()
	os.Exit(m.Run())
}

func fakeBroker(coordAddr *net.TCPAddr, expectedMsgs, responses chan common.Sendable, brokerDeath chan bool) {
	conn, _ := net.DialTCP("tcp", nil, coordAddr)
	reader := msgp.NewReader(conn)
	writer := msgp.NewWriter(conn)
	for {
		msg, _ := common.MessageFromDecoderMsgp(reader)
		log.WithField("message", msg).Debug("Broker received message")
		if msg != nil {
			expectedMsgs <- msg
		}
		responseMsg, ok := <-responses
		if !ok {
			brokerDeath <- true
			return
		}
		if responseMsg != nil {
			responseMsg.Encode(writer)
			writer.Flush()
		}
	}
}

func AssertMFBChanEmpty(assert *require.Assertions, channel chan *MessageFromBroker) {
	select {
	case <-channel:
		assert.Fail("Channel not empty")
	default:
	}
}
