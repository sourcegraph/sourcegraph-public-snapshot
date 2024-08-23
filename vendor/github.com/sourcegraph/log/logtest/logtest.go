package logtest

import (
	"flag"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/internal/configurable"
	"github.com/sourcegraph/log/internal/globallogger"
	"github.com/sourcegraph/log/internal/sinkcores/outputcore"
	"github.com/sourcegraph/log/internal/stderr"
	"github.com/sourcegraph/log/output"
)

// stdTestInit guards the initialization of the standard library testing package.
var stdTestInit sync.Once

// Init can be used to instantiate the log package for running tests, to be called in
// TestMain for the relevant package. Remember to call (*testing.M).Run() after initializing
// the logger!
//
// testing.M is an unused argument, used to indicate this function should be called in
// TestMain.
func Init(_ *testing.M) {
	stdTestInit.Do(func() {
		// ensure Verbose and standard library testing stuff is set up
		testing.Init()
		flag.Parse()
	})

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
	// send output from package-scope loggers to stderr (we can't write to testing here)
	w, err := stderr.Open()
	if err != nil {
		panic(err)
	}
	core := outputcore.NewCore(w, level, output.FormatConsole, zap.SamplingConfig{}, nil, true)
	// use an empty resource, we don't log output Resource in dev mode anyway
	globallogger.Init(log.Resource{}, true, []zapcore.Core{core})
}

type CapturedLog struct {
	Time    time.Time
	Scope   string
	Level   log.Level
	Message string
	Fields  map[string]interface{}
}

type CapturedLogs []CapturedLog

// Messages aggregates all messages (excluding fields) from the captured log entries.
func (cl CapturedLogs) Messages() []string {
	var messages []string
	for _, l := range cl {
		messages = append(messages, l.Message)
	}
	return messages
}

// Filter returns captured logs that match the condition.
func (cl CapturedLogs) Filter(condition func(l CapturedLog) bool) CapturedLogs {
	var filtered CapturedLogs
	for _, l := range cl {
		if condition(l) {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

// Contains asserts that at least one entry matching the condition exists in the captured
// logs.
func (cl CapturedLogs) Contains(condition func(l CapturedLog) bool) bool {
	for _, l := range cl {
		if condition(l) {
			return true
		}
	}
	return false
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

	root := log.Scoped(t.Name())

	// Cast into internal API
	cl := configurable.Cast(root)

	// Hook test output
	return cl.WithCore(func(c zapcore.Core) zapcore.Core {
		var level zapcore.LevelEnabler = c // by default, use the parent core's leveller
		if options.Level != "" {
			level = zap.NewAtomicLevelAt(options.Level.Parse())
		}

		// replace the core entirely
		return outputcore.NewCore(&testingWriter{
			t:          t,
			markFailed: options.FailOnErrorLogs,
		}, level, output.FormatConsole, zap.SamplingConfig{}, nil, true)
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
func Captured(t testing.TB) (logger log.Logger, exportLogs func() CapturedLogs) {
	// Cast into internal APIs
	cl := configurable.Cast(scopedTestLogger(t, LoggerOptions{}))

	observerCore, entries := observer.New(zap.DebugLevel) // capture all levels

	logger = cl.WithCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, observerCore)
	})

	return logger, func() CapturedLogs {
		observerCore.Sync()

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
func NoOp(_ *testing.T) log.Logger {
	return log.NoOp()
}
