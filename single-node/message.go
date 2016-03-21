package main

import (
	"errors"
	"gopkg.in/vmihailenco/msgpack.v2"
)

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
