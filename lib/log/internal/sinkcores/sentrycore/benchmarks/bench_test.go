package benchmarks_test

import (
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
	"github.com/sourcegraph/sourcegraph/lib/log/sinks"
	"github.com/stretchr/testify/assert"
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
	hub, tr := newTestHub(t)
	sink := sinks.NewSentrySinkCore(hub)
	logger, exportLogs := logtest.Captured(t, sink)
	return logger, tr, func() { _ = exportLogs() }
}

func newTestHub(t testing.TB) (*sentry.Hub, *sentrycore.TransportMock) {
	transport := &sentrycore.TransportMock{}
	c, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
	assert.NoError(t, err)
	hub := sentry.NewHub(c, sentry.NewScope())
	return hub, transport
}
