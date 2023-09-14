package events

import (
	"context"
	"math/rand"
	"time"
)

// DelayedLogger wraps a logger in an arbitrary delay, just for fun, or testing.
type DelayedLogger struct{ Logger }

var _ Logger = &DelayedLogger{}

func NewDelayedLogger(logger Logger) Logger {
	return &instrumentedLogger{
		Scope:  "delayedLogger",
		Logger: &DelayedLogger{Logger: logger},
	}
}

func (l *DelayedLogger) LogEvent(spanCtx context.Context, event Event) error {
	// Sleep for a semi-realistic BigQuery-like delay
	time.Sleep(15*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond)
	return l.Logger.LogEvent(spanCtx, event)
}
