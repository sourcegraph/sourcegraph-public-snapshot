package events

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type bufferedEvent struct {
	spanCtx context.Context
	Event
}

type BufferedLogger struct {
	log log.Logger

	// handler is the underlying event logger to which events are submitted.
	handler Logger

	// bufferC is a buffered channel of events to be logged.
	bufferC chan bufferedEvent
	// timeout is the max duration to wait to submit an event.
	timeout time.Duration
	// bufferClosed indicates if the buffer has been closed.
	bufferClosed *atomic.Bool
	// flushedC is a channel that is closed when the buffer is emptied.
	flushedC chan struct{}
}

var _ Logger = &BufferedLogger{}
var _ goroutine.BackgroundRoutine = &BufferedLogger{}

// NewBufferedLogger wraps handler with a buffered logger that submits events
// in the background instead of in the hot-path of a request. It implements
// goroutine.BackgroundRoutine that must be started.
func NewBufferedLogger(logger log.Logger, handler Logger, queueSize int) *BufferedLogger {
	return &BufferedLogger{
		log: logger.Scoped("bufferedLogger", "buffered events logger"),

		handler: handler,

		bufferC:      make(chan bufferedEvent, queueSize),
		timeout:      100 * time.Millisecond,
		bufferClosed: &atomic.Bool{},
		flushedC:     make(chan struct{}),
	}
}

// LogEvent implements event.Logger by submitting the event to a buffer for processing.
func (l *BufferedLogger) LogEvent(spanCtx context.Context, event Event) error {
	// If buffer is closed, make a best-effort attempt to log the event directly.
	if l.bufferClosed.Load() {
		trace.Logger(spanCtx, l.log).Warn("buffer is closed: logging event directly")
		return l.handler.LogEvent(spanCtx, event)
	}

	select {
	case l.bufferC <- bufferedEvent{spanCtx: spanCtx, Event: event}:
		return nil

	case <-time.After(l.timeout):
		return errors.Newf("failed to insert event in %s: buffer full: %d items pending",
			l.timeout.String(), len(l.bufferC))
	}
}

// Start starts buffered logger's background processing job.
func (l *BufferedLogger) Start() {
	go func() {
		for event := range l.bufferC {
			if err := l.handler.LogEvent(event.spanCtx, event.Event); err != nil {
				l.log.Error("failed to log buffered event", log.Error(err))
			}
		}

		l.log.Info("events buffer emptied")
		close(l.flushedC)
	}()
}

// Stop stops buffered logger's background processing job and flushes its buffer.
func (l *BufferedLogger) Stop() {
	l.bufferClosed.Store(true)
	close(l.bufferC)
	l.log.Info("buffer closed - waiting for events to flush")
	<-l.flushedC
	l.log.Info("shutdown complete")
}
