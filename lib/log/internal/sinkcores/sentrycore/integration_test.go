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
		// Levels that exit are annoying to test, it would required to fire up a subprocess, so
		// instead, we just check the result of the Enabled() method in another subtest.
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

func TestTags(t *testing.T) {
	e := errors.New("test error")
	t.Run("scope", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger = logger.Scoped("my-scope", "testing scope tags")
		logger.Error("msg", log.Error(e))
		sync()
		assert.Len(t, tr.Events(), 1)
		assert.Equal(t, tr.Events()[0].Tags["scope"], "TestTags/scope.my-scope")
	})

	t.Run("transient", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger.Warn("msg", log.Error(e))
		sync()
		assert.Len(t, tr.Events(), 1)
		assert.Equal(t, tr.Events()[0].Tags["transient"], "true")
	})
}

func TestWith(t *testing.T) {
	a := errors.New("A")
	b := errors.New("B")
	c := errors.New("C")
	t.Run("multiple errors", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger.With(log.Error(a), log.Error(b)).Warn("msg", log.Error(c))
		sync()
		assert.Len(t, tr.Events(), 3)
	})
}

func TestFields(t *testing.T) {
	assertEventLogCtx := func(t *testing.T, tr *sentrycore.TransportMock, cb func(map[string]interface{})) {
		assert.Len(t, tr.Events(), 1)
		if len(tr.Events()) < 1 {
			t.FailNow()
		}
		e := tr.Events()[0]
		assert.IsType(t, map[string]interface{}{}, e.Contexts["log"])
		cb(tr.Events()[0].Contexts["log"].(map[string]interface{}))
	}
	e := errors.New("test error")

	t.Run("int", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger.With(log.Int("int", 4)).Error("msg", log.Error(e))
		sync()
		assertEventLogCtx(t, tr, func(ctx map[string]interface{}) {
			assert.Equal(t, int64(4), ctx["int"])
		})
	})
	t.Run("int64", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger.With(log.Int64("int", 4)).Error("msg", log.Error(e))
		sync()
		assertEventLogCtx(t, tr, func(ctx map[string]interface{}) {
			assert.Equal(t, int64(4), ctx["int"])
		})
	})
	t.Run("string", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger.With(log.String("string", "foo")).Error("msg", log.Error(e))
		sync()
		assertEventLogCtx(t, tr, func(ctx map[string]interface{}) {
			assert.Equal(t, "foo", ctx["string"])
		})
	})
	t.Run("object", func(t *testing.T) {
		logger, tr, sync := newTestLogger(t)
		logger.With(log.Object("object", log.String("string", "foo"), log.Int("int", 4))).Error("msg", log.Error(e))
		sync()
		assertEventLogCtx(t, tr, func(ctx map[string]interface{}) {
			assert.Equal(t, map[string]interface{}{"int": int64(4), "string": "foo"}, ctx["object"])
		})
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
		wg.Add(10)
		f := func() {
			for i := 0; i < 10; i++ {
				logger.With(log.Error(e)).Warn("msg")
			}
			wg.Done()
		}
		for i := 0; i < 10; i++ {
			go f()
		}
		wg.Wait()
		_sync()
		assert.Len(t, tr.Events(), 100)
	})
}

func TestNeverBlock(t *testing.T) {
	go withTimeout(t, 2*time.Second)

	e := errors.New("test error")
	hub, _ := newTestHub(t)
	c := sentrycore.NewCore(hub)
	c.Stop()

	for i := 0; i < 2048; i++ {
		c.Write(zapcore.Entry{Level: zapcore.ErrorLevel, Message: "should not block"}, []zapcore.Field{log.Error(e)})
	}
}

// TestFlush ensures that even with a huge backlog of events, the Flush functions returns.
func TestFlush(t *testing.T) {
	go withTimeout(t, 10*time.Second)
	e := errors.New("test error")
	hub, tr := newTestHub(t)
	core := sentrycore.NewCore(hub)
	go func() {
		for {
			// Without this sleep, we're hitting the max goroutine count that the race detector can handle, which
			// causes it to just abort because it can't run with that many go routines.
			// https://github.com/golang/go/issues/38184
			time.Sleep(2 * time.Millisecond)
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
	transport := &sentrycore.TransportMock{}
	sink := log.NewSentrySinkWithOptions(sentry.ClientOptions{Transport: transport})
	logger, exportLogs := logtest.Captured(t, sink)
	return logger, transport, func() { _ = exportLogs() }
}

func newTestHub(t testing.TB) (*sentry.Hub, *sentrycore.TransportMock) {
	transport := &sentrycore.TransportMock{}
	c, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
	assert.NoError(t, err)
	hub := sentry.NewHub(c, sentry.NewScope())
	return hub, transport
}
