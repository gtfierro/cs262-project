package common

import (
	"bytes"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io/ioutil"
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

func BenchmarkDecodeQueryShort(b *testing.B) {
	bytes, _ := msgpack.Marshal("Key1 = 'Val1'")
	bytes = append([]byte{byte(QUERYMSG)}, bytes...)
	c := NewCopyReader(bytes)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := msgpack.NewDecoder(c)
		MessageFromDecoder(dec)
	}
}

func BenchmarkMsgpDecodeQueryShort(b *testing.B) {
	v := QueryMessage{Query: "Key1 = 'Val1'"}
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

func BenchmarkEncodeQueryShort(b *testing.B) {
	query := QueryMessage{Query: "Key1 = 'Val1'"}
	encoder := msgpack.NewEncoder(ioutil.Discard)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(query)
	}
}

func BenchmarkDecodeQueryLong(b *testing.B) {
	var query = strings.Repeat("Key1 = 'Val1'", 50)
	bytes, _ := msgpack.Marshal(query)
	bytes = append([]byte{byte(QUERYMSG)}, bytes...)
	c := NewCopyReader(bytes)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := msgpack.NewDecoder(c)
		MessageFromDecoder(dec)
	}
}

func BenchmarkMsgpDecodeQueryLong(b *testing.B) {
	v := QueryMessage{Query: strings.Repeat("Key1 = 'Val1'", 50)}
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

func BenchmarkEncodeQueryLong(b *testing.B) {
	var query_string = strings.Repeat("Key1 = 'Val1'", 50)
	query := QueryMessage{Query: query_string}
	encoder := msgpack.NewEncoder(ioutil.Discard)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(query)
	}
}

func BenchmarkDecodePublishNoMetadata(b *testing.B) {
	var msg = PublishMessage{UUID: "f58c8216-fa71-11e5-b77e-1002b58053c7",
		Value: 1459780334680233928}
	bytes, _ := msgpack.Marshal(msg)
	bytes = append([]byte{byte(PUBLISHMSG)}, bytes...)
	c := NewCopyReader(bytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := msgpack.NewDecoder(c)
		MessageFromDecoder(dec)
	}
}

func BenchmarkEncodePublishNoMetadata(b *testing.B) {
	var msg = PublishMessage{UUID: "f58c8216-fa71-11e5-b77e-1002b58053c7",
		Value: 1459780334680233928}
	encoder := msgpack.NewEncoder(ioutil.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(msg)
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

func BenchmarkDecodePublishWithMetadata(b *testing.B) {
	var msg = PublishMessage{UUID: "f58c8216-fa71-11e5-b77e-1002b58053c7",
		Metadata: testmetadata,
		Value:    1459780334680233928}
	bytes, _ := msgpack.Marshal(msg)
	bytes = append([]byte{byte(PUBLISHMSG)}, bytes...)
	c := NewCopyReader(bytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := msgpack.NewDecoder(c)
		MessageFromDecoder(dec)
	}
}

func BenchmarkEncodePublishWithMetadata(b *testing.B) {
	var msg = PublishMessage{UUID: "f58c8216-fa71-11e5-b77e-1002b58053c7",
		Metadata: testmetadata,
		Value:    1459780334680233928}
	encoder := msgpack.NewEncoder(ioutil.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(msg)
	}
}
