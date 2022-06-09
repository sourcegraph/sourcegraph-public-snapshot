// Package logging carries logic related Sourcegraph's legacy logger, and is DEPRECATED.
// All new logging should use lib/log, and existing logging should be opportunistically
// migrated to the new logger. See https://docs.sourcegraph.com/dev/how-to/add_logging
package logging

// ErrorLogger captures the method required for logging an error.
//
// Deprecated: All logging should use lib/log instead. See https://docs.sourcegraph.com/dev/how-to/add_logging
type ErrorLogger interface {
	Error(msg string, ctx ...any)
}

// Log logs the given message and context when the given error is defined.
//
// Deprecated: All logging should use lib/log instead. See https://docs.sourcegraph.com/dev/how-to/add_logging
func Log(lg ErrorLogger, msg string, err *error, ctx ...any) {
	if lg == nil || err == nil || *err == nil {
		return
	}

	lg.Error(msg, append(append(make([]any, 0, 2+len(ctx)), "error", *err), ctx...)...)
}
