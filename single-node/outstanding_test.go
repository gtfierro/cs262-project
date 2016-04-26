package main

import (
	"github.com/gtfierro/cs262-project/common"
	"sync"
	"testing"
	"time"
)

func BenchmarkOutstandingGotMessage(b *testing.B) {
	om := newOutstandingManager()
	messages := make([]common.Message, b.N)
	for i := 0; i < b.N; i++ {
		bpm := &common.BrokerPublishMessage{}
		bpm.SetID(common.GetMessageID())
		messages[i] = bpm
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		om.GotMessage(messages[i])
	}
}

func BenchmarkOutstandingWaitMessage(b *testing.B) {
	om := newOutstandingManager()
	messages := make([]common.Message, b.N)
	for i := 0; i < b.N; i++ {
		bpm := &common.BrokerPublishMessage{}
		bpm.SetID(common.GetMessageID())
		messages[i] = bpm
	}

	for i := 0; i < b.N; i++ {
		om.GotMessage(messages[i])
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		om.WaitForMessage(messages[i].GetID())
		if i%1000 == 0 {
			println(i, b.N)
		}
	}
}

func TestOutstandingWaitMessage1(t *testing.T) {
	var (
		wg   sync.WaitGroup
		msg2 common.Message
		err  error
	)
	wg.Add(1)

	om := newOutstandingManager()

	msg := &common.BrokerPublishMessage{}
	msg.SetID(common.GetMessageID())

	go func() {
		msg2, err = om.WaitForMessage(msg.GetID())
		if err != nil {
			t.Error(err)
		}
		if msg2 != msg {
			t.Errorf("Expected %v but got %v", msg2, msg)
		}
		wg.Done()
	}()

	time.Sleep(200 * time.Millisecond)
	om.GotMessage(msg)
	wg.Wait()
}

func TestOutstandingWaitMessage2(t *testing.T) {
	var (
		wg   sync.WaitGroup
		msg2 common.Message
		err  error
	)
	wg.Add(1)

	om := newOutstandingManager()

	msg := &common.BrokerPublishMessage{}
	msg.SetID(common.GetMessageID())

	om.GotMessage(msg)

	go func() {
		msg2, err = om.WaitForMessage(msg.GetID())
		if err != nil {
			t.Error(err)
		}
		if msg2 != msg {
			t.Errorf("Expected %v but got %v", msg2, msg)
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestManyWait1000(t *testing.T) {
	om := newOutstandingManager()
	messages := make([]common.Message, 1000)
	for i := 0; i < 1000; i++ {
		bpm := &common.BrokerPublishMessage{}
		bpm.SetID(common.GetMessageID())
		messages[i] = bpm
	}

	for i := 0; i < 1000; i++ {
		om.GotMessage(messages[i])
	}

	for i := 0; i < 1000; i++ {
		msg, _ := om.WaitForMessage(messages[i].GetID())
		if msg.GetID() != messages[i].GetID() {
			t.Error("Wrong ID")
		}
	}
}
