package common

import (
	"bytes"
	"github.com/tinylib/msgp/msgp"
	"strings"
	"testing"
)

type CopyReader struct {
	b []byte
}

func NewCopyReader(backing []byte) CopyReader {
	return CopyReader{b: backing}
}

func (cr CopyReader) Read(b []byte) (n int, e error) {
	n = copy(cr.b, b)
	return
}
func BenchmarkMsgpDecodeQueryShort(b *testing.B) {
	v := QueryMessage("Key1 = 'Val1'")
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)
	b.SetBytes(int64(buf.Len()))
	rd := msgp.NewEndlessReader(buf.Bytes(), b)
	dc := msgp.NewReader(rd)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := v.DecodeMsg(dc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMsgpDecodeQueryLong(b *testing.B) {
	v := QueryMessage(strings.Repeat("Key1 = 'Val1'", 50))
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)
	b.SetBytes(int64(buf.Len()))
	rd := msgp.NewEndlessReader(buf.Bytes(), b)
	dc := msgp.NewReader(rd)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := v.DecodeMsg(dc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var testmetadata = map[string]interface{}{
	"key1":  "val1",
	"key2":  "val2",
	"key3":  "val3",
	"key4":  "val4",
	"key5":  "val5",
	"key6":  "val6",
	"key7":  "val7",
	"key8":  "val8",
	"key9":  "val9",
	"key10": "val10",
	"key11": "val11",
}
