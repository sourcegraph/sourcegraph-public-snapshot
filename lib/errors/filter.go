package errors

import (
	"context"
)

// Ignore filters out any errors that match pred.
func Ignore(err error, pred ErrorPredicate) error {
	if pred(err) {
		return nil
	}
	var filteredErr error
	for unwrapped := Unwrap(err); unwrapped != nil; unwrapped = Unwrap(unwrapped) {
		if !pred(unwrapped) {
			filteredErr = CombineErrors(filteredErr, unwrapped)
		}
	}
	return filteredErr
}

// ErrorPredicate is a function type that returns whether an error matches a given condition
type ErrorPredicate func(error) bool

// HasTypePred returns an ErrorPredicate that returns true for errors that unwrap to an error with the same type as target
func HasTypePred(target error) ErrorPredicate {
	return func(err error) bool {
		return HasType(err, target)
	}
}

// IsPred returns an ErrorPredicate that returns true for errors that uwrap to the target error
func IsPred(target error) ErrorPredicate {
	return func(err error) bool {
		return Is(err, target)
	}
}

var IsContextCanceled = IsPred(context.Canceled)
