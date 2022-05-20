// sinkcores provides a few additional cores that adds side effects to logging.
//
// See SentryCore for more information about turning log messages whose level is equal or above zapcore.WarnLevel
// into Sentry reports if there is an error attached with log.Error(err).
package sinkcores

import (
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"go.uber.org/zap/zapcore"
)

// baseContext contains the data surrounding an error, that is shared by all errors attached to the current core.
type baseContext struct {
	Key     string
	Scope   string
	Level   zapcore.Level
	Message string

	Fields []zapcore.Field
}

// clone safely copies a baseContext
func (b *baseContext) clone() *baseContext {
	c := *b
	c.Key = b.Key
	c.Scope = b.Scope
	c.Level = b.Level
	c.Message = b.Message
	copy(c.Fields, b.Fields)
	return &c
}

// ErrorContext is an error and its associated context that is accumulated during the core lifetime.
type ErrorContext struct {
	baseContext
	Error error
}

// SentryCore turns any log message that comes with at least one error into one or more error reports. All
// error reports will share the same metadata, with the exception of those attached onto the errors themselves.
type SentryCore struct {
	base baseContext
	// errs accumulates the errors fed to the core as attributes.
	errs   []error
	worker *sentryWorker
}

var _ zapcore.Core = &SentryCore{}

// NewSentryCore returns a new SentryCore with a ready to use worker. It should be called only once, when attaching
// this core onto the global logger that is then used to create scoped loggers in other parts of the codebase.
func NewSentryCore(hub *sentry.Hub) *SentryCore {
	return &SentryCore{
		worker: &sentryWorker{
			hub:     hub.Clone(), // Avoid accidental side effects if the hub is modified elsewhere.
			ErrorsC: make(chan ErrorContext, 16),
			done:    make(chan struct{}),
		},
	}
}

// sentryWorker encapsulate the process of sending events to Sentry.
// The internal implementation used a buffered approach that batches sending events. The default batch size is
// 30, so for the first iteration, it's good enough approach. We may want to reconsider if we observe events
// being dropped because the batch is full.
//
// See https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/getsentry/sentry-go%24+file:%5Etransport%5C.go+defaultBufferSize&patternType=literal for more details.
type sentryWorker struct {
	// hub is an isolated Sentry context used to send out events to their API.
	hub *sentry.Hub
	// ErrorsC is the channel used to pass errors and their associated context to the go routine sending out events.
	ErrorsC chan ErrorContext
	// done is channel which when closed shuts down the worker immediately.
	done chan struct{}
	sync.Mutex
}

// clone returns a copy of the core, carrying all previously accumulated context, but that can be safely be
// modified without affecting other core instances.
func (s *SentryCore) clone() *SentryCore {
	c := SentryCore{
		worker: s.worker,
		base:   *s.base.clone(),
	}
	copy(c.errs, s.errs)

	return &c
}

// With stores fields passed to the core into a new core that will be then used to contruct the final error report.
//
// It does not capture errors, because we may get additional context in a subsequent With or Write call
// that will also need to be included.
func (s *SentryCore) With(fields []zapcore.Field) zapcore.Core {
	s = s.clone()
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				s.errs = append(s.errs, enc.Source)
				continue
			}
		}
		s.base.Fields = append(s.base.Fields, f)
	}
	return s
}

// Check will inspect e to see if it needs to be sent to Sentry.
// TODO
func (s *SentryCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(e, s)
}

// Write will asynchronoulsy send out all errors and the fields that have been accumulated during the
// lifetime of the core.
func (s *SentryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Clone ourselves, so we don't affect other loggers
	s = s.clone()
	go func() {
		fmt.Println("write")
		s.base.Scope = entry.LoggerName
		s.base.Message = entry.Message

		n := 0
		for _, f := range fields {
			if f.Type == zapcore.ErrorType {
				if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
					// If we find one of our errors, we remove it from the fields so our error reports are not including
					// their own error as an attribute, which would a useless repetition.
					s.errs = append(s.errs, enc.Source)
					continue
				}
			}
			fields[n] = f
			n++
		}
		fields = fields[:n] // account for the filtered out elements.
		s.base.Fields = append(fields, s.base.Fields...)

		// queue the error reporting
		for _, err := range s.errs {
			ec := ErrorContext{baseContext: s.base}
			ec.Error = err
			s.worker.ErrorsC <- ec
		}
	}()
	return nil
}

// Starts launches the go routine responsible for consuming ErrorContext that needs to be submitted to Sentry.
func (s *SentryCore) Start() {
	s.worker.start()
}

func (w *sentryWorker) start() {
	println("hereeee start")
	go func() {
		println("inside start")
		for {
			select {
			case err := <-w.ErrorsC:
				println("weeee")
				w.capture(err) // it takes between 250µs and 150µs on my machine.
			case <-w.done:
				println("done")
				w.hub.Flush(5 * time.Second)
				return
			}
		}
	}()
}

// capture submits an ErrorContext to Sentry.
func (w *sentryWorker) capture(errC ErrorContext) {
	fmt.Println("caputre")
	if w.hub == nil {
		return
	}
	// Extract a sentry event from the error itself. If the error is an errors.Error, it will
	// include a stack trace and additional details.
	event, extraDetails := errors.BuildSentryReport(errC.Error)
	// Prepend the log message to the description, to increase visibility.
	// This does not change how the errors are grouped.
	event.Message = fmt.Sprintf("%s\n--\n%s", errC.Message, event.Message)

	if len(event.Exception) > 0 {
		// Sentry uses the Type of the first exception as the issue title. By default,
		// "github.com/cockroachdb/errors" uses "<filename>:<lineno> (<functionname>)"
		// which is very sensitive to move up/down lines. Using the original error
		// string would be much more readable. We are also not losing location
		// information because that is also encoded in the stack trace.
		event.Exception[0].Type = errors.Cause(errC.Error).Error()
	}

	// Tags are indexed fields that can be used to filter errors with.
	tags := map[string]string{
		"scope": errC.Scope,
	}
	if errC.Level == zapcore.DPanicLevel {
		// If the error being reported panics in development, let's tag it
		// so we can distinguish it from other levels and easily identify them
		tags["panic_in_development"] = "true"
	}

	// Extra are fields that are added to the error as annotation, still registering
	// as the same error when counted.
	for _, f := range errC.Fields {
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
	switch errC.Level {
	case zapcore.WarnLevel:
		level = sentry.LevelWarning
	case zapcore.ErrorLevel:
		level = sentry.LevelError
	case zapcore.FatalLevel, zapcore.PanicLevel:
		level = sentry.LevelFatal
	case zapcore.DPanicLevel:
		level = sentry.LevelError
	}

	// Submit the event itself.
	w.hub.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		scope.SetLevel(level)
		w.hub.CaptureEvent(event)
	})
}

// Enabled returns false when the log level is below the Warn level.
func (s *SentryCore) Enabled(level zapcore.Level) bool {
	return level >= zapcore.WarnLevel
}

// Sync ensure that the remaining event are flushed, but has a hard limit of TODO seconds
// after which it will stop blocking to avoid interruping application shutdown.
func (s *SentryCore) Sync() error {
	close(s.worker.done)
	return nil
}
