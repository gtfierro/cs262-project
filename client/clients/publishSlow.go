package main

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"os"
	"time"
)

var log *logging.Logger

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

func main() {
	brokerIP := os.Args[1]
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
		ID:                 client.UUIDFromName("C"),
	}
	C, err := client.NewClient(config)
	if err != nil {
		log.Criticalf("Could not create client: %v", err)
		os.Exit(1)
	}

	pub := C.AddPublisher(client.UUIDFromName("publishSlow"))

	go func() {
		i := 0
		pub.AddMetadata(map[string]interface{}{"Room": "410"})
		for {
			if err := pub.Publish(i); err != nil {
				log.Error(err)
			}
			time.Sleep(1 * time.Second)
			i++
		}
	}()

	<-C.Stop
	stats := pub.GetStats()
	log.Infof("Sent %d msg. %.2f per sec", stats.MessagesAttempted, float64(stats.MessagesAttempted)/float64(30))
}
