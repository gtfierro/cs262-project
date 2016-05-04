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
	"sync/atomic"
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

	maxDuration := (1000 * time.Second).Nanoseconds()

	numBrokers := 5
	queriesPerBroker := 20
	queryFrequency := 500 * time.Millisecond // overall; per broker will be (queryFrequency * numBrokers)
	delayBetweenQueries := time.Duration(numBrokers)*queryFrequency // per broker so requests will arrive at (delayBetweenQueries / numBrokers)
	wg.Add(numBrokers * queriesPerBroker)
	publisherWorkGroup.Add(numBrokers)

	var initialDiffLatencies = make([]int64, numBrokers*queriesPerBroker)
	for brokerNum := 0; brokerNum < numBrokers; brokerNum++ {
		go func(bNum int) {
			hasPublisher := false
			queryNum := 0
			connectCallback := func(broker *client.SimulatedBroker) {
				connectMsg := &common.BrokerConnectMessage{
					MessageIDStruct: common.GetMessageIDStruct(),
					BrokerInfo: common.BrokerInfo{
						BrokerID: client.UUIDFromName(fmt.Sprintf("broker%d", bNum)),
						ClientBrokerAddr: fmt.Sprintf("0.0.0.%d:4444", bNum),
						CoordBrokerAddr: fmt.Sprintf("0.0.0.%d:5505", bNum),
					},
				}
				if err := broker.Send(connectMsg); err != nil {		
					return 
				} else {
					time.Sleep(500*time.Millisecond)
				}
				if !hasPublisher {
					err := broker.Send(&common.BrokerPublishMessage{
						MessageIDStruct: common.GetMessageIDStruct(),
						UUID:            client.UUIDFromName(fmt.Sprintf("broker%d-publisher", bNum)),
						Metadata:        map[string]interface{}{"Room": "410", "BuildingNum": bNum},
					})
					if err != nil {
						return
					}
					hasPublisher = true
					time.Sleep(5 * time.Second) // give it time to stabilize
					publisherWorkGroup.Done()
					publisherWorkGroup.Wait()
					time.Sleep(time.Duration(bNum) * queryFrequency)
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
					log.Infof("Sent query %v from broker %v", queryNum, bNum)
					initialDiffLatencies[queryNum*numBrokers+bNum] = time.Now().UnixNano()
					time.Sleep(delayBetweenQueries)
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
					startTime := initialDiffLatencies[qNum*numBrokers+bNum]
					latency := time.Now().UnixNano() - startTime
					if startTime < maxDuration {
						log.Errorf("Received a duplicate message on broker %d about query %d", bNum, qNum)
					} else {
						initialDiffLatencies[qNum*numBrokers+bNum] = latency
						wg.Done()
					}
					fin := atomic.AddUint32(&queriesFinished, 1)
					log.Infof("Finished %d queries; query %v on broker %v had %v latency", fin, qNum, bNum, latency)
				}
			}
			broker, err := client.NewSimulatedBroker(connectCallback, msgHandler, client.UUIDFromName("broker"+string(bNum)), coordinatorAddress)
			if err != nil {
				log.Criticalf("Could not create client: %v", err)
				os.Exit(1)
			}
			broker.Start()
		}(brokerNum)
	}
	wg.Wait()
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

	fmt.Printf("Settings: queryFrequency %s, numBrokers %d, queriesPerBroker %d\n", queryFrequency.String(), numBrokers, queriesPerBroker)
	fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
	fmt.Printf("Mean: %s   Variance %s\n", time.Duration(mean).String(), time.Duration(variance).String())
	fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
	fmt.Print("\n\nIndividual Latency Count:\nQuery Number, Latency (Nano)\n")
	for queryNum, latencyNano := range initialDiffLatencies {
		fmt.Printf("%d,%d\n", queryNum, latencyNano)
	}
}
