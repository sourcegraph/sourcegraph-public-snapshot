package log

import (
	"os"

	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/globallogger"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

const (
	envSrcDevelopment = "SRC_DEVELOPMENT"
	envSrcLogFormat   = "SRC_LOG_FORMAT"
	envSrcLogLevel    = "SRC_LOG_LEVEL"
)

type Resource = otfields.Resource

// Init initializes the log package's global logger as a logger of the given resource.
// It must be called on service startup, i.e. 'main()', NOT on an 'init()' function.
// Subsequent calls will panic, so do not call this within a non-service context.
//
// Init returns a callback, sync, that should be called before application exit.
//
// For testing, you can use 'logtest.Init' to initialize the logging library.
//
// If Init is not called, Get will panic.
func Init(r Resource) (sync func() error) {
	if globallogger.IsInitialized() {
		panic("log.Init initialized multiple times")
	}

	level := zap.NewAtomicLevelAt(Level(os.Getenv(envSrcLogLevel)).Parse())
	format := encoders.ParseOutputFormat(os.Getenv(envSrcLogFormat))
	development := os.Getenv(envSrcDevelopment) == "true"
	return globallogger.Init(r, level, format, development)
}
