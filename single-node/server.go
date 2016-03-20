package main

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net"
)

type UUID string

type Server struct {
	address  *net.TCPAddr
	listener *net.TCPListener
	metadata *MetadataStore
	broker   *Broker
}

// Create a new server instance using the configuration
func NewServer(c *Config) *Server {
	var (
		address string
		err     error
		s       = &Server{}
	)
	if c.Server.Global {
		address = fmt.Sprintf("0.0.0.0:%d", *c.Server.Port)
	} else {
		address = fmt.Sprintf(":%d", *c.Server.Port)
	}

	// parse the config into an address
	s.address, err = net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.WithFields(logrus.Fields{
			"port": *c.Server.Port, "global": c.Server.Global, "error": err.Error(),
		}).Fatal("Could not resolve the generated TCP address")
	}

	// listen on the address
	s.listener, err = net.ListenTCP("tcp", s.address)
	if err != nil {
		log.WithFields(logrus.Fields{
			"address": s.address, "error": err.Error(),
		}).Fatal("Could not listen on the provided address")
	}

	s.metadata = NewMetadataStore(c)
	s.broker = NewBroker(s.metadata)
	return s
}

// This method listens for incoming connections and handles them. It does NOT return
func (s *Server) listenAndDispatch() {
	var (
		conn net.Conn
		err  error
	)
	log.WithFields(logrus.Fields{
		"address": s.address,
	}).Info("Broker Listening!")

	// loop on the TCP connection and hand new connections to the dispatcher
	for {
		conn, err = s.listener.Accept()
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Error accepting connection")
		}
		//TODO: revisit this. Do we use a worker pattern
		// 		like in https://nesv.github.io/golang/2014/02/25/worker-queues-in-go.html ?
		//		This needs benchmarking! But this is the simplest way to do it for now
		go s.dispatch(conn)
	}
}

// This method partial parses the message from the connection and sends it to the
// correct handler. Because of how MsgPack works and how we have set up the message
// structure, we can just look at the first byte of a connection:
// If it's a string, its a subscribe. If it's an array, its a publish.
func (s *Server) dispatch(conn net.Conn) {
	log.WithFields(logrus.Fields{
		"from": conn.RemoteAddr(),
	}).Debug("Got a new message!")
	var r = bufio.NewReader(conn)
	firstByte, err := r.Peek(1)
	if err != nil {
		log.WithFields(logrus.Fields{
			"address": conn.RemoteAddr(),
		}).Error("Closing connection because we couldn't read first byte")
		// close connection
		conn.Close()
		return
	}

	if isMsgPackArray(firstByte[0]) {
		s.handlePublish(r, conn)
	} else if isMsgPackString(firstByte[0]) {
		s.handleSubscribe(r, conn)
	} else {
		log.WithFields(logrus.Fields{
			"firstByte": firstByte[0],
		}).Error("Could not decode the packet. Not valid MsgPack!")
		conn.Close()
	}
}

// We have buffered the initial contents into bufio.Reader, so we pass that
// to this handler so we can finish decoding it. We also pass in the connection
// so that we can transmit back to the client.
func (s *Server) handleSubscribe(r *bufio.Reader, conn net.Conn) {
	var (
		query string
		err   error
		dec   = msgpack.NewDecoder(r)
	)
	// decode string
	query, err = dec.DecodeString()
	if err != nil && err.Error() != "EOF" {
		log.WithFields(logrus.Fields{
			"from": conn.RemoteAddr(), "error": err,
		}).Error("Could not decode subscription")
		conn.Close()
		return
	}

	log.WithFields(logrus.Fields{
		"from": conn.RemoteAddr(), "query": query,
	}).Debug("Got a new Subscription!")
	s.NewSubscription(query, conn)
}

func (s *Server) handlePublish(r *bufio.Reader, conn net.Conn) {
	var (
		dec = msgpack.NewDecoder(r)
		msg = new(Message)
		err error
	)

	// decode the incoming message
	err = msg.DecodeMsgpack(dec)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Could not decode message")
		conn.Close()
		return
	}

	// log that we got it (if DEBUG is on)
	log.WithFields(logrus.Fields{
		"from": conn.RemoteAddr(), "uuid": msg.UUID,
		"metadata": msg.Metadata, "value": msg.Value,
	}).Debug("Got a new Publish!")

	// save the metadata
	err = s.metadata.Save(msg)
	if err != nil {
		log.WithFields(logrus.Fields{
			"message": msg, "error": err,
		}).Error("Could not save metadata")
		conn.Close()
		return
	}

	//TODO: update subscriptions
	//TODO: forward to clients
}
