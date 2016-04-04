package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net"
)

type Client struct {
	// the connection back to the client
	conn *net.Conn
	// the query this client is subscribed to
	query string
	// buffer of messages to send out
	buffer chan common.Sendable
	// whether or not client is actively sending messages
	active bool
	stop   chan bool
	// send on this channel when we die
	death   chan<- *Client
	encoder *msgpack.Encoder
}

func NewClient(query string, conn *net.Conn, death chan<- *Client) *Client {
	return &Client{
		query:   query,
		conn:    conn,
		buffer:  make(chan common.Sendable, 10), // TODO buffer size?
		active:  true,
		stop:    make(chan bool),
		encoder: msgpack.NewEncoder(*conn),
		death:   death,
	}
}

// queues a message to be sent
func (c *Client) Send(m common.Sendable) {
	select {
	case c.buffer <- m: // if we have space in the buffer
	default: // drop it otherwise
		log.Debugf("Dropping %v", m)
	}
}

func (c *Client) dosend() {
	for {
		select {
		case <-c.stop:
			c.active = false
			break
		case m := <-c.buffer:
			log.WithFields(log.Fields{
				"query": c.query, "message": m,
			}).Debug("Forwarding message")
			err := m.EncodeMsgpack(c.encoder)
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err,
					"message": m,
				}).Error("Error sending Message to client!")
				c.death <- c
				return
			}
		}
	}
}
