package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	gouuid "github.com/nu7hatch/gouuid"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"runtime"
	"time"
)

type Server struct {
	address     *net.TCPAddr
	listener    *net.TCPListener
	metadata    *common.MetadataStore
	broker      Broker
	brokerID    common.UUID
	coordinator *Coordinator
	local       bool
	// server signals on this channel when it has stopped
	stopped chan bool
	closed  bool
}

// Create a new server instance using the configuration
func NewServer(c *common.Config) *Server {
	var (
		address string
		err     error
		s       = &Server{}
	)
	if c.Server.Global {
		address = fmt.Sprintf("0.0.0.0:%d", c.Server.Port)
	} else {
		address = fmt.Sprintf(":%d", c.Server.Port)
	}

	s.local = c.Server.LocalEvaluation

	// parse the config into an address
	s.address, err = net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.WithFields(log.Fields{
			"port": c.Server.Port, "global": c.Server.Global, "error": err.Error(),
		}).Fatal("Could not resolve the generated TCP address")
	}

	// listen on the address
	s.listener, err = net.ListenTCP("tcp", s.address)
	if err != nil {
		log.WithFields(log.Fields{
			"address": s.address, "error": err.Error(),
		}).Fatal("Could not listen on the provided address")
	}

	s.metadata = common.NewMetadataStore(c)
	s.coordinator = ConnectCoordinator(c.Server, s)
	s.broker = NewBroker(c.Server, s.metadata, s.coordinator)
	bid, _ := gouuid.NewV4()
	s.brokerID = common.UUID(bid.String())
	s.stopped = make(chan bool)
	s.closed = false

	// print up some server stats
	go func() {
		for {
			time.Sleep(5 * time.Second)
			log.Infof("Number of active goroutines %v", runtime.NumGoroutine())
		}
	}()
	return s
}

func (s *Server) stop() {
	log.Info("Stopping Server")
	s.closed = true
	s.listener.Close()
	time.Sleep(50 * time.Millisecond) // brief pause to let TCP close
	log.Info("Stopped Server")
	s.stopped <- true
}

// This method listens for incoming connections and handles them. It does NOT return
func (s *Server) listenAndDispatch() {
	var (
		conn net.Conn
		err  error
	)
	log.WithFields(log.Fields{
		"address": s.address,
	}).Info("Broker Listening!")

	// loop on the TCP connection and hand new connections to the dispatcher
	for {
		conn, err = s.listener.Accept()
		if err != nil {
			if s.closed {
				return // exit
			}
			log.WithFields(log.Fields{
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
	log.WithFields(log.Fields{
		"from": conn.RemoteAddr(),
	}).Debug("Got a new message!")

	var dec = msgp.NewReader(conn)
	msg, err := common.MessageFromDecoderMsgp(dec)
	if err == io.EOF {
		conn.Close()
		return
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "address": conn.RemoteAddr(),
		}).Error("Could not decode msgpack on connection. Closing!")
		// close connection
		conn.Close()
		return
	}
	switch m := msg.(type) {
	case *common.QueryMessage:
		s.handleSubscribe(*m, dec, conn)
	case *common.PublishMessage:
		s.handlePublish(m, dec, conn)
	default:
		log.WithField("message", msg).Warn("Server received unexpected message type!")
	}
}

// We have buffered the initial contents into bufio.Reader, so we pass that
// to this handler so we can finish decoding it. We also pass in the connection
// so that we can transmit back to the client.
func (s *Server) handleSubscribe(query common.QueryMessage, dec *msgp.Reader, conn net.Conn) {
	log.WithFields(log.Fields{
		"from": conn.RemoteAddr(), "query": query,
	}).Debug("Got a new Subscription!")

	if s.local {
		s.broker.NewSubscription(string(query), conn)
	} else {
		s.coordinator.forwardSubscription(query, conn)
	}
}

func (s *Server) handlePublish(first *common.PublishMessage, dec *msgp.Reader, conn net.Conn) {
	s.broker.HandleProducer(first, dec, conn)
}
