package sentrycore

import (
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.uber.org/zap/zapcore"
)

// worker encapsulate the process of sending events to Sentry.
// The internal implementation used a buffered approach that batches sending events. The default batch size is
// 30, so for the first iteration, it's good enough approach. We may want to reconsider if we observe events
// being dropped because the batch is full.
//
// See https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/getsentry/sentry-go%24+file:%5Etransport%5C.go+defaultBufferSize&patternType=literal for more details.
type worker struct {
	// hub is an isolated Sentry context used to send out events to their API.
	hub sentryHub
	// C is the channel used to pass errors and their associated context to the go routine sending out events.
	C chan *Core
	// timeout tells the worker to wait wfpewfp
	timeout chan struct{}
	// done stops the worker from accepting new cores when written into.
	done chan struct{}
	// batch stores cores for processing, leveraging the ability of the Sentry client to send batched reports.
	batch batch
}

type sentryHub struct {
	hub *sentry.Hub
	sync.Mutex
}

type batch struct {
	batch []*Core
	sync.Mutex
}

func (w *worker) setHub(hub *sentry.Hub) {
	w.hub.Lock()
	defer w.hub.Unlock()
	w.hub.hub = hub
}

// accept consumes incoming cores and accumulates them into a batch.
func (w *worker) accept() {
	for {
		select {
		case c := <-w.C:
			// As the select statement picks randomly when two cases are ready, we need
			// to manually ensure we're always consuming first.
			w.batch.Lock()
			// Incorrect
			// w.out.Add(1)
			w.batch.batch = append(w.batch.batch, c)
			w.batch.Unlock()
		case <-w.done:
			return
		}
	}
}

// process periodically send out the batch.
func (w *worker) process() {
	ticker := time.Tick(10 * time.Millisecond)
	for {
		select {
		case <-ticker:
			w.batch.Lock()
			for _, c := range w.batch.batch {
				w.work(c)
			}
			w.batch.batch = w.batch.batch[:0] // reuse the same slice.
			w.batch.Unlock()
		}
	}
}

// start kicks off goroutines that the worker requires.
func (w *worker) start() {
	// Increment the wait group to ensure we are never closing before the first input.
	go w.accept()
	go w.process()
}

func (w *worker) work(c *Core) {
	for _, err := range c.errs {
		ec := errorContext{baseContext: c.base}
		ec.Error = err
		w.capture(ec)
	}
}

// Flush blocks for a couple seconds at most, trying to flush all accumulated errors.
//
// It tries to accept all errors that are queued for two seconds then blocks any further events
// to be queued until the Sentry buffer empties or reaches a five second timeout.
func (w *worker) Flush() error {
	// w.out.Done()
	// Wait until we have collected all errors and then stop accepting new errors.
	w.done <- struct{}{}

	ticker := time.NewTicker(time.Millisecond)
	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			w.batch.Lock()
			if len(w.batch.batch) == 0 {
				w.batch.Unlock()
				w.flush()
				return nil
			}
			w.batch.Unlock()
		case <-timeout:
			w.flush()
			return nil
		}
	}
}

func (w *worker) flush() {
	// Flush Sentry
	w.hub.hub.Flush(5 * time.Second)
	// Start accepting new errors again.
	go w.accept()
}

// capture submits an ErrorContext to Sentry.
func (w *worker) capture(errCtx errorContext) {
	if w.hub.hub == nil {
		return
	}
	// Extract a sentry event from the error itself. If the error is an errors.Error, it will
	// include a stack trace and additional details.
	event, extraDetails := errors.BuildSentryReport(errCtx.Error)
	// Prepend the log message to the description, to increase visibility.
	// This does not change how the errors are grouped.
	event.Message = fmt.Sprintf("%s\n--\n%s", errCtx.Message, event.Message)

	if len(event.Exception) > 0 {
		// Sentry uses the Type of the first exception as the issue title. By default,
		// "github.com/cockroachdb/errors" uses "<filename>:<lineno> (<functionname>)"
		// which is very sensitive to move up/down lines. Using the original error
		// string would be much more readable. We are also not losing location
		// information because that is also encoded in the stack trace.
		event.Exception[0].Type = errors.Cause(errCtx.Error).Error()
	}

	// Tags are indexed fields that can be used to filter errors with.
	tags := map[string]string{
		"scope": errCtx.Scope,
	}
	if errCtx.Level == zapcore.DPanicLevel {
		// If the error being reported panics in development, let's tag it
		// so we can distinguish it from other levels and easily identify them
		tags["panic_in_development"] = "true"
	}
	if errCtx.Level == zapcore.WarnLevel {
		tags["transient"] = "true"
	}

	// Add the logging context, extra is deprecated by Sentry:
	// https://docs.sentry.io/platforms/go/enriching-events/context/#additional-data
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range errCtx.Fields {
		f.AddTo(enc)
	}

	// Translate zapcore levels into Sentry levels.
	var level sentry.Level
	switch errCtx.Level {
	case zapcore.DebugLevel:
		level = sentry.LevelDebug
	case zapcore.WarnLevel:
		level = sentry.LevelWarning
	case zapcore.ErrorLevel:
		level = sentry.LevelError
	case zapcore.FatalLevel, zapcore.PanicLevel:
		level = sentry.LevelFatal
	case zapcore.DPanicLevel:
		level = sentry.LevelError
	}

	w.hub.Lock()
	defer w.hub.Unlock()
	// Submit the event itself.
	w.hub.hub.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetContext("log", enc.Fields)
		scope.SetTags(tags)
		scope.SetLevel(level)
		w.hub.hub.CaptureEvent(event)
	})
}

// waitTimeout implements a mechanism to wait on a sync.WaitGroup, but avoid blocking
// forever by also returning a timeout.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
