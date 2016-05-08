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
	"time"
)

var log *logging.Logger
var r rand.Source
var wg sync.WaitGroup
var rate time.Duration
var waitTime time.Duration
var numClients int
var gather chan []float64

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
	r = rand.NewSource(time.Now().UnixNano())
	rate = time.Duration(100) * time.Millisecond
	waitTime = 20 * time.Minute
	numClients = 10
	gather = make(chan []float64, numClients)
}

func runPubSubPair(rate time.Duration, subConfig, pubConfig *client.Config) {
	// 1 client, 1 publisher. Measure tiem between one publishing, one subscribing
	mdkey := fmt.Sprintf("Room%d", r.Int63())
	mdval := fmt.Sprintf("%d", r.Int63())
	query := fmt.Sprintf("%s = '%s'", mdkey, mdval)
	publisher_seed := fmt.Sprintf("%d", r.Int63())
	subscriber_seed := fmt.Sprintf("%d", r.Int63())

	latencies := []float64{}

	c, err := client.NewClient(client.UUIDFromName(subscriber_seed), query, subConfig)
	c.AttachPublishHandler(func(m *common.PublishMessage) {
		recv_time := time.Now().UnixNano()
		sent_time := m.Value.(int64)
		diff := recv_time - sent_time
		latencies = append(latencies, float64(diff))
		if len(latencies)%100 == 0 {
			fmt.Printf("Message: %v. Diff %s\n", len(latencies), time.Duration(diff).String())
		}
	})
	c.Start()
	time.Sleep(500 * time.Millisecond)

	publisher, err := client.NewPublisher(client.UUIDFromName(publisher_seed), func(pub *client.Publisher) {
		ticker := time.Tick(rate)
		pub.AddMetadata(map[string]interface{}{mdkey: mdval})
		for now := range ticker {
			if err := pub.Publish(now.UnixNano()); err != nil {
				log.Error(err)
			}
		}
	}, pubConfig)
	if err != nil {
		log.Criticalf("Could not create publisher (%v)", err)
		os.Exit(1)
	}
	publisher.Start()

	c.StopIn(waitTime)
	<-c.Stop
	wg.Done()
	gather <- latencies
}

func main() {
	brokerIP := os.Args[1]
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
	}

	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go runPubSubPair(rate, config, config)
	}
	wg.Wait()
	close(gather)

	var latencies [][]float64
	for data := range gather {
		latencies = append(latencies, data)
	}

	quartiles, _ := stats.Quartile(latencies[0])
	q1 := time.Duration(quartiles.Q1)
	q2 := time.Duration(quartiles.Q2)
	q3 := time.Duration(quartiles.Q3)
	mean, _ := stats.Mean(latencies[0])
	percent99, _ := stats.Percentile(latencies[0], 99)
	fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
	fmt.Printf("Mean: %s\n", time.Duration(mean).String())
	fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
	client.ManyColumnCSV(latencies, "many_clients.csv")
}
