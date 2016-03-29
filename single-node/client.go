package main

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net"
)

type Client struct {
	// the connection back to the client
	conn *net.Conn
	// the query this client is subscribed to
	query string
	// buffer of messages to send out
	buffer chan interface{}
	// whether or not client is actively sending messages
	active  bool
	stop    chan bool
	encoder *msgpack.Encoder
}

func NewClient(query string, conn *net.Conn) *Client {
	return &Client{
		query:   query,
		conn:    conn,
		buffer:  make(chan interface{}, 10),
		active:  true,
		stop:    make(chan bool),
		encoder: msgpack.NewEncoder(*conn),
	}
}

// queues a message to be sent
func (c *Client) Send(m interface{}) {
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
			c.encoder.Encode(m)
		}
	}
}
