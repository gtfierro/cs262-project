package common

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/gcfg.v1"
)

// configuration for the logging
type LoggingConfig struct {
	// whether or not log outputs in JSON
	UseJSON bool
	Level   string
}

// server configuration
type ServerConfig struct {
	// if true, listens on 0.0.0.0
	Global bool
	// Client Interface
	Port int
	// the public-facing address of the broker
	Host string
	// A unique key for this Broker
	BrokerID UUID
	// the name of the coordinator server
	CoordinatorHost string
	// the port of the coordinator server
	CoordinatorPort int
	// if true, then the broker evaluates metadata locally
	// (distribution option #2). If false, then it forwards
	// queries to the coordinator (distribution option #1).
	// Note that if running in single-node mode, without
	// a coordinator, this must be TRUE.
	LocalEvaluation bool
}

// Coordinator configuration
type CoordinatorConfig struct {
	Port              int
	Global            bool
	HeartbeatInterval int    // seconds
	CoordinatorCount  int    // number of coordinators total
	EtcdAddresses     string // comma-separated list
	InstanceId        string // AWS instance id, e.g. i-1a2b3c4d
	Region            string // AWS region name, e.g. us-west-1
	ElasticIP         string // the elastic IP to fight over
}

// MongoDB configuration
type MongoConfig struct {
	Port     int
	Host     string
	Database string
}

// Debugging configuration
type DebugConfig struct {
	Enable        bool
	ProfileType   string
	ProfileLength int
}
type BenchmarkConfig struct {
	BrokerURL         string
	BrokerPort        int
	StepSpacing       int    // How long between increasing client/publisher counts (seconds)
	ConfigurationName string // Named bundle of query/metadata
}

type Config struct {
	Logging     LoggingConfig
	Server      ServerConfig
	Mongo       MongoConfig
	Coordinator CoordinatorConfig
	Debug       DebugConfig
	Benchmark   BenchmarkConfig
}

// Don't want to log anything since this is called before SetupLogging;
// return the desired log message to be logged later if desired
// TODO this makes it so that default_config.ini overwrites config.ini...
func LoadConfig(filename string) (config *Config, logmsg string) {
	config = new(Config)
	err := gcfg.ReadFileInto(config, filename)
	if err == nil {
		logmsg = fmt.Sprintf("Using local config at %v", filename)
		return
	}
	err = gcfg.ReadFileInto(config, "./config.ini")
	if err == nil {
		return
	}
	err = gcfg.ReadFileInto(config, "./default_config.ini")
	if err == nil {
		logmsg = "Using default configuration found at ./default_config.ini. "
	} else {
		logmsg = fmt.Sprintf("Unable to find any configuration file at ./default_config.ini, ./config.ini, or %v", filename)
	}
	return
}

func SetupLogging(config *Config) {
	if config.Logging.UseJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	loglevel, err := log.ParseLevel(config.Logging.Level)
	log.Infof("Using log level %v", loglevel)
	if err != nil {
		log.Error(err)
		loglevel = log.InfoLevel // default to Info
	}
	log.SetLevel(loglevel)
}
