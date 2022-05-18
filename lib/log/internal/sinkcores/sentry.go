package sinkcores

// TODO doc, dpanic tags

import (
	"fmt"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"go.uber.org/zap/zapcore"
)

type ErrorContext struct {
	Key    string
	Scope  string
	Error  error
	Level  zapcore.Level
	Fields []zapcore.Field
}

type ErrorsContext struct {
	ErrorContext
	errs []error
}

type SentryCore struct {
	hub *sentry.Hub

	ec ErrorsContext

	ErrorsC chan ErrorContext
	started sync.Once
	done    chan struct{}
}

var _ zapcore.Core = &SentryCore{}

func (s *SentryCore) clone() *SentryCore {
	return &SentryCore{
		hub:     s.hub,
		ec:      s.ec,
		ErrorsC: s.ErrorsC,
		started: s.started,
		done:    s.done,
	}
}

// With only accumulate errors passed to the logger without sending them, as we do not
// have the full context required to annotate the error being captured by Sentry.
func (s *SentryCore) With(fields []zapcore.Field) zapcore.Core {
	s = s.clone()
	s.ec.Fields = append(s.ec.Fields, fields...)
	return s
}

// Check will inspect e to see if it needs to be sent to Sentry.
// TODO
func (s *SentryCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(e, s)
}

func (s *SentryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	s.started.Do(s.process)
	s.ec.Scope = entry.LoggerName
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				// queue in background so that we don't block
				println(enc.Error())
				println("fields len", len(fields))

				go func() {
					for _, err := range s.ec.errs {
						ec := s.ec.ErrorContext
						ec.Error = err
						s.ErrorsC <- ec
					}
					ec := s.ec.ErrorContext
					ec.Error = enc.Source
					s.ErrorsC <- ec
				}()
			}
		}
	}
	return nil
}

func (s *SentryCore) process() {
	println("process")
	s.ErrorsC = make(chan ErrorContext)
	go func() {
		for {
			select {
			case <-s.done:
				return
			case err := <-s.ErrorsC:
				s.capture(err)
			}
		}
	}()
}

func (s *SentryCore) capture(errC ErrorContext) {
	event, extraDetails := errors.BuildSentryReport(errC.Error)

	// Sentry uses the Type of the first exception as the issue title. By default,
	// "github.com/cockroachdb/errors" uses "<filename>:<lineno> (<functionname>)"
	// which is very sensitive to move up/down lines. Using the original error
	// string would be much more readable. We are also not losing location
	// information because that is also encoded in the stack trace.
	if len(event.Exception) > 0 {
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
	println("len fields", len(errC.Fields))
	for _, f := range errC.Fields {
		println("fields key", f.Key)
		switch f.Type {
		case zapcore.StringType:
			extraDetails[fmt.Sprintf("log.%s", f.Key)] = f.String
		case zapcore.Int8Type, zapcore.Int16Type, zapcore.Int32Type, zapcore.Int64Type:
			extraDetails[fmt.Sprintf("log.%s", f.Key)] = f.Integer
		default:
			extraDetails[fmt.Sprintf("log.%s", f.Key)] = fmt.Sprintf("%v", f.Interface)
		}
	}

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

	s.hub.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		scope.SetLevel(level)
		s.hub.CaptureEvent(event)
	})
}

// Enabled returns false when the log level is below the Warn level.
func (s *SentryCore) Enabled(level zapcore.Level) bool {
	// return level >= zapcore.WarnLevel
	return true
}

// Sync ensure that the remaining event are flushed, but has a hard limit of TODO seconds
// after which it will stop blocking to avoid interruping application shutdown.
func (s *SentryCore) Sync() error {
	// TODO something flush flush
	return nil
}
