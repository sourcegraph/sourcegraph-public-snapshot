package sinkcores

// TODO doc, dpanic tags

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"go.uber.org/zap/zapcore"
)

type baseContext struct {
	Key     string
	Scope   string
	Level   zapcore.Level
	Message string

	Fields []zapcore.Field
}

func (b *baseContext) clone() *baseContext {
	c := *b
	c.Key = b.Key
	c.Scope = b.Scope
	c.Level = b.Level
	c.Message = b.Message
	copy(c.Fields, b.Fields)
	return &c
}

type ErrorContext struct {
	baseContext
	Error error
}

type SentryCore struct {
	hub  *sentry.Hub
	base baseContext
	errs []error

	ErrorsC chan ErrorContext
	done    chan struct{}
}

var _ zapcore.Core = &SentryCore{}

func (s *SentryCore) clone() *SentryCore {
	c := SentryCore{
		hub:     s.hub,
		base:    *s.base.clone(),
		ErrorsC: s.ErrorsC,
		done:    s.done,
	}
	copy(c.errs, s.errs)

	return &c
}

// With only accumulate errors passed to the logger without sending them, as we do not
// have the full context required to annotate the error being captured by Sentry.
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

func (s *SentryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	s.base.Scope = entry.LoggerName
	s.base.Message = entry.Message

	n := 0
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				s.errs = append(s.errs, enc.Source)
				continue
			}
		}
		fields[n] = f
		n++
	}
	fields = fields[:n]
	fields = append(fields, s.base.Fields...)

	// queue the error reporting
	go func() {
		for _, err := range s.errs {
			ec := ErrorContext{baseContext: s.base}
			ec.Error = err
			s.ErrorsC <- ec
		}
	}()
	return nil
}

func (s *SentryCore) Start() {
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
	if s.hub == nil {
		return
	}
	event, extraDetails := errors.BuildSentryReport(errC.Error)
	// Prepend the log message to the description, to increase visibility.
	// This does not change how the errors are grouped.
	event.Message = fmt.Sprintf("%s\n--\n%s", errC.Message, event.Message)

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

	extraDetails["log.msg"] = errC.Message

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
	return level >= zapcore.WarnLevel
}

// Sync ensure that the remaining event are flushed, but has a hard limit of TODO seconds
// after which it will stop blocking to avoid interruping application shutdown.
func (s *SentryCore) Sync() error {
	// TODO something flush flush
	close(s.done)
	return nil
}
