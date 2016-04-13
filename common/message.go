//go:generate msgp
package common

import (
	"errors"
	"fmt"
	"github.com/tinylib/msgp/msgp"
	"sync"
)

type UUID string

type Sendable interface {
	Encode(enc *msgp.Writer) error
}

type SendableWithID interface {
	Encode(enc *msgp.Writer) error
	GetID() uint32
}

type MessageIDStruct struct {
	MessageID uint32
}

func (sendable *MessageIDStruct) GetID() uint32 {
	return sendable.GetID()
}

type MessageType uint8

const (
	PUBLISHMSG MessageType = iota
	QUERYMSG
	SUBSCRIPDIFFMSG
	MATCHPRODMSG
	BROKERCONNECTMSG
)

type ProducerState uint

const (
	ProdStateOld ProducerState = iota
	ProdStateNew
	ProdStateSame
)

/////////////////////////////////////////////////////
/***************** Sent by Clients *****************/
/////////////////////////////////////////////////////

/***** BrokerRequestMessage *****/

// Sent from clients / publishers -> coordinator when they cannot contact their
// local/home broker
type BrokerRequestMessage struct {
	// this is the broker that clients are expecting
	// They send it to Coordinator so that when this broker comes back online,
	// it knows which clients to inform to reconnect
	// "ip:port"
	LocalBrokerAddr string
}

/***** QueryMessage *****/
// Client starts a query with this
type QueryMessage string

///////////////////////////////////////
/********** Publish Message **********/
///////////////////////////////////////

type PublishMessage struct {
	UUID     UUID
	Metadata map[string]interface{}
	Value    interface{}
	L        sync.RWMutex `msg:"-"`
}

func (m *PublishMessage) FromArray(array []interface{}) error {
	var (
		ok     bool
		uuid_s string // temporary for decoding
		tmpmap map[interface{}]interface{}
	)
	// check array length
	if len(array) != 3 {
		return errors.New("Length of publish array is not 3")
	}

	// decode UUID, should be first part of the slice
	if uuid_s, ok = array[0].(string); !ok {
		return errors.New("UUID in array[0] was not a string")
	}
	m.UUID = UUID(uuid_s)

	if tmpmap, ok = array[1].(map[interface{}]interface{}); !ok {
		return errors.New("Map in array[1] was not a map")
	} else if len(tmpmap) > 0 {
		m.Metadata = make(map[string]interface{})
		for k, v := range tmpmap {
			k_str, k_ok := k.(string)
			if !k_ok {
				return fmt.Errorf("Key in metadata was not a string (%v)", k)
			}
			m.Metadata[k_str] = v
		}
	}

	// no decoding of value
	m.Value = array[2]

	return nil
}

func (m *PublishMessage) IsEmpty() bool {
	return m.UUID == ""
}

//////////////////////////////////////
/***** SubscriptionDiff Message *****/
//////////////////////////////////////
type SubscriptionDiffMessage map[string][]UUID

func (m *SubscriptionDiffMessage) FromProducerState(state map[UUID]ProducerState) {
	(*m)["New"] = make([]UUID, len(state))
	i := 0
	for uuid, _ := range state {
		(*m)["New"][i] = uuid
		i += 1
	}
}

////////////////////////////////////
/***** ForwardRequest Message *****/
////////////////////////////////////

// Sent from coordinator -> brokers to tell the broker to create a forwarding route
// from one broker to another
type ForwardRequestMessage struct {
	MessageIDStruct
	// list of publishers whose messages should be forwarded
	PublisherList []UUID
	// the destination broker
	//TODO: need to allocate this on the broker somehow
	// ETK: or should we just refer to brokers by their ip:port string?
	BrokerID UUID
	// the query string which defines this forward request
	Query string
}

/////////////////////////////////////////////////////
/***************** Sent by COORDINATOR *****************/
/////////////////////////////////////////////////////

/***** CancelForwardRequest *****/

// Sent from coordinator -> brokers to cancel the forwarding route created by a
// ForwardRequest; used when clients cancel their subscription/disappear
type CancelForwardRequest struct {
	MessageIDStruct
	// the query that has been cancelled
	Query string
	// TODO: not sure why this is here?
	BrokerID UUID
}

/***** BrokerSubscriptionDiff Message *****/

// Analogous to SubscriptionDiffMessage, but used for internal comm., i.e. when
// coordinator notifies a broker to talk to its client
type BrokerSubscriptionDiffMessage map[string][]UUID

func (m *BrokerSubscriptionDiffMessage) FromProducerState(state map[UUID]ProducerState) {
	(*m)["New"] = make([]UUID, len(state))
	i := 0
	for uuid, _ := range state {
		(*m)["New"][i] = uuid
		i += 1
	}
}

/***** BrokerAssignmentMessage *****/

// Sent from coordinator -> clients/publishers to let them know which failover
// broker they should contact
type BrokerAssignmentMessage struct {
	// the address of the failover broker: "ip:port"
	BrokerAddr string
	// id of the failover broker
	BrokerID UUID
}

/***** BrokerDeathMessage *****/

// Sent from coordinator -> all brokers when it determines that a broker is
// offline, notifying other brokers they should stop attempting to forward to
// that broker
type BrokerDeathMessage struct {
	MessageIDStruct
	// addr of the failed broker "ip:port"
	BrokerAddr string
	// id of the failed broker
	BrokerID UUID
}

/***** ClientTerminationRequest *****/
// Sent from coordinator -> broker when coordinator wants the broker to break the
// connection with a specific client (i.e., when the broker is a
// failover and the local broker comes back online)
type ClientTerminationRequest struct {
	MessageIDStruct
	// "ip:port"
	ClientAddr string
}

/***** PublisherTerminationRequest *****/
// Sent from coordinator -> broker when coordinator wants the broker to break the
// connection with a specific publisher (i.e., when the broker is a
// failover and the local broker comes back online)
type PublisherTerminationRequest struct {
	MessageID   uint32
	PublisherID UUID
}

/***** HeartbeatMessage *****/
// Sent from coordinator -> broker every x seconds to ensure that the broker is still alive

type HeartbeatMessage struct {
	MessageIDStruct
}

/////////////////////////////////////////////////////
/***************** Sent by BROKER *****************/
/////////////////////////////////////////////////////

/***** BrokerPublishMessage *****/

// Analogous to PublishMessage, but used for internal communication, i.e. when
// a broker forwards a PublishMessage to another broker
type BrokerPublishMessage struct {
	UUID     UUID
	Metadata map[string]interface{}
	Value    interface{}
	L        sync.RWMutex `msg:"-"`
}

/***** ClientTermination Message *****/

// Sent from broker -> coordinator when a client connection / subscription is
// terminated
type ClientTerminationMessage struct {
	MessageIDStruct
	// the client that has left
	// "ip:port"
	ClientAddr string
}

/****** PublisherTermination Message *****/

// Sent from broker -> coordinator when a publisher connection is terminated
type PublisherTerminationMessage struct {
	MessageIDStruct
	// the publisher that has left
	PublisherID UUID
}

/***** BrokerConnectMessage *****/

// Sent from broker -> Coordinator whenever a broker comes online
type BrokerConnectMessage struct {
	MessageIDStruct
	// where incoming requests from clients/publishers should be routed to
	// "ip:port"
	BrokerAddr string
}

/***** BrokerTerminateMessage *****/

// Sent from broker -> coordinator if it is going offline permanently
type BrokerTerminateMessage struct {
	MessageID  uint32
	BrokerAddr string
	BrokerID   UUID
}

/////////////////////////////////
/***** Acknowledge Message *****/
/////////////////////////////////

// Used for communication between coordinator and brokers to confirm that a
// message was received. The sender should keep track of unacknowledged
// messages and remove them from some sort of buffer when an ack is received.
type AcknowledgeMessage struct {
	MessageIDStruct
}
