package main

import (
	"gopkg.in/vmihailenco/msgpack.v2"
	"net"
	"testing"
)

var testConfig = Config{
	Logging: LoggingConfig{UseJSON: false, Level: "debug"},
	Server:  ServerConfig{Port: 4444, Global: false},
	Mongo:   MongoConfig{Port: 27017, Host: "0.0.0.0"},
	Debug:   DebugConfig{Enable: false, ProfileLength: 0},
}

func BenchmarkDispatchNoMetadata(b *testing.B) {
	s := NewServer(&testConfig)
	go s.listenAndDispatch()
	address, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:4444")
	conn, err := net.DialTCP("tcp4", nil, address)
	if err != nil {
		b.Fatal(err)
	}
	encoder := msgpack.NewEncoder(conn)
	message := &Message{UUID: "cd47df06-f451-11e5-873b-9b450be7df8d", Value: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.Value = i
		encoder.Encode(message)
	}
	go func() {
		s.stop()
	}()
	<-s.stopped
}

func BenchmarkDispatchWithMetadata(b *testing.B) {
	s := NewServer(&testConfig)
	go s.listenAndDispatch()
	address, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:4444")
	conn, err := net.DialTCP("tcp4", nil, address)
	if err != nil {
		b.Fatal(err)
	}
	encoder := msgpack.NewEncoder(conn)
	message := &Message{UUID: "cd47df06-f451-11e5-873b-9b450be7df8d", Value: 0, Metadata: map[string]interface{}{"ABC": "123"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message.Value = i
		encoder.Encode(message)
	}
	go func() {
		s.stop()
	}()
	<-s.stopped
}
