package common

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *AcknowledgeMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageID":
			{
				var tmp uint32
				tmp, err = dc.ReadUint32()
				z.MessageID = MessageIDType(tmp)
			}
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z AcknowledgeMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageID))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z AcknowledgeMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageID))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AcknowledgeMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageID":
			{
				var tmp uint32
				tmp, bts, err = msgp.ReadUint32Bytes(bts)
				z.MessageID = MessageIDType(tmp)
			}
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z AcknowledgeMessage) Msgsize() (s int) {
	s = 1 + 10 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerAssignmentMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BrokerInfo":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, err = dc.ReadString()
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, err = dc.ReadString()
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *BrokerAssignmentMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "BrokerInfo"
	// map header, size 2
	// write "BrokerID"
	err = en.Append(0x81, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.BrokerInfo.BrokerID))
	if err != nil {
		return
	}
	// write "BrokerAddr"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.BrokerInfo.BrokerAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BrokerAssignmentMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "BrokerInfo"
	// map header, size 2
	// string "BrokerID"
	o = append(o, 0x81, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.BrokerInfo.BrokerID))
	// string "BrokerAddr"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.BrokerInfo.BrokerAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerAssignmentMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BrokerInfo":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, bts, err = msgp.ReadStringBytes(bts)
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *BrokerAssignmentMessage) Msgsize() (s int) {
	s = 1 + 11 + 1 + 9 + msgp.StringPrefixSize + len(string(z.BrokerInfo.BrokerID)) + 11 + msgp.StringPrefixSize + len(z.BrokerInfo.BrokerAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerConnectMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "BrokerInfo":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, err = dc.ReadString()
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, err = dc.ReadString()
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *BrokerConnectMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "BrokerInfo"
	// map header, size 2
	// write "BrokerID"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.BrokerInfo.BrokerID))
	if err != nil {
		return
	}
	// write "BrokerAddr"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.BrokerInfo.BrokerAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BrokerConnectMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "BrokerInfo"
	// map header, size 2
	// string "BrokerID"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.BrokerInfo.BrokerID))
	// string "BrokerAddr"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.BrokerInfo.BrokerAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerConnectMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "BrokerInfo":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, bts, err = msgp.ReadStringBytes(bts)
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *BrokerConnectMessage) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 11 + 1 + 9 + msgp.StringPrefixSize + len(string(z.BrokerInfo.BrokerID)) + 11 + msgp.StringPrefixSize + len(z.BrokerInfo.BrokerAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerDeathMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "BrokerInfo":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, err = dc.ReadString()
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, err = dc.ReadString()
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *BrokerDeathMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "BrokerInfo"
	// map header, size 2
	// write "BrokerID"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.BrokerInfo.BrokerID))
	if err != nil {
		return
	}
	// write "BrokerAddr"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.BrokerInfo.BrokerAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BrokerDeathMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "BrokerInfo"
	// map header, size 2
	// string "BrokerID"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.BrokerInfo.BrokerID))
	// string "BrokerAddr"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.BrokerInfo.BrokerAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerDeathMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "BrokerInfo":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, bts, err = msgp.ReadStringBytes(bts)
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *BrokerDeathMessage) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 11 + 1 + 9 + msgp.StringPrefixSize + len(string(z.BrokerInfo.BrokerID)) + 11 + msgp.StringPrefixSize + len(z.BrokerInfo.BrokerAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerInfo) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BrokerID":
			{
				var tmp string
				tmp, err = dc.ReadString()
				z.BrokerID = UUID(tmp)
			}
			if err != nil {
				return
			}
		case "BrokerAddr":
			z.BrokerAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z BrokerInfo) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "BrokerID"
	err = en.Append(0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.BrokerID))
	if err != nil {
		return
	}
	// write "BrokerAddr"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.BrokerAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z BrokerInfo) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "BrokerID"
	o = append(o, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.BrokerID))
	// string "BrokerAddr"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.BrokerAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerInfo) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BrokerID":
			{
				var tmp string
				tmp, bts, err = msgp.ReadStringBytes(bts)
				z.BrokerID = UUID(tmp)
			}
			if err != nil {
				return
			}
		case "BrokerAddr":
			z.BrokerAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z BrokerInfo) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(string(z.BrokerID)) + 11 + msgp.StringPrefixSize + len(z.BrokerAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerPublishMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "UUID":
			{
				var tmp string
				tmp, err = dc.ReadString()
				z.UUID = UUID(tmp)
			}
			if err != nil {
				return
			}
		case "Metadata":
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Metadata == nil && msz > 0 {
				z.Metadata = make(map[string]interface{}, msz)
			} else if len(z.Metadata) > 0 {
				for key, _ := range z.Metadata {
					delete(z.Metadata, key)
				}
			}
			for msz > 0 {
				msz--
				var xvk string
				var bzg interface{}
				xvk, err = dc.ReadString()
				if err != nil {
					return
				}
				bzg, err = dc.ReadIntf()
				if err != nil {
					return
				}
				z.Metadata[xvk] = bzg
			}
		case "Value":
			z.Value, err = dc.ReadIntf()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *BrokerPublishMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "UUID"
	err = en.Append(0x83, 0xa4, 0x55, 0x55, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.UUID))
	if err != nil {
		return
	}
	// write "Metadata"
	err = en.Append(0xa8, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Metadata)))
	if err != nil {
		return
	}
	for xvk, bzg := range z.Metadata {
		err = en.WriteString(xvk)
		if err != nil {
			return
		}
		err = en.WriteIntf(bzg)
		if err != nil {
			return
		}
	}
	// write "Value"
	err = en.Append(0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteIntf(z.Value)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BrokerPublishMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "UUID"
	o = append(o, 0x83, 0xa4, 0x55, 0x55, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.UUID))
	// string "Metadata"
	o = append(o, 0xa8, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.Metadata)))
	for xvk, bzg := range z.Metadata {
		o = msgp.AppendString(o, xvk)
		o, err = msgp.AppendIntf(o, bzg)
		if err != nil {
			return
		}
	}
	// string "Value"
	o = append(o, 0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	o, err = msgp.AppendIntf(o, z.Value)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerPublishMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "UUID":
			{
				var tmp string
				tmp, bts, err = msgp.ReadStringBytes(bts)
				z.UUID = UUID(tmp)
			}
			if err != nil {
				return
			}
		case "Metadata":
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Metadata == nil && msz > 0 {
				z.Metadata = make(map[string]interface{}, msz)
			} else if len(z.Metadata) > 0 {
				for key, _ := range z.Metadata {
					delete(z.Metadata, key)
				}
			}
			for msz > 0 {
				var xvk string
				var bzg interface{}
				msz--
				xvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				bzg, bts, err = msgp.ReadIntfBytes(bts)
				if err != nil {
					return
				}
				z.Metadata[xvk] = bzg
			}
		case "Value":
			z.Value, bts, err = msgp.ReadIntfBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *BrokerPublishMessage) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(string(z.UUID)) + 9 + msgp.MapHeaderSize
	if z.Metadata != nil {
		for xvk, bzg := range z.Metadata {
			_ = bzg
			s += msgp.StringPrefixSize + len(xvk) + msgp.GuessSize(bzg)
		}
	}
	s += 6 + msgp.GuessSize(z.Value)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerQueryMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "QueryMessage":
			z.QueryMessage, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ClientAddr":
			z.ClientAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z BrokerQueryMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "QueryMessage"
	err = en.Append(0x82, 0xac, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.QueryMessage)
	if err != nil {
		return
	}
	// write "ClientAddr"
	err = en.Append(0xaa, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z BrokerQueryMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "QueryMessage"
	o = append(o, 0x82, 0xac, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	o = msgp.AppendString(o, z.QueryMessage)
	// string "ClientAddr"
	o = append(o, 0xaa, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.ClientAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerQueryMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "QueryMessage":
			z.QueryMessage, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ClientAddr":
			z.ClientAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z BrokerQueryMessage) Msgsize() (s int) {
	s = 1 + 13 + msgp.StringPrefixSize + len(z.QueryMessage) + 11 + msgp.StringPrefixSize + len(z.ClientAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerRequestMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "LocalBrokerAddr":
			z.LocalBrokerAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z BrokerRequestMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "LocalBrokerAddr"
	err = en.Append(0x81, 0xaf, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.LocalBrokerAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z BrokerRequestMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "LocalBrokerAddr"
	o = append(o, 0x81, 0xaf, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.LocalBrokerAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerRequestMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "LocalBrokerAddr":
			z.LocalBrokerAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z BrokerRequestMessage) Msgsize() (s int) {
	s = 1 + 16 + msgp.StringPrefixSize + len(z.LocalBrokerAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerSubscriptionDiffMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "NewPublishers":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.NewPublishers) >= int(xsz) {
				z.NewPublishers = z.NewPublishers[:xsz]
			} else {
				z.NewPublishers = make([]UUID, xsz)
			}
			for bai := range z.NewPublishers {
				{
					var tmp string
					tmp, err = dc.ReadString()
					z.NewPublishers[bai] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "DelPublishers":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.DelPublishers) >= int(xsz) {
				z.DelPublishers = z.DelPublishers[:xsz]
			} else {
				z.DelPublishers = make([]UUID, xsz)
			}
			for cmr := range z.DelPublishers {
				{
					var tmp string
					tmp, err = dc.ReadString()
					z.DelPublishers[cmr] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "Query":
			z.Query, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *BrokerSubscriptionDiffMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "NewPublishers"
	err = en.Append(0x83, 0xad, 0x4e, 0x65, 0x77, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.NewPublishers)))
	if err != nil {
		return
	}
	for bai := range z.NewPublishers {
		err = en.WriteString(string(z.NewPublishers[bai]))
		if err != nil {
			return
		}
	}
	// write "DelPublishers"
	err = en.Append(0xad, 0x44, 0x65, 0x6c, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.DelPublishers)))
	if err != nil {
		return
	}
	for cmr := range z.DelPublishers {
		err = en.WriteString(string(z.DelPublishers[cmr]))
		if err != nil {
			return
		}
	}
	// write "Query"
	err = en.Append(0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Query)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BrokerSubscriptionDiffMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "NewPublishers"
	o = append(o, 0x83, 0xad, 0x4e, 0x65, 0x77, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.NewPublishers)))
	for bai := range z.NewPublishers {
		o = msgp.AppendString(o, string(z.NewPublishers[bai]))
	}
	// string "DelPublishers"
	o = append(o, 0xad, 0x44, 0x65, 0x6c, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.DelPublishers)))
	for cmr := range z.DelPublishers {
		o = msgp.AppendString(o, string(z.DelPublishers[cmr]))
	}
	// string "Query"
	o = append(o, 0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	o = msgp.AppendString(o, z.Query)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerSubscriptionDiffMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "NewPublishers":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.NewPublishers) >= int(xsz) {
				z.NewPublishers = z.NewPublishers[:xsz]
			} else {
				z.NewPublishers = make([]UUID, xsz)
			}
			for bai := range z.NewPublishers {
				{
					var tmp string
					tmp, bts, err = msgp.ReadStringBytes(bts)
					z.NewPublishers[bai] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "DelPublishers":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.DelPublishers) >= int(xsz) {
				z.DelPublishers = z.DelPublishers[:xsz]
			} else {
				z.DelPublishers = make([]UUID, xsz)
			}
			for cmr := range z.DelPublishers {
				{
					var tmp string
					tmp, bts, err = msgp.ReadStringBytes(bts)
					z.DelPublishers[cmr] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "Query":
			z.Query, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *BrokerSubscriptionDiffMessage) Msgsize() (s int) {
	s = 1 + 14 + msgp.ArrayHeaderSize
	for bai := range z.NewPublishers {
		s += msgp.StringPrefixSize + len(string(z.NewPublishers[bai]))
	}
	s += 14 + msgp.ArrayHeaderSize
	for cmr := range z.DelPublishers {
		s += msgp.StringPrefixSize + len(string(z.DelPublishers[cmr]))
	}
	s += 6 + msgp.StringPrefixSize + len(z.Query)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BrokerTerminateMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *BrokerTerminateMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x81, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BrokerTerminateMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x81, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BrokerTerminateMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *BrokerTerminateMessage) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CancelForwardRequest) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "PublisherList":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.PublisherList) >= int(xsz) {
				z.PublisherList = z.PublisherList[:xsz]
			} else {
				z.PublisherList = make([]UUID, xsz)
			}
			for ajw := range z.PublisherList {
				{
					var tmp string
					tmp, err = dc.ReadString()
					z.PublisherList[ajw] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "Query":
			z.Query, err = dc.ReadString()
			if err != nil {
				return
			}
		case "BrokerInfo":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, err = dc.ReadString()
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, err = dc.ReadString()
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *CancelForwardRequest) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x84, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "PublisherList"
	err = en.Append(0xad, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.PublisherList)))
	if err != nil {
		return
	}
	for ajw := range z.PublisherList {
		err = en.WriteString(string(z.PublisherList[ajw]))
		if err != nil {
			return
		}
	}
	// write "Query"
	err = en.Append(0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Query)
	if err != nil {
		return
	}
	// write "BrokerInfo"
	// map header, size 2
	// write "BrokerID"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.BrokerInfo.BrokerID))
	if err != nil {
		return
	}
	// write "BrokerAddr"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.BrokerInfo.BrokerAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CancelForwardRequest) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x84, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "PublisherList"
	o = append(o, 0xad, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74)
	o = msgp.AppendArrayHeader(o, uint32(len(z.PublisherList)))
	for ajw := range z.PublisherList {
		o = msgp.AppendString(o, string(z.PublisherList[ajw]))
	}
	// string "Query"
	o = append(o, 0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	o = msgp.AppendString(o, z.Query)
	// string "BrokerInfo"
	// map header, size 2
	// string "BrokerID"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.BrokerInfo.BrokerID))
	// string "BrokerAddr"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.BrokerInfo.BrokerAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CancelForwardRequest) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "PublisherList":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.PublisherList) >= int(xsz) {
				z.PublisherList = z.PublisherList[:xsz]
			} else {
				z.PublisherList = make([]UUID, xsz)
			}
			for ajw := range z.PublisherList {
				{
					var tmp string
					tmp, bts, err = msgp.ReadStringBytes(bts)
					z.PublisherList[ajw] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "Query":
			z.Query, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "BrokerInfo":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, bts, err = msgp.ReadStringBytes(bts)
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *CancelForwardRequest) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 14 + msgp.ArrayHeaderSize
	for ajw := range z.PublisherList {
		s += msgp.StringPrefixSize + len(string(z.PublisherList[ajw]))
	}
	s += 6 + msgp.StringPrefixSize + len(z.Query) + 11 + 1 + 9 + msgp.StringPrefixSize + len(string(z.BrokerInfo.BrokerID)) + 11 + msgp.StringPrefixSize + len(z.BrokerInfo.BrokerAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ClientTerminationMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "ClientAddr":
			z.ClientAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ClientTerminationMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "ClientAddr"
	err = en.Append(0xaa, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ClientTerminationMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "ClientAddr"
	o = append(o, 0xaa, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.ClientAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ClientTerminationMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "ClientAddr":
			z.ClientAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *ClientTerminationMessage) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 11 + msgp.StringPrefixSize + len(z.ClientAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ClientTerminationRequest) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "ClientAddr":
			z.ClientAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ClientTerminationRequest) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "ClientAddr"
	err = en.Append(0xaa, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientAddr)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ClientTerminationRequest) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "ClientAddr"
	o = append(o, 0xaa, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.ClientAddr)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ClientTerminationRequest) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "ClientAddr":
			z.ClientAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *ClientTerminationRequest) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 11 + msgp.StringPrefixSize + len(z.ClientAddr)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ForwardRequestMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "PublisherList":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.PublisherList) >= int(xsz) {
				z.PublisherList = z.PublisherList[:xsz]
			} else {
				z.PublisherList = make([]UUID, xsz)
			}
			for wht := range z.PublisherList {
				{
					var tmp string
					tmp, err = dc.ReadString()
					z.PublisherList[wht] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "BrokerInfo":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, err = dc.ReadString()
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, err = dc.ReadString()
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "Query":
			z.Query, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ForwardRequestMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x84, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "PublisherList"
	err = en.Append(0xad, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.PublisherList)))
	if err != nil {
		return
	}
	for wht := range z.PublisherList {
		err = en.WriteString(string(z.PublisherList[wht]))
		if err != nil {
			return
		}
	}
	// write "BrokerInfo"
	// map header, size 2
	// write "BrokerID"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.BrokerInfo.BrokerID))
	if err != nil {
		return
	}
	// write "BrokerAddr"
	err = en.Append(0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.BrokerInfo.BrokerAddr)
	if err != nil {
		return
	}
	// write "Query"
	err = en.Append(0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Query)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ForwardRequestMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x84, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "PublisherList"
	o = append(o, 0xad, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74)
	o = msgp.AppendArrayHeader(o, uint32(len(z.PublisherList)))
	for wht := range z.PublisherList {
		o = msgp.AppendString(o, string(z.PublisherList[wht]))
	}
	// string "BrokerInfo"
	// map header, size 2
	// string "BrokerID"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x82, 0xa8, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.BrokerInfo.BrokerID))
	// string "BrokerAddr"
	o = append(o, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.BrokerInfo.BrokerAddr)
	// string "Query"
	o = append(o, 0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	o = msgp.AppendString(o, z.Query)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ForwardRequestMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "PublisherList":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.PublisherList) >= int(xsz) {
				z.PublisherList = z.PublisherList[:xsz]
			} else {
				z.PublisherList = make([]UUID, xsz)
			}
			for wht := range z.PublisherList {
				{
					var tmp string
					tmp, bts, err = msgp.ReadStringBytes(bts)
					z.PublisherList[wht] = UUID(tmp)
				}
				if err != nil {
					return
				}
			}
		case "BrokerInfo":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "BrokerID":
					{
						var tmp string
						tmp, bts, err = msgp.ReadStringBytes(bts)
						z.BrokerInfo.BrokerID = UUID(tmp)
					}
					if err != nil {
						return
					}
				case "BrokerAddr":
					z.BrokerInfo.BrokerAddr, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "Query":
			z.Query, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *ForwardRequestMessage) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 14 + msgp.ArrayHeaderSize
	for wht := range z.PublisherList {
		s += msgp.StringPrefixSize + len(string(z.PublisherList[wht]))
	}
	s += 11 + 1 + 9 + msgp.StringPrefixSize + len(string(z.BrokerInfo.BrokerID)) + 11 + msgp.StringPrefixSize + len(z.BrokerInfo.BrokerAddr) + 6 + msgp.StringPrefixSize + len(z.Query)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MessageIDStruct) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageID":
			{
				var tmp uint32
				tmp, err = dc.ReadUint32()
				z.MessageID = MessageIDType(tmp)
			}
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MessageIDStruct) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageID))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MessageIDStruct) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageID))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MessageIDStruct) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageID":
			{
				var tmp uint32
				tmp, bts, err = msgp.ReadUint32Bytes(bts)
				z.MessageID = MessageIDType(tmp)
			}
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z MessageIDStruct) Msgsize() (s int) {
	s = 1 + 10 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MessageIDType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var tmp uint32
		tmp, err = dc.ReadUint32()
		(*z) = MessageIDType(tmp)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MessageIDType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteUint32(uint32(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MessageIDType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendUint32(o, uint32(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MessageIDType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var tmp uint32
		tmp, bts, err = msgp.ReadUint32Bytes(bts)
		(*z) = MessageIDType(tmp)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z MessageIDType) Msgsize() (s int) {
	s = msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MessageType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var tmp uint8
		tmp, err = dc.ReadUint8()
		(*z) = MessageType(tmp)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MessageType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteUint8(uint8(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MessageType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendUint8(o, uint8(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MessageType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var tmp uint8
		tmp, bts, err = msgp.ReadUint8Bytes(bts)
		(*z) = MessageType(tmp)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z MessageType) Msgsize() (s int) {
	s = msgp.Uint8Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ProducerState) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var tmp uint
		tmp, err = dc.ReadUint()
		(*z) = ProducerState(tmp)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ProducerState) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteUint(uint(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ProducerState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendUint(o, uint(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ProducerState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var tmp uint
		tmp, bts, err = msgp.ReadUintBytes(bts)
		(*z) = ProducerState(tmp)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z ProducerState) Msgsize() (s int) {
	s = msgp.UintSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PublishMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "UUID":
			{
				var tmp string
				tmp, err = dc.ReadString()
				z.UUID = UUID(tmp)
			}
			if err != nil {
				return
			}
		case "Metadata":
			var msz uint32
			msz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Metadata == nil && msz > 0 {
				z.Metadata = make(map[string]interface{}, msz)
			} else if len(z.Metadata) > 0 {
				for key, _ := range z.Metadata {
					delete(z.Metadata, key)
				}
			}
			for msz > 0 {
				msz--
				var hct string
				var cua interface{}
				hct, err = dc.ReadString()
				if err != nil {
					return
				}
				cua, err = dc.ReadIntf()
				if err != nil {
					return
				}
				z.Metadata[hct] = cua
			}
		case "Value":
			z.Value, err = dc.ReadIntf()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PublishMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "UUID"
	err = en.Append(0x83, 0xa4, 0x55, 0x55, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.UUID))
	if err != nil {
		return
	}
	// write "Metadata"
	err = en.Append(0xa8, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Metadata)))
	if err != nil {
		return
	}
	for hct, cua := range z.Metadata {
		err = en.WriteString(hct)
		if err != nil {
			return
		}
		err = en.WriteIntf(cua)
		if err != nil {
			return
		}
	}
	// write "Value"
	err = en.Append(0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteIntf(z.Value)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PublishMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "UUID"
	o = append(o, 0x83, 0xa4, 0x55, 0x55, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.UUID))
	// string "Metadata"
	o = append(o, 0xa8, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.Metadata)))
	for hct, cua := range z.Metadata {
		o = msgp.AppendString(o, hct)
		o, err = msgp.AppendIntf(o, cua)
		if err != nil {
			return
		}
	}
	// string "Value"
	o = append(o, 0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	o, err = msgp.AppendIntf(o, z.Value)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PublishMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "UUID":
			{
				var tmp string
				tmp, bts, err = msgp.ReadStringBytes(bts)
				z.UUID = UUID(tmp)
			}
			if err != nil {
				return
			}
		case "Metadata":
			var msz uint32
			msz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Metadata == nil && msz > 0 {
				z.Metadata = make(map[string]interface{}, msz)
			} else if len(z.Metadata) > 0 {
				for key, _ := range z.Metadata {
					delete(z.Metadata, key)
				}
			}
			for msz > 0 {
				var hct string
				var cua interface{}
				msz--
				hct, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				cua, bts, err = msgp.ReadIntfBytes(bts)
				if err != nil {
					return
				}
				z.Metadata[hct] = cua
			}
		case "Value":
			z.Value, bts, err = msgp.ReadIntfBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *PublishMessage) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(string(z.UUID)) + 9 + msgp.MapHeaderSize
	if z.Metadata != nil {
		for hct, cua := range z.Metadata {
			_ = cua
			s += msgp.StringPrefixSize + len(hct) + msgp.GuessSize(cua)
		}
	}
	s += 6 + msgp.GuessSize(z.Value)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PublisherTerminationMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, err = dc.ReadUint32()
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "PublisherID":
			{
				var tmp string
				tmp, err = dc.ReadString()
				z.PublisherID = UUID(tmp)
			}
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PublisherTerminationMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "MessageIDStruct"
	// map header, size 1
	// write "MessageID"
	err = en.Append(0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(uint32(z.MessageIDStruct.MessageID))
	if err != nil {
		return
	}
	// write "PublisherID"
	err = en.Append(0xab, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.PublisherID))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PublisherTerminationMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "MessageIDStruct"
	// map header, size 1
	// string "MessageID"
	o = append(o, 0x82, 0xaf, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x81, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, uint32(z.MessageIDStruct.MessageID))
	// string "PublisherID"
	o = append(o, 0xab, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.PublisherID))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PublisherTerminationMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageIDStruct":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "MessageID":
					{
						var tmp uint32
						tmp, bts, err = msgp.ReadUint32Bytes(bts)
						z.MessageIDStruct.MessageID = MessageIDType(tmp)
					}
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "PublisherID":
			{
				var tmp string
				tmp, bts, err = msgp.ReadStringBytes(bts)
				z.PublisherID = UUID(tmp)
			}
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *PublisherTerminationMessage) Msgsize() (s int) {
	s = 1 + 16 + 1 + 10 + msgp.Uint32Size + 12 + msgp.StringPrefixSize + len(string(z.PublisherID))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PublisherTerminationRequest) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageID":
			z.MessageID, err = dc.ReadUint32()
			if err != nil {
				return
			}
		case "PublisherID":
			{
				var tmp string
				tmp, err = dc.ReadString()
				z.PublisherID = UUID(tmp)
			}
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PublisherTerminationRequest) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "MessageID"
	err = en.Append(0x82, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteUint32(z.MessageID)
	if err != nil {
		return
	}
	// write "PublisherID"
	err = en.Append(0xab, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.PublisherID))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z PublisherTerminationRequest) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "MessageID"
	o = append(o, 0x82, 0xa9, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44)
	o = msgp.AppendUint32(o, z.MessageID)
	// string "PublisherID"
	o = append(o, 0xab, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x49, 0x44)
	o = msgp.AppendString(o, string(z.PublisherID))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PublisherTerminationRequest) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "MessageID":
			z.MessageID, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "PublisherID":
			{
				var tmp string
				tmp, bts, err = msgp.ReadStringBytes(bts)
				z.PublisherID = UUID(tmp)
			}
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z PublisherTerminationRequest) Msgsize() (s int) {
	s = 1 + 10 + msgp.Uint32Size + 12 + msgp.StringPrefixSize + len(string(z.PublisherID))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *QueryMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var tmp string
		tmp, err = dc.ReadString()
		(*z) = QueryMessage(tmp)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z QueryMessage) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z QueryMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *QueryMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var tmp string
		tmp, bts, err = msgp.ReadStringBytes(bts)
		(*z) = QueryMessage(tmp)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z QueryMessage) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SubscriptionDiffMessage) DecodeMsg(dc *msgp.Reader) (err error) {
	var msz uint32
	msz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(SubscriptionDiffMessage, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		msz--
		var pks string
		var jfb []UUID
		pks, err = dc.ReadString()
		if err != nil {
			return
		}
		var xsz uint32
		xsz, err = dc.ReadArrayHeader()
		if err != nil {
			return
		}
		if cap(jfb) >= int(xsz) {
			jfb = jfb[:xsz]
		} else {
			jfb = make([]UUID, xsz)
		}
		for cxo := range jfb {
			{
				var tmp string
				tmp, err = dc.ReadString()
				jfb[cxo] = UUID(tmp)
			}
			if err != nil {
				return
			}
		}
		(*z)[pks] = jfb
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SubscriptionDiffMessage) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for eff, rsw := range z {
		err = en.WriteString(eff)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(rsw)))
		if err != nil {
			return
		}
		for xpk := range rsw {
			err = en.WriteString(string(rsw[xpk]))
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SubscriptionDiffMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for eff, rsw := range z {
		o = msgp.AppendString(o, eff)
		o = msgp.AppendArrayHeader(o, uint32(len(rsw)))
		for xpk := range rsw {
			o = msgp.AppendString(o, string(rsw[xpk]))
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SubscriptionDiffMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var msz uint32
	msz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && msz > 0 {
		(*z) = make(SubscriptionDiffMessage, msz)
	} else if len((*z)) > 0 {
		for key, _ := range *z {
			delete((*z), key)
		}
	}
	for msz > 0 {
		var dnj string
		var obc []UUID
		msz--
		dnj, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		var xsz uint32
		xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			return
		}
		if cap(obc) >= int(xsz) {
			obc = obc[:xsz]
		} else {
			obc = make([]UUID, xsz)
		}
		for snv := range obc {
			{
				var tmp string
				tmp, bts, err = msgp.ReadStringBytes(bts)
				obc[snv] = UUID(tmp)
			}
			if err != nil {
				return
			}
		}
		(*z)[dnj] = obc
	}
	o = bts
	return
}

func (z SubscriptionDiffMessage) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for kgt, ema := range z {
			_ = ema
			s += msgp.StringPrefixSize + len(kgt) + msgp.ArrayHeaderSize
			for pez := range ema {
				s += msgp.StringPrefixSize + len(string(ema[pez]))
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UUID) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var tmp string
		tmp, err = dc.ReadString()
		(*z) = UUID(tmp)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z UUID) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UUID) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UUID) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var tmp string
		tmp, bts, err = msgp.ReadStringBytes(bts)
		(*z) = UUID(tmp)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z UUID) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}
