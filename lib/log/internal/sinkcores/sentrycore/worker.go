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
	hub   *sentry.Hub
	muHub sync.Mutex
	// C is the channel used to pass errors and their associated context to the go routine sending out events.
	C chan *Core
	// done stops the worker from accepting new cores when written into.
	done chan struct{}
	// batch stores cores for processing, leveraging the ability of the Sentry client to send batched reports.
	batch   []*Core
	muBatch sync.Mutex
	// out tracks the outgoing cores count, i.e those which have to be sent out.
	out sync.WaitGroup
}

func (w *worker) setHub(hub *sentry.Hub) {
	w.muHub.Lock()
	defer w.muHub.Unlock()
	w.hub = hub
}

// accept consumes incoming cores and accumulates them into a batch.
func (w *worker) accept() {
	for {
		select {
		case c := <-w.C:
			w.muBatch.Lock()
			w.batch = append(w.batch, c)
			w.out.Add(1)
			w.muBatch.Unlock()
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
			w.muBatch.Lock()
			for _, c := range w.batch {
				w.work(c)
				w.out.Done()
			}
			w.batch = w.batch[:0] // reuse the same slice.
			w.muBatch.Unlock()
		}
	}
}

func (w *worker) start() {
	go w.accept()
	go w.process()
}

func (w *worker) work(c *Core) {
	for _, err := range c.errs {
		ec := ErrorContext{baseContext: c.base}
		ec.Error = err
		w.capture(ec)
	}
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

// Flush blocks for a couple seconds at most, trying to flush all accumulated errors.
//
// It tries to accept all errors that are queued for two seconds then blocks any further events
// to be queued until the Sentry buffer empties or reaches a five second timeout.
func (w *worker) Flush() error {
	// Wait until we have collected all errors and then stop accepting new errors.
	// waitTimeout(&w.in, 2*time.Second)
	w.done <- struct{}{}
	// Wait until we have processed all errors
	w.out.Wait()
	// Make sure Sentry has flushed everything.
	w.hub.Flush(5 * time.Second)
	// Start accepting new errors again.
	go w.accept()
	return nil
}

// capture submits an ErrorContext to Sentry.
func (w *worker) capture(errCtx ErrorContext) {
	if w.hub == nil {
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

	// Extra are fields that are added to the error as annotation, still registering
	// as the same error when counted.
	for _, f := range errCtx.Fields {
		switch f.Type {
		case zapcore.StringType:
			extraDetails[fmt.Sprintf("log.%s", f.Key)] = f.String
		case zapcore.Int8Type, zapcore.Int16Type, zapcore.Int32Type, zapcore.Int64Type:
			extraDetails[fmt.Sprintf("log.%s", f.Key)] = f.Integer
		default:
			// Because the log package only exposes base types or sliced versions, using %v is a
			// good enough way to print the values for extra attributes display in Sentry.
			extraDetails[fmt.Sprintf("log.%s", f.Key)] = fmt.Sprintf("%v", f.Interface)
		}
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

	w.muHub.Lock()
	defer w.muHub.Unlock()
	// Submit the event itself.
	w.hub.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		scope.SetLevel(level)
		w.hub.CaptureEvent(event)
	})
}
