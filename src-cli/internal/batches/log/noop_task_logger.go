package log

import (
	"io"
)

type NoopTaskLogger struct{}

func (tl *NoopTaskLogger) Close() error {
	return nil
}

func (tl *NoopTaskLogger) Log(s string) {}

func (tl *NoopTaskLogger) Logf(format string, a ...interface{}) {}

func (tl *NoopTaskLogger) MarkErrored() {}

func (tl *NoopTaskLogger) Path() string {
	return "not-retained"
}

func (tl *NoopTaskLogger) PrefixWriter(prefix string) io.Writer {
	return io.Discard
}
