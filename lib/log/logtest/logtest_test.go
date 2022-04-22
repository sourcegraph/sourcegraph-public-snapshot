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
	assert.Equal(t, logs[0].Scope, "TestExport")
	assert.Equal(t, logs[0].Message, "hello world")
	assert.Equal(t, logs[0].Fields["Attributes"], map[string]interface{}{"key": "value"})
}
