package main

import (
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
	minClientCount                int
	maxClientCount                int
	clientStepSize                int
}

var Standard = Layout{
	name: "standard",
	fractionPublishersFast:        0.1,
	clientsUseSameQuery:           false,
	tenthsOfClientsTouchedByQuery: 1,
	minPublisherCount:             50,
	maxPublisherCount:             200,
	publisherStepSize:             50,
	minClientCount:                50,
	maxClientCount:                200,
	clientStepSize:                50,
}

var allLayouts = []Layout{Standard}

func GetLayoutByName(name string) (layout *Layout) {
	for _, l := range allLayouts {
		if l.name == name {
			return &l
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
		"minClientCount":               l.minClientCount,
		"maxClientCount":               l.maxClientCount,
		"clientStepSize":               l.clientStepSize,
	}).Info("Current Layout in use is")
}
