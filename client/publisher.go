package client

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"net"
	"sync"
)

type Publisher struct {
	ID common.UUID

	brokerAddress    *net.TCPAddr
	brokerConn       *net.TCPConn
	brokerEncoder    *msgp.Writer
	brokerEncodeLock sync.Mutex

	coordAddress    *net.TCPAddr
	coordConn       *net.TCPConn
	coordEncoder    *msgp.Writer
	coordEncodeLock sync.Mutex

	// metadata for this client if it acts as a publisher
	metadataLock sync.RWMutex
	// if true, then there is metadata this client needs to send
	// to the broker in its next publish message
	dirtyMetadata bool
	metadata      map[string]interface{}
}

func NewPublisher(id common.UUID, brokerAddress, coordAddress *net.TCPAddr) *Publisher {
	var err error
	p := &Publisher{
		ID:            id,
		brokerAddress: brokerAddress,
		coordAddress:  coordAddress,
		metadata:      make(map[string]interface{}),
	}
	// TODO: failover
	// TODO: connect to coordinator
	if p.brokerConn, err = net.DialTCP("tcp", nil, brokerAddress); err != nil {
		log.Error(errors.Wrap(err, "Could not dial broker"))
	}
	p.brokerEncoder = msgp.NewWriter(p.brokerConn)

	return p
}

func (p *Publisher) SendBroker(m common.Sendable) error {
	p.brokerEncodeLock.Lock()
	defer p.brokerEncodeLock.Unlock()
	if err := m.Encode(p.brokerEncoder); err != nil {
		return errors.Wrap(err, "Could not encode message")
	}
	if err := p.brokerEncoder.Flush(); err != nil {
		return errors.Wrap(err, "Could not send message to broker")
	}
	return nil
}

func (p *Publisher) Publish(value interface{}) error {
	m := &common.PublishMessage{
		UUID:  "12345",
		Value: value,
	}
	p.metadataLock.RLock()
	if p.dirtyMetadata {
		m.Metadata = p.metadata
	}
	p.dirtyMetadata = false
	p.metadataLock.RUnlock()
	return p.SendBroker(m)
}

func (p *Publisher) AddMetadata(newm map[string]interface{}) {
	// TODO: do we need to do anything beyond a blind overwrite?
	p.metadataLock.Lock()
	for k, v := range newm {
		// deletes are handled by making v nil
		p.metadata[k] = v
	}
	p.dirtyMetadata = true
	p.metadataLock.Unlock()
}
