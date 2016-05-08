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
var query string
var clientLatencies [][]float64
var clientLock sync.Mutex

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	rate = time.Duration(200) * time.Millisecond
	waitTime = 5 * time.Minute
	numPublishers = 12
	numClients = 10
	mdkey = "Room"
	//mdkey = fmt.Sprintf("Room%d", r.Int63())
	wg.Add(numClients * numPublishers) // wait for everyone
	clientLatencies = make([][]float64, numClients)
}

func runPub(idx, mdval int, rate time.Duration, pubConfig *client.Config) {
	publisher_seed := fmt.Sprintf("%d", r.Int63())
	// change mdval to change the metadata

	log.Errorf("UUID %v", client.UUIDFromName(publisher_seed))
	publisher, err := client.NewPublisher(client.UUIDFromName(publisher_seed), func(pub *client.Publisher) {
		ticker := time.Tick(rate)
		val := fmt.Sprintf("%d", mdval)
		pub.AddMetadata(map[string]interface{}{"Bucket": fmt.Sprintf("%d", idx), mdkey: val})
		log.Info(map[string]interface{}{"Bucket": fmt.Sprintf("%d", idx), mdkey: val})
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
	go func(pub *client.Publisher) {
		for {
			time.Sleep((30 * time.Second) + time.Duration(r.Int31n(30))*time.Second)
			old := mdval
			mdval = (mdval + 1) % 12
			val := fmt.Sprintf("%d", mdval)
			log.Infof("Changing metadata room from %d to %d", old, mdval)
			pub.AddMetadata(map[string]interface{}{"Bucket": fmt.Sprintf("%d", idx), mdkey: val})
		}
	}(publisher)
	publisher.Start()

	publisher.StopIn(waitTime)
	<-publisher.Stop
	wg.Done()
}

func run1sub10pub(idx int, subscribeChunk string, rate time.Duration, pubConfig, subConfig *client.Config) {
	clientLock.Lock()
	clientLatencies[idx] = []float64{}
	clientLock.Unlock()

	switch subscribeChunk {
	case "1":
		query = fmt.Sprintf("Bucket = '%d' and (%s = '0' or %s = '1' or %s = '2' or %s = '3')", idx, mdkey, mdkey, mdkey, mdkey)
	case "2":
		query = fmt.Sprintf("Bucket = '%d' and (%s = '4' or %s = '5' or %s = '6' or %s = '7')", idx, mdkey, mdkey, mdkey, mdkey)
	case "3":
		query = fmt.Sprintf("Bucket = '%d' and (%s = '8' or %s = '9' or %s = '10' or %s = '11')", idx, mdkey, mdkey, mdkey, mdkey)
	}
	subscriber_seed := fmt.Sprintf("%d", r.Int63())
	log.Info(query)
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
		if num%100 == 0 {
			fmt.Printf("Message: %v. Diff %s\n", num, time.Duration(diff).String())
		}
	})
	if err != nil {
		log.Criticalf("Could not create client (%v)", err)
		os.Exit(1)
	}
	c.Start()

	for i := 0; i < numPublishers; i++ {
		i := i
		time.Sleep(time.Duration(r.Int31n(100)) * time.Millisecond)
		go runPub(idx, i, rate, pubConfig)
	}

}

func main() {
	brokerIP := os.Args[1]
	subscribeChunk := os.Args[2]
	config := &client.Config{
		BrokerAddress:      fmt.Sprintf("%s:4444", brokerIP),
		CoordinatorAddress: "cs262.cal-sdb.org:5505",
	}

	for i := 0; i < numClients; i++ {
		idx := i
		go run1sub10pub(idx, subscribeChunk, rate, config, config)
	}
	wg.Wait()
	client.ManyColumnCSV(clientLatencies, []string{}, fmt.Sprintf("mdchange_data_%s.csv", subscribeChunk))
}
