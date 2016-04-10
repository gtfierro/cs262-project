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

type MessageType uint8

const (
	PUBLISHMSG MessageType = iota
	QUERYMSG
	SUBSCRIPDIFFMSG
	MATCHPRODMSG
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

// Sent from clients / publishers -> central when they cannot contact their
// local/home broker
type BrokerRequestMessage struct {
    // this is the broker that clients are expecting
    // They send it to Central so that when this broker comes back online,
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

func MessageFromDecoderMsgp(dec *msgp.Reader) (Sendable, error) {
	msgtype_tmp, err := dec.ReadByte()
	if err != nil {
		return nil, err
	}
	msgtype := uint8(msgtype_tmp)
	switch MessageType(msgtype) {
	case PUBLISHMSG:
		msg := new(PublishMessage)
		err = msg.DecodeMsg(dec)
		return msg, err
	case QUERYMSG:
		qm := new(QueryMessage)
		qm.DecodeMsg(dec)
		return qm, err
	case SUBSCRIPDIFFMSG:
		sdm := make(SubscriptionDiffMessage)
		sdm.DecodeMsg(dec)
		return &sdm, err
	default:
		return nil, errors.New(fmt.Sprintf("MessageType unknown: %v", msgtype))
	}
}

////////////////////////////////////
/***** ForwardRequest Message *****/
////////////////////////////////////

// Sent from central -> brokers to tell the broker to create a forwarding route
// from one broker to another
type ForwardRequestMessage struct {
	MessageID uint32
	// list of publishers whose messages should be forwarded
	PublisherList []UUID
	// the destination broker
	//TODO: need to allocate this on the broker somehow
	BrokerID UUID
	// the query string which defines this forward request
	Query string
}

/////////////////////////////////////////////////////
/***************** Sent by CENTRAL *****************/
/////////////////////////////////////////////////////


/***** CancelForwardRequest *****/

// Sent from central -> brokers to cancel the forwarding route created by a
// ForwardRequest; used when clients cancel their subscription/disappear
type CancelForwardRequest struct {
    MessageID uint32
    // the query that has been cancelled
    Query   string
    // TODO: not sure why this is here?
    BrokerID UUID
}

/***** BrokerSubscriptionDiff Message *****/

// Analogous to SubscriptionDiffMessage, but used for internal comm., i.e. when
// central notifies a broker to talk to its client
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

// Sent from central -> clients/publishers to let them know which failover
// broker they should contact
type BrokerAssignmentMessage struct {
    // the address of the failover broker: "ip:port"
    BrokerAddr  string
    // id of the failover broker
    BrokerID    UUID
}

/***** BrokerDeathMessage *****/

// Sent from central -> all brokers when it determines that a broker is
// offline, notifying other brokers they should stop attempting to forward to
// that broker 
type BrokerDeathMessage struct {
    MessageID   uint32
    // addr of the failed broker "ip:port"
    BrokerAddr string
    // id of the failed broker
    BrokerID    UUID
}

/***** ClientTerminationRequest *****/
// Sent from central -> broker when central wants the broker to break the
// connection with a specific client (i.e., when the broker is a
// failover and the local broker comes back online)
type ClientTerminationRequest struct {
    MessageID   uint32
    // "ip:port"
    ClientAddr  string
}

/***** PublisherTerminationRequest *****/
// Sent from central -> broker when central wants the broker to break the
// connection with a specific publisher (i.e., when the broker is a
// failover and the local broker comes back online)
type PublisherTerminationRequest struct {
    MessageID   uint32
    PublisherID  UUID
}

/***** HeartbeatMessage *****/
// Sent from central -> broker every x seconds to ensure that the broker is still alive

type HeartbeatMessage struct {
    MessageID uint32
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

// Sent from broker -> central when a client connection / subscription is
// terminated
type ClientTerminationMessage struct {
	MessageID uint32
    // the client that has left
    // "ip:port"
	ClientAddr    string
}

/****** PublisherTermination Message *****/

// Sent from broker -> central when a publisher connection is terminated
type PublisherTerminationMessage struct {
	MessageID   uint32
    // the publisher that has left
	PublisherID UUID
}

/***** BrokerConnectMessage *****/

// Sent from broker -> Central whenever a broker comes online
type BrokerConnectMessage struct {
    MessageID  uint32
    // where incoming requests from clients/publishers should be routed to
    // "ip:port"
    BrokerAddr string
}

/***** BrokerTerminateMessage *****/

// Sent from broker -> central if it is going offline permanently
type BrokerTerminateMessage struct {
    MessageID   uint32
    BrokerAddr string
    BerokerID   UUID
}


/////////////////////////////////
/***** Acknowledge Message *****/
/////////////////////////////////

// Used for communication between central and brokers to confirm that a
// message was received. The sender should keep track of unacknowledged
// messages and remove them from some sort of buffer when an ack is received.
type AcknowledgeMessage struct {
	MessageID uint32
}
