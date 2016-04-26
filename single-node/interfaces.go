package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
)

type Broker interface {
	HandleProducer(msg *common.PublishMessage, dec *msgp.Reader, conn net.Conn)
	NewSubscription(query string, clientID common.UUID, conn net.Conn) *Client
	ForwardMessage(msg *common.PublishMessage)
}
