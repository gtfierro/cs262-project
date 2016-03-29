package main

import (
	"cs262-project/common"
	"net"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
	"time"
)

type Client struct {
	BrokerURL string
	BrokerPort int
	Query string
	latencyChan chan int64
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
	qmsg := common.QueryMessage{c.Query}
	err = (&qmsg).EncodeMsgpack(enc)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Error sending subscription query to broker")
	}

	dec := msgpack.NewDecoder(conn)
	for {
		msg, err := common.MessageFromDecoder(dec)
		arrivalTime := time.Now().UnixNano()
		if err == io.EOF {
			return // connection is closed
		} else if err != nil {
			log.WithField("error", err).Fatal("Error while decoding message")
		}
		switch m := msg.(type) {
		case *common.PublishMessage:
			c.latencyChan <- arrivalTime - int64(m.Value.(uint64)) // TODO ETK weird thing is happening -
		// I'm sending this as an int64 but it's arriving as a uint64???
		case *common.SubscriptionDiffMessage:
		case *common.MatchingProducersMessage:
			// Can safely ignore both of these
		}
	}
}
