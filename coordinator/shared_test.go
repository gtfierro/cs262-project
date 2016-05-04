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

func sendDummyMessage(conn *net.TCPConn, expectMsgChan chan common.Sendable) {
	w1 := msgp.NewWriter(conn)
	(&common.AcknowledgeMessage{}).Encode(w1)
	w1.Flush()
	<-expectMsgChan
}

func AssertMFBChanEmpty(assert *require.Assertions, channel chan *MessageFromBroker) {
	select {
	case <-channel:
		assert.Fail("Channel not empty")
	default:
	}
}

type PassthroughCommConn struct {
	conn   *net.TCPConn
	writer *msgp.Writer
	reader *msgp.Reader
}

func NewPassthroughCommConn(conn *net.TCPConn) CommConn {
	pcc := new(PassthroughCommConn)
	pcc.conn = conn
	pcc.writer = msgp.NewWriter(conn)
	pcc.reader = msgp.NewReader(conn)
	return pcc
}

func (pcc *PassthroughCommConn) Send(msg common.Sendable) error {
	err := msg.Encode(pcc.writer)
	pcc.writer.Flush()
	return err
}
func (pcc *PassthroughCommConn) ReceiveMessage() (msg common.Sendable, err error) {
	return common.MessageFromDecoderMsgp(pcc.reader)
}
func (pcc *PassthroughCommConn) GetBrokerConn(brokerID common.UUID) CommConn {
	return pcc
}
func (pcc *PassthroughCommConn) Close() {
	pcc.conn.Close()
}
func (pcc *PassthroughCommConn) GetPendingMessages() map[common.MessageIDType]common.SendableWithID {
	return make(map[common.MessageIDType]common.SendableWithID)
}
