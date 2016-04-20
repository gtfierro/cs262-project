package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/nu7hatch/gouuid"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"runtime"
	"time"
)

type Server struct {
	address  *net.TCPAddr
	listener *net.TCPListener
	metadata *common.MetadataStore
	broker   *Broker
	// server signals on this channel when it has stopped
	stopped chan bool
	closed  bool

	// the address of the coordinator server
	coordinatorAddress *net.TCPAddr
	// connection to the coordinator server
	coordinatorConn *net.TCPConn
	// the number of seconds to wait before retrying to
	// contact coordinator server
	coordinatorRetryTime int
	// the maximum interval to increase to between attempts to contact
	// the coordinator server
	coordinatorRetryTimeMAX int
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

	// parse the config into an address
	coordinatorAddress := fmt.Sprintf("%s:%d", c.Server.CoordinatorHost, c.Server.CoordinatorPort)
	s.coordinatorAddress, err = net.ResolveTCPAddr("tcp", coordinatorAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"addr": coordinatorAddress, "error": err,
		}).Fatal("Could not resolve the generated TCP address")
	}

	s.metadata = common.NewMetadataStore(c)
	s.broker = NewBroker(s.metadata)
	s.stopped = make(chan bool)
	s.closed = false

	s.coordinatorRetryTime = 1
	s.coordinatorRetryTimeMAX = 60

	//go s.talkToManagement()

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
	s.broker.NewSubscription(string(query), conn)
}

func (s *Server) handlePublish(first *common.PublishMessage, dec *msgp.Reader, conn net.Conn) {
	s.broker.HandleProducer(first, dec, conn)
}

func (s *Server) talkToManagement() {
	var err error
	// Dial a connection to the Coordinator server
	s.coordinatorConn, err = net.DialTCP("tcp", nil, s.coordinatorAddress)
	// if connection fails, we do exponential retry
	for err != nil {
		log.WithFields(log.Fields{
			"err": err, "server": s.coordinatorAddress, "retry": s.coordinatorRetryTime,
		}).Error("Failed to contact coordinator server. Retrying")
		time.Sleep(time.Duration(s.coordinatorRetryTime) * time.Second)
		// increase retry window by factor of 2
		if s.coordinatorRetryTime*2 < s.coordinatorRetryTimeMAX {
			s.coordinatorRetryTime *= 2
		} else {
			s.coordinatorRetryTime = s.coordinatorRetryTimeMAX
		}
		// Dial a connection to the Coordinator server
		s.coordinatorConn, err = net.DialTCP("tcp", nil, s.coordinatorAddress)
	}
	// if we were successful, reset the wait timer
	s.coordinatorRetryTime = 1
	encoder := msgp.NewWriter(s.coordinatorConn)

	// when we come online, send the BrokerConnectMessage to inform the coordinator
	// server where it should send clients
	//TODO: how do we allocate MessageIDs?!?!?!?
	// these need to be globally unique. Prepend w/ some hash of broker?
	uuid, _ := uuid.NewV4()
	brokerID := uuid.String()
	// TODO should do something else for the BrokerID since we want it to persist after restarts
	bcm := &common.BrokerConnectMessage{BrokerInfo: common.BrokerInfo{
		BrokerID:   common.UUID(brokerID),
		BrokerAddr: s.address.String(),
	}, MessageIDStruct: common.GetMessageIDStruct()}
	bcm.Encode(encoder)
}
