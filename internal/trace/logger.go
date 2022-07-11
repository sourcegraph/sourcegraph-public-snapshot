package trace

import (
	"context"

	"github.com/sourcegraph/log"
)

// Logger will set the TraceContext on l if ctx has one, and also assign the trace
// family as a scope if a trace family is found. This is an expanded convenience function
// around l.WithTrace for the common case.
//
// If you already set the family manually on the logger scope, then you might want to use
// trace.Context(ctx) instead.
func Logger(ctx context.Context, l log.Logger) log.Logger {
	if t := TraceFromContext(ctx); t != nil {
		if t.family != "" {
			l = l.Scoped(t.family, "trace family")
		}
		if tc := Context(ctx); tc != nil {
			l = l.WithTrace(*tc)
		}
	}
	return l
}
