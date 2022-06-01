package benchmarks_test

import (
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func BenchmarkWithSentry(b *testing.B) {
	logger, _, _ := newTestLogger(b)

	err := errors.New("foobar")
	for n := 0; n < b.N; n++ {
		logger.With(log.Error(err)).Warn("msg", log.Int("key", 5))
	}
}

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
