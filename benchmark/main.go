package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/nu7hatch/gouuid"
)

const FastPublishFrequency = 100 // per minute
const SlowPublishFrequency = 6

// config flags
var configfile = flag.String("c", "config.ini", "Path to configuration file")

func init() {
	// set up logging
	log.SetOutput(os.Stderr)
}

func main() {
	config := common.LoadConfig(*configfile)
	common.SetupLogging(config)
	layout := GetLayoutByName(config.Benchmark.ConfigurationName)

	publishers := initializePublishers(layout, *config.Benchmark.BrokerURL, *config.Benchmark.BrokerPort)

	latencyChan := make(chan int64, 1e6) // TODO proper sizing?
	clients := initializeClients(layout, *config.Benchmark.BrokerURL, *config.Benchmark.BrokerPort, latencyChan)

	var wg sync.WaitGroup

	// Start up the initial clients and publishers
	startClientsAndPublishers(0, layout.minClientCount, 0, layout.minPublisherCount, clients, publishers, wg)
	clientsRunning := layout.minClientCount
	publishersRunning := layout.minPublisherCount

	var latencySum *int64 = new(int64)
	var latencyCount *int32 = new(int32)
	go logLatencyAverages(latencySum, latencyCount)

	// TODO starting multiple in the hopes of keeping up, no idea if that's
	// necessary or helpful
	for i := 0; i < 5; i++ {
		go pollAndIncrementLatencyValues(latencyChan, latencySum, latencyCount)
	}

	time.Sleep(time.Second * time.Duration(*config.Benchmark.StepSpacing))

	// Every StepSpacing seconds, spin up more publishers and clients
	// until the max is reached
	for clientsRunning < layout.maxClientCount || publishersRunning < layout.maxPublisherCount {
		clientGoal := min(clientsRunning+layout.clientStepSize, layout.maxClientCount)
		publisherGoal := min(publishersRunning+layout.publisherStepSize, layout.maxPublisherCount)
		startClientsAndPublishers(clientsRunning, clientGoal, publishersRunning, publisherGoal, clients, publishers, wg)
		clientsRunning = clientGoal
		publishersRunning = publisherGoal
		time.Sleep(time.Second * time.Duration(*config.Benchmark.StepSpacing))
	}

	for _, p := range publishers {
		p.stop <- true
	}

	wg.Wait()
}

func min(a int, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

// Start up clients and publishers as necessary until there are clientGoal clients and
// publisherGoal publishers running.
func startClientsAndPublishers(clientsRunning int, clientGoal int, publishersRunning int,
	publisherGoal int, clients []Client, publishers []Publisher, wg sync.WaitGroup) {
	for currentClients := clientsRunning; currentClients < clientGoal; currentClients++ {
		go clients[currentClients].subscribe()
	}
	for currentPublishers := publishersRunning; currentPublishers < publisherGoal; currentPublishers++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			publishers[i].publishContinuously()
		}(currentPublishers)
	}
}

func initializeClients(layout *Layout, brokerURL string, brokerPort int, latencyChan chan int64) []Client {
	var query string
	if layout.clientsUseSameQuery {
		// Generate one query for all of them
		query = genQueryString(layout.tenthsOfClientsTouchedByQuery)
	}
	clients := make([]Client, layout.maxClientCount)
	for i := 0; i < layout.maxClientCount; i++ {
		if !layout.clientsUseSameQuery {
			// New query each time
			query = genQueryString(layout.tenthsOfClientsTouchedByQuery)
		}
		clients[i] = Client{
			BrokerURL:   brokerURL,
			BrokerPort:  brokerPort,
			Query:       query,
			latencyChan: latencyChan,
		}
	}
	return clients
}

func initializePublishers(layout *Layout, brokerURL string, brokerPort int) []Publisher {
	var (
		freq     int
		u        *uuid.UUID
		metadata = make(map[string]interface{})
	)
	roomNumber := 0
	publishers := make([]Publisher, layout.maxPublisherCount)
	for i := 0; i < layout.maxPublisherCount; i++ {
		u, _ = uuid.NewV4()
		if rand.Float64() < layout.fractionPublishersFast {
			freq = FastPublishFrequency
		} else {
			freq = SlowPublishFrequency
		}
		metadata["Building"] = "Soda"
		roomNumber++
		metadata["Room"] = string('a' + (roomNumber % 10)) // to easily select fractions of publishers
		metadata["Type"] = "Counter"
		publishers[i] = Publisher{
			BrokerURL:  brokerURL,
			BrokerPort: brokerPort,
			Metadata:   metadata,
			Frequency:  freq,
			uuid:       common.UUID(u.String()),
			stop:       make(chan bool),
		}
	}
	return publishers
}

func pollAndIncrementLatencyValues(latencyChan chan int64, latencySum *int64, latencyCount *int32) {
	for {
		newLatency := <-latencyChan
		atomic.AddInt64(latencySum, newLatency)
		atomic.AddInt32(latencyCount, 1)
	}
}

func logLatencyAverages(latencySum *int64, latencyCount *int32) {
	for {
		time.Sleep(time.Second)
		oldLatencySum := atomic.SwapInt64(latencySum, 0)
		oldLatencyCount := atomic.SwapInt32(latencyCount, 0)
		log.WithFields(log.Fields{
			"messageCount":     oldLatencyCount,
			"averageLatencyNS": float64(oldLatencySum) / float64(oldLatencyCount),
		}).Info("Received message count and average latency over the last second")
	}
}

func genQueryString(tenthsOfPublishers int) string {
	//Introduce a portion of the query that is always true and contains
	//a random string so that it will be treated as a separate query by
	//the broker
	randChars := make([]string, 20)
	for i := 0; i < len(randChars); i++ {
		randChars[i] = string('A' + rand.Int31n('Z'-'A'))
	}
	randStr := strings.Join(randChars, "")
	if tenthsOfPublishers == 0 {
		return fmt.Sprintf("Building = 'Foobar' and Type != '%v'", randStr)
	} else if tenthsOfPublishers == 10 {
		return fmt.Sprintf("Building = 'Soda' and Type != '%v'", randStr)
	}
	// Since we can't do range queries, just matching against characters
	charsUsed := make(map[string]bool)
	charArray := make([]string, tenthsOfPublishers)
	for len(charsUsed) < tenthsOfPublishers {
		c := string('a' + rand.Int31n(10))
		if _, ok := charsUsed[c]; !ok {
			charArray[len(charsUsed)] = c
			charsUsed[c] = true
		}
	}
	//TODO ETK this should be changed to set inclusion once that's available
	return fmt.Sprintf("Room = '%v' and Type != '%v'",
		strings.Join(charArray, "' or Room = '"), randStr)
}
