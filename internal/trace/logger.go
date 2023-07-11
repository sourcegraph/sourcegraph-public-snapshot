package trace

import (
	"context"

	"github.com/sourcegraph/log"
)

// Logger will set the TraceContext on l if ctx has one. This is an expanded
// convenience function around l.WithTrace for the common case.
func Logger(ctx context.Context, l log.Logger) log.Logger {
	// Attach any trace (WithTrace no-ops if empty trace is provided)
	return l.WithTrace(Context(ctx))
}
