package eazydb

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func initLogger(logger *logrus.Logger, enabled bool) *logrus.Logger {
	if logger != nil {
		return logger
	}

	levelstr := os.Getenv("EAZYDB_LOG_LEVEL")
	var level logrus.Level = 4
	switch levelstr {
	case "panic":
		level = 0
	case "fatal":
		level = 1
	case "error":
		level = 2
	case "warning":
		level = 4
	case "info":
		level = 5
	case "debug":
		level = 6
	case "trace":
		level = 7
	}

	log := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	if !enabled {
		log.Out = io.Discard
	}

	return log
}
