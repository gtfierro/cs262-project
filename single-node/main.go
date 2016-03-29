//go:generate go tool yacc -o query.go -p cqbs query.y
package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"cs262-project/common"
	"github.com/pkg/profile"
	"os"
	"time"
)

// config flags
var configfile = flag.String("c", "config.ini", "Path to configuration file")

func init() {
	// set up logging
	log.SetOutput(os.Stderr)
}

func main() {
	flag.Parse()
	// configure logging instance
	config := common.LoadConfig(*configfile)
	if config.Logging.UseJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	loglevel, err := log.ParseLevel(*config.Logging.Level)
	if err != nil {
		loglevel = log.InfoLevel // default to Info
	}
	log.SetLevel(loglevel)

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
