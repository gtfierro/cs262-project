package client

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"net"
	"sync"
	"time"
)

type SimulatedBroker struct {
	// the IP:Port of the coordinator that we talk to
	CoordinatorAddress *net.TCPAddr
	coordConn          *net.TCPConn
	coordEncoder       *msgp.Writer
	coordEncodeLock    sync.RWMutex

	// signals on this channel when it is done
	Stop chan bool

	uuid common.UUID

	connectCallback func(*SimulatedBroker)
	msgHandler      func(common.Sendable)
}

func NewSimulatedBroker(connectCallback func(*SimulatedBroker), msgHandler func(common.Sendable), id common.UUID,
	coordAddr string) (sb *SimulatedBroker, err error) {

	sb = new(SimulatedBroker)
	sb.connectCallback = connectCallback
	sb.msgHandler = msgHandler
	sb.Stop = make(chan bool)
	sb.uuid = id

	if sb.CoordinatorAddress, err = net.ResolveTCPAddr("tcp", coordAddr); err != nil {
		return
	}

	sb.connectCoordinator(false)
	return
}

func (bc *SimulatedBroker) Start() {
	go bc.listenCoordinator()
	bc.connectCallback(bc)
}

// after the duration expires, stop the client by signalling on c.Stop
func (bc *SimulatedBroker) StopIn(d time.Duration) {
	go func(c *SimulatedBroker) {
		time.Sleep(d)
		c.Stop <- true
	}(bc)
}

// Loop until we can finally connect to the coordinator.
// This blocks indefinitely until it is successful
func (bc *SimulatedBroker) connectCoordinator(startListener bool) {
	var (
		err      error
		waitTime = 500 * time.Millisecond
		maxWait  = 10 * time.Second
	)
	bc.coordEncodeLock.Lock()
	defer bc.coordEncodeLock.Unlock()
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
	bc.coordEncoder = msgp.NewWriter(bc.coordConn)
	if startListener {
		go bc.listenCoordinator()
		go bc.connectCallback(bc)
	}
}

func (bc *SimulatedBroker) listenCoordinator() {
	bc.coordEncodeLock.RLock()
	defer bc.coordEncodeLock.RUnlock()
	heartbeat := common.HeartbeatMessage{}
	reader := msgp.NewReader(net.Conn(bc.coordConn))
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		if err != nil {
			log.Warn("connection issue. Will attempt to reconnect")
			go bc.connectCoordinator(true)
			return
		}

		switch m := msg.(type) {
		case *common.RequestHeartbeatMessage:
			go bc.Send(&heartbeat)
		case *common.AcknowledgeMessage:
			// Ignore
		case common.SendableWithID:
			go bc.Send(&common.AcknowledgeMessage{MessageID: m.GetID()})
			bc.msgHandler(m)
		default:
			bc.msgHandler(m)
		}
	}
}

func (bc *SimulatedBroker) Send(m common.Sendable) error {
	bc.coordEncodeLock.RLock()
	defer bc.coordEncodeLock.RUnlock()
	if bc.coordEncoder == nil {
		return errors.Wrap(nil, "Nil encoder")
	}
	if err := m.Encode(bc.coordEncoder); err != nil {
		return errors.Wrap(err, "Could not encode message")
	}
	if err := bc.coordEncoder.Flush(); err != nil {
		return errors.Wrap(err, "Could not send message to coordinator")
	}
	return nil
}
