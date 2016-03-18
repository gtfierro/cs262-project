package main

import (
	"github.com/Sirupsen/logrus"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	// configuration for the logging
	Logging struct {
		// whether or not log outputs in JSON
		UseJSON bool
		Level   *string
	}

	// server configuration
	Server struct {
		Port *int
		// if true, listens on 0.0.0.0
		Global bool
	}

	// MongoDB configuration
	Mongo struct {
		Port *int
		Host *string
	}
}

func LoadConfig(filename string) (config *Config) {
	config = new(Config)
	err := gcfg.ReadFileInto(config, filename)
	if err != nil {
		log.WithFields(logrus.Fields{"location": filename}).Error("No configuration file found at given location. Trying local ./config.ini")
	} else {
		return
	}
	err = gcfg.ReadFileInto(config, "./config.ini")
	if err != nil {
		log.Fatal("No configuration file found at ./config.ini")
	}
	return
}
