package main

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"os"
	"strconv"
	"time"
)

var log *logging.Logger

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

func main() {
	var dur int64
	brokerIP := os.Args[1]
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
		ID:                 client.UUIDFromName("C"),
	}
	if len(os.Args) > 2 {
		dur, _ = strconv.ParseInt(os.Args[2], 10, 64)
	} else {
		dur = 1000
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
			time.Sleep(time.Duration(dur) * time.Nanosecond)
			i++
		}
	}()
	C.StopIn(60 * time.Second)

	<-C.Stop
	stats := pub.GetStats()
	log.Infof("Sent %d msg. %.2f per sec", stats.MessagesAttempted, float64(stats.MessagesAttempted)/float64(30))
}
