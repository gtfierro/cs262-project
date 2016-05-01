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
	// the unique client identifier
	ID common.UUID
	// the query this client is subscribed to
	query string
	// buffer of messages to send out
	buffer chan common.Sendable
	// whether or not client is actively sending messages
	active  *AtomicBool
	timeout *time.Timer
	pause   chan bool
	// send on this channel when we die
	death   chan<- *Client
	encoder *msgp.Writer
}

func NewClient(query string, clientID common.UUID, conn *net.Conn, death chan<- *Client) *Client {
	return &Client{
		query:   query,
		ID:      clientID,
		conn:    conn,
		buffer:  make(chan common.Sendable, 1e6), // TODO buffer size?
		active:  &AtomicBool{0},
		timeout: time.NewTimer(10 * time.Second),
		// we don't want to block on sending/receiving the pause
		// signal because its handled in the same goroutine, so we buffer
		pause:   make(chan bool, 1),
		encoder: msgp.NewWriter(*conn),
		death:   death,
	}
}

// This creates a new client reference that represents an external broker.
// "address" should be in the form of "ip:port"
func ClientFromBrokerString(brokerID common.UUID, query, address string, death chan<- *Client) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.WithFields(log.Fields{
			"address": address, "error": err,
		}).Error("Could not dial remote broker")
		return nil, err
	}
	return NewClient(query, brokerID, &conn, death), nil
}

// queues a message to be sent
func (c *Client) Send(m common.Sendable) {
	// if our event loop isn't started, then start it
	if !c.active.Get() {
		c.active.Set(true)
		c.timeout.Reset(10 * time.Second)
		go c.dosend()
	}
	select {
	case c.buffer <- m: // if we have space in the buffer
	default: // drop it otherwise
		log.Debugf("Dropping %v", m)
	}
}

// tells this client to kill itself so that references
// to it can be safely removed
func (c *Client) Die() {
	c.timeout.Stop()
	(*c.conn).Close()
	c.death <- c
}

// This is the client event loop that handles forwarding messages from
// the broker to the client. It uses channels to communicate. This method
// is most likely going to run in a goroutine, so we need a way to terminate
// it and start it when the clients leave/appear again. This is explained below.
func (c *Client) dosend() {
	for {
		select {
		// When clients leave, we don't know that they have failed until we try to forward
		// them a message and it fails. If we don't do anything about this, then we will
		// leak goroutines for clients that leave and never come back. So, when the timeout
		// fires, we pause the client and send a signal on the pause channel
		case <-c.timeout.C:
			log.Info("Client timeout!")
			c.pause <- true
			// When we receive a message on the pause channel, we mark ourselves as inactive
			// and pause the timeout timer. We then return so that this goroutine is terminated.
		case <-c.pause:
			c.active.Set(false)
			c.timeout.Stop()
			return
			// When we receive a message on the incoming buffer, we we call the Encode() method
			// on it, and then make sure to Flush() the encoder to actually do the send.
			// TODO: there is an opportunity here to batch multiple messages into a single send,
			// by calling Encode() on multiple messages before doing the final flush to send
			// them over the network. This would be a good thing to look at if it turns out
			// during benchmarking that we are spending a lot of time on network write syscalls
		case m := <-c.buffer:
			// reset the timer to the new default value when we send
			c.timeout.Reset(10 * time.Second)
			//log.WithFields(log.Fields{
			//	"query": c.query, "message": m,
			//}).Debug("Forwarding message")
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
				// if we error out, stop the timer
				// to make sure that its goroutine stops, and then queue ourselves
				// for deletion by sending on the DEATH channel
				c.timeout.Stop()
				c.death <- c
				return
			}
		}
	}
}
