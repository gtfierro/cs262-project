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

type QueryMessage string
type SubscriptionDiffMessage map[string][]UUID
type MatchingProducersMessage map[UUID]ProducerState

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

// another type of message
type ProducerList struct {
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
	case MATCHPRODMSG:
		mpm := make(MatchingProducersMessage)
		mpm.DecodeMsg(dec)
		return &mpm, err
	default:
		return nil, errors.New(fmt.Sprintf("MessageType unknown: %v", msgtype))
	}
}
