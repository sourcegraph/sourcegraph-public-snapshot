package errors

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
)

// MultiError is a container for groups of errors.
//
// This is very inspired by https://github.com/knz/shakespeare/blob/master/pkg/cmd/errors.go
type MultiError struct {
	Errors []error
}

func (m *MultiError) initialize() {
	if m.Errors == nil {
		m.Errors = []error{}
	}
}

var _ error = (*MultiError)(nil)
var _ fmt.Formatter = (*MultiError)(nil)

func combineNonNilErrors(err1 error, err2 error) *MultiError {
	multi1, ok1 := err1.(*MultiError)
	multi2, ok2 := err2.(*MultiError)
	// guard against typed nils
	if ok1 && multi1 == nil {
		multi1 = new(MultiError)
	}
	if ok2 && multi2 == nil {
		multi2 = new(MultiError)
	}
	// ensure fields are non-nil
	multi1.initialize()
	multi2.initialize()
	// flatten
	if ok1 && ok2 {
		return &MultiError{
			Errors: append(multi1.Errors, multi2.Errors...),
		}
	} else if ok1 {
		return &MultiError{
			Errors: append(multi1.Errors, err2),
		}
	} else if ok2 {
		return &MultiError{
			Errors: append([]error{err1}, multi2.Errors...),
		}
	}
	return &MultiError{Errors: []error{err1, err2}}
}

func CombineErrors(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return combineNonNilErrors(err1, err2)
}

func (e *MultiError) Error() string { return fmt.Sprintf("%v", e) }

func (e *MultiError) Cause() error  { return e.Errors[len(e.Errors)-1] }
func (e *MultiError) Unwrap() error { return e.Errors[len(e.Errors)-1] }

func (e *MultiError) Format(s fmt.State, verb rune) { errors.FormatError(e, s, verb) }

func (e *MultiError) FormatError(p errors.Printer) error {
	if len(e.Errors) > 1 {
		p.Printf("%d errors occured:", len(e.Errors))
	}

	// Simple output
	var buf bytes.Buffer
	for _, err := range e.Errors {
		if len(e.Errors) > 1 {
			p.Print("\n\t* ")
		}
		buf.Reset()
		fmt.Fprintf(&buf, "%v", err)
		p.Print(strings.ReplaceAll(buf.String(), "\n", "\n    "))
	}

	// Print additional details
	if p.Detail() {
		p.Print("-- details follow")
		for i, err := range e.Errors {
			p.Printf("\n(%d) %+v", i+1, err)
		}
	}

	return nil
}

func (e *MultiError) Is(refError error) bool {
	if errors.Is(e, refError) {
		return true
	}
	for _, err := range e.Errors {
		if errors.Is(err, refError) {
			return true
		}
	}
	return false
}

// ErrorOrNil is a no-op that just returns self - all multierror constructors should return
// nil when len(e.Errors) == 0.
//
// It exists as a hangover from go-multierror days.
func (e *MultiError) ErrorOrNil() error {
	return e
}
