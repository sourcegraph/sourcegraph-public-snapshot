package logtest

import (
	"flag"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/globallogger"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

// Init can be used to instantiate the log package for running tests, to be called in
// TestMain for the relevant package. Remember to call (*testing.M).Run() after initializing
// the logger!
//
// testing.M is an unused argument, used to indicate this function should be called in
// TestMain.
func Init(_ *testing.M) {
	// ensure Verbose is set up
	testing.Init()
	flag.Parse()
	// set reasonable defaults
	if testing.Verbose() {
		initGlobal(zapcore.DebugLevel)
	} else {
		initGlobal(zapcore.WarnLevel)
	}
}

// InitWithLevel does the same thing as Init, but uses the provided log level to configur
// the log level for this package's tests, which can be helpful for exceptionally noisy
// tests.
//
// If your loggers are parameterized, you can also use logtest.NoOp to silence output for
// specific tests.
func InitWithLevel(_ *testing.M, level log.Level) {
	initGlobal(level.Parse())
}

func initGlobal(level zapcore.Level) {
	// use an empty resource, we don't log output Resource in dev mode anyway
	globallogger.Init(otfields.Resource{}, level, encoders.OutputConsole, true)
}

// configurableAdapter exposes internal APIs on zapAdapter.
type configurableAdapter interface {
	log.Logger

	WithCore(func(c zapcore.Core) zapcore.Core) log.Logger
}

type CapturedLog struct {
	Time    time.Time
	Scope   string
	Level   log.Level
	Message string
	Fields  map[string]any
}

type LoggerOptions struct {
	// Level configures the minimum log level to output.
	Level log.Level
	// FailOnErrorLogs indicates that the test should fail if an error log is output.
	FailOnErrorLogs bool
}

func scopedTestLogger(t testing.TB, options LoggerOptions) log.Logger {
	// initialize just in case - the underlying call to log.Init is no-op if this has
	// already been done. We allow this in testing for convenience.
	Init(nil)

	// On cleanup, flush the global logger.
	t.Cleanup(func() { globallogger.Get(true).Sync() })

	root := log.Scoped(t.Name(), "")

	// Cast into internal API
	configurable := root.(configurableAdapter)

	// Hook test output
	return configurable.WithCore(func(c zapcore.Core) zapcore.Core {
		var level zapcore.LevelEnabler = c // by default, use the parent core's leveller
		if options.Level != "" {
			level = zap.NewAtomicLevelAt(options.Level.Parse())
		}

		return newTestingCore(t, level, options.FailOnErrorLogs) // replace the core entirely
	})
}

// Scoped retrieves a logger scoped to the the given test. It writes to testing.TB.
//
// Unlike log.Scoped(), logtest.Scoped() is safe to use without initialization.
func Scoped(t testing.TB) log.Logger {
	return scopedTestLogger(t, LoggerOptions{})
}

// Scoped retrieves a logger scoped to the the given test, configured with additional
// options. It writes to testing.TB.
//
// Unlike log.Scoped(), logtest.Scoped() is safe to use without initialization.
func ScopedWith(t testing.TB, options LoggerOptions) log.Logger {
	return scopedTestLogger(t, options)
}

// Captured retrieves a logger from scoped to the the given test, and returns a callback,
// dumpLogs, which flushes the logger buffer and returns log entries.
func Captured(t testing.TB) (logger log.Logger, exportLogs func() []CapturedLog) {
	// Cast into internal APIs
	configurable := scopedTestLogger(t, LoggerOptions{}).(configurableAdapter)

	observerCore, entries := observer.New(zap.DebugLevel) // capture all levels
	logger = configurable.WithCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(observerCore, c)
	})

	return logger, func() []CapturedLog {
		entries := entries.TakeAll()
		logs := make([]CapturedLog, len(entries))
		for i, e := range entries {
			logs[i] = CapturedLog{
				Time:    e.Time,
				Scope:   e.LoggerName,
				Level:   log.Level(e.Level.String()),
				Message: e.Message,
				Fields:  e.ContextMap(),
			}
		}
		return logs
	}
}

// NoOp returns a no-op Logger, useful for silencing all output in a specific test.
func NoOp(t *testing.T) log.Logger {
	return Scoped(t).IncreaseLevel("noop", "no-op logger", log.LevelNone)
}
