package benchmarks_test

import (
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

// BenchmarkWithSentry-10           2253642              5205 ns/op            9841 B/op         87 allocs/op
func BenchmarkWithSentry(b *testing.B) {
	logger, _, _ := newTestLogger(b)

	err := errors.New("foobar")
	for n := 0; n < b.N; n++ {
		logger.With(log.Error(err)).Warn("msg", log.Int("key", 5))
	}
}

// BenchmarkWithoutSentry-10        2656189              4537 ns/op            6334 B/op         44 allocs/op
func BenchmarkWithoutSentry(b *testing.B) {
	logger, _ := logtest.Captured(b)
	err := errors.New("foobar")
	for n := 0; n < b.N; n++ {
		logger.With(log.Error(err), log.Int("key", 5)).Warn("msg")
	}
}

func newTestLogger(t testing.TB) (log.Logger, *sentrycore.TransportMock, func()) {
	transport := &sentrycore.TransportMock{}
	sink := log.NewSentrySinkWithOptions(sentry.ClientOptions{Transport: transport})
	logger, exportLogs := logtest.Captured(t, sink)
	return logger, transport, func() { _ = exportLogs() }
}
