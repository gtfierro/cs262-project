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
	// Client Interface
	Port int
	// if true, listens on 0.0.0.0
	Global bool
	// the name of the central server
	CentralHost string
	// the port of the central server
	CentralPort int
}

// MongoDB configuration
type MongoConfig struct {
	Port int
	Host string
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
	StepSpacing       int    // How long between increasing client/producer counts (seconds)
	ConfigurationName string // Named bundle of query/metadata
}

type Config struct {
	Logging   LoggingConfig
	Server    ServerConfig
	Mongo     MongoConfig
	Debug     DebugConfig
	Benchmark BenchmarkConfig
}

// Don't want to log anything since this is called before SetupLogging;
// return the desired log message to be logged later if desired
func LoadConfig(filename string) (config *Config, logmsg string) {
	config = new(Config)
	err := gcfg.ReadFileInto(config, "./default_config.ini")
	defaultConfigMsg := ""
	if err == nil {
		defaultConfigMsg = "Using default configuration found at ./default_config.ini. "
	}
	err = gcfg.ReadFileInto(config, filename)
	if err == nil {
		logmsg = fmt.Sprintf("%vUsing local config at %v", defaultConfigMsg, filename)
		return
	}
	err = gcfg.ReadFileInto(config, "./config.ini")
	if err != nil && defaultConfigMsg == "" {
		logmsg = fmt.Sprintf("Unable to find any configuration file at ./default_config.ini, ./config.ini, or %v", filename)
	} else {
		logmsg = fmt.Sprintf("%vUsing local config at ./config.ini", defaultConfigMsg)
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
