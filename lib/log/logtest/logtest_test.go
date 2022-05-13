package logtest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/log"
)

func TestExport(t *testing.T) {
	logger, exportLogs := Captured(t)
	assert.NotNil(t, logger)

	logger.Info("hello world", log.String("key", "value"))

	logs := exportLogs()
	assert.Len(t, logs, 1)
	assert.Equal(t, "TestExport", logs[0].Scope)
	assert.Equal(t, "hello world", logs[0].Message)
	assert.Equal(t, map[string]any{"key": "value"}, logs[0].Fields)
}
