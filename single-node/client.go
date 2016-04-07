package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
	"time"
)

type Client struct {
	// the connection back to the client
	conn *net.Conn
	// the query this client is subscribed to
	query string
	// buffer of messages to send out
	buffer chan common.Sendable
	// whether or not client is actively sending messages
	active  bool
	timeout *time.Timer
	stop    chan bool
	// send on this channel when we die
	death   chan<- *Client
	encoder *msgp.Writer
}

func NewClient(query string, conn *net.Conn, death chan<- *Client) *Client {
	return &Client{
		query:   query,
		conn:    conn,
		buffer:  make(chan common.Sendable, 1e6), // TODO buffer size?
		timeout: time.NewTimer(10 * time.Second), // TODO config file
		active:  true,
		stop:    make(chan bool),
		encoder: msgp.NewWriter(*conn),
		death:   death,
	}
}

// queues a message to be sent
func (c *Client) Send(m common.Sendable) {
	if !c.active {
		go c.dosend()
	}
	select {
	case c.buffer <- m: // if we have space in the buffer
	default: // drop it otherwise
		log.Debugf("Dropping %v", m)
	}
}

func (c *Client) dosend() {
	for {
		select {
		case <-c.timeout.C:
			log.Info("client timeout")
			c.stop <- true
		case <-c.stop:
			c.active = false
			c.timeout.Stop()
			return
		case m := <-c.buffer:
			c.timeout.Reset(10 * time.Second)
			log.WithFields(log.Fields{
				"query": c.query, "message": m,
			}).Debug("Forwarding message")
			err := m.Encode(c.encoder)
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err,
					"message": m,
				}).Error("Error serializing message")
			}
			err = c.encoder.Flush()
			if err != nil {
				log.WithFields(log.Fields{
					"error":   err,
					"message": m,
				}).Error("Error sending Message to client!")
				c.active = false
				c.timeout.Stop()
				c.death <- c
				return
			}
		}
	}
}
