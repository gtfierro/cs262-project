package main

import (
	"cs262-project/common"
	"net"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
)

type Client struct {
	BrokerURL string
	BrokerPort int
	Query string
	msgCount chan int
}

func (c *Client) subscribe() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", c.BrokerURL, c.BrokerPort))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("Error creating connection to %v:%v", c.BrokerURL, c.BrokerPort)
	}

	enc := msgpack.NewEncoder(conn)
	log.WithField("query", c.Query).Debug("Submitting query")
	err = enc.EncodeString(c.Query)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Error sending subscription query to broker")
	}

	dec := msgpack.NewDecoder(conn)
	msg := new(common.Message)
	msgCount := 0
	for {
		err = msg.DecodeMsgpack(dec)
		// Simply accept and ignore the message unless we reach EOF
		if err == io.EOF {
			c.msgCount <- msgCount
			return // connection is closed
		} else if msgCount % 100 == 0 && msgCount > 0 {
			c.msgCount <- msgCount
		}
		msgCount += 1
	}
}
