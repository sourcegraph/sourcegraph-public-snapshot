package configurable

import (
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/log"
)

// Logger exposes internal APIs that must be implemented on
// github.com/sourcegraph/log.zapAdapter
type Logger interface {
	log.Logger

	// WithCore is an internal API used to allow packages like logtest to hook into
	// underlying zap logger's core.
	WithCore(func(c zapcore.Core) zapcore.Core) log.Logger
}

// Cast provides a configurable logger API for testing purposes.
func Cast(l log.Logger) Logger { return l.(Logger) }
