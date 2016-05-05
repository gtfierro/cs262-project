package main

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/client"
	"github.com/gtfierro/cs262-project/common"
	"math/rand"
	"os"
	"sync"
	"time"
)

// N publishers to a single subscriber

var log *logging.Logger
var r *rand.Rand
var wg sync.WaitGroup
var rate time.Duration
var waitTime time.Duration
var numPublishers int
var numClients int
var mdkey string
var mdval string
var clientLatencies [][]float64
var clientLock sync.Mutex

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	rate = time.Duration(200) * time.Millisecond
	waitTime = 10 * time.Minute
	numPublishers = 10
	numClients = 100
	mdkey = fmt.Sprintf("Room%d", r.Int63())
	mdval = fmt.Sprintf("%d", r.Int63())
	wg.Add(numClients * numPublishers) // wait for everyone
	clientLatencies = make([][]float64, numClients)
}

func runPub(val string, rate time.Duration, pubConfig *client.Config) {
	publisher_seed := fmt.Sprintf("%d", r.Int63())

	publisher, err := client.NewPublisher(client.UUIDFromName(publisher_seed), func(pub *client.Publisher) {
		ticker := time.Tick(rate)
		pub.AddMetadata(map[string]interface{}{mdkey: val})
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

	publisher.StopIn(waitTime)
	<-publisher.Stop
	wg.Done()
}

func run1sub10pub(idx int, rate time.Duration, pubConfig, subConfig *client.Config) {
	clientLock.Lock()
	clientLatencies[idx] = []float64{}
	clientLock.Unlock()

	subscriber_seed := fmt.Sprintf("%d", r.Int63())
	val := fmt.Sprintf("%d", idx)
	query := fmt.Sprintf("%s = '%s'", mdkey, val)
	c, err := client.NewClient(client.UUIDFromName(subscriber_seed), query, subConfig)
	num := 0
	c.AttachPublishHandler(func(m *common.PublishMessage) {
		num += 1
		recv_time := time.Now().UnixNano()
		sent_time := m.Value.(int64)
		diff := recv_time - sent_time
		clientLock.Lock()
		clientLatencies[idx] = append(clientLatencies[idx], float64(diff))
		clientLock.Unlock()
		if num%1000 == 0 {
			fmt.Printf("Message: %v. Diff %s\n", num, time.Duration(diff).String())
		}
	})
	if err != nil {
		log.Criticalf("Could not create client (%v)", err)
		os.Exit(1)
	}
	c.Start()

	for i := 0; i < numPublishers; i++ {
		time.Sleep(time.Duration(r.Int31n(100)) * time.Millisecond)
		go runPub(val, rate, pubConfig)
	}

}

func main() {
	brokerIP := os.Args[1]
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
	}

	for i := 0; i < numClients; i++ {
		idx := i
		time.Sleep(1 * time.Second)
		go run1sub10pub(idx, rate, config, config)
	}
	wg.Wait()
	client.ManyColumnCSV(clientLatencies, []string{}, "data.csv")
}
