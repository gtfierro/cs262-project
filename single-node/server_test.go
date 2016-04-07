package main

import (
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"net"
	"testing"
)

var testConfig = &common.Config{
	Logging: common.LoggingConfig{UseJSON: false, Level: "error"},
	Server:  common.ServerConfig{Port: 4444, Global: false},
	Mongo:   common.MongoConfig{Port: 27017, Host: "0.0.0.0"},
	Debug:   common.DebugConfig{Enable: false, ProfileLength: 0},
}

func BenchmarkDispatchNoMetadata(b *testing.B) {
	s := NewServer(testConfig)
	go s.listenAndDispatch()
	address, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:4444")
	conn, err := net.DialTCP("tcp4", nil, address)
	if err != nil {
		b.Fatal(err)
	}
	encoder := msgp.NewWriter(conn)
	message := &common.PublishMessage{UUID: "cd47df06-f451-11e5-873b-9b450be7df8d", Value: 0}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.Value = i
		message.Encode(encoder)
	}
	go func() {
		s.stop()
	}()
	<-s.stopped
}

func BenchmarkDispatchWithMetadata(b *testing.B) {
	s := NewServer(testConfig)
	go s.listenAndDispatch()
	address, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:4444")
	conn, err := net.DialTCP("tcp4", nil, address)
	if err != nil {
		b.Fatal(err)
	}
	encoder := msgp.NewWriter(conn)
	message := &common.PublishMessage{UUID: "cd47df06-f451-11e5-873b-9b450be7df8d", Value: 0, Metadata: map[string]interface{}{"ABC": "123"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.Value = i
		message.Encode(encoder)
	}
	go func() {
		s.stop()
	}()
	<-s.stopped
}
