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

	coordinatorAddress := "cs262.cal-sdb.org:5505"

	var wg sync.WaitGroup
	var publisherWorkGroup sync.WaitGroup

	delayBetweenQueries := time.Second // per broker so requests will arrive at (delayBetweenQueries / numBrokers)
	numBrokers := 10
	queriesPerBroker := 100
	wg.Add(numBrokers * queriesPerBroker)
	publisherWorkGroup.Add(numBrokers)

	var initialDiffLatencies = make([]int64, numBrokers*queriesPerBroker)
	for brokerNum := 0; brokerNum < numBrokers; brokerNum++ {
		go func() {
			hasPublisher := false
			queryNum := 0
			connectCallback := func(broker *client.SimulatedBroker) {
				if !hasPublisher {
					err := broker.Send(&common.BrokerPublishMessage{
						MessageIDStruct: common.GetMessageIDStruct(),
						UUID:            client.UUIDFromName(fmt.Sprintf("broker%d-publisher", brokerNum)),
						Metadata:        map[string]interface{}{"Room": "410", "BuildingNum": brokerNum},
					})
					if err != nil {
						return
					}
					hasPublisher = true
					time.Sleep(5 * time.Second) // give it time to stabilize
					publisherWorkGroup.Done()
					publisherWorkGroup.Wait()
				}
				for ; queryNum < queriesPerBroker; queryNum++ {
					queryMsg := &common.BrokerQueryMessage{
						MessageIDStruct: common.GetMessageIDStruct(),
						Query:           fmt.Sprintf("Room = '410' and QueryNum != '%d'", queryNum), // encode querynum in query
						UUID:            client.UUIDFromName(fmt.Sprintf("broker%d-query%d", brokerNum, queryNum)),
					}
					initialDiffLatencies[queryNum+brokerNum*queriesPerBroker] = time.Now().UnixNano()
					if err := broker.Send(queryMsg); err != nil {
						return
					}
					time.Sleep(delayBetweenQueries)
				}
			}
			msgHandler := func(msg common.Sendable) {
				if subDiff, ok := msg.(*common.BrokerSubscriptionDiffMessage); ok {
					var qNum int
					if _, err := fmt.Sscanf(subDiff.Query, "Room = '410' and QueryNum != '%d'", &qNum); err != nil {
						log.Fatal("Error while scanning query num from query")
					}
					startTime := initialDiffLatencies[qNum+brokerNum*queriesPerBroker]
					initialDiffLatencies[qNum+brokerNum*queriesPerBroker] = time.Now().UnixNano() - startTime
					wg.Done()
				}
			}
			broker, err := client.NewSimulatedBroker(connectCallback, msgHandler, client.UUIDFromName("broker"+string(brokerNum)), coordinatorAddress)
			if err != nil {
				log.Criticalf("Could not create client: %v", err)
				os.Exit(1)
			}
			broker.Start()
		}()
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
	fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
	fmt.Printf("Mean: %s   Variance %s\n", time.Duration(mean).String(), time.Duration(variance).String())
	fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
	fmt.Print("\n\nIndividual Latency Count:\nQuery Number, Latency (Nano)\n")
	for queryNum, latencyNano := range initialDiffLatencies {
		fmt.Printf("%d,%d\n", queryNum, latencyNano)
	}
}
