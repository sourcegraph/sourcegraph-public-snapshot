package events

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type nonBlockingLogger struct {
	l       log.Logger
	handler Logger
}

var _ Logger = &nonBlockingLogger{}

func NewNonBlockingLogger(l log.Logger, handler Logger) Logger {
	return &nonBlockingLogger{l: l.Scoped("events.nonblocking", "background events submission"), handler: handler}
}

func (b *nonBlockingLogger) LogEvent(spanCtx context.Context, event Event) error {
	go func() {
		ctx := backgroundContextWithSpan(spanCtx)
		if err := b.handler.LogEvent(ctx, event); err != nil {
			trace.Logger(ctx, b.l).Error("failed to submit event", log.Error(err))
		}
	}()
	return nil
}
