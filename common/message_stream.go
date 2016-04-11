package common

import (
	"errors"
	"fmt"
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

func (m *BrokerConnectMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERCONNECTMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
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
	case BROKERCONNECTMSG:
		bcm := new(BrokerConnectMessage)
		bcm.DecodeMsg(dec)
		return bcm, err
	default:
		return nil, errors.New(fmt.Sprintf("MessageType unknown: %v", msgtype))
	}
}
