//go:generate msgp
package common

import (
	"errors"
	"fmt"
	"github.com/tinylib/msgp/msgp"
	"math/rand"
	"reflect"
	"sync"
)

type UUID string

type BrokerInfo struct {
	BrokerID         UUID
	ClientBrokerAddr string // "ip:port" to be contacted at
	CoordBrokerAddr  string // "ip:port" to be contacted at
}

func GetMessageType(msg Sendable) string {
	if msg == nil {
		return "nil"
	}
	return reflect.TypeOf(msg).Elem().Name()
}

type Sendable interface {
	Encode(enc *msgp.Writer) error
	Marshal() (o []byte, err error)
}

type Message interface {
	GetID() MessageIDType
}

type SendableWithID interface {
	Encode(enc *msgp.Writer) error
	Marshal() (o []byte, err error)
	GetID() MessageIDType
	SetID(MessageIDType)
	Copy() SendableWithID
}

type MessageIDType uint32

type MessageIDStruct struct {
	MessageID MessageIDType
}

func (sendable *MessageIDStruct) GetID() MessageIDType {
	return sendable.MessageID
}

func (sendable *MessageIDStruct) SetID(id MessageIDType) {
	sendable.MessageID = id
}

func GetMessageIDStruct() MessageIDStruct {
	return MessageIDStruct{MessageID: GetMessageID()} // TODO
}

func GetMessageID() MessageIDType {
	return MessageIDType(rand.Uint32())
}

type MessageType uint8

const (
	// External messages
	PUBLISHMSG MessageType = iota
	QUERYMSG
	SUBSCRIPDIFFMSG
	BROKERREQUESTMSG

	// Internal messages
	BROKERCONNECTMSG
	FORWARDREQUESTMSG
	CANCELFORWARDREQUESTMSG
	BROKERSUBSCRIPDIFFMSG
	BROKERASSIGNMSG
	BROKERDEATHMSG
	CLIENTTERMREQUESTMSG
	PUBTERMREQUESTMSG
	REQHEARTBEATMSG
	HEARTBEATMSG
	BROKERPUBLISHMSG
	BROKERQUERYMSG
	CLIENTTERMMSG
	PUBTERMMSG
	BROKERTERMMSG
	ACKMSG

	LEADERCHANGEMSG
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
	IsPublisher     bool // false if a client
	UUID            UUID // UUID of client or publisher
}

/***** QueryMessage *****/
// Client starts a query with this
type QueryMessage struct {
	UUID  UUID
	Query string
}

///////////////////////////////////////
/********** Publish Message **********/
///////////////////////////////////////

type PublishMessage struct {
	UUID     UUID
	Metadata map[string]interface{}
	Value    interface{}
	L        sync.RWMutex `msg:"-"`
}

func (m *PublishMessage) ToBroker() *BrokerPublishMessage {
	bpm := new(BrokerPublishMessage)
	bpm.FromPublishMessage(m)
	return bpm
}

func (m *PublishMessage) FromBroker(bpm *BrokerPublishMessage) {
	m.UUID = bpm.UUID
	m.Metadata = bpm.Metadata
	m.Value = bpm.Value
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
	BrokerInfo
	// the query string which defines this forward request
	Query string
}

func (m *ForwardRequestMessage) Copy() SendableWithID {
	return &ForwardRequestMessage{
		MessageIDStruct: m.MessageIDStruct,
		PublisherList:   m.PublisherList,
		BrokerInfo:      m.BrokerInfo,
		Query:           m.Query,
	}
}

/////////////////////////////////////////////////////
/***************** Sent by COORDINATOR *****************/
/////////////////////////////////////////////////////

/***** CancelForwardRequest *****/

// Sent from coordinator -> brokers to cancel the forwarding route created by a
// ForwardRequest; used when clients cancel their subscription/disappear
type CancelForwardRequest struct {
	MessageIDStruct
	// list of publishers whose messages should be cancelled
	PublisherList []UUID
	// the query that has been cancelled
	Query string
	// Necessary so you know who you need to stop forwarding to, since you
	// may be forwarding to multiple brokers for the same query
	BrokerInfo
}

func (m *CancelForwardRequest) Copy() SendableWithID {
	return &CancelForwardRequest{
		MessageIDStruct: m.MessageIDStruct,
		PublisherList:   m.PublisherList,
		Query:           m.Query,
		BrokerInfo:      m.BrokerInfo,
	}
}

/***** BrokerSubscriptionDiff Message *****/

// Analogous to SubscriptionDiffMessage, but used for internal comm., i.e. when
// coordinator notifies a broker to talk to its client
type BrokerSubscriptionDiffMessage struct {
	MessageIDStruct
	NewPublishers []UUID
	DelPublishers []UUID
	Query         string
}

func (m *BrokerSubscriptionDiffMessage) Copy() SendableWithID {
	return &BrokerSubscriptionDiffMessage{
		MessageIDStruct: m.MessageIDStruct,
		NewPublishers:   m.NewPublishers,
		DelPublishers:   m.DelPublishers,
		Query:           m.Query,
	}
}

func (m *BrokerSubscriptionDiffMessage) FromProducerState(state map[UUID]ProducerState) {
	m.NewPublishers = make([]UUID, len(state))
	i := 0
	for uuid, _ := range state {
		m.NewPublishers[i] = uuid
		i += 1
	}
}

/***** BrokerAssignmentMessage *****/

// Sent from coordinator -> clients/publishers to let them know which failover
// broker they should contact
type BrokerAssignmentMessage struct {
	// the ID and address of the failover broker: "ip:port"
	BrokerInfo
}

/***** BrokerDeathMessage *****/

// Sent from coordinator -> all brokers when it determines that a broker is
// offline, notifying other brokers they should stop attempting to forward to
// that broker
type BrokerDeathMessage struct {
	MessageIDStruct
	BrokerInfo
}

func (m *BrokerDeathMessage) Copy() SendableWithID {
	return &BrokerDeathMessage{
		MessageIDStruct: m.MessageIDStruct,
		BrokerInfo:      m.BrokerInfo,
	}
}

/***** ClientTerminationRequest *****/
// Sent from coordinator -> broker when coordinator wants the broker to break the
// connection with a specific client (i.e., when the broker is a
// failover and the local broker comes back online)
type ClientTerminationRequest struct {
	MessageIDStruct
	// "ip:port"
	ClientIDs []UUID
}

func (m *ClientTerminationRequest) Copy() SendableWithID {
	return &ClientTerminationRequest{
		MessageIDStruct: m.MessageIDStruct,
		ClientIDs:       m.ClientIDs,
	}
}

/***** PublisherTerminationRequest *****/
// Sent from coordinator -> broker when coordinator wants the broker to break the
// connection with a specific publisher (i.e., when the broker is a
// failover and the local broker comes back online)
type PublisherTerminationRequest struct {
	MessageIDStruct
	PublisherIDs []UUID
}

func (m *PublisherTerminationRequest) Copy() SendableWithID {
	return &PublisherTerminationRequest{
		MessageIDStruct: m.MessageIDStruct,
		PublisherIDs:    m.PublisherIDs,
	}
}

/***** RequestHeartbeatMessage *****/
// Sent from coordinator -> broker to request a heartbeat to confirm broker is alive

type RequestHeartbeatMessage struct{}

/////////////////////////////////////////////////////
/***************** Sent by BROKER *****************/
/////////////////////////////////////////////////////

/***** HeartbeatMessage *****/
// Sent from broker -> coordinator every x seconds to confirm that the broker is still alive

type HeartbeatMessage struct{}

/***** BrokerPublishMessage *****/

// Analogous to PublishMessage, but used for internal communication, i.e. when
// a broker forwards a PublishMessage to another broker
type BrokerPublishMessage struct {
	MessageIDStruct
	UUID     UUID
	Metadata map[string]interface{}
	Value    interface{}
	L        sync.RWMutex `msg:"-"`
}

func (m *BrokerPublishMessage) Copy() SendableWithID {
	return &BrokerPublishMessage{
		MessageIDStruct: m.MessageIDStruct,
		UUID:            m.UUID,
		Metadata:        m.Metadata,
		Value:           m.Value,
	}
}

func (m *BrokerPublishMessage) ToRegular() *PublishMessage {
	pm := new(PublishMessage)
	pm.FromBroker(m)
	return pm
}

func (m *BrokerPublishMessage) FromPublishMessage(pm *PublishMessage) {
	m.UUID = pm.UUID
	m.Metadata = pm.Metadata
	m.Value = pm.Value
	m.SetID(GetMessageID())
}

/***** BrokerQueryMessage ********/

// Analogous to QueryMessage, but used when sending messages
// from brokers to the coordinator so that it is possible
// to tell which client the query is attached to
type BrokerQueryMessage struct {
	MessageIDStruct
	Query string
	UUID  UUID
}

func (m *BrokerQueryMessage) Copy() SendableWithID {
	return &BrokerQueryMessage{
		MessageIDStruct: m.MessageIDStruct,
		Query:           m.Query,
		UUID:            m.UUID,
	}
}

/***** ClientTermination Message *****/

// Sent from broker -> coordinator when a client connection / subscription is
// terminated
type ClientTerminationMessage struct {
	MessageIDStruct
	// the client that has left
	ClientID UUID
}

func (m *ClientTerminationMessage) Copy() SendableWithID {
	return &ClientTerminationMessage{
		MessageIDStruct: m.MessageIDStruct,
		ClientID:        m.ClientID,
	}
}

/****** PublisherTermination Message *****/

// Sent from broker -> coordinator when a publisher connection is terminated
type PublisherTerminationMessage struct {
	MessageIDStruct
	// the publisher that has left
	PublisherID UUID
}

func (m *PublisherTerminationMessage) Copy() SendableWithID {
	return &PublisherTerminationMessage{
		MessageIDStruct: m.MessageIDStruct,
		PublisherID:     m.PublisherID,
	}
}

/***** BrokerConnectMessage *****/

// Sent from broker -> Coordinator whenever a broker comes online
type BrokerConnectMessage struct {
	MessageIDStruct
	BrokerInfo // its own ID and where incoming requests should be routed to
}

func (m *BrokerConnectMessage) Copy() SendableWithID {
	return &BrokerConnectMessage{
		MessageIDStruct: m.MessageIDStruct,
		BrokerInfo:      m.BrokerInfo,
	}
}

/***** BrokerTerminateMessage *****/

// Sent from broker -> coordinator if it is going offline permanently
type BrokerTerminateMessage struct {
	MessageIDStruct
}

func (m *BrokerTerminateMessage) Copy() SendableWithID {
	return &BrokerTerminateMessage{
		MessageIDStruct: m.MessageIDStruct,
	}
}

/////////////////////////////////
/***** Acknowledge Message *****/
/////////////////////////////////

// Used for communication between coordinator and brokers to confirm that a
// message was received. The sender should keep track of unacknowledged
// messages and remove them from some sort of buffer when an ack is received.
type AcknowledgeMessage struct {
	MessageID MessageIDType
}

func (m *AcknowledgeMessage) GetID() MessageIDType {
	return m.MessageID
}

///////////////////////////////////////
/***** Leadership Change Message *****/
///////////////////////////////////////

// Used for the log only to mark that a leadership change occurred
// at that point in the log
type LeaderChangeMessage struct{}
