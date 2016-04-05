package main

import (
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Producer struct {
	ID common.UUID
	// decoder
	dec  *msgpack.Decoder
	C    chan *common.PublishMessage
	stop chan bool
}

func NewProducer(id common.UUID, decoder *msgpack.Decoder) *Producer {
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
	var (
		err error
		msg = new(common.PublishMessage)
	)
	for {
		if msgtype, err := p.dec.DecodeUint8(); common.MessageType(msgtype) != common.PUBLISHMSG || err != nil && err != io.EOF {
			log.WithFields(log.Fields{
				"error":   err,
				"msgtype": msgtype,
			}).Error("Error decoding incoming PublishMessage")
			return
		}

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
			log.Warn("Dropping incoming message from publisher")
			continue // drop if buffer is full
		}
	}
}
