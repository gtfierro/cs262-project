//go:generate go tool yacc -o query.go -p cqbs query.y
package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
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
	common.SetupLogging(config)

	if config.Debug.Enable {
		p := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
		time.AfterFunc(time.Duration(config.Debug.ProfileLength)*time.Second, func() {
			p.Stop()
			os.Exit(0)
		})
	}

	server := NewServer(config)
	server.listenAndDispatch()
}
