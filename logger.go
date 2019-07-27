package generalprobe

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func setupLogger() {
	level := os.Getenv("GENERALPROBE_LOG_LEVEL")

	switch {
	case strings.EqualFold(level, "debug"):
		logger.SetLevel(logrus.DebugLevel)
	case strings.EqualFold(level, "info"):
		logger.SetLevel(logrus.InfoLevel)
	case strings.EqualFold(level, "warn"):
		logger.SetLevel(logrus.WarnLevel)
	case strings.EqualFold(level, "error"):
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.WarnLevel)
	}
}

// SetLoggerTraceLevel changes logging level to Trace
func SetLoggerTraceLevel() { logger.SetLevel(logrus.TraceLevel) }

// SetLoggerDebugLevel changes logging level to Debug
func SetLoggerDebugLevel() { logger.SetLevel(logrus.DebugLevel) }

// SetLoggerInfoLevel changes logging level to Info
func SetLoggerInfoLevel() { logger.SetLevel(logrus.InfoLevel) }

// SetLoggerWarnLevel changes logging level to Warn
func SetLoggerWarnLevel() { logger.SetLevel(logrus.WarnLevel) }
