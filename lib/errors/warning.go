package errors

// Warning embeds an error. It's purpose is to indicate that this error is not a critical error and
// maybe ignored. Additionally, the recommended log level for this kind of an error is Warn.
type Warning interface {
	error
	IsWarn() bool
}

type warning struct {
	Err error
}

// Ensure that warning always implements the error interface.
var _ error = (*warning)(nil)

// Ensure that warning always implements the Warning interface.
var _ Warning = (*warning)(nil)

func NewWarningError(err error) error {
	return &warning{
		Err: err,
	}
}

func (ce *warning) Error() string {
	return ce.Err.Error()
}

// IsWarn always returns true. It exists to differentiate regular errors with Warning errors. That
// is, all Warning type objects are error types, but not all error types are Warning types.
func (w *warning) IsWarn() bool {
	return true
}
