package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	common.SetupTestLogging()
	os.Exit(m.Run())
}

func fakeBroker(coordAddr *net.TCPAddr, expectedMsgs, responses chan common.Sendable) {
	conn, _ := net.DialTCP("tcp", nil, coordAddr)
	reader := msgp.NewReader(conn)
	writer := msgp.NewWriter(conn)
	for {
		msg, _ := common.MessageFromDecoderMsgp(reader)
		log.WithField("message", msg).Debug("Broker received message")
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
