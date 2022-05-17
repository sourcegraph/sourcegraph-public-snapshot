package sinkcores

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"go.uber.org/zap/zapcore"
)

type ContextError struct {
	Key   string
	Scope string
	Error error
}

type SentryCore struct {
	WithErrors  []ContextError
	withErrsMux sync.Mutex

	ErrorsC chan ContextError
}

var _ zapcore.Core = &SentryCore{}

func (s *SentryCore) With(fields []zapcore.Field) zapcore.Core {
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				s.withErrsMux.Lock()
				s.WithErrors = append(s.WithErrors, ContextError{
					Key:   f.Key,
					Error: enc.Source,
				})
				s.withErrsMux.Unlock()
			}
		}
	}
	return s
}

func (s *SentryCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(e, s)
}

func (s *SentryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			if enc, ok := f.Interface.(*encoders.ErrorEncoder); ok {
				// queue in background so that we don't block
				println(enc.Error())

				go func() {
					s.ErrorsC <- ContextError{
						Key:   f.Key,
						Scope: entry.LoggerName,
						Error: enc.Source,
					}
				}()
			}
		}
	}
	return nil
}

func (s *SentryCore) Enabled(zapcore.Level) bool { return true }

func (s *SentryCore) Sync() error {
	// TODO something flush flush
	return nil
}
