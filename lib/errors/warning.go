package errors

// Warning embeds an error. Its purpose is to indicate that this error is not a critical error and
// may be ignored. Additionally, it **must** be logged only as a warning. If it cannot be logged as a
// warning, then these are not the droids you're looking for.
type Warning interface {
	error
	// IsWarning should always return true. It exists to differentiate regular errors with Warning
	// errors. That is, all Warning type objects are error types, but not all error types are
	// Warning types.
	IsWarning() bool
}

// warning is the error that wraps an error that is meant to be handled as a warning and not a
// critical error.
//
// AUTHOR'S NOTE
//
// @indradhanush: This type does not need a method `As(any) bool` and can be "asserted" with
// errors.As (see example below) when the underlying package being used is cockroachdb/errors. The
// `As` method from the cockroachdb/errors library is able to distinguish between warning and native
// error types.
//
// When writing this part of the code, I had implemented an `As(any) bool` method into this struct
// but it never got invoked and the corresponding tests in TestWarningError still pass the
// assertions. However after further deliberations during code review, I'm choosing to keep it as
// part of the method list of this type with an aim for interoperability in the future. But the
// method is a NOOP. The good news is that I've also added a test for this method in
// TestWarningError.
type warning struct {
	error error
}

// Ensure that warning always implements the Warning error interface.
var _ Warning = (*warning)(nil)

// NewWarningError will return an error of type warning. This should be used to wrap errors where we
// do not intend to return an error, increment an error metric. That is, if an error is returned and
// it is not critical and / or expected to be intermittent and / or nothing we can do about
// (example: 404 errors from upstream code host APIs in repo syncing), we should wrap the error with
// NewWarningError.
//
// Consumers of these errors should then use errors.As to check if the error is of a warning type
// and based on that, should just log it as a warning. For example:
//
//	var ref errors.Warning
//	err := someFunctionThatReturnsAWarningErrorOrACriticalError()
//	if err != nil && errors.As(err, &ref) {
//	    log.Warnf("failed to do X: %v", err)
//	}
//
//	if err != nil {
//	    return err
//	}
func NewWarningError(err error) *warning {
	return &warning{
		error: err,
	}
}

func (w *warning) Error() string {
	return w.error.Error()
}

// IsWarning always returns true. It exists to differentiate regular errors with Warning
// errors. That is, all Warning type objects are error types, but not all error types are Warning
// types.
func (w *warning) IsWarning() bool {
	return true
}

// Unwrap returns the underlying error of the warning.
func (w *warning) Unwrap() error {
	return w.error
}

// As will return true if the target is of type warning.
//
// However, this method is not invoked when `errors.As` is invoked. See note in the docstring of the
// warning struct for more context.
func (w *warning) As(target any) bool {
	if _, ok := target.(*warning); ok {
		return true
	}

	return false
}

// IsWarning is a helper to check whether the specified err is a Warning
func IsWarning(err error) bool {
	var ref Warning
	if As(err, &ref) {
		return ref.IsWarning()
	}
	return false
}
