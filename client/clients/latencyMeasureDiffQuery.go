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

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
	r = rand.NewSource(time.Now().UnixNano())
}

func main() {
	brokerIP := os.Args[1]
	var keychoose string
	// configuring if the query keys are different or if they are on the same ke
	if len(os.Args) > 2 && os.Args[2] == "diff" {
		log.Info("Queries have diff keys")
		keychoose = "diff"
	} else {
		log.Info("Queries have same keys")
		keychoose = "same"
	}
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
	}

	// pregenerate the queries
	var iterations = 100
	var queries = make([]string, iterations)
	var metadatas = make([]map[string]interface{}, iterations)
	var nonce = fmt.Sprintf("%d", r.Int63())
	for i := 0; i < iterations; i++ {
		var query string
		var md map[string]interface{}
		if keychoose == "diff" {
			query = fmt.Sprintf("Room%s%d = '%d'", nonce, i, i)
			md = map[string]interface{}{fmt.Sprintf("Room%s%d", nonce, i): fmt.Sprintf("%d", i)}
		} else {
			query = fmt.Sprintf("Room%s = '%d'", nonce, i)
			md = map[string]interface{}{fmt.Sprintf("Room%s", nonce): fmt.Sprintf("%d", i)}
		}
		queries[i] = query
		metadatas[i] = md
	}

	var wg sync.WaitGroup
	wg.Add(iterations)
	for iter := 0; iter < iterations; iter++ {
		time.Sleep(200 * time.Millisecond)
		iter := iter // local shadow
		// before we begin, need to put some metadata in and maintain the connection or it will not give us a response
		pub, err := client.NewPublisher(client.UUIDFromName(nonce+string(iter)), func(pub *client.Publisher) {
			waitforever := make(chan bool)
			pub.AddMetadata(metadatas[iter])
			if err := pub.Publish(1); err != nil {
				log.Error(err)
			}
			wg.Done()
			<-waitforever

		}, config)

		if err != nil {
			log.Criticalf("Could not publish (%v)", err)
			os.Exit(1)
		}
		pub.Start()
	}
	wg.Wait()
	wg.Add(iterations)

	var initialDiffLatencies = make([]float64, iterations)
	go func() {
		for i := 0; i < iterations; i++ {
			i := i
			c, err := client.NewClient(client.UUIDFromName(nonce+string(i)+"subscriber"), queries[i], config)
			if err != nil {
				log.Criticalf("Could not create client: %v", err)
				os.Exit(1)
			}
			start := time.Now()
			counted := false
			waitforme := make(chan bool)
			c.AttachDiffHandler(func(m *common.SubscriptionDiffMessage) {
				counted = true
				waitforme <- true
			})
			c.Start()
			<-waitforme
			wg.Done()
			diff := time.Since(start)
			if i%10 == 0 {
				log.Infof("iteration %d: %s", i, diff.String())
			}
			initialDiffLatencies[i] = float64(diff)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	wg.Wait()
	quartiles, _ := stats.Quartile(initialDiffLatencies)
	q1 := time.Duration(quartiles.Q1)
	q2 := time.Duration(quartiles.Q2)
	q3 := time.Duration(quartiles.Q3)
	mean, _ := stats.Mean(initialDiffLatencies)
	percent99, _ := stats.Percentile(initialDiffLatencies, 99)
	fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
	fmt.Printf("Mean: %s\n", time.Duration(mean).String())
	fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
	client.ManyColumnCSV([][]float64{initialDiffLatencies}, "latency_subscriptiondiff_diffquery.csv")
}
