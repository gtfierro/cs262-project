//go:generate go tool yacc -o query.go -p cqbs query.y
package main

import (
	"flag"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/profile"
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
	config, configLogmsg := common.LoadConfig(*configfile)
	common.SetupLogging(config)
	log.Info(configLogmsg)

	if config.Debug.Enable {
		var p interface {
			Stop()
		}
		switch config.Debug.ProfileType {
		case "cpu", "profile":
			p = profile.Start(profile.CPUProfile, profile.ProfilePath("."))
		case "block":
			p = profile.Start(profile.BlockProfile, profile.ProfilePath("."))
		case "mem":
			p = profile.Start(profile.MemProfile, profile.ProfilePath("."))
		}
		time.AfterFunc(time.Duration(config.Debug.ProfileLength)*time.Second, func() {
			p.Stop()
			os.Exit(0)
		})
	}

	server := NewServer(config)
	server.listenAndDispatch()
}
