package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
)

type Broker interface {
	RemapProducer(p *Producer, message *common.PublishMessage)
	HandleProducer(msg *common.PublishMessage, dec *msgp.Reader, conn net.Conn)
	NewSubscription(querystring string, conn net.Conn) *Client
	ForwardMessage(msg *common.PublishMessage)
}
