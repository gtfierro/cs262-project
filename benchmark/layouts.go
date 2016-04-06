package main

type Layout struct {
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
	fractionPublishersFast:        0.9,
	clientsUseSameQuery:           false,
	tenthsOfClientsTouchedByQuery: 1,
	minPublisherCount:             50,
	maxPublisherCount:             200,
	publisherStepSize:             50,
	minClientCount:                50,
	maxClientCount:                200,
	clientStepSize:                50,
}

func GetLayoutByName(name *string) (layout *Layout) {
	switch *name {
	case "standard":
		layout = &Standard
	}
	return
}
