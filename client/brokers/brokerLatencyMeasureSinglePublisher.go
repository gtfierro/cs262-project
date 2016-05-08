package main

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"github.com/montanaflynn/stats"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var log *logging.Logger

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

func main() {

	coordinatorAddress := "cs262.cal-sdb.org:5505"

	var wg sync.WaitGroup
	var publisherWorkGroup sync.WaitGroup
	var queriesFinished uint32 = 0

	numBrokers := 10
	queriesPerBroker := 50
	minQueryDelay := 10 * time.Millisecond  // overall; per broker will be (queryDelay * numBrokers)
	maxQueryDelay := 100 * time.Millisecond // where queryDelay is uniformly distributed btwn min and max
	timeout := 2 * time.Minute

	minBrokerQueryDelayNano := int64(numBrokers) * minQueryDelay.Nanoseconds()
	queryDelayRangeNano := (int64(numBrokers) * maxQueryDelay.Nanoseconds()) - minBrokerQueryDelayNano
	wg.Add(numBrokers * queriesPerBroker)
	publisherWorkGroup.Add(numBrokers)

	var sendTime = make([]int64, numBrokers*queriesPerBroker)
	var initialDiffLatencies = make([]int64, numBrokers*queriesPerBroker)
	for brokerNum := 0; brokerNum < numBrokers; brokerNum++ {
		go func(bNum int) {
			brokerID := common.UUID(fmt.Sprintf("broker%d", bNum)) //client.UUIDFromName(fmt.Sprintf("broker%d", bNum))
			hasPublisher := false
			queryNum := 0
			connectCallback := func(broker *client.SimulatedBroker) {
				if !hasPublisher {
					err := broker.Send(&common.BrokerPublishMessage{
						MessageIDStruct: common.GetMessageIDStruct(),
						UUID:            client.UUIDFromName(fmt.Sprintf("broker%d-publisher", bNum)),
						Metadata:        map[string]interface{}{"Room": "410", "BuildingNum": fmt.Sprintf("%v", bNum)},
						Value:           "val",
					})
					if err != nil {
						return
					}
					hasPublisher = true
					time.Sleep(2 * time.Second) // give it time to stabilize
					publisherWorkGroup.Done()
					publisherWorkGroup.Wait()
					time.Sleep(time.Duration(minBrokerQueryDelayNano + rand.Int63n(queryDelayRangeNano)))
				}
				for ; queryNum < queriesPerBroker; queryNum++ {
					queryMsg := &common.BrokerQueryMessage{
						MessageIDStruct: common.GetMessageIDStruct(),
						Query:           fmt.Sprintf("Room = '410' and Query != '%d' and Broker != '%d'", queryNum, bNum), // encode querynum in query
						UUID:            client.UUIDFromName(fmt.Sprintf("broker%d-query%d", bNum, queryNum)),
					}
					if err := broker.Send(queryMsg); err != nil {
						queryNum -= 1
						return
					}
					log.Infof("Sent query %v from broker %v (globalQueryNum %v)", queryNum, bNum, queryNum*numBrokers+bNum)
					if queryNum*numBrokers+bNum >= len(initialDiffLatencies) {
						log.Errorf("INDEX OUT OF BOUND: query %d, numBrokers %d, bNum %d, index %d, len %d", queryNum, numBrokers, bNum, queryNum*numBrokers+bNum, len(initialDiffLatencies))
					}
					sendTime[queryNum*numBrokers+bNum] = time.Now().UnixNano()
					time.Sleep(time.Duration(minBrokerQueryDelayNano + rand.Int63n(queryDelayRangeNano)))
				}
			}
			msgHandler := func(msg common.Sendable) {
				if subDiff, ok := msg.(*common.BrokerSubscriptionDiffMessage); ok {
					var qNum int
					var bNumQuery int
					if _, err := fmt.Sscanf(subDiff.Query, "Room = '410' and Query != '%d' and Broker != '%d'", &qNum, &bNumQuery); err != nil {
						log.Fatal("Error while scanning query num from query")
					}
					if bNumQuery != bNum {
						log.Errorf("bNumQuery: %d, bNum: %d (queryNum %d)", bNumQuery, bNum, qNum)
					}
					latency := time.Now().UnixNano() - sendTime[qNum*numBrokers+bNum]
					if initialDiffLatencies[qNum*numBrokers+bNum] != 0 {
						log.Errorf("Received a duplicate message on broker %d about query %d", bNum, qNum)
					} else {
						initialDiffLatencies[qNum*numBrokers+bNum] = latency
						wg.Done()
						fin := atomic.AddUint32(&queriesFinished, 1)
						log.Infof("Finished %d queries; query %v on broker %v had %v latency", fin, qNum, bNum, latency/1000000)
					}
				}
			}
			broker, err := client.NewSimulatedBroker(connectCallback, func() {}, msgHandler, brokerID, coordinatorAddress)
			if err != nil {
				log.Criticalf("Could not create client: %v", err)
				os.Exit(1)
			}
			broker.Start()
		}(brokerNum)
	}
	waitChan := make(chan bool)
	go func() {
		wg.Wait()
		close(waitChan)
	}()
	select {
	case <-time.After(timeout):
	case <-waitChan:
	}
	floatLatencies := make([]float64, numBrokers*queriesPerBroker)
	for idx, latency := range initialDiffLatencies {
		floatLatencies[idx] = float64(latency)
	}
	quartiles, _ := stats.Quartile(floatLatencies)
	q1 := time.Duration(quartiles.Q1)
	q2 := time.Duration(quartiles.Q2)
	q3 := time.Duration(quartiles.Q3)
	mean, _ := stats.Mean(floatLatencies)
	variance, _ := stats.Variance(floatLatencies)
	percent99, _ := stats.Percentile(floatLatencies, 99)

	fmt.Printf("Settings: numBrokers %d, queriesPerBroker %d, minQueryDelay %s, maxQueryDelay %s\n", numBrokers, queriesPerBroker, minQueryDelay.String(), maxQueryDelay.String())
	fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
	fmt.Printf("Mean: %s   Variance %s\n", time.Duration(mean).String(), time.Duration(variance).String())
	fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
	fmt.Print("\n\nIndividual Latency Count:\nQuery Number, Latency (Nano)\n")
	for queryNum, latencyNano := range initialDiffLatencies {
		fmt.Printf("%d,%d,%d\n", queryNum, sendTime[queryNum], latencyNano)
	}
}
