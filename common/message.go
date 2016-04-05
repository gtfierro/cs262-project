//go:generate msgp
package common

import (
	"errors"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type UUID string

type Sendable interface {
	EncodeMsgpack(enc *msgpack.Encoder) error
	DecodeMsgpack(dec *msgpack.Decoder) error
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

func MessageFromDecoder(dec *msgpack.Decoder) (Sendable, error) {
	msgtype, err := dec.DecodeUint8()
	if err != nil {
		return nil, err
	}
	switch MessageType(msgtype) {
	case PUBLISHMSG:
		msg := new(PublishMessage)
		err = msg.DecodeMsgpack(dec)
		return msg, err
	case QUERYMSG:
		qm := new(QueryMessage)
		qm.DecodeMsgpack(dec)
		return qm, err
	case SUBSCRIPDIFFMSG:
		sdm := make(SubscriptionDiffMessage)
		sdm.DecodeMsgpack(dec)
		return &sdm, err
	case MATCHPRODMSG:
		mpm := make(MatchingProducersMessage)
		mpm.DecodeMsgpack(dec)
		return &mpm, err
	default:
		return nil, errors.New(fmt.Sprintf("MessageType unknown: %v", msgtype))
	}
}

type QueryMessage struct {
	Query string
}
type SubscriptionDiffMessage map[string][]UUID
type MatchingProducersMessage map[UUID]ProducerState

func (m *QueryMessage) EncodeMsgpack(enc *msgpack.Encoder) error {
	if err := enc.EncodeUint8(uint8(QUERYMSG)); err != nil {
		return err
	}
	return enc.EncodeString(m.Query)
}

func (m *QueryMessage) DecodeMsgpack(dec *msgpack.Decoder) (err error) {
	m.Query, err = dec.DecodeString()
	if err != nil {
		return err
	}
	return nil
}

func (m *MatchingProducersMessage) EncodeMsgpack(enc *msgpack.Encoder) (err error) {
	if err := enc.EncodeUint8(uint8(MATCHPRODMSG)); err != nil {
		return err
	}
	if err := enc.EncodeMapLen(len(*m)); err != nil {
		return err
	}
	for k, v := range *m {
		if err = enc.EncodeString(string(k)); err != nil {
			return
		}
		if err = enc.EncodeUint(uint(v)); err != nil {
			return
		}
	}
	return
}

func (mpm *MatchingProducersMessage) DecodeMsgpack(dec *msgpack.Decoder) (err error) {
	var (
		len int
		key string
		val uint
	)
	if len, err = dec.DecodeMapLen(); err != nil {
		return
	}
	for i := 0; i < len; i++ {
		if key, err = dec.DecodeString(); err != nil {
			return
		}
		if val, err = dec.DecodeUint(); err != nil {
			return
		}
		(*mpm)[UUID(key)] = ProducerState(val)
	}
	return
}

func (m *SubscriptionDiffMessage) EncodeMsgpack(enc *msgpack.Encoder) (err error) {
	if err = enc.EncodeUint8(uint8(SUBSCRIPDIFFMSG)); err != nil {
		return
	}
	if err = enc.EncodeMapLen(len(*m)); err != nil {
		return
	}
	for k, v := range *m {
		if err = enc.EncodeString(string(k)); err != nil {
			return
		}
		if err = enc.EncodeArrayLen(len(v)); err != nil {
			return
		}
		for uuid := range v {
			if err = enc.EncodeString(string(uuid)); err != nil {
				return
			}
		}
	}
	return
}

func (sdm *SubscriptionDiffMessage) DecodeMsgpack(dec *msgpack.Decoder) (err error) {
	var (
		len  int
		key  string
		len2 int
		val  string
	)
	if len, err = dec.DecodeMapLen(); err != nil {
		return
	}
	for i := 0; i < len; i++ {
		if key, err = dec.DecodeString(); err != nil {
			return
		}
		if len2, err = dec.DecodeSliceLen(); err != nil {
			return
		}
		tmpslice := make([]UUID, len2)
		for i := 0; i < len2; i++ {
			if val, err = dec.DecodeString(); err != nil {
				return
			}
			tmpslice[i] = UUID(val)
		}
		(*sdm)[string(key)] = tmpslice
	}
	return
}

type PublishMessage struct {
	UUID     UUID
	Metadata map[string]interface{}
	Value    interface{}
}

func (m *PublishMessage) EncodeMsgpack(enc *msgpack.Encoder) error {
	if err := enc.EncodeUint8(uint8(PUBLISHMSG)); err != nil {
		return err
	}
	if err := enc.EncodeArrayLen(3); err != nil {
		return err
	}
	return enc.Encode(m.UUID, m.Metadata, m.Value)
}

func (m *PublishMessage) DecodeMsgpack(dec *msgpack.Decoder) error {
	var (
		err    error
		uuid_s string // temporary for decoding
		arrlen int
		maplen int
		key    string
		val    interface{}
	)
	// decode the actual slice
	arrlen, err = dec.DecodeSliceLen()
	if err != nil {
		return err
	}
	if arrlen != 3 {
		return errors.New("Length of publish array is not 3")
	}

	// decode UUID, should be first part of the slice
	if uuid_s, err = dec.DecodeString(); err != nil {
		return err
	}
	m.UUID = UUID(uuid_s)

	// decode metadata
	if maplen, err = dec.DecodeMapLen(); err != nil {
		return err
	}

	if maplen > 0 {
		m.Metadata = make(map[string]interface{})
	}

	for i := 0; i < maplen; i++ {
		if key, err = dec.DecodeString(); err != nil {
			return err
		}
		if val, err = dec.DecodeInterface(); err != nil {
			return err
		}
		m.Metadata[key] = val
	}

	// no decoding of value
	m.Value, err = dec.DecodeInterface()
	return err
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
