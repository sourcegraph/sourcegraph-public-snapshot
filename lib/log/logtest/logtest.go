package logtest

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/global"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

// Init can be used to instantiate the log package for running tests, to be called in
// TestMain for the relevant package. Remember to call (*testing.M).Run() after initializing
// the logger! Initialization sets the resource name to the name of the calling package.
//
// testing.M is an unused argument, used to indicate this function should be called in
// TestMain.
//
// level can be used to configure the log level for this package's tests, which can be
// helpful for exceptionally noisy tests. You can also consider using 'libtest.Get(t)'
// and printing logs manually with 'DumpLogs'
func Init(_ *testing.M, level log.Level) {
	initGlobal(level.Parse())
}

func initGlobal(level zapcore.Level) {
	// use an empty resource, we don't log output Resource in dev mode anyway
	global.Init(otfields.Resource{}, zap.NewAtomicLevelAt(level), encoders.OutputConsole, true)
}

// configurableAdapter exposes internal APIs on zapAdapter
type configurableAdapter interface {
	log.Logger

	WithOptions(options ...zap.Option) log.Logger
}

type CapturedLog struct {
	Time    time.Time
	Scope   string
	Level   log.Level
	Message string
	Fields  map[string]interface{}
}

// Get retrieves a logger from log.Get with the test's name and returns a callback,
// dumpLogs, which flushes the logger buffer and returns log entries. The returned logger
// is scoped to the test name.
//
// Unlike log.Get(), logtest.Get() is safe to use without initialization.
func Get(t testing.TB) (logger log.Logger, exportLogs func() []CapturedLog) {
	// initialize just in case - the underlying call to log.Init is no-op if this has
	// already been done. We allow this in testing for convenience.
	initGlobal(zapcore.DebugLevel)

	root := log.Get(t.Name())

	// Cast into internal API
	configurable := root.(configurableAdapter)

	observerCore, entries := observer.New(zap.DebugLevel) // capture all levels
	logger = configurable.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		// Set up AttributesNamespace to mirror the underlying core created by log.Get()
		observeCore := observerCore.With([]zapcore.Field{otfields.AttributesNamespace})
		// Tee to both the underlying core, and our observer core
		return zapcore.NewTee(observeCore, c)
	}))

	return logger, func() []CapturedLog {
		logger.Sync()
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

// Dump dumps a JSON summary of each log entry.
func Dump(t testing.TB, logs []CapturedLog) {
	for _, log := range logs {
		b, err := json.Marshal(&log)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Log(string(b))
	}
}

// DumpLogsIfFailed calls Dump if the test failed, otherwise does nothing.
func DumpIfFailed(t testing.TB, logs []CapturedLog) {
	if !t.Failed() {
		return
	}
	Dump(t, logs)
}
