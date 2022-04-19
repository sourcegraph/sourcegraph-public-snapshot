package logtest

import (
	"runtime"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/global"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

// Init can be used to instantiate the log package for running tests, to be called in
// TestMain for the relevant package. Remember to call (*testing.M).Run() after initializing
// the logger! Initialization sets the resource name to the name of the calling package.
//
// testing.M is an unused argument, used to indicate this function should be called in
// TestMain.
func Init(_ *testing.M, level log.Level) {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	nameParts := strings.Split(details.Name(), "/")
	packageName := strings.Join(nameParts[:len(nameParts)-1], "/")
	global.Init(otfields.Resource{
		Name:      packageName,
		Namespace: "test",
	}, zap.NewAtomicLevelAt(level.Parse()), "console", true)
}

type configurableAdapter interface {
	log.Logger

	WithOptions(options ...zap.Option) log.Logger
}

type CapturedLog struct {
	Scope      string
	Level      log.Level
	Message    string
	Attributes map[string]interface{}
}

// Get retrieves a logger from log.Get with the test's name and returns a callback,
// dumpLogs, which flushes the logger buffer and returns log entries.
//
// The logger is scoped to the test name.
func Get(t testing.TB) (logger log.Logger, dumpLogs func() []CapturedLog) {
	root := log.Get(t.Name())

	configurable := root.(configurableAdapter)
	observeCore, entries := observer.New(zap.DebugLevel)
	logger = configurable.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(observeCore, c)
	}))

	return logger, func() []CapturedLog {
		logger.Sync()
		entries := entries.TakeAll()
		logs := make([]CapturedLog, len(entries))
		for i, e := range entries {
			logs[i] = CapturedLog{
				Scope:      e.LoggerName,
				Level:      log.Level(e.Level.String()),
				Message:    e.Message,
				Attributes: e.ContextMap(),
			}
		}
		return logs
	}
}
