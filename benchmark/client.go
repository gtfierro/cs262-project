package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"time"
)

type Client struct {
	BrokerURL   string
	BrokerPort  int
	Query       string
	latencyChan chan int64
}

func (c *Client) subscribe() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", c.BrokerURL, c.BrokerPort))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("Error creating connection to %v:%v", c.BrokerURL, c.BrokerPort)
	}

	enc := msgp.NewWriter(conn)
	log.WithField("query", c.Query).Debug("Submitting query")
	qmsg := common.QueryMessage(c.Query)
	err = qmsg.Encode(enc)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Error marshalling query")
	}
	err = enc.Flush()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Error sending subscription query to broker")
	}

	dec := msgp.NewReader(conn)
	for {
		msg, err := common.MessageFromDecoderMsgp(dec)
		arrivalTime := time.Now().UnixNano()
		if err == io.EOF {
			return // connection is closed
		} else if err != nil {
			log.WithField("error", err).Fatal("Error while decoding message")
		}
		switch m := msg.(type) {
		case *common.PublishMessage:
			c.latencyChan <- arrivalTime - m.Value.(int64)
		case *common.SubscriptionDiffMessage:
			// Can safely ignore this
		}
	}
}
