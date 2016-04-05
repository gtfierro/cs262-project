package main

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Publisher struct {
	BrokerURL  string
	BrokerPort int
	Metadata   map[string]interface{}
	uuid       common.UUID
	Frequency  int // per minute
	stop       chan bool
}

func (p *Publisher) publishContinuously() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", p.BrokerURL, p.BrokerPort))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("Error creating connection to %v:%v", p.BrokerURL, p.BrokerPort)
	}

	spacingMs := 60e3 / float64(p.Frequency)

	enc := msgpack.NewEncoder(conn)
	msg := common.PublishMessage{UUID: p.uuid, Metadata: p.Metadata, Value: time.Now().UnixNano()}
	msg.EncodeMsgpack(enc)
	msg.Metadata = nil // Only send metadata on initial connection
Loop:
	for {
		select {
		case <-p.stop:
			break Loop
		case <-time.After(time.Millisecond * time.Duration(spacingMs)):
			msg.Value = time.Now().UnixNano()
			msg.EncodeMsgpack(enc)
		}
	}
	conn.Close()
}
