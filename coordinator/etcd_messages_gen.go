package main

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *Client) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "ClientID":
			err = z.ClientID.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "CurrentBrokerID":
			err = z.CurrentBrokerID.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "HomeBrokerID":
			err = z.HomeBrokerID.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "QueryString":
			z.QueryString, err = dc.ReadString()
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
func (z *Client) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "ClientID"
	err = en.Append(0x84, 0xa8, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = z.ClientID.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "CurrentBrokerID"
	err = en.Append(0xaf, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = z.CurrentBrokerID.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "HomeBrokerID"
	err = en.Append(0xac, 0x48, 0x6f, 0x6d, 0x65, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = z.HomeBrokerID.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "QueryString"
	err = en.Append(0xab, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67)
	if err != nil {
		return err
	}
	err = en.WriteString(z.QueryString)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Client) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "ClientID"
	o = append(o, 0x84, 0xa8, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44)
	o, err = z.ClientID.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "CurrentBrokerID"
	o = append(o, 0xaf, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o, err = z.CurrentBrokerID.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "HomeBrokerID"
	o = append(o, 0xac, 0x48, 0x6f, 0x6d, 0x65, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o, err = z.HomeBrokerID.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "QueryString"
	o = append(o, 0xab, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67)
	o = msgp.AppendString(o, z.QueryString)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Client) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "ClientID":
			bts, err = z.ClientID.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "CurrentBrokerID":
			bts, err = z.CurrentBrokerID.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "HomeBrokerID":
			bts, err = z.HomeBrokerID.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "QueryString":
			z.QueryString, bts, err = msgp.ReadStringBytes(bts)
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

func (z *Client) Msgsize() (s int) {
	s = 1 + 9 + z.ClientID.Msgsize() + 16 + z.CurrentBrokerID.Msgsize() + 13 + z.HomeBrokerID.Msgsize() + 12 + msgp.StringPrefixSize + len(z.QueryString)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Publisher) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "PublisherID":
			err = z.PublisherID.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "CurrentBrokerID":
			err = z.CurrentBrokerID.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "HomeBrokerID":
			err = z.HomeBrokerID.DecodeMsg(dc)
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
func (z *Publisher) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "PublisherID"
	err = en.Append(0x84, 0xab, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = z.PublisherID.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "CurrentBrokerID"
	err = en.Append(0xaf, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = z.CurrentBrokerID.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "HomeBrokerID"
	err = en.Append(0xac, 0x48, 0x6f, 0x6d, 0x65, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = z.HomeBrokerID.EncodeMsg(en)
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
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Publisher) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "PublisherID"
	o = append(o, 0x84, 0xab, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72, 0x49, 0x44)
	o, err = z.PublisherID.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "CurrentBrokerID"
	o = append(o, 0xaf, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o, err = z.CurrentBrokerID.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "HomeBrokerID"
	o = append(o, 0xac, 0x48, 0x6f, 0x6d, 0x65, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x44)
	o, err = z.HomeBrokerID.MarshalMsg(o)
	if err != nil {
		return
	}
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
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Publisher) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "PublisherID":
			bts, err = z.PublisherID.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "CurrentBrokerID":
			bts, err = z.CurrentBrokerID.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "HomeBrokerID":
			bts, err = z.HomeBrokerID.UnmarshalMsg(bts)
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

func (z *Publisher) Msgsize() (s int) {
	s = 1 + 12 + z.PublisherID.Msgsize() + 16 + z.CurrentBrokerID.Msgsize() + 13 + z.HomeBrokerID.Msgsize() + 9 + msgp.MapHeaderSize
	if z.Metadata != nil {
		for xvk, bzg := range z.Metadata {
			_ = bzg
			s += msgp.StringPrefixSize + len(xvk) + msgp.GuessSize(bzg)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SerializableBroker) DecodeMsg(dc *msgp.Reader) (err error) {
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
			err = z.BrokerInfo.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Alive":
			z.Alive, err = dc.ReadBool()
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
func (z *SerializableBroker) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "BrokerInfo"
	err = en.Append(0x82, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f)
	if err != nil {
		return err
	}
	err = z.BrokerInfo.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Alive"
	err = en.Append(0xa5, 0x41, 0x6c, 0x69, 0x76, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Alive)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SerializableBroker) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "BrokerInfo"
	o = append(o, 0x82, 0xaa, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f)
	o, err = z.BrokerInfo.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Alive"
	o = append(o, 0xa5, 0x41, 0x6c, 0x69, 0x76, 0x65)
	o = msgp.AppendBool(o, z.Alive)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SerializableBroker) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
			bts, err = z.BrokerInfo.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Alive":
			z.Alive, bts, err = msgp.ReadBoolBytes(bts)
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

func (z *SerializableBroker) Msgsize() (s int) {
	s = 1 + 11 + z.BrokerInfo.Msgsize() + 6 + msgp.BoolSize
	return
}
