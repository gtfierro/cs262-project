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
	sendLock           sync.Mutex

	// signals on this channel when it is done
	Stop chan bool

	uuid common.UUID

	connectCallback func(*SimulatedBroker)
	errorCallback   func()
	msgHandler      func(common.Sendable)
}

func NewSimulatedBroker(connectCallback func(*SimulatedBroker), errorCallback func(), msgHandler func(common.Sendable), id common.UUID,
	coordAddr string) (sb *SimulatedBroker, err error) {

	sb = new(SimulatedBroker)
	sb.connectCallback = connectCallback
	sb.errorCallback = errorCallback
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
	bc.sendConnectMessage()
	go bc.listenCoordinator()
	go bc.connectCallback(bc)
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
		maxWait  = 5 * time.Second
	)
	bc.coordEncodeLock.Lock()
	bc.coordConn, err = net.DialTCP("tcp", nil, bc.CoordinatorAddress)
	for err != nil {
		log.Warningf("Retrying coordinator connection to %v with delay %v (%s)", bc.CoordinatorAddress, waitTime, bc.uuid)
		time.Sleep(waitTime)
		waitTime *= 2
		if waitTime > maxWait {
			waitTime = maxWait
		}
		bc.coordConn, err = net.DialTCP("tcp", nil, bc.CoordinatorAddress)
	}
	log.Debugf("Connected to coordinator (%s)", bc.uuid)
	bc.coordEncoder = msgp.NewWriter(bc.coordConn)
	bc.coordEncodeLock.Unlock()
	if startListener {
		if err := bc.sendConnectMessage(); err != nil {
			go bc.connectCoordinator(true)
		} else {
			go bc.listenCoordinator()
			go bc.connectCallback(bc)
		}
	}
}

func (bc *SimulatedBroker) sendConnectMessage() error {
	connectMsg := &common.BrokerConnectMessage{
		MessageIDStruct: common.GetMessageIDStruct(),
		BrokerInfo: common.BrokerInfo{
			BrokerID:         bc.uuid,
			ClientBrokerAddr: "0.0.0.0:4444", // TODO
			CoordBrokerAddr:  "0.0.0.0:5505",
		},
	}
	if err := bc.Send(connectMsg); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (bc *SimulatedBroker) listenCoordinator() {
	bc.coordEncodeLock.RLock()
	defer bc.coordEncodeLock.RUnlock()
	heartbeat := common.HeartbeatMessage{}
	reader := msgp.NewReader(net.Conn(bc.coordConn))
	for {
		msg, err := common.MessageFromDecoderMsgp(reader)
		if err != nil {
			log.Warnf("connection issue. Will attempt to reconnect (%s)", bc.uuid)
			bc.errorCallback()
			go bc.connectCoordinator(true)
			return
		}

		switch m := msg.(type) {
		case *common.RequestHeartbeatMessage:
			go bc.Send(&heartbeat)
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
	bc.sendLock.Lock()
	defer bc.sendLock.Unlock()
	if bc.coordEncoder == nil {
		log.Infof("Null encoder when sending (%s)", bc.uuid)
		return errors.New("Nil encoder")
	}
	if err := m.Encode(bc.coordEncoder); err != nil {
		log.Infof("Error when encoding (%s)", bc.uuid)
		return errors.Wrap(err, "Could not encode message")
	}
	if err := bc.coordEncoder.Flush(); err != nil {
		log.Infof("Error when sending (%s)", bc.uuid)
		return errors.Wrap(err, "Could not send message to coordinator")
	}
	return nil
}
