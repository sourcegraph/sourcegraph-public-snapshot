package errcode

import "errors"

type nonRetryableError struct {
	error
}

func (nonRetryableError) NonRetryable() bool {
	return true
}

// MakeNonRetryable makes any error non-retryable.
func MakeNonRetryable(err error) error {
	return nonRetryableError{err}
}

// IsNonRetryable will check if err or one of its causes is a error that cannot be retried.
func IsNonRetryable(err error) bool {
	var e interface{ NonRetryable() bool }
	return errors.As(err, &e) && e.NonRetryable()
}
