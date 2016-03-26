//go:generate go tool yacc -o query.go -p cqbs query.y
package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/profile"
	"os"
	"time"
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

	if config.Debug.Enable {
		p := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
		time.AfterFunc(time.Duration(*config.Debug.ProfileLength)*time.Second, func() {
			p.Stop()
			os.Exit(0)
		})
	}

	server := NewServer(config)
	server.listenAndDispatch()
}
