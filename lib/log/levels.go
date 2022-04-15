package log

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const envSrcLogLevel = "SRC_LOG_LEVEL"

func parseLevel(l string) zapcore.Level {
	switch strings.ToLower(l) {
	case "debug", "dbug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error", "eror":
		return zapcore.ErrorLevel
	case "crit":
		return zapcore.DPanicLevel
	}
	// Quietly fall back to info
	return zapcore.InfoLevel
}

func watchLogLevel(logLevel zap.AtomicLevel) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		}

		level := parseLevel(os.Getenv(envSrcLogLevel))
		logLevel.SetLevel(level)
	}
}
