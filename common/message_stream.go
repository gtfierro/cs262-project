package common

import (
	"github.com/tinylib/msgp/msgp"
)

func (m *QueryMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(QUERYMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *MatchingProducersMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(MATCHPRODMSG))
	if err != nil {
		return err
	}
	err = enc.WriteArrayHeader(uint32(len(*m)))
	for uuid, _ := range *m {
		err = enc.WriteString(string(uuid))
	}
	return err
}

func (m *MatchingProducersMessage) DecodeMsg(dec *msgp.Reader) error {
	size, err := dec.ReadArrayHeader()
	if err != nil {
		return err
	}
	for i := 0; uint32(i) < size; i++ {
		uuid, err := dec.ReadString()
		if err != nil {
			return err
		}
		(*m)[UUID(uuid)] = ProdStateSame
	}
	return err
}

func (m *SubscriptionDiffMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(SUBSCRIPDIFFMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *PublishMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(PUBLISHMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}
