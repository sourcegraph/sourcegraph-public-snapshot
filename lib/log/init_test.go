package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitLogger(t *testing.T) {
	root := initLogger(
		&Resource{
			Name: "test",
		},
		zap.NewAtomicLevelAt(zap.DebugLevel),
		"console",
		true,
	)
	assert.NotNil(t, root)

	// Capture output - TODO export as API
	adapted := &zapAdapter{Logger: root.With(attributesNamespace)}
	observeCore, logs := observer.New(zap.DebugLevel)
	logger := Logger(adapted.withOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(observeCore, c)
	})))

	logger.Debug("a debug message") // 1

	logger = logger.With(String("some", "field"))
	logger.Info("hello world", String("hello", "world")) // 2

	logger = logger.WithTrace(TraceContext{TraceID: "asdf"})
	logger.Info("goodbye", String("world", "hello")) // 3
	logger.Warn("another message")                   // 4

	logger.Sync()

	lines := logs.All()
	assert.Len(t, lines, 4)

	// TODO stronger assertions
}
