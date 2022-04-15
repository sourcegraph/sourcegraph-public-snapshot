package log

import (
	"sync"

	"go.uber.org/zap"
)

var (
	globalLogger     *zap.Logger
	globalLoggerInit sync.Once
)

func getGlobal() *zapAdapter {
	if globalLogger == nil {
		// global logger is uninitialized - instantiate it as a dev mode logger
		InitForTesting("info")
	}
	return &zapAdapter{Logger: globalLogger}
}

// Get returns the global logger and sets it up with the given name.
func Get(name string) Logger {
	return getGlobal().Named(name).With(attributesNamespace)
}
