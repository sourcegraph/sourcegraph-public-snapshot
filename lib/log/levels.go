package log

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

const envSrcLogLevel = "SRC_LOG_LEVEL"

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"

	// LevelNone silences all log output.
	LevelNone Level = "none"
)

func (l Level) Parse() zapcore.Level {
	switch Level(strings.ToLower(string(l))) {
	case LevelDebug, "dbug":
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError, "eror", "crit":
		return zapcore.ErrorLevel
	case LevelNone:
		// Logger does not export anything at the fatal level, so this effectively
		// silences all output.
		return zapcore.FatalLevel
	}
	// Quietly fall back to info
	return zapcore.InfoLevel
}
