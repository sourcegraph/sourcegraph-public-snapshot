package errors

import (
	"fmt"

	"github.com/cockroachdb/errors" //nolint:depguard // needed for implementation of multiError.As
)

// MultiError is a container for groups of errors.
type MultiError interface {
	error
	// Errors returns all errors carried by this MultiError, or an empty slice otherwise.
	Errors() []error
}

// multiError is our default underlying implementation for MultiError. It is compatible
// with cockroachdb.Error's formatting, printing, etc. and supports introspecting via
// As, Is, and friends.
//
// Implementation is based on https://github.com/knz/shakespeare/blob/master/pkg/cmd/errors.go
type multiError struct {
	errs []error
}

var _ MultiError = (*multiError)(nil)
var _ Typed = (*multiError)(nil)

func combineNonNilErrors(err1 error, err2 error) MultiError {
	multi1, ok1 := err1.(MultiError)
	multi2, ok2 := err2.(MultiError)
	// flatten
	var errs []error
	if ok1 && ok2 {
		errs = append(multi1.Errors(), multi2.Errors()...)
	} else if ok1 {
		errs = append(multi1.Errors(), err2)
	} else if ok2 {
		errs = append([]error{err1}, multi2.Errors()...)
	} else {
		errs = []error{err1, err2}
	}
	return &multiError{errs: errs}
}

// CombineErrors returns a MultiError from err1 and err2. If both are nil, nil is returned.
func CombineErrors(err1, err2 error) MultiError {
	if err1 == nil && err2 == nil {
		return nil
	}
	if err1 == nil {
		if multi, ok := err2.(MultiError); ok {
			return multi
		}
		return &multiError{errs: []error{err2}}
	} else if err2 == nil {
		if multi, ok := err1.(MultiError); ok {
			return multi
		}
		return &multiError{errs: []error{err1}}
	}
	return combineNonNilErrors(err1, err2)
}

// Append returns a MultiError created from all given errors, skipping errs that are nil.
// If no non-nil errors are provided, nil is returned.
func Append(err error, errs ...error) MultiError {
	multi := CombineErrors(err, nil)
	for _, e := range errs {
		if e != nil {
			multi = CombineErrors(multi, e)
		}
	}
	return multi
}

func (e *multiError) Error() string { return fmt.Sprintf("%v", e) }
func (e *multiError) Errors() []error {
	if e == nil || e.errs == nil {
		return nil
	}
	return e.errs
}

func (e *multiError) Cause() error  { return e.errs[len(e.errs)-1] }
func (e *multiError) Unwrap() error { return e.errs[len(e.errs)-1] }

func (e *multiError) Is(refError error) bool {
	if e == refError {
		return true
	}
	for _, err := range e.errs {
		if Is(err, refError) {
			return true
		}
	}
	return false
}

func (e *multiError) As(target any) bool {
	if m, ok := target.(*multiError); ok {
		*m = *e
		return true
	}
	for _, err := range e.errs {
		// To conform to the Typed interface, 'target' has to be of type
		// any. This means we cannot use our custom As wrapper which has
		// a generic argument, so use cockroachdb's As instead.
		if errors.As(err, target) {
			return true
		}
	}
	return false
}
