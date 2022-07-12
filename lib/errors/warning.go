package errors

type Warning interface {
	error
	IsWarn() bool
}

// warning is the error that wraps an error with an error level.
type warning struct {
	error error
}

// Ensure that classifiedError always implements the error interface.
var _ error = (*warning)(nil)

func NewWarningError(err error) error {
	return &warning{
		error: err,
	}
}

func (ce *warning) Error() string {
	return ce.error.Error()
}

// IsWarn always returns true. It exists to differentiate regular errors with Warning errors. That
// is, all Warning type objects are error types, but not all error types are Warning types.
func (w *warning) IsWarn() bool {
	return true
}
