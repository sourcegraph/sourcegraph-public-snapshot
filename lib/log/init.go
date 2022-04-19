package log

import (
	"os"

	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/lib/log/internal/encoders"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/global"
	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

type Resource = otfields.Resource

// Init initializes the log package's global logger as a logger of the given resource.
// It must be called on service startup, i.e. 'main()', NOT on an 'init()' function.
//
// Subsequent calls will be a no-op, so do not call this within a non-service context.
// For testing, you can use 'logtest.Init' to initialize the logging library.
//
// If Init is not called, Get will panic.
func Init(r Resource) {
	level := zap.NewAtomicLevelAt(Level(os.Getenv(envSrcLogLevel)).Parse())
	format := encoders.ParseOutputFormat(os.Getenv(envSrcLogFormat))
	global.Init(r, level, format, os.Getenv("SRC_DEVELOPMENT") == "true")
}
