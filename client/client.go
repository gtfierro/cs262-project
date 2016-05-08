package client

import (
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/errors"
)

var log *logging.Logger

type Client struct {
	BrokerConnection
	// unique client identifier
	ID common.UUID

	config *Config

	// if true, then publishHandler is non-null
	hasPublishHandler bool
	// attach a publish handler using AttachPublishHandler
	publishHandler func(m *common.PublishMessage)

	hasDiffHandler bool
	diffHandler    func(m *common.SubscriptionDiffMessage)

	// the query this client is subscribed to
	query string
}

// Creates a new client with the given configuration
func NewClient(clientID common.UUID, query string, cfg *Config) (c *Client, err error) {
	c = &Client{
		query:             query,
		ID:                clientID,
		hasPublishHandler: false,
		hasDiffHandler:    false,
	}
	err = (&c.BrokerConnection).initialize(c.subscribe, c.messageHandler, false, clientID, cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// This function is called whenever the client receives a published message
func (c *Client) AttachPublishHandler(f func(m *common.PublishMessage)) {
	c.hasPublishHandler = true
	c.publishHandler = f
}

func (c *Client) AttachDiffHandler(f func(m *common.SubscriptionDiffMessage)) {
	c.hasDiffHandler = true
	c.diffHandler = f
}

func (c *Client) messageHandler(msg common.Sendable) {
	switch m := msg.(type) {
	case *common.PublishMessage:
		if c.hasPublishHandler {
			c.publishHandler(m)
		} else {
			log.Infof("Got publish message %v", m)
		}
	case *common.SubscriptionDiffMessage:
		if c.hasDiffHandler {
			c.diffHandler(m)
		} else {
			log.Infof("Got diff message %v", m)
		}
	default:
		log.Infof("Got %T message %v", m, m)
	}
}

// subscribes the client to the given query via the broker specified in the
// client's configuration (or whatever next broker if the client has experienced
// a failover). Use AttachPublishHandler to do special handling of the received
// published messages
func (c *Client) subscribe() {
	// cache the query we are subscribing to
	msg := &common.QueryMessage{
		UUID:  c.ID,
		Query: c.query,
	}
	if err := c.sendBroker(msg); err != nil {
		log.Errorf("Error? %v", errors.Cause(err))
	}
}
