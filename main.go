package main

import (
	"flag"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	if os.Getenv("LOGFORMAT") == "json" {
		log.SetFormatter(&log.JSONFormatter{
			FieldMap: log.FieldMap{
				log.FieldKeyTime:  "timestamp",
				log.FieldKeyLevel: "level",
				log.FieldKeyMsg:   "message",
			},
		})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}
	log.SetOutput(os.Stdout)

	loglevel := strings.ToLower(os.Getenv("LOGLEVEL"))
	if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "error" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	configFilePathArg := flag.String("config", "/etc/ouretl/default.conf", "file path for ouretl configuration")
	flag.Parse()

	config, err := newDefaultConfigFromTOMLFile(*configFilePathArg)
	if err != nil {
		log.Fatal("Config file could not be read. Check that the file exist, that it has the correct file permissions, and that it's in a valid TOML format.")
	}

	intercom := make(chan *defaultDataMessage)
	newWorkerPoolFromConfig(intercom, config)
	newHandlerPoolFromConfig(intercom, config)
}
