package main

// function to test of the byte matches a MsgPack string encoding
// Uses byte types from https://github.com/msgpack/msgpack/blob/master/spec.md#formats
func isMsgPackString(b byte) bool {
	return (0xbf&b) == 0xbf || b == 0xd9 || b == 0xda || b == 0xdb
}

func isMsgPackArray(b byte) bool {
	return (0x9f&b) == 0x9f || b == 0xdc || b == 0xdd
}
