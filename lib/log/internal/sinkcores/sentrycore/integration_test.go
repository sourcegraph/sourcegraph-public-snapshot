package sentrycore_test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores/sentrycore"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
	"github.com/sourcegraph/sourcegraph/lib/log/sinks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLevelFiltering(t *testing.T) {
	e := errors.New("test error")
	tt := []struct {
		level      zapcore.Level
		wantReport bool
	}{
		{level: zapcore.DebugLevel, wantReport: false},
		{level: zapcore.InfoLevel, wantReport: false},
		{level: zapcore.WarnLevel, wantReport: true},
		{level: zapcore.ErrorLevel, wantReport: true},
		// Those are annoying to test, it would required to fire up a subprocess, so
		// instead, we just check the result of the Enabled() method in another subtest.
		// {level: zapcore.FatalLevel, wantReport: true},
		// {level: zapcore.DPanicLevel, wantReport: true},
	}
	for _, test := range tt {
		var desc string
		if test.wantReport {
			desc = "has report"
		} else {
			desc = "no report"
		}
		t.Run(fmt.Sprintf("%s, %s", test.level.CapitalString(), desc), func(t *testing.T) {
			logger, tr, sync := newTestLogger(t)
			logWithLevel(logger, test.level, "msg", log.Error(e))
			var count int
			if test.wantReport {
				count = 1
			}
			sync()
			assert.Len(t, tr.Events(), count)
		})
		t.Run(fmt.Sprintf("%s, %s (with)", test.level.CapitalString(), desc), func(t *testing.T) {
			logger, tr, sync := newTestLogger(t)
			logWithLevel(logger.With(log.Error(e)), test.level, "msg")
			var count int
			if test.wantReport {
				count = 1
			}
			sync()
			assert.Len(t, tr.Events(), count)
		})
	}

	t.Run("FATAL has report", func(t *testing.T) {
		hub, _ := newTestHub(t)
		core := sentrycore.NewCore(hub)
		got := core.Enabled(zapcore.FatalLevel)
		assert.True(t, got)
	})

	t.Run("DPANIC has report", func(t *testing.T) {
		hub, _ := newTestHub(t)
		core := sentrycore.NewCore(hub)
		got := core.Enabled(zapcore.FatalLevel)
		assert.True(t, got)
	})
}

func TestFieldsFiltering(t *testing.T) {
	tt := []struct {
		level      zapcore.Level
		wantReport bool
	}{
		{level: zapcore.DebugLevel, wantReport: false},
		{level: zapcore.InfoLevel, wantReport: false},
		{level: zapcore.WarnLevel, wantReport: false},
		{level: zapcore.ErrorLevel, wantReport: false},
	}
	for _, test := range tt {
		var desc string
		if test.wantReport {
			desc = "has report"
		} else {
			desc = "no report"
		}
		t.Run(fmt.Sprintf("%s, %s", test.level.CapitalString(), desc), func(t *testing.T) {
			logger, tr, sync := newTestLogger(t)
			logWithLevel(logger.With(log.String("foo", "bar")), test.level, "msg")
			var count int
			if test.wantReport {
				count = 1
			}
			sync()
			assert.Len(t, tr.Events(), count)
		})
	}
}

func TestConcurrentLogging(t *testing.T) {
	e := errors.New("test error")
	t.Run("2 goroutines, 50 msg each", func(t *testing.T) {
		logger, tr, _sync := newTestLogger(t)
		var wg sync.WaitGroup
		wg.Add(2)
		f := func() {
			for i := 0; i < 50; i++ {
				logger.With(log.Error(e)).Warn("msg")
			}
			wg.Done()
		}
		go f()
		go f()
		wg.Wait()
		_sync()
		assert.Len(t, tr.Events(), 100)
	})
}

func TestFlush(t *testing.T) {
	go withTimeout(t, 10*time.Second)
	e := errors.New("test error")
	hub, tr := newTestHub(t)
	core := sentrycore.NewCore(hub)
	core.Start()
	go func() {
		for {
			err := core.Write(zapcore.Entry{Level: zapcore.InfoLevel, Message: "msg"}, []zapcore.Field{log.Error(e)})
			assert.NoError(t, err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	core.Sync()
	assert.Greater(t, len(tr.Events()), 1)
}

func withTimeout(t *testing.T, timeout time.Duration) {
	t.Helper()

	testFinished := make(chan struct{})
	t.Cleanup(func() { close(testFinished) })

	select {
	case <-testFinished:
	case <-time.After(timeout):
		t.Errorf("test timed out after %s", timeout)
		os.Exit(1)
	}
}

func logWithLevel(logger log.Logger, level zapcore.Level, msg string, fields ...zapcore.Field) {
	switch level {
	case zapcore.DebugLevel:
		logger.Debug(msg, fields...)
	case zapcore.InfoLevel:
		logger.Info(msg, fields...)
	case zapcore.WarnLevel:
		logger.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		logger.Error(msg, fields...)
	case zapcore.FatalLevel:
		logger.Fatal(msg, fields...)
	case zapcore.DPanicLevel, zapcore.PanicLevel:
		panic("not implemented")
	}
}

func newTestLogger(t testing.TB) (log.Logger, *sentrycore.TransportMock, func()) {
	hub, tr := newTestHub(t)
	sink := sinks.NewSentrySink(hub)
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
