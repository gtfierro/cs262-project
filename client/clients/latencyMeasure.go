// This client measures the forwarding latency by running both a publisher and a subscriber and
// measuring the time for operations.
// Operations:
// - ping to the intended broker. Establish a baseline for network communication.
// - Time for initial subscription diff: send subscription and see how long before we get a diff
//   (should be empty)
// - Publish Message: send new metadata and measure latency to receive diff
// - Publish Message: no metadata, measure latency to receive message
package main

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"github.com/montanaflynn/stats"
	"os"
	"sync"
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
	}

	// before we begin, need to put some metadata in or it will not give us a response

	pub, err := client.NewPublisher(client.UUIDFromName("publishSlow"), func(pub *client.Publisher) {
		waitforever := make(chan bool)
		pub.AddMetadata(map[string]interface{}{"Room": "420"})
		if err := pub.Publish(1); err != nil {
			log.Error(err)
		}
		<-waitforever

	}, config)

	if err != nil {
		log.Criticalf("Could not publish (%v)", err)
		os.Exit(1)
	}
	pub.Start()

	var wg sync.WaitGroup

	var iterations = 100
	wg.Add(iterations)

	var initialDiffLatencies = make([]float64, iterations)
	go func() {
		for i := 0; i < iterations; i++ {
			c, err := client.NewClient(client.UUIDFromName(string(i)+"subscriber"), "Room = '420'", config)
			if err != nil {
				log.Criticalf("Could not create client: %v", err)
				os.Exit(1)
			}
			start := time.Now()
			counted := false
			c.AttachDiffHandler(func(m *common.SubscriptionDiffMessage) {
				if !counted {
					counted = true
					wg.Done()
					c.Stop <- true
				}
			})
			c.Start()
			<-c.Stop
			diff := time.Since(start)
			if i%10 == 0 {
				log.Infof("iteration %d: %s", i, diff.String())
			}
			initialDiffLatencies[i] = float64(diff)
			time.Sleep(500 * time.Millisecond)
		}
	}()
	wg.Wait()
	quartiles, _ := stats.Quartile(initialDiffLatencies)
	q1 := time.Duration(quartiles.Q1)
	q2 := time.Duration(quartiles.Q2)
	q3 := time.Duration(quartiles.Q3)
	mean, _ := stats.Mean(initialDiffLatencies)
	variance, _ := stats.Variance(initialDiffLatencies)
	percent99, _ := stats.Percentile(initialDiffLatencies, 99)
	fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
	fmt.Printf("Mean: %s   Variance %s\n", time.Duration(mean).String(), time.Duration(variance).String())
	fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
}
