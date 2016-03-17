package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"os"
)

// global logging instance
var log *logrus.Logger

// config flags
var configfile = flag.String("c", "config.ini", "Path to configuration file")

func init() {
	// set up logging
	log = logrus.New()
	log.Out = os.Stderr
}

func main() {
	flag.Parse()
	// configure logging instance
	config := LoadConfig(*configfile)
	if config.Logging.UseJSON {
		log.Formatter = &logrus.JSONFormatter{}
	}

	server := NewServer(config)
	server.listenAndDispatch()
}
