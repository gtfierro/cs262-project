package common

import (
	"errors"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type UUID string

type Sendable interface {
	EncodeMsgpack(enc *msgpack.Encoder)
}

type Message struct {
	UUID     UUID
	Metadata map[string]interface{}
	Value    interface{}
}

func (m *Message) DecodeMsgpack(dec *msgpack.Decoder) error {
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

func (m *Message) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode([]interface{}{m.UUID, m.Metadata, m.Value})
}

func (m *Message) FromArray(array []interface{}) error {
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

func (m *Message) IsEmpty() bool {
	return m.UUID == ""
}

// another type of message
type ProducerList struct {
}
