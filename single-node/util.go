package main

import (
	"sync/atomic"
)

// function to test of the byte matches a MsgPack string encoding
// Uses byte types from https://github.com/msgpack/msgpack/blob/master/spec.md#formats
func isMsgPackString(b byte) bool {
	return (0xbf&b) == b || b == 0xd9 || b == 0xda || b == 0xdb
}

func isMsgPackArray(b byte) bool {
	return (0x9f&b) == b || b == 0xdc || b == 0xdd
}

type AtomicBool struct {
	val uint32
}

func (b *AtomicBool) Set(v bool) {
	if v {
		atomic.StoreUint32(&b.val, uint32(1))
	} else {
		atomic.StoreUint32(&b.val, uint32(0))
	}
}

func (b *AtomicBool) Get() bool {
	return atomic.LoadUint32(&b.val) == 1
}
