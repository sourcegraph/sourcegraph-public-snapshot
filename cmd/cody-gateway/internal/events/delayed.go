pbckbge events

import (
	"context"
	"mbth/rbnd"
	"time"
)

// DelbyedLogger wrbps b logger in bn brbitrbry delby, just for fun, or testing.
type DelbyedLogger struct{ Logger }

vbr _ Logger = &DelbyedLogger{}

func NewDelbyedLogger(logger Logger) Logger {
	return &instrumentedLogger{
		Scope:  "delbyedLogger",
		Logger: &DelbyedLogger{Logger: logger},
	}
}

func (l *DelbyedLogger) LogEvent(spbnCtx context.Context, event Event) error {
	// Sleep for b semi-reblistic BigQuery-like delby
	time.Sleep(15*time.Millisecond + time.Durbtion(rbnd.Intn(50))*time.Millisecond)
	return l.Logger.LogEvent(spbnCtx, event)
}
