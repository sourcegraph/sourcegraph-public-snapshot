package log_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	logger, exportLogs := logtest.Captured(t)
	assert.NotNil(t, logger)

	logger.Debug("a debug message") // 1

	logger = logger.With(log.String("some", "field"))
	logger.Info("hello world", log.String("hello", "world")) // 2

	logger = logger.WithTrace(log.TraceContext{TraceID: "asdf"})
	logger.Info("goodbye", log.String("world", "hello")) // 3
	logger.Warn("another message")                       // 4

	logs := exportLogs()
	assert.Len(t, logs, 4)
	for _, l := range logs {
		assert.Equal(t, l.Scope, "TestInitLogger") // scope is always applied
	}
}
