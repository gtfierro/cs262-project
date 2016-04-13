package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"net"
	"sync"
	"time"
)

type MessageHandler func(common.Sendable)

type Broker struct {
	address string
}

type BrokerConnectionManager struct {
	brokerMap         map[*Broker]*BrokerConnection
	mapLock           sync.RWMutex
	heartbeatInterval int // seconds
	clock             common.Clock
	//BrokerDeath       chan *Broker // send on this channel when a Broker can't be contacted
}

func NewBrokerConnectionManager(heartbeatInterval int) *BrokerConnectionManager {
	bcm := new(BrokerConnectionManager)
	bcm.brokerMap = make(map[*Broker]*BrokerConnection)
	bcm.mapLock = sync.RWMutex{}
	bcm.heartbeatInterval = heartbeatInterval
	bcm.clock = new(common.RealClock)
	//bcm.BrokerDeath = make(chan *Broker)
	return bcm
}

// If broker already exists in the mapping, this should be a reconnection
//   so simply update with the new connection and restart
// Otherwise, create a new BrokerConnection, put it into the map, and start it
func (bcm *BrokerConnectionManager) ConnectBroker(broker Broker, conn *net.TCPConn,
	messageHandler MessageHandler) (err error) {
	bcm.mapLock.RLock()
	existingBrokerConn, ok := bcm.brokerMap[&broker]
	bcm.mapLock.RUnlock()
	if ok {
		existingBrokerConn.WaitForCleanup()
		existingBrokerConn.StartAsynchronously(conn)
	} else {
		brokerConn := NewBrokerConnection(&broker, messageHandler,
			time.Duration(bcm.heartbeatInterval)*time.Second, bcm.clock)
		bcm.brokerMap[&broker] = brokerConn
		brokerConn.StartAsynchronously(conn)
	}
	return
}

// Terminate the BrokerConnection and remove from our map
func (bcm *BrokerConnectionManager) TerminateBroker(broker Broker) {
	bcm.mapLock.RLock()
	bconn, ok := bcm.brokerMap[&broker]
	bcm.mapLock.RUnlock()
	if !ok {
		log.WithField("broker", broker).Warn("Attempted to Terminate a nonexistent broker")
		return
	}
	bconn.Terminate()
	bcm.mapLock.Lock()
	delete(bcm.brokerMap, &broker)
	bcm.mapLock.Unlock()
}

// Asynchronously send a message to the given broker
func (bcm *BrokerConnectionManager) SendToBroker(broker Broker, message common.Sendable) (err error) {
	bcm.mapLock.RLock()
	defer bcm.mapLock.RUnlock()
	brokerConn, ok := bcm.brokerMap[&broker]
	if !ok {
		log.WithField("broker", broker).Warn("Attempted to send message to nonexistent broker")
		return errors.New("Broker was unable to be found")
	}
	brokerConn.Send(message)
	return
}

// Asynchronously send a message to all currently living  brokers
func (bcm *BrokerConnectionManager) BroadcastToBrokers(message common.Sendable) {
	bcm.mapLock.RLock()
	defer bcm.mapLock.RUnlock()
	for _, brokerConn := range bcm.brokerMap {
		brokerConn.Send(message)
	}
}

// Send a request for a heartbeat to broker; should be sent if contacted by a client
// saying that they couldn't contact the broker
func (bcm *BrokerConnectionManager) RequestHeartbeat(broker Broker) (err error) {
	bcm.mapLock.RLock()
	defer bcm.mapLock.RUnlock()
	brokerConn, ok := bcm.brokerMap[&broker]
	if !ok {
		log.WithField("broker", broker).Warn("Attempted to send heartbeat request to nonexistent broker")
		return errors.New("Broker was unable to be found")
	}
	brokerConn.RequestHeartbeat()
	return

}
