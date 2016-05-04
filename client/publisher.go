package client

import (
	"github.com/gtfierro/cs262-project/common"
	"sync"
	"sync/atomic"
)

type Publisher struct {
	BrokerConnection
	ID common.UUID

	// metadata for this client if it acts as a publisher
	metadataLock sync.RWMutex
	// if true, then there is metadata this client needs to send
	// to the broker in its next publish message
	dirtyMetadata bool
	metadata      map[string]interface{}

	sentMessagesAttempt uint32
	sentMessagesTotal   uint32
}

// publishLoop should be a loop which will start whenever a connection to a broker
// is established; it should exit whenever an error is encountered and it will be
// restarted once a reconnection is reestablished
func NewPublisher(id common.UUID, publishLoop func(*Publisher), cfg *Config) (p *Publisher, err error) {
	p = &Publisher{
		ID:                  id,
		metadata:            make(map[string]interface{}),
		sentMessagesAttempt: 0,
		sentMessagesTotal:   0,
	}
	connectCallback := func() {
		go publishLoop(p)
	}
	err = (&p.BrokerConnection).initialize(connectCallback, func(common.Sendable) {}, cfg)
	if err != nil {
		p = nil
		return
	}

	return
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
	if err := p.sendBroker(m); err != nil {
		return err
	}
	atomic.AddUint32(&p.sentMessagesTotal, 1)
	return nil
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
