package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/gtfierro/cs262-project/common"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"
	"net"
	"strings"
	"testing"
	"time"
)

func makeConfig(port int) *common.Config {
	return &common.Config{
		Mongo: common.MongoConfig{
			Host:     "0.0.0.0",
			Port:     27017,
			Database: fmt.Sprintf("coordTest_%v", port),
		},
		Coordinator: common.CoordinatorConfig{
			Port:              port,
			Global:            true,
			HeartbeatInterval: 200,
			CoordinatorCount:  3,
			EtcdAddresses:     "127.0.0.1:2379",
		},
	}
}

func TestReplication(t *testing.T) {
	assert := require.New(t)

	etcdConn := NewEtcdConnection([]string{"127.0.0.1:2379"})
	etcdConn.kv.Delete(etcdConn.GetCtx(), "", etcdc.WithPrefix()) // Clear out etcd

	coords := []*Server{NewServer(makeConfig(5001)), NewServer(makeConfig(5002)), NewServer(makeConfig(5003))}
	for _, c := range coords {
		defer c.metadata.DropDatabase()
		go c.Run()
	}

	// Wait for a leader election and maintain that
	var leaderCoord *Server
	for leaderCoord == nil {
		for _, c := range coords {
			if c.leaderService.IsLeader() {
				leaderCoord = c
			}
		}
		time.Sleep(time.Second)
	}

	addrString := strings.Replace(leaderCoord.addrString, "0.0.0.0", "127.0.0.1", -1)
	addr, _ := net.ResolveTCPAddr("tcp", addrString)
	conn1, _ := net.DialTCP("tcp", nil, addr)
	read1 := msgp.NewReader(conn1)
	write1 := msgp.NewWriter(conn1)
	conn2, _ := net.DialTCP("tcp", nil, addr)
	read2 := msgp.NewReader(conn2)
	write2 := msgp.NewWriter(conn2)

	bconnect1 := &common.BrokerConnectMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		BrokerInfo: common.BrokerInfo{
			BrokerID:         common.UUID("broker1"),
			CoordBrokerAddr:  "127.0.0.1:1001",
			ClientBrokerAddr: "127.0.0.1:2001",
		},
	}
	bconnect2 := &common.BrokerConnectMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		BrokerInfo: common.BrokerInfo{
			BrokerID:         common.UUID("broker2"),
			CoordBrokerAddr:  "127.0.0.1:1002",
			ClientBrokerAddr: "127.0.0.1:2002",
		},
	}

	bconnect1.Encode(write1)
	write1.Flush()

	ack, _ := common.MessageFromDecoderMsgp(read1)
	common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: bconnect1.MessageID}, ack)
	bconnect2.Encode(write2)
	write2.Flush()
	ack, _ = common.MessageFromDecoderMsgp(read2)
	common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: bconnect2.MessageID}, ack)

	publish1 := &common.BrokerPublishMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		UUID:            common.UUID("pub1"),
		Metadata:        make(map[string]interface{}),
		Value:           "5",
	}
	publish1.Metadata["Room"] = "410"
	publish1.Encode(write1)
	write1.Flush()

	ack, _ = common.MessageFromDecoderMsgp(read1)
	common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: publish1.MessageID}, ack)

	sub1 := &common.BrokerQueryMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		Query:           "Room = '410'",
		UUID:            common.UUID("sub1"),
	}
	sub1.Encode(write2)
	write2.Flush()

	for sawAck, sawDiffMsg := false, false; sawAck && sawDiffMsg; {
		msg, _ := common.MessageFromDecoderMsgp(read2)
		switch m := msg.(type) {
		case *common.AcknowledgeMessage:
			common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: sub1.MessageID}, m)
			sawAck = true
		case *common.BrokerSubscriptionDiffMessage:
			assert.Equal("Room = '410'", m.Query)
			sawDiffMsg = true
			(&common.AcknowledgeMessage{MessageID: m.MessageID}).Encode(write2)
			write2.Flush()
		}
	}

	for sawFwd := false; sawFwd; {
		msg, _ := common.MessageFromDecoderMsgp(read1)
		switch m := msg.(type) {
		case *common.ForwardRequestMessage:
			assert.Equal(common.UUID("pub1"), m.PublisherList[0])
			assert.Equal(common.UUID("broker2"), m.BrokerID)
			assert.Equal("Room = '410'", m.Query)
			(&common.AcknowledgeMessage{MessageID: m.MessageID}).Encode(write1)
			write1.Flush()
		case common.SendableWithID:
			(&common.AcknowledgeMessage{MessageID: m.GetID()}).Encode(write1)
			write1.Flush()
		}
	}

	leaderCoord.Shutdown() // kill and ensure that a new leader is elected
	//oldLeader := leaderCoord

	// Wait for a leader election
	leaderCoord = nil
	for leaderCoord == nil {
		for _, c := range coords {
			if !c.stopped && c.leaderService.IsLeader() {
				leaderCoord = c
			}
		}
		time.Sleep(time.Second)
	}

	// Now ensure that the new leader is in the current state and correctly proceeds
	addrString = strings.Replace(leaderCoord.addrString, "0.0.0.0", "127.0.0.1", -1)
	addr, _ = net.ResolveTCPAddr("tcp", addrString)
	conn1a, _ := net.DialTCP("tcp", nil, addr)
	read1a := msgp.NewReader(conn1a)
	write1a := msgp.NewWriter(conn1a)
	conn2a, _ := net.DialTCP("tcp", nil, addr)
	read2a := msgp.NewReader(conn2a)
	write2a := msgp.NewWriter(conn2a)

	bconnect1.MessageID = common.GetMessageID()
	bconnect1.Encode(write1a)
	write1a.Flush()
	ack, _ = common.MessageFromDecoderMsgp(read1a)
	common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: bconnect1.MessageID}, ack)
	bconnect2.MessageID = common.GetMessageID()
	bconnect2.Encode(write2a)
	write2a.Flush()
	ack, _ = common.MessageFromDecoderMsgp(read2a)
	common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: bconnect2.MessageID}, ack)

	sub2 := &common.BrokerQueryMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		Query:           "Room = '410'",
		UUID:            common.UUID("sub2"),
	}
	sub2.Encode(write1a)
	write1a.Flush()

	msg, _ := common.MessageFromDecoderMsgp(read2a)
	(&common.AcknowledgeMessage{MessageID: msg.(common.SendableWithID).GetID()}).Encode(write2a)
	write2a.Flush()

	for sawAck, sawFwd1, sawFwd2, sawBDM, sawDiffMsg := false, false, false, false, false; sawAck && sawFwd1 && sawFwd2 && sawDiffMsg && sawBDM; {
		msg, _ := common.MessageFromDecoderMsgp(read1a)
		switch m := msg.(type) {
		case *common.AcknowledgeMessage:
			common.AssertStrEqual(assert, &common.AcknowledgeMessage{MessageID: sub2.MessageID}, m)
			sawAck = true
		case *common.BrokerSubscriptionDiffMessage:
			assert.Equal("Room = '410'", m.Query)
			(&common.AcknowledgeMessage{MessageID: m.MessageID}).Encode(write1a)
			sawDiffMsg = true
			write1a.Flush()
		case *common.ForwardRequestMessage:
			assert.Equal(common.UUID("pub1"), m.PublisherList[0])
			assert.Equal("Room = '410'", m.Query)
			if m.BrokerID == common.UUID("broker1") {
				sawFwd1 = true
			} else if m.BrokerID == common.UUID("broker2") {
				sawFwd2 = true
			}
			(&common.AcknowledgeMessage{MessageID: m.MessageID}).Encode(write1a)
			write1a.Flush()
		case *common.BrokerDeathMessage:
			assert.Equal(common.UUID("broker2"), m.BrokerID)
			(&common.AcknowledgeMessage{MessageID: m.MessageID}).Encode(write1a)
			write1a.Flush()
			sawBDM = true
		case *common.CancelForwardRequest:
			(&common.AcknowledgeMessage{MessageID: m.MessageID}).Encode(write1a)
			write1a.Flush()
		default:
			logrus.Errorf("Message: %v, messagetype: %v", m, common.GetMessageType(m))
		}
	}
}

// test where we get > 100 messages in the log, make a server completely rebuild
