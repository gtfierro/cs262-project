package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
)

type BrokerClient struct {
	// the connection back to the client
	conn *net.TCPConn
	// the unique client identifier
	ID common.UUID
	// buffer of messages to send out
	buffer  chan common.Sendable
	encoder *msgp.Writer
}

func NewBrokerClient(brokerID common.UUID, query, address string) (*BrokerClient, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.WithFields(log.Fields{
			"address": address, "error": err,
		}).Error("Could not dial remote broker")
		return nil, err
	}
	bc := &BrokerClient{
		ID:      brokerID,
		conn:    conn,
		buffer:  make(chan common.Sendable, 1e6),
		encoder: msgp.NewWriter(conn),
	}
	go bc.dosend()
	return bc, nil
}

// queues a message to be sent
func (bc *BrokerClient) Send(m common.Sendable) {
	select {
	case bc.buffer <- m: // if we have space in the buffer
	default: // drop it otherwise
		log.Debugf("Dropping %v", m)
	}
}

func (bc *BrokerClient) dosend() {
	for m := range bc.buffer {
		log.Warnf("Sending message %v to %v", m, (*bc.conn).RemoteAddr())
		err := m.Encode(bc.encoder)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"message": m,
			}).Error("Error serializing message")
		}
		err = bc.encoder.Flush()
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"message": m,
			}).Error("Error sending Message to client!")
			// if we error out, stop the timer
			// to make sure that its goroutine stops, and then queue ourselves
			// for deletion by sending on the DEATH channel
			return
		}
	}
}
