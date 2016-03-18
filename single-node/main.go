//go:generate go tool yacc -o query.go -p cqbs query.y
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
	loglevel, err := logrus.ParseLevel(*config.Logging.Level)
	if err != nil {
		loglevel = logrus.InfoLevel // default to Info
	}
	log.Level = loglevel

	server := NewServer(config)
	server.listenAndDispatch()
}
