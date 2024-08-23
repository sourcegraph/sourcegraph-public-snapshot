package logr

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/log"
)

// New instantiates a new logr.Logger that sends entries to the given
// log.Logger.
func New(l log.Logger) logr.Logger { return logr.New(&LogSink{Logger: l}) }

// LogSink implements logr.LogSink, backed by a sourcegraph/log.Logger.
type LogSink struct{ log.Logger }

var _ logr.LogSink = &LogSink{}

// Init receives optional information about the logr library for LogSink
// implementations that need it.
func (s *LogSink) Init(info logr.RuntimeInfo) {
	// This method mutates, so we need a pointer receiver and an update to
	// its Logger.
	s.Logger = s.AddCallerSkip(info.CallDepth)
}

// Enabled tests whether this LogSink is enabled at the specified V-level.
// For example, commandline flags might be used to set the logging
// verbosity and disable some info logs.
func (s LogSink) Enabled(level int) bool {
	// Let underlying Logger handle enabling/disabling entries
	return true
}

// Info logs a non-error message with the given key/value pairs as context.
// The level argument is provided for optional logging.  This method will
// only be called when Enabled(level) is true. See Logger.Info for more
// details.
func (s LogSink) Info(level int, msg string, keysAndValues ...any) {
	fields := toLogFields(keysAndValues)
	zl := toZapLevel(level)

	switch {
	case zl >= zapcore.ErrorLevel:
		s.Logger.Error(msg, fields...)
	case zl == zapcore.WarnLevel:
		s.Logger.Warn(msg, fields...)
	case zl == zapcore.InfoLevel:
		s.Logger.Info(msg, fields...)
	default:
		s.Logger.Debug(msg, fields...)
	}
}

// Error logs an error, with the given message and key/value pairs as
// context.  See Logger.Error for more details.
func (s LogSink) Error(err error, msg string, keysAndValues ...any) {
	fields := toLogFields(keysAndValues)
	s.Logger.Error(msg, append(fields, log.Error(err))...)
}

// WithValues returns a new LogSink with additional key/value pairs.  See
// Logger.WithValues for more details.
func (s LogSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &LogSink{s.Logger.With(toLogFields(keysAndValues)...)}
}

// WithName returns a new LogSink with the specified name appended.  See
// Logger.WithName for more details.
func (s LogSink) WithName(name string) logr.LogSink {
	return &LogSink{s.Logger.Scoped(name)}
}
