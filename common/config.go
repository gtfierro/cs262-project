package common

import (
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
	Port int
	// if true, listens on 0.0.0.0
	Global bool
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
	BrokerURL         *string
	BrokerPort        *int
	StepSpacing       *int    // How long between increasing client/producer counts (seconds)
	ConfigurationName *string // Named bundle of query/metadata
}

type Config struct {
	Logging   LoggingConfig
	Server    ServerConfig
	Mongo     MongoConfig
	Debug     DebugConfig
	Benchmark BenchmarkConfig
}

func LoadConfig(filename string) (config *Config) {
	config = new(Config)
	defaultConfigFound := false
	err := gcfg.ReadFileInto(config, "./default_config.ini")
	if err != nil {
		log.WithFields(log.Fields{
			"location": filename,
			"error":    err,
		}).Info("Couldn't load configuration file ./default_config.ini; continuing to try given location")
	} else {
		defaultConfigFound = true
		log.Info("Using default values from ./default_config.ini")
	}
	err = gcfg.ReadFileInto(config, filename)
	if err != nil {
		log.WithFields(log.Fields{
			"location": filename,
			"error":    err,
		}).Warn("Couldn't load configuration file at given location. Trying local ./config.ini")
	} else {
		return
	}
	err = gcfg.ReadFileInto(config, "./config.ini")
	if err != nil && !defaultConfigFound {
		log.WithField("error", err).Warn("Couldn't load configuration file at ./config.ini")
	}
	return
}

func SetupLogging(config *Config) {
	if config.Logging.UseJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	loglevel, err := log.ParseLevel(config.Logging.Level)
	if err != nil {
		log.Error(err)
		loglevel = log.InfoLevel // default to Info
	}
	log.SetLevel(loglevel)
}
