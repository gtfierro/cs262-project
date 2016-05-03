package main

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"os"
)

var log *logging.Logger

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

func main() {
	var id string
	brokerIP := os.Args[1]
	if len(os.Args) > 2 {
		id = os.Args[2]
	} else {
		id = "client"
	}

	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
		ID:                 client.UUIDFromName(id),
	}
	C, err := client.NewClient(config)
	if err != nil {
		log.Criticalf("Could not create client: %v", err)
		os.Exit(1)
	}

	received := 0
	C.AttachPublishHandler(func(m *common.PublishMessage) {
		received += 1
		log.Infof("MESSAGE %d %v", received, m)
	})

	C.Subscribe("Room = '410'")

	<-C.Stop
}
