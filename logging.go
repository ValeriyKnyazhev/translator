package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func CreateLogger() (*log.Logger, error) {
	logger := log.New()
	logger.Out = os.Stdout
	logger.SetLevel(log.DebugLevel)
	logger.Formatter = &log.TextFormatter{
		ForceColors: true,
	}
	return logger, nil
}
