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

	logger.Debug("a debug message") // 0

	logger = logger.With(log.String("some", "field"))

	logger.Info("hello world", log.String("hello", "world")) // 1

	logger = logger.WithTrace(log.TraceContext{TraceID: "asdf"})
	logger.Info("goodbye", log.String("world", "hello")) // 2
	logger.Warn("another message")                       // 3

	logger.Error("object of fields", // 4
		log.Object("object",
			log.String("field1", "value"),
			log.String("field2", "value"),
		))

	logs := exportLogs()
	assert.Len(t, logs, 5)
	for _, l := range logs {
		assert.Equal(t, l.Scope, "TestInitLogger") // scope is always applied
	}

	assert.Equal(t, map[string]interface{}{
		"some":  "field",
		"hello": "world",
	}, logs[1].Fields["Attributes"])

	assert.Equal(t, "asdf", logs[2].Fields["TraceId"])
	assert.Equal(t, map[string]interface{}{
		"some":  "field",
		"world": "hello",
	}, logs[2].Fields["Attributes"])

	assert.Equal(t, map[string]interface{}{
		"some": "field",
		"object": map[string]interface{}{
			"field1": "value",
			"field2": "value",
		},
	}, logs[4].Fields["Attributes"])
}
