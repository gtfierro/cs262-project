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

func (m *BrokerRequestMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERREQUESTMSG))
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

func (m *ForwardRequestMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(FORWARDREQUESTMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *CancelForwardRequest) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(CANCELFORWARDREQUESTMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *BrokerSubscriptionDiffMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERSUBSCRIPDIFFMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *BrokerAssignmentMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERASSIGNMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *BrokerDeathMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERDEATHMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *ClientTerminationRequest) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(CLIENTTERMREQUESTMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *PublisherTerminationRequest) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(PUBTERMREQUESTMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *HeartbeatMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(HEARTBEATMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *BrokerPublishMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERPUBLISHMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *ClientTerminationMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(CLIENTTERMMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *PublisherTerminationMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(PUBTERMMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *BrokerTerminateMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(BROKERTERMMSG))
	if err != nil {
		return err
	}
	return m.EncodeMsg(enc)
}

func (m *AcknowledgeMessage) Encode(enc *msgp.Writer) error {
	err := enc.WriteUint8(uint8(ACKMSG))
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
		msg := new(QueryMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case SUBSCRIPDIFFMSG:
		msg := make(SubscriptionDiffMessage)
		msg.DecodeMsg(dec)
		return &msg, err
	case BROKERREQUESTMSG:
		msg := new(BrokerRequestMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case BROKERCONNECTMSG:
		msg := new(BrokerConnectMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case FORWARDREQUESTMSG:
		msg := new(ForwardRequestMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case CANCELFORWARDREQUESTMSG:
		msg := new(CancelForwardRequest)
		msg.DecodeMsg(dec)
		return msg, err
	case BROKERSUBSCRIPDIFFMSG:
		msg := new(BrokerSubscriptionDiffMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case BROKERASSIGNMSG:
		msg := new(BrokerAssignmentMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case BROKERDEATHMSG:
		msg := new(BrokerDeathMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case CLIENTTERMREQUESTMSG:
		msg := new(ClientTerminationRequest)
		msg.DecodeMsg(dec)
		return msg, err
	case PUBTERMREQUESTMSG:
		msg := new(PublisherTerminationRequest)
		msg.DecodeMsg(dec)
		return msg, err
	case HEARTBEATMSG:
		msg := new(HeartbeatMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case BROKERPUBLISHMSG:
		msg := new(BrokerPublishMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case CLIENTTERMMSG:
		msg := new(ClientTerminationMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case PUBTERMMSG:
		msg := new(PublisherTerminationMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case BROKERTERMMSG:
		msg := new(BrokerTerminateMessage)
		msg.DecodeMsg(dec)
		return msg, err
	case ACKMSG:
		msg := new(AcknowledgeMessage)
		msg.DecodeMsg(dec)
		return msg, err
	default:
		return nil, errors.New(fmt.Sprintf("MessageType unknown: %v", msgtype))
	}
}
