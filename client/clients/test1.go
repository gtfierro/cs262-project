package main

import (
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"os"
	"sync/atomic"
	"time"
)

var log *logging.Logger

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

func main() {
	config := &client.Config{
		BrokerAddress:      "0.0.0.0:4444",
		CoordinatorAddress: "0.0.0.0:5505",
		ID:                 client.UUIDFromName("C"),
	}
	C, err := client.NewClient(config)
	if err != nil {
		log.Criticalf("Could not create client: %v", err)
		os.Exit(1)
	}
	C.AttachPublishHandler(func(m *common.PublishMessage) {
		log.Infof("MESSAGE %v", m)
	})
	C.Subscribe("Room = '410'")
	pub := C.AddPublisher(client.UUIDFromName("pub1"))

	var sent uint32 = 0
	go func() {
		i := 0
		pub.AddMetadata(map[string]interface{}{"Room": "410"})
		for {
			if err := pub.Publish(i); err != nil {
				log.Error(err)
			}
			time.Sleep(1 * time.Second)
			atomic.AddUint32(&sent, 1)
			i++
		}
	}()

	C.StopIn(300 * time.Second)

	<-C.Stop
	numPublish := atomic.LoadUint32(&sent)
	log.Infof("Sent %d msg. %.2f per sec", numPublish, float64(numPublish)/float64(300))
}
