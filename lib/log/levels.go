package log

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"

	// LevelNone silences all log output.
	LevelNone Level = "none"
)

// Parse parses the given level string as a supported output format, while trying to
// maintain some degree of back-compat with the intent of previously supported log levels.
//
// This is exported only for internal use.
func (l Level) Parse() zapcore.Level {
	switch Level(strings.ToLower(string(l))) {
	case LevelDebug, "dbug":
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError, "eror",
		// We do not want to introduce 'Critical' support yet, since there are already an
		// abundance of log levels. Users of 'Critical' should just use 'Error' instead.
		"crit":
		return zapcore.ErrorLevel
	case LevelNone:
		// Logger does not export anything at the fatal level, so this effectively
		// silences all output.
		return zapcore.FatalLevel
	}
	// Quietly fall back to info
	return zapcore.InfoLevel
}
