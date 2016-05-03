package main

import (
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"os"
	"strconv"
	"time"
)

var log *logging.Logger

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

func main() {
	var (
		numClients int64
		err        error
	)
	if len(os.Args) > 1 {
		numClients, _ = strconv.ParseInt(os.Args[1], 10, 64)
	} else {
		numClients = 10
	}
	config := &client.Config{
		BrokerAddress:      "0.0.0.0:4444",
		CoordinatorAddress: "cs262.cal-sdb.org:5055",
	}

	var clients = make([]*client.Client, numClients)
	var counts = make([]int64, numClients)
	for i, _ := range clients {
		config.ID = client.UUIDFromName(string(i))
		clients[i], err = client.NewClient(config)
		if err != nil {
			log.Criticalf("Could not create client: %v", err)
			os.Exit(1)
		}
		clients[i].AttachPublishHandler(func(m *common.PublishMessage) {
			counts[i] += 1
		})
		clients[i].Subscribe("Room = '410'")
	}

	time.Sleep(60 * time.Second)
}
