package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/tinylib/msgp/msgp"
	"math/rand"
	"net"
	"time"
)

type Publisher struct {
	BrokerURL               string
	BrokerPort              int
	BaseMetadata            map[string]interface{}
	AdditionalMetadataSize  int
	MetadataRefreshInterval int  // seconds between metadata refreshes; -1 to never refresh
	MetadataRefreshRandom   bool // if true, wait a random interval between 0 and MetadataRefreshInterval
	uuid                    common.UUID
	Frequency               int // per minute
	stop                    chan bool
}

func (p *Publisher) publishContinuously() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", p.BrokerURL, p.BrokerPort))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("Error creating connection to %v:%v", p.BrokerURL, p.BrokerPort)
	}

	spacingMs := 60e3 / float64(p.Frequency)

	metadata := createFullMetadata(p.BaseMetadata, p.AdditionalMetadataSize)

	enc := msgp.NewWriter(conn)
	msg := common.PublishMessage{UUID: p.uuid, Metadata: metadata, Value: time.Now().UnixNano()}
	msg.EncodeMsg(enc)
	msg.Metadata = nil // Clear MD after sending

	mdRefreshStop := make(chan bool)
	newMD := make(chan *map[string]interface{})
	if p.MetadataRefreshInterval >= 0 {
		go func() {
		MDRefreshLoop:
			for {
				var nextInterval time.Duration
				if p.MetadataRefreshRandom {
					nextInterval = time.Second * time.Duration(int(rand.Float32()*float32(p.MetadataRefreshInterval)))
				} else {
					nextInterval = time.Second * time.Duration(p.MetadataRefreshInterval)
				}
				select {
				case <-mdRefreshStop:
					break MDRefreshLoop
				case <-time.After(nextInterval):
					md := createFullMetadata(p.BaseMetadata, p.AdditionalMetadataSize)
					newMD <- &md
				}
			}
		}()
	}
Loop:
	for {
		select {
		case <-p.stop:
			mdRefreshStop <- true
			break Loop
		case <-time.After(time.Millisecond * time.Duration(spacingMs)):
			select {
			case md := <-newMD:
				msg.Metadata = *md
			default:
				msg.Metadata = nil
			}
			msg.Value = time.Now().UnixNano()
			msg.Encode(enc)
			enc.Flush()
		}
	}
	conn.Close()
}

// Combine the baseMetadata with another additionalSize new keys all with randomly
// generated values
func createFullMetadata(baseMetadata map[string]interface{}, additionalSize int) map[string]interface{} {
	var retmap = make(map[string]interface{}, len(baseMetadata)+additionalSize)
	for k, v := range baseMetadata {
		retmap[k] = v
	}
	for i := 0; i < additionalSize; i++ {
		retmap[fmt.Sprintf("key%v", i)] = fmt.Sprintf("%v", rand.Int())
	}
	return retmap
}
