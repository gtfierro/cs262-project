package common

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func SetupTestLogging() {
	logLevelStr := flag.String("l", "FATAL", "Log level during testing")
	flag.Parse()
	logLevel, _ := log.ParseLevel(*logLevelStr)
	log.SetLevel(logLevel)
}

// Poor man's deep equality check because I can't get anything else to work
func AssertStrEqual(assert *require.Assertions, a, b interface{}) {
	assert.Equal(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

func AssertSendableChanEmpty(assert *require.Assertions, channel chan Sendable) {
	select {
	case <-channel:
		assert.Fail("Channel not empty")
	default:
	}
}
