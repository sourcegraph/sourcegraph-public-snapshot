package std

import (
	"bytes"
	stdlog "log"

	"github.com/sourcegraph/sourcegraph/lib/log"
)

// NewLogger creates a standard library logger that writes to logger at the designated
// level. This is useful for providing loggers to libraries that only accept the standard
// library logger.
func NewLogger(logger log.Logger, level log.Level) *stdlog.Logger {
	return stdlog.New(&logWriter{
		// stdlogger.Print -> stdlogger.Output -> Write -> logger
		logger: logger.AddCallerSkip(3),
		level:  level,
	}, "", 0)
}

// logWriter is an io.Writer that doesn't really implement io.Writer correctly, but
// implements it correctly enough to satisfy the needs of stdlog.Logger's usage of
// io.Writer. Notably, stdlog.Logger:
//
// - does not use the bytes written return value
// - guarantees that each call to Write is a separate message
type logWriter struct {
	logger log.Logger
	level  log.Level
}

func (w *logWriter) Write(p []byte) (int, error) {
	msg := string(bytes.TrimSuffix(p, []byte("\n")))
	switch w.level {
	case log.LevelDebug:
		w.logger.Debug(msg)
	case log.LevelInfo:
		w.logger.Info(msg)
	case log.LevelWarn:
		w.logger.Warn(msg)
	case log.LevelError:
		w.logger.Error(msg)
	}
	return len(msg), nil
}
