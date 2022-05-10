package logging

// ErrorLogger captures the method required for logging an error.
type ErrorLogger interface {
	Error(msg string, ctx ...any)
}

// Log logs the given message and context when the given error is defined.
func Log(lg ErrorLogger, msg string, err *error, ctx ...any) {
	if lg == nil || err == nil || *err == nil {
		return
	}

	lg.Error(msg, append(append(make([]any, 0, 2+len(ctx)), "error", *err), ctx...)...)
}
