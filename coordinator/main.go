package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/profile"
	"golang.org/x/net/context"
	"os"
	"runtime/trace"
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
		case "trace":
			f, err := os.Create("trace.out")
			if err != nil {
				log.Fatal(err)
			}
			trace.Start(f)
		}
		time.AfterFunc(time.Duration(config.Debug.ProfileLength)*time.Second, func() {
			if p != nil {
				p.Stop()
			}
			if config.Debug.ProfileType == "trace" {
				trace.Stop()
			}

			os.Exit(0)
		})
	}

	server := NewServer(config)
	logStartKey := server.rebuildIfNecessary()
	go server.handleLeadership()
	go server.handleBrokerMessages()
	go server.monitorLog(logStartKey)
	go server.monitorGeneralConnections()
	server.listenAndDispatch()
}
