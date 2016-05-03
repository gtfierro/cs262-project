package client

import (
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/errors"
	uuidlib "github.com/satori/go.uuid"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var log *logging.Logger
var NamespaceUUID = uuidlib.FromStringOrNil("85ce106e-0ccf-11e6-81fc-0cc47a0f7eea")

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

// Creates a deterministic UUID from a given name. Names are easier to remember
// than UUIDs, so this should make writing scripts easier
func UUIDFromName(name string) common.UUID {
	return common.UUID(uuidlib.NewV5(NamespaceUUID, name).String())
}

// configuration for a client
type Config struct {
	// ip:port of the initial, local broker
	BrokerAddress string
	// ip:port of the coordinator
	CoordinatorAddress string
	// the client identifier. Must be Unique!
	ID common.UUID
}

type Client struct {
	// unique client identifier
	ID common.UUID

	// Handling the connection to the local broker
	// the IP:Port of the local broker we talk to
	BrokerAddress    *net.TCPAddr
	brokerConn       *net.TCPConn
	brokerEncoder    *msgp.Writer
	brokerEncodeLock sync.Mutex

	// the IP:Port of the coordinator that we fall back to
	CoordinatorAddress *net.TCPAddr
	coordConn          *net.TCPConn
	coordEncoder       *msgp.Writer
	coordEncodeLock    sync.Mutex

	// if true, then publishHandler is non-null
	hasPublishHandler bool
	// attach a publish handler using AttachPublishHandler
	publishHandler func(m *common.PublishMessage)
	publishersLock sync.RWMutex
	publishers     map[common.UUID]*Publisher

	hasDiffHandler bool
	diffHandler    func(m *common.SubscriptionDiffMessage)

	brokerDead bool

	// client signals on this channel when it is done
	Stop chan bool

	// the query this client is subscribed to
	query string
}

// Creates a new client with the given configuration
func NewClient(cfg *Config) (*Client, error) {
	var err error
	c := &Client{
		ID:                cfg.ID,
		Stop:              make(chan bool),
		hasPublishHandler: false,
		publishers:        make(map[common.UUID]*Publisher),
		hasDiffHandler:    false,
		brokerDead:        true,
	}
	c.BrokerAddress, err = net.ResolveTCPAddr("tcp", cfg.BrokerAddress)
	if err != nil {
		return c, errors.Wrap(err, "Could not resolve local broker address")
	}
	if err = c.connectBroker(c.BrokerAddress); err != nil {
		return c, errors.Wrap(err, "Could not connect to local broker")
	}

	c.CoordinatorAddress, err = net.ResolveTCPAddr("tcp", cfg.CoordinatorAddress)
	if err != nil {
		return c, errors.Wrap(err, "Could not resolve coordinator address")
	}

	// start listening for messages from the broker
	go c.listen()

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

// This should be triggered when we can no longer contact our local broker. In this
// case, we sent a BrokerRequestMessage to the coordinator
func (c *Client) doFailover() {
	c.brokerDead = true
	// establish the coordinator connection
	c.connectCoordinator()
	// prepare the BrokerRequestMessage
	brm := &common.BrokerRequestMessage{
		LocalBrokerAddr: c.BrokerAddress.String(),
		IsPublisher:     false,
		UUID:            "392c1b18-0c37-11e6-b352-1002b58053c7",
	}
	// loop until we can contact the coordinator
	err := c.sendCoordinator(brm)
	for err != nil {
		time.Sleep(1)
		err = c.sendCoordinator(brm)
	}
}

func (c *Client) connectBroker(address *net.TCPAddr) error {
	var err error
	if c.brokerConn, err = net.DialTCP("tcp", nil, address); err != nil {
		return errors.Wrap(err, "Could not dial broker")
	}
	c.brokerEncodeLock.Lock()
	c.brokerEncoder = msgp.NewWriter(c.brokerConn)
	c.brokerEncodeLock.Unlock()
	c.brokerDead = false
	return nil
}

// Loop until we can finally connect to the coordinator.
// This blocks indefinitely until it is successful
func (c *Client) connectCoordinator() {
	var (
		err      error
		waitTime = 1 * time.Second
		maxWait  = 30 * time.Second
	)
	c.coordConn, err = net.DialTCP("tcp", nil, c.CoordinatorAddress)
	for err != nil {
		log.Warningf("Retrying coordinator connection to %v with delay %v", c.CoordinatorAddress, waitTime)
		time.Sleep(waitTime)
		waitTime *= 2
		if waitTime > maxWait {
			waitTime = maxWait
		}
		c.coordConn, err = net.DialTCP("tcp", nil, c.CoordinatorAddress)
	}
	log.Debug("Connected to coordinator")
	go c.listenCoordinator()
	c.coordEncoder = msgp.NewWriter(c.coordConn)
}

//TODO: implement
// This function should contact the coordinator to get the new broker
func (c *Client) configureNewBroker(m *common.BrokerAssignmentMessage) {
	var err error
	c.BrokerAddress, err = net.ResolveTCPAddr("tcp", m.ClientBrokerAddr)
	if err != nil {
		log.Critical(errors.Wrap(err, "Could not resolve local broker address"))
	}
	if err = c.connectBroker(c.BrokerAddress); err != nil {
		log.Critical(errors.Wrap(err, "Could not connect to local broker"))
	}
	// send our subscription once we're back
	c.resubscribe()
	go c.listen()
}

// Sends message to the currently configured broker
func (c *Client) sendBroker(m common.Sendable) error {
	c.brokerEncodeLock.Lock()
	defer c.brokerEncodeLock.Unlock()
	if err := m.Encode(c.brokerEncoder); err != nil {
		return errors.Wrap(err, "Could not encode message")
	}
	if err := c.brokerEncoder.Flush(); err != nil {
		// do failover if we fail to send
		go c.doFailover()
		return errors.Wrap(err, "Could not send message to broker")
	}
	return nil
}

func (c *Client) sendCoordinator(m common.Sendable) error {
	c.coordEncodeLock.Lock()
	if err := m.Encode(c.coordEncoder); err != nil {
		c.coordEncodeLock.Unlock()
		return errors.Wrap(err, "Could not encode message")
	}
	if err := c.coordEncoder.Flush(); err != nil {
		c.coordEncodeLock.Unlock()
		c.connectCoordinator()
		return errors.Wrap(err, "Could not send message to coordinator")
	}
	c.coordEncodeLock.Unlock()
	return nil
}

func (c *Client) listen() {
	reader := msgp.NewReader(net.Conn(c.brokerConn))
	for {
		if c.brokerDead {
			return
		}
		msg, err := common.MessageFromDecoderMsgp(reader)
		if err == io.EOF {
			log.Warn("connection closed. Do failover")
			c.doFailover()
			break
		}
		if err != nil {
			log.Warn(errors.Wrap(err, "Could not decode message"))
		}

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
}

func (c *Client) listenCoordinator() {
	if c.coordConn == nil {
		return
	}
	reader := msgp.NewReader(net.Conn(c.coordConn))
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		if err == io.EOF {
			log.Warn("connection closed. Do failover")
			c.doFailover()
			break
		}
		if err != nil {
			log.Warn(errors.Wrap(err, "Could not decode message"))
		}

		log.Infof("Got %T message %v from coordinator", msg, msg)
		switch m := msg.(type) {
		case *common.BrokerAssignmentMessage:
			c.configureNewBroker(m)
		default:
			log.Infof("Got %T message %v from coordinator", m, m)
		}
	}
}

// after the duration expires, stop the client by signalling on c.Stop
func (c *Client) StopIn(d time.Duration) {
	go func(c *Client) {
		time.Sleep(d)
		c.Stop <- true
	}(c)
}

// subscribes the client to the given query via the broker specified in the
// client's configuration (or whatever next broker if the client has experienced
// a failover). Use AttachPublishHandler to do special handling of the received
// published messages
func (c *Client) Subscribe(query string) {
	// cache the query we are subscribing to
	c.query = query
	msg := &common.QueryMessage{
		UUID:  c.ID,
		Query: c.query,
	}
	if err := c.sendBroker(msg); err != nil {
		log.Errorf("Error? %v", errors.Cause(err))
	}
}

// call to resend subscription
func (c *Client) resubscribe() {
	c.Subscribe(c.query)
}

// Called externally, this adds new metadata to this client if it is acting
// as a publisher. If the client is not a publisher, this function runs but does
// not affect any part of the subscription operation.
// To DELETE a metadata key, use a value of nil for the key you want to delete. It
// will get folded into the next publish message, and then the keys will be removed
// from the local metadata map
func (c *Client) AddMetadata(pubid common.UUID, newm map[string]interface{}) {
	if len(newm) == 0 {
		return
	}
	c.publishersLock.Lock()
	if pub, found := c.publishers[pubid]; found {
		pub.AddMetadata(newm)
	}
	c.publishersLock.Unlock()
}

func (c *Client) AddPublisher(id common.UUID) *Publisher {
	pub := NewPublisher(id, c.BrokerAddress, c.CoordinatorAddress)
	c.publishersLock.Lock()
	c.publishers[id] = pub
	c.publishersLock.Unlock()
	return pub
}
