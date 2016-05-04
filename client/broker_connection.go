package client

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"sync"
	"time"
)

type BrokerConnection struct {
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

	// signals on this channel when it is done
	Stop       chan bool
	brokerDead bool

	connectCallback func()
	msgHandler      func(common.Sendable)
}

func (bc *BrokerConnection) initialize(connectCallback func(), msgHandler func(common.Sendable), cfg *Config) (err error) {
	bc.brokerDead = true
	bc.connectCallback = connectCallback
	bc.msgHandler = msgHandler
	bc.Stop = make(chan bool)

	if bc.BrokerAddress, err = net.ResolveTCPAddr("tcp", cfg.BrokerAddress); err != nil {
		return
	}

	if err = bc.connectBroker(bc.BrokerAddress); err != nil {
		return
	}

	bc.CoordinatorAddress, err = net.ResolveTCPAddr("tcp", cfg.CoordinatorAddress)
	return
}

func (bc *BrokerConnection) Start() {
	go bc.listen()
	bc.connectCallback()
}

// after the duration expires, stop the client by signalling on c.Stop
func (bc *BrokerConnection) StopIn(d time.Duration) {
	go func(c *BrokerConnection) {
		time.Sleep(d)
		c.Stop <- true
	}(bc)
}

func (bc *BrokerConnection) listen() {
	reader := msgp.NewReader(net.Conn(bc.brokerConn))
	for {
		if bc.brokerDead {
			return
		}
		msg, err := common.MessageFromDecoderMsgp(reader)
		if err == io.EOF {
			log.Warn("connection closed. Do failover")
			bc.doFailover()
			break
		}
		if err != nil {
			log.Warn(errors.Wrap(err, "Could not decode message"))
		}

		bc.msgHandler(msg)
	}
}

// This function should contact the coordinator to get the new broker
func (bc *BrokerConnection) configureNewBroker(m *common.BrokerAssignmentMessage) {
	var err error
	bc.BrokerAddress, err = net.ResolveTCPAddr("tcp", m.ClientBrokerAddr)
	if err != nil {
		log.Critical(errors.Wrap(err, "Could not resolve local broker address"))
	}
	if err = bc.connectBroker(bc.BrokerAddress); err != nil {
		log.Critical(errors.Wrap(err, "Could not connect to local broker"))
	}
	// send our subscription once we're back
	bc.connectCallback()
	go bc.listen()
}

func (bc *BrokerConnection) connectBroker(address *net.TCPAddr) error {
	var err error
	if bc.brokerConn, err = net.DialTCP("tcp", nil, address); err != nil {
		return errors.Wrap(err, "Could not dial broker")
	}
	bc.brokerEncodeLock.Lock()
	bc.brokerEncoder = msgp.NewWriter(bc.brokerConn)
	bc.brokerEncodeLock.Unlock()
	bc.brokerDead = false
	return nil
}

// Sends message to the currently configured broker
func (bc *BrokerConnection) sendBroker(m common.Sendable) error {
	bc.brokerEncodeLock.Lock()
	defer bc.brokerEncodeLock.Unlock()
	if err := m.Encode(bc.brokerEncoder); err != nil {
		return errors.Wrap(err, "Could not encode message")
	}
	if err := bc.brokerEncoder.Flush(); err != nil {
		return errors.Wrap(err, "Could not send message to broker")
	}
	return nil
}

// This should be triggered when we can no longer contact our local broker. In this
// case, we sent a BrokerRequestMessage to the coordinator
func (bc *BrokerConnection) doFailover() {
	bc.brokerDead = true
	// establish the coordinator connection
	bc.connectCoordinator()
	// prepare the BrokerRequestMessage
	brm := &common.BrokerRequestMessage{
		LocalBrokerAddr: bc.BrokerAddress.String(),
		IsPublisher:     false,                                  // TODO
		UUID:            "392c1b18-0c37-11e6-b352-1002b58053c7", // TODO
	}
	// loop until we can contact the coordinator
	err := bc.sendCoordinator(brm)
	for err != nil {
		time.Sleep(1)
		err = bc.sendCoordinator(brm)
	}
}

// Loop until we can finally connect to the coordinator.
// This blocks indefinitely until it is successful
func (bc *BrokerConnection) connectCoordinator() {
	var (
		err      error
		waitTime = 1 * time.Second
		maxWait  = 30 * time.Second
	)
	bc.coordConn, err = net.DialTCP("tcp", nil, bc.CoordinatorAddress)
	for err != nil {
		log.Warningf("Retrying coordinator connection to %v with delay %v", bc.CoordinatorAddress, waitTime)
		time.Sleep(waitTime)
		waitTime *= 2
		if waitTime > maxWait {
			waitTime = maxWait
		}
		bc.coordConn, err = net.DialTCP("tcp", nil, bc.CoordinatorAddress)
	}
	log.Debug("Connected to coordinator")
	go bc.listenCoordinator()
	bc.coordEncoder = msgp.NewWriter(bc.coordConn)
}

func (bc *BrokerConnection) listenCoordinator() {
	if bc.coordConn == nil {
		return
	}
	reader := msgp.NewReader(net.Conn(bc.coordConn))
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		if err == io.EOF {
			log.Warn("connection closed. Do failover")
			bc.doFailover()
			break
		}
		if err != nil {
			log.Warn(errors.Wrap(err, "Could not decode message"))
		}

		log.Infof("Got %T message %v from coordinator", msg, msg)
		switch m := msg.(type) {
		case *common.BrokerAssignmentMessage:
			bc.configureNewBroker(m)
			bc.coordConn.Close()
			return
		default:
			log.Infof("Got %T message %v from coordinator", m, m)
		}
	}
}

func (bc *BrokerConnection) sendCoordinator(m common.Sendable) error {
	bc.coordEncodeLock.Lock()
	if err := m.Encode(bc.coordEncoder); err != nil {
		bc.coordEncodeLock.Unlock()
		return errors.Wrap(err, "Could not encode message")
	}
	if err := bc.coordEncoder.Flush(); err != nil {
		bc.coordEncodeLock.Unlock()
		bc.connectCoordinator()
		return errors.Wrap(err, "Could not send message to coordinator")
	}
	bc.coordEncodeLock.Unlock()
	return nil
}
