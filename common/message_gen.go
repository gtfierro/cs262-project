package common

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

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
func (z *PublishMessage) MarshalMsg(b []byte) (o []byte, err error) {
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

func (z *PublishMessage) Msgsize() (s int) {
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
func (z *QueryMessage) DecodeMsg(dc *msgp.Reader) (err error) {
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
func (z QueryMessage) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Query"
	err = en.Append(0x81, 0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
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
func (z QueryMessage) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Query"
	o = append(o, 0x81, 0xa5, 0x51, 0x75, 0x65, 0x72, 0x79)
	o = msgp.AppendString(o, z.Query)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *QueryMessage) UnmarshalMsg(bts []byte) (o []byte, err error) {
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

func (z QueryMessage) Msgsize() (s int) {
	s = 1 + 6 + msgp.StringPrefixSize + len(z.Query)
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
		var wht string
		var hct []UUID
		wht, err = dc.ReadString()
		if err != nil {
			return
		}
		var xsz uint32
		xsz, err = dc.ReadArrayHeader()
		if err != nil {
			return
		}
		if cap(hct) >= int(xsz) {
			hct = hct[:xsz]
		} else {
			hct = make([]UUID, xsz)
		}
		for cua := range hct {
			{
				var tmp string
				tmp, err = dc.ReadString()
				hct[cua] = UUID(tmp)
			}
			if err != nil {
				return
			}
		}
		(*z)[wht] = hct
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SubscriptionDiffMessage) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for xhx, lqf := range z {
		err = en.WriteString(xhx)
		if err != nil {
			return
		}
		err = en.WriteArrayHeader(uint32(len(lqf)))
		if err != nil {
			return
		}
		for daf := range lqf {
			err = en.WriteString(string(lqf[daf]))
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
	for xhx, lqf := range z {
		o = msgp.AppendString(o, xhx)
		o = msgp.AppendArrayHeader(o, uint32(len(lqf)))
		for daf := range lqf {
			o = msgp.AppendString(o, string(lqf[daf]))
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
		var pks string
		var jfb []UUID
		msz--
		pks, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		var xsz uint32
		xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
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
				tmp, bts, err = msgp.ReadStringBytes(bts)
				jfb[cxo] = UUID(tmp)
			}
			if err != nil {
				return
			}
		}
		(*z)[pks] = jfb
	}
	o = bts
	return
}

func (z SubscriptionDiffMessage) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for eff, rsw := range z {
			_ = rsw
			s += msgp.StringPrefixSize + len(eff) + msgp.ArrayHeaderSize
			for xpk := range rsw {
				s += msgp.StringPrefixSize + len(string(rsw[xpk]))
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
