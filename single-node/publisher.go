package main

import (
	"github.com/Sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io"
)

type Producer struct {
	ID UUID
	// decoder
	dec  *msgpack.Decoder
	C    chan *Message
	stop chan bool
}

func NewProducer(id UUID, decoder *msgpack.Decoder) *Producer {
	p := &Producer{
		ID:   id,
		dec:  decoder,
		C:    make(chan *Message, 10),
		stop: make(chan bool),
	}

	go p.dorecv()

	return p
}

func (p *Producer) dorecv() {
	var (
		err error
		msg = new(Message)
	)
	for {
		err = msg.DecodeMsgpack(p.dec)
		if err == io.EOF {
			p.stop <- true
			return // connection is closed
		}
		if msg.isEmpty() {
			continue
		}
		if err != nil {
			log.WithFields(logrus.Fields{
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
