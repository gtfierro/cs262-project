package client_test

import (
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"log"
	"os"
)

func ExampleSubscribe() {
	config := &client.Config{
		BrokerAddress:      "1.2.3.4:4444",
		CoordinatorAddress: "5.6.7.8:5505",
		ID:                 client.UUIDFromName("Client Name"),
	}
	C, err := client.NewClient(config)
	if err != nil {
		log.Panicf("Could not create client: %v", err)
		os.Exit(1)
	}

	received := 0
	C.AttachPublishHandler(func(m *common.PublishMessage) {
		received += 1
		log.Printf("MESSAGE %d %v\n", received, m)
	})

	C.Subscribe("Room = '410' and Type = 'Temperature Sensor'")

	<-C.Stop
	log.Printf("Received %d messages\n", received)
}
