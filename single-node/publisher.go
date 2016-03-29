package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
)

type Producer struct {
	ID common.UUID
	// decoder
	dec  *msgpack.Decoder
	C    chan *common.Message
	stop chan bool
}

func NewProducer(id common.UUID, decoder *msgpack.Decoder) *Producer {
	p := &Producer{
		ID:   id,
		dec:  decoder,
		C:    make(chan *common.Message, 10),
		stop: make(chan bool),
	}

	go p.dorecv()

	return p
}

func (p *Producer) dorecv() {
	var (
		err error
		msg = new(common.Message)
	)
	for {
		err = msg.DecodeMsgpack(p.dec)
		if err == io.EOF {
			p.stop <- true
			return // connection is closed
		}
		if msg.IsEmpty() {
			continue
		}
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error decoding msgpack producer")
		}
		select {
		case p.C <- msg: // buffer message
		default:
			continue // drop if buffer is full
		}
	}
}
