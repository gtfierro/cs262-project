package main

import (
	"cs262-project/common"
	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
	"flag"
	"os"
	"math/rand"
	"fmt"
	"strings"
	"time"
	"sync"
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
	var (
		freq int
		u *uuid.UUID
		metadata = make(map[string]interface{})
		query string
	)
	config := common.LoadConfig(*configfile)
	common.SetupLogging(config)
	layout := GetLayoutByName(config.Benchmark.ConfigurationName)

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
			BrokerURL: *config.Benchmark.BrokerURL,
			BrokerPort: *config.Benchmark.BrokerPort,
			Metadata: metadata,
			Frequency: freq,
			uuid: common.UUID(u.String()),
			stop: make(chan bool),
		}
	}

	msgCountChan := make(chan int, 50)
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
			BrokerURL: *config.Benchmark.BrokerURL,
			BrokerPort: *config.Benchmark.BrokerPort,
			Query: query,
			msgCount: msgCountChan,
		}
	}

	var wg sync.WaitGroup
	//wg.Add(layout.maxPublisherCount)

	// Start up the initial clients and publishers
	clientsRunning := layout.minClientCount
	for i := 0; i < layout.minClientCount; i++ {
		go clients[i].subscribe()
	}
	publishersRunning := 0
	publishersRunning = layout.minPublisherCount
	for i := 0; i < layout.minPublisherCount; i++ {
		wg.Add(1)
		go func (index int) {
			defer wg.Done()
			publishers[index].publishContinuously()
		}(i)
	}
	time.Sleep(time.Second * time.Duration(*config.Benchmark.StepSpacing))

	go func() {
		for {
			amt := <-msgCountChan
			log.WithField("count", amt).Debug("Some client received messages!")
		}
	}()

	// Every StepSpacing seconds, spin up more publishers and clients
	// until the max is reached
	for clientsRunning < layout.maxClientCount ||
	publishersRunning < layout.maxPublisherCount {
		nextClientGoal := clientsRunning + layout.clientStepSize
		for clientsRunning < nextClientGoal && clientsRunning < layout.maxClientCount {
			go clients[clientsRunning].subscribe()
			clientsRunning++
		}
		nextPublisherGoal := publishersRunning + layout.publisherStepSize
		for publishersRunning < nextPublisherGoal && publishersRunning < layout.maxPublisherCount {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				publishers[i].publishContinuously()
			}(publishersRunning)
			publishersRunning++
		}
		time.Sleep(time.Second * time.Duration(*config.Benchmark.StepSpacing))
	}

	for _, p := range publishers {
		p.stop <- true
	}

	wg.Wait()
}

func genQueryString(tenthsOfPublishers int) string {
	//Introduce a portion of the query that is always true and contains
	//a random string so that it will be treated as a separate query by
	//the broker
	randChars := make([]string, 20)
	for i := 0; i < len(randChars); i++ {
		randChars[i] = string('A' + rand.Int31n('Z' - 'A'))
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