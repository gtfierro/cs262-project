package main

import (
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Layout struct {
	name                          string
	fractionPublishersFast        float64
	clientsUseSameQuery           bool
	tenthsOfClientsTouchedByQuery int
	minPublisherCount             int
	maxPublisherCount             int
	publisherStepSize             int
	publisherMDRefreshInterval    int  // seconds between publisher MD updates; set to -1 to never refresh
	publisherMDRefreshRandom      bool // if true, refreshes after a random interval between
	// [0, publisherMDRefreshInterval)
	publisherMDSize int // The number of key-value pairs to pad the publisher MD with
	minClientCount  int
	maxClientCount  int
	clientStepSize  int
}

var allLayouts = make([]*Layout, 0)

func init() {
	var standard = Layout{
		name: "Standard",
		fractionPublishersFast:        0.1,
		clientsUseSameQuery:           false,
		tenthsOfClientsTouchedByQuery: 1,
		minPublisherCount:             50,
		maxPublisherCount:             200,
		publisherStepSize:             50,
		publisherMDRefreshInterval:    -1,
		publisherMDSize:               5,
		minClientCount:                50,
		maxClientCount:                200,
		clientStepSize:                50,
	}
	allLayouts = append(allLayouts, &standard)
	var standardMDRefresh = standard
	standardMDRefresh.name = "StandardMDRefresh"
	standardMDRefresh.publisherMDRefreshInterval = 10
	standardMDRefresh.publisherMDRefreshRandom = true
	allLayouts = append(allLayouts, &standardMDRefresh)
}

func GetLayoutByName(name string) (layout *Layout) {
	for _, l := range allLayouts {
		if strings.EqualFold(l.name, name) {
			return l
		}
	}
	return nil
}

func (l *Layout) logSelf() {
	// Log as separate fields to more easily JSON-parse
	log.WithFields(log.Fields{
		"name": l.name,
		"fractionPublishersFast":       l.fractionPublishersFast,
		"clientsUseSameQuery":          l.clientsUseSameQuery,
		"tentsOfClientsTouchedByQuery": l.tenthsOfClientsTouchedByQuery,
		"minPublisherCount":            l.minPublisherCount,
		"maxPublisherCount":            l.maxPublisherCount,
		"publisherStepSize":            l.publisherStepSize,
		"publisherMDRefreshInterval":   l.publisherMDRefreshInterval,
		"publisherMDRefreshRandom":     l.publisherMDRefreshRandom,
		"publisherMDSize":              l.publisherMDSize,
		"minClientCount":               l.minClientCount,
		"maxClientCount":               l.maxClientCount,
		"clientStepSize":               l.clientStepSize,
	}).Info("Current Layout in use is")
}
