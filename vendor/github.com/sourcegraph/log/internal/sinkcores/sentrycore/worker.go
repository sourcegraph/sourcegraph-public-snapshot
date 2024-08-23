package sentrycore

import (
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log/internal/encoders"
	"github.com/sourcegraph/log/internal/otelfields"
	"go.uber.org/zap/zapcore"
)

const (
	// flushDelay defines the brief window in which log messages are still accepted
	// before flushing.
	flushDelay = 500 * time.Millisecond
	// sentryTimeout defines how much time Sentry has to send the events.
	sentryTimeout = 5 * time.Second
)

// worker encapsulate the process of sending events to Sentry by asynchronously
// consuming Cores, turning them into capturable, annotated errors.
//
// See https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/getsentry/sentry-go%24+file:%5Etransport%5C.go+defaultBufferSize&patternType=literal for more details.
type worker struct {
	// hub is an isolated Sentry context used to send out events to their API.
	hub sentryHub
	// C is the channel used to pass errors and their associated context to the go routine sending out events.
	C chan *errorContext
	// timeout tells the worker to wait wfpewfp
	timeout chan struct{}
	// done stops the worker from accepting new cores when written into.
	done chan struct{}
}

type sentryHub struct {
	hub *sentry.Hub
	sync.Mutex
}

func (w *worker) setHub(hub *sentry.Hub) {
	w.hub.Lock()
	defer w.hub.Unlock()
	w.hub.hub = hub
}

// start kicks off the consuming go routine.
func (w *worker) start() {
	go w.consume()
}

// consume consumes incoming cores and turn them into reportable errors.
func (w *worker) consume() {
	ticker := time.NewTicker(flushDelay)
	defer ticker.Stop()
	for {
		select {
		case errC := <-w.C:
			w.capture(errC)
		case <-ticker.C:
			// We only check if we're closing periodically, to make sure we have
			// consumed the last few events that were sent.
			select {
			case <-w.done:
				return
			default:
			}
		}
	}
}

// Flush blocks for a couple seconds at most, trying to flush all accumulated errors.
//
// It will keep consuming events based for a duration defined by flushDelay and
// then tells Sentry to flush for a max duration of sentryTimeout.
func (w *worker) Flush() error {
	w.done <- struct{}{}
	w.flush()
	return nil
}

func (w *worker) flush() {
	// Flush Sentry
	w.hub.hub.Flush(sentryTimeout)
	// Start accepting new errors again.
	go w.consume()
}

func (w *worker) stop() {
	w.done <- struct{}{}
}

// capture submits an ErrorContext to Sentry.
func (w *worker) capture(errCtx *errorContext) {
	if w.hub.hub == nil {
		return
	}
	// Extract a sentry event from the error itself. If the error is an errors.Error, it will
	// include a stack trace and additional details.
	event, extraDetails := errors.BuildSentryReport(errCtx.Error)
	// Prepend the log message to the description, to increase visibility.
	// This does not change how the errors are grouped.
	event.Message = fmt.Sprintf("%s: %s\n--\n%s", errCtx.Scope, errCtx.Message, event.Message)

	if len(event.Exception) > 0 {
		// Sentry uses the Type of the first exception as the issue title. By default,
		// "github.com/cockroachdb/errors" uses "<filename>:<lineno> (<functionname>)"
		// which is very sensitive to move up/down lines. Using the original error
		// string would be much more readable. We are also not losing location
		// information because that is also encoded in the stack trace.
		//
		// Additionally we include prefix the error with the log message, because usually
		// when we log an error it is analagous to "wrapping" the error with a fixed
		// description. And of course we include the scope, because we want to promote
		// its use as ways to identify and understand the context of observability
		// elements.
		//
		// HOWEVER!!!! It turns out when Sentry says "first exception" it ACTUALLY means
		// the LAST exception:
		//
		// 	// Note that "first exception" is the last item in the slice,
		// 	// because... Sentry is annoying.
		//
		// Source: https://github.com/cockroachdb/errors/blob/26622367a22260fa287d2f7aa2a085b0324c74ee/report/report.go#L324-L325
		// We've observed this behaviour in practice as well. So here, make sure we are
		// overwriting Type on the LAST exception instead of the first.
		event.Exception[len(event.Exception)-1].Type = fmt.Sprintf("[%s] %s: %s",
			errCtx.Scope, errCtx.Message, errors.Cause(errCtx.Error).Error())
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

	// Extract the service name, if present in the fields.
	for _, f := range errCtx.Fields {
		if f.Key == otelfields.ResourceFieldKey {
			if r, ok := f.Interface.(*encoders.ResourceEncoder); ok {
				tags["resource.service.name"] = r.Name
				if r.Version != "" {
					tags["resource.service.version"] = r.Version
					event.Release = r.Version
				}
				if r.Namespace != "" {
					tags["resource.service.namespace"] = r.Namespace
					event.Environment = r.Namespace
				}
				if r.InstanceID != "" {
					tags["resource.service.instance.id"] = r.InstanceID
					event.ServerName = r.InstanceID
				}
			}
		}
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
		if enc.Fields != nil {
			// Also reflect trace into Sentry's specialized trace context
			// https://develop.sentry.dev/sdk/event-payloads/contexts/#trace-context
			scope.SetContext("trace", sentry.Context{
				"trace_id": enc.Fields["TraceId"],
				"span_id":  enc.Fields["SpanId"],
			})
		}
		scope.SetTags(tags)
		scope.SetLevel(level)
		w.hub.hub.CaptureEvent(event)
	})
}
