package main

import (
	"github.com/gtfierro/cs262-project/common"
	"math/rand"
	"net"
	"strings"
	"testing"
)

func BenchmarkNewSubscriptionAllUniqueWarm(b *testing.B) {
	var testConfig = &common.Config{
		Logging: common.LoggingConfig{UseJSON: false, Level: "fatal"},
		Server:  common.ServerConfig{Port: 4444, Global: false},
		Mongo:   common.MongoConfig{Port: 27017, Host: "0.0.0.0"},
		Debug:   common.DebugConfig{Enable: false, ProfileLength: 0},
	}
	common.SetupLogging(testConfig)
	conn := &net.TCPConn{}
	metadata := NewMetadataStore(testConfig)
	broker := NewBroker(metadata)

	querystrings := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		randChars := make([]string, 20)
		for i := 0; i < len(randChars); i++ {
			randChars[i] = string('A' + rand.Int31n('Z'-'A'))
		}
		randStr := strings.Join(randChars, "")
		querystrings[i] = "Room = '" + randStr + "'"
		// we run these first to warm them up
		broker.NewSubscription(querystrings[i], conn)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broker.NewSubscription(querystrings[i], conn)
	}
}

func BenchmarkNewSubscriptionAllUniqueCold(b *testing.B) {
	var testConfig = &common.Config{
		Logging: common.LoggingConfig{UseJSON: false, Level: "fatal"},
		Server:  common.ServerConfig{Port: 4444, Global: false},
		Mongo:   common.MongoConfig{Port: 27017, Host: "0.0.0.0"},
		Debug:   common.DebugConfig{Enable: false, ProfileLength: 0},
	}
	common.SetupLogging(testConfig)
	conn := &net.TCPConn{}
	metadata := NewMetadataStore(testConfig)
	broker := NewBroker(metadata)

	querystrings := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		randChars := make([]string, 20)
		for i := 0; i < len(randChars); i++ {
			randChars[i] = string('A' + rand.Int31n('Z'-'A'))
		}
		randStr := strings.Join(randChars, "")
		querystrings[i] = "Room = '" + randStr + "'"
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broker.NewSubscription(querystrings[i], conn)
	}
}
