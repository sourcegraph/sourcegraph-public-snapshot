package slog

import (
	"context"
	"log"
	"strings"
)

// Stdlib creates a standard library logger from the given logger.
//
// All logs will be logged at the level set by the logger and the
// given ctx will be passed to the logger's Log method, thereby
// logging all fields and tracing info in the context.
//
// You can redirect the stdlib default logger with log.SetOutput
// to the Writer on the logger returned by this function.
// See the example.
func Stdlib(ctx context.Context, l Logger, level Level) *log.Logger {
	l.skip += 2

	l = l.Named("stdlib")

	w := &stdlogWriter{
		ctx:   ctx,
		l:     l,
		level: level,
	}

	return log.New(w, "", 0)
}

type stdlogWriter struct {
	ctx   context.Context
	l     Logger
	level Level
}

func (w stdlogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	// stdlib includes a trailing newline on the msg that
	// we do not want.
	msg = strings.TrimSuffix(msg, "\n")

	w.l.log(w.ctx, w.level, msg, Map{})

	return len(p), nil
}
