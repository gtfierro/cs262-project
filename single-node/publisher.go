package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"io"
)

type Producer struct {
	ID common.UUID
	// decoder
	dec  *msgp.Reader
	C    chan *common.PublishMessage
	stop chan bool
}

func NewProducer(id common.UUID, decoder *msgp.Reader) *Producer {
	p := &Producer{
		ID:   id,
		dec:  decoder,
		C:    make(chan *common.PublishMessage, 100), // TODO buffer size?
		stop: make(chan bool),
	}

	go p.dorecv()

	return p
}

func (p *Producer) dorecv() {
	for {
		m, err := common.MessageFromDecoderMsgp(p.dec)

		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"msgtype": fmt.Sprintf("%T", m),
			}).Error("Error decoding incoming message")
			return
		}
		switch msg := m.(type) {
		case *common.PublishMessage:
			if err == io.EOF {
				p.stop <- true
				return // connection is closed
			}
			if msg.IsEmpty() {
				continue
			}
			select {
			case p.C <- msg: // buffer message
			default:
				//log.Warn("Dropping incoming message from publisher")
				continue // drop if buffer is full
			}
		default:
			continue
		}

	}
}
