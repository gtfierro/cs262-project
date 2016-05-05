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

// 1 publisher, N subscribers. Each subscriber has a different query For a
// benchmark time of T seconds, every T/N seconds, the publisher changes its
// metadata to qualify for a new subscriber.

var (
	log        *logging.Logger
	numClients = 100
	runTime    = 15 * time.Minute
	rate       = 1 * time.Second
	r          rand.Source
	mdkey      string
)

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
	r = rand.NewSource(time.Now().UnixNano())
	// key used in query
	mdkey = fmt.Sprintf("Room%d", r.Int63())
}

func main() {
	brokerIP := os.Args[1]
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
	}

	var l sync.Mutex
	clientLatencies := map[int][]float64{}
	clientTimes := map[int][]int64{}
	for i := 0; i < numClients; i++ {
		i := i
		latencies := []float64{}
		times := []int64{}
		query := fmt.Sprintf("%s = '%s'", mdkey, fmt.Sprintf("%d", i))
		log.Infof("QUERY: %s", query)
		c, err := client.NewClient(client.UUIDFromName(string(r.Int63())), query, config)
		c.AttachPublishHandler(func(m *common.PublishMessage) {
			recv_time := time.Now().UnixNano()
			sent_time := m.Value.(int64)
			diff := recv_time - sent_time
			latencies = append(latencies, float64(diff))
			times = append(times, sent_time)
			l.Lock()
			clientLatencies[i] = latencies
			clientTimes[i] = times
			l.Unlock()
			fmt.Printf("%d Message: %v. Diff %s\n", i, len(latencies), time.Duration(diff).String())
		})
		if err != nil {
			log.Criticalf("Could not create client (%v)", err)
			os.Exit(1)
		}
		c.Start()
	}
	time.Sleep(1 * time.Second)

	// period for switching metadata
	switching := runTime / time.Duration(numClients)
	mdIdx := 0 // which "room" we are on
	publisher, err := client.NewPublisher(client.UUIDFromName(string(r.Int63())), func(pub *client.Publisher) {
		ticker := time.Tick(switching)
		pub.AddMetadata(map[string]interface{}{mdkey: fmt.Sprintf("%d", mdIdx)})
		log.Infof("New metadata: %v", map[string]interface{}{mdkey: fmt.Sprintf("%d", mdIdx)})
		for _ = range ticker {
			mdIdx += 1
			pub.AddMetadata(map[string]interface{}{mdkey: fmt.Sprintf("%d", mdIdx)})
			log.Infof("New metadata: %v", map[string]interface{}{mdkey: fmt.Sprintf("%d", mdIdx)})
		}
	}, config)
	if err != nil {
		log.Criticalf("Could not create publisher (%v)", err)
		os.Exit(1)
	}
	publisher.Start()
	publisher.StopIn(runTime)
	go func(pub *client.Publisher) {
		ticker := time.Tick(rate)
		for now := range ticker {
			if err := pub.Publish(now.UnixNano()); err != nil {
				log.Error(err)
			}
		}
	}(publisher)
	<-publisher.Stop
	l.Lock()

	log.Info(clientTimes)
	log.Info(clientLatencies)
	timeseries := make([][]float64, 2)
	data := make([]float64, len(clientTimes[0])*numClients)
	for i, times := range clientTimes {
		for j, p := range times {
			data[j+i*len(times)] = float64(p)
		}
	}
	timeseries[0] = data
	log.Info(data)

	data2 := make([]float64, len(clientLatencies[0])*numClients)
	for i, latencies := range clientLatencies {
		for j, p := range latencies {
			data2[j+i*len(latencies)] = float64(p)
		}
	}
	timeseries[1] = data2
	log.Info(data2)
	log.Info(timeseries)
	client.ManyColumnCSV(timeseries, []string{"time", "latency"}, "metadata_change_clients.csv")

	for _, data := range clientLatencies {
		quartiles, _ := stats.Quartile(data)
		q1 := time.Duration(quartiles.Q1)
		q2 := time.Duration(quartiles.Q2)
		q3 := time.Duration(quartiles.Q3)
		mean, _ := stats.Mean(data)
		percent99, _ := stats.Percentile(data, 99)
		fmt.Printf("Quartiles. 25%% %s, 50%% %s, 75%% %s\n", q1.String(), q2.String(), q3.String())
		fmt.Printf("Mean: %s\n", time.Duration(mean).String())
		fmt.Printf("99 percentile: %s\n", time.Duration(percent99).String())
	}
	l.Unlock()
}
