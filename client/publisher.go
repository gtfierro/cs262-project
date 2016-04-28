package client

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"net"
	"sync"
	"sync/atomic"
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

	sentMessagesAttempt uint32
	sentMessagesTotal   uint32
}

func NewPublisher(id common.UUID, brokerAddress, coordAddress *net.TCPAddr) *Publisher {
	p := &Publisher{
		ID:                  id,
		brokerAddress:       brokerAddress,
		coordAddress:        coordAddress,
		metadata:            make(map[string]interface{}),
		sentMessagesAttempt: 0,
		sentMessagesTotal:   0,
	}
	// TODO: failover
	// TODO: connect to coordinator
	for p.connectBroker(p.brokerAddress) != nil {
	}

	return p
}

// What is our failover technique?
// TODO
func (p *Publisher) doFailover() {
}

func (p *Publisher) connectBroker(address *net.TCPAddr) error {
	var err error
	if p.brokerConn, err = net.DialTCP("tcp", nil, address); err != nil {
		return errors.Wrap(err, "Could not dial broker")
	}
	p.brokerEncodeLock.Lock()
	p.brokerEncoder = msgp.NewWriter(p.brokerConn)
	p.brokerEncodeLock.Unlock()
	return nil
}

func (p *Publisher) sendBroker(m common.Sendable) error {
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
	atomic.AddUint32(&p.sentMessagesAttempt, 1)
	m := &common.PublishMessage{
		UUID:  p.ID,
		Value: value,
	}
	p.metadataLock.RLock()
	if p.dirtyMetadata {
		m.Metadata = p.metadata
	}
	p.dirtyMetadata = false
	p.metadataLock.RUnlock()
	// if we fail to send, then reconnect
	var err error
	if err = p.sendBroker(m); err != nil {
		p.doFailover()
	}
	atomic.AddUint32(&p.sentMessagesTotal, 1)
	return err
}

func (p *Publisher) GetStats() PublisherStats {
	stats := PublisherStats{
		MessagesAttempted:  atomic.LoadUint32(&p.sentMessagesAttempt),
		MessagesSuccessful: atomic.LoadUint32(&p.sentMessagesTotal),
	}
	return stats
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

type PublisherStats struct {
	MessagesAttempted  uint32
	MessagesSuccessful uint32
}
