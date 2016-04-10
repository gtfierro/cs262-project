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
