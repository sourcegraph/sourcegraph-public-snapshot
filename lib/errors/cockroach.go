package errors

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/cockroachdb/errors" //nolint:depguard
	"github.com/cockroachdb/redact"
)

func init() {
	registerCockroachSafeTypes()
}

var (
	// Safe is a arg marker for non-PII arguments.
	Safe = redact.Safe

	New = errors.New
	// Newf assumes all args are unsafe PII, except for types in registerCockroachSafeTypes.
	// Use Safe to mark non-PII args. Contents of format are retained.
	Newf = errors.Newf
	// Errorf is the same as Newf. It assumes all args are unsafe PII, except for types
	// in registerCockroachSafeTypes. Use Safe to mark non-PII args. Contents of format
	// are retained.
	Errorf = errors.Newf

	Wrap = errors.Wrap
	// Wrapf assumes all args are unsafe PII, except for types in registerCockroachSafeTypes.
	// Use Safe to mark non-PII args. Contents of format are retained.
	Wrapf = errors.Wrapf
	// WithMessage is the same as Wrap.
	WithMessage = errors.Wrap

	// WithStack annotates err with a stack trace at the point WithStack was
	// called. Useful for sentinel errors.
	WithStack = errors.WithStack

	// WithSafeDetails annotates an error with the given reportable details.
	// The format is made available as a PII-free string, alongside
	// with a PII-free representation of every additional argument.
	// Arguments can be reported as-is (without redaction) by wrapping
	// them using the Safe() function.
	//
	// If the format is empty and there are no arguments, the
	// error argument is returned unchanged.
	//
	// Detail is shown:
	// - when formatting with `%+v`.
	// - in Sentry reports.
	WithSafeDetails = errors.WithSafeDetails

	// Is checks if the error tree err is equal to the value target.
	//
	// For error types which do not contain any data, Is is equivalent to As.
	//
	// For error types which contain data, it's possible that Is
	// returns true for a value other than the one returned by As,
	// since an error tree can contain multiple errors of the same
	// concrete type but with different data.
	Is        = errors.Is
	IsAny     = errors.IsAny
	Cause     = errors.Cause
	Unwrap    = errors.Unwrap
	UnwrapAll = errors.UnwrapAll

	BuildSentryReport = errors.BuildSentryReport
)

// As checks if the error tree err is of type target, and if so,
// sets target to the value of the error.
//
// If looking for an error of concrete type T, then the second
// argument must be a non-nil pointer of type *T. This implies that
// if the error interface is implemented with a pointer receiver,
// then target must be of type **MyConcreteType.
//
// For error types which do not contain any data, As is equivalent to Is.
//
// For error types which contain data, As will return an arbitrary
// error of the target type, in case there are multiple errors of the
// same concrete type in the error tree.
//
// Compared to errors.As, this method uses a generic argument to prevent
// a runtime panic when target is not a pointer to an error type.
//
// Use AsInterface over this function for interface targets.
func As[T error](err error, target *T) bool {
	return errors.As(err, target)
}

// AsInterface checks if the error tree err is of type target (which must be
// an interface type), and if so, sets target to the value of the error.
//
// In general, 'I' may be any interface, not just an error interface.
// See internal/errcode/code.go for some examples.
//
// Use As over this function for concrete types.
func AsInterface[I any](err error, target *I) bool {
	if target == nil {
		panic("Expected non-nil pointer to interface")
	}
	if typ := reflect.TypeOf(target); typ.Elem().Kind() != reflect.Interface {
		panic("Expected pointer to interface")
	}
	return errors.As(err, target)
}

// HasType checks if the error tree err has a node of type T.
//
// CAVEAT: HasType is implemented via As. So strictly speaking, it is
// possible that HasType returns true via some implementation of
// `interface { As(target any) bool }` in the error tree that
// doesn't actually check the type.
func HasType[T error](err error) bool {
	// At the moment, the cockroachdb/errors package's implementation
	// of HasType does not correctly handle multi-errors, whereas As does,
	// so we implement HasType via As.
	// (See https://github.com/cockroachdb/errors/issues/145)
	var zero T
	return As(err, &zero)
}

// Extend multiError to work with cockroachdb errors. Implement here to keep imports in
// one place.

var _ fmt.Formatter = (*multiError)(nil)

func (e *multiError) Format(s fmt.State, verb rune) { errors.FormatError(e, s, verb) }

var _ errors.Formatter = (*multiError)(nil)

func (e *multiError) FormatError(p errors.Printer) error {
	if len(e.errs) > 1 {
		p.Printf("%d errors occurred:", len(e.errs))
	}

	// Simple output
	for _, err := range e.errs {
		if len(e.errs) > 1 {
			p.Print("\n\t* ")
		}
		p.Printf("%v", err)
	}

	// Print additional details
	if p.Detail() {
		p.Print("-- details follow")
		for i, err := range e.errs {
			p.Printf("\n(%d) %+v", i+1, err)
		}
	}

	return nil
}

// registerSafeTypes registers types that should not be considered PII by
// cockroachdb/errors.
//
// Sourced from https://sourcegraph.com/github.com/cockroachdb/cockroach/-/blob/pkg/util/log/redact.go?L141
func registerCockroachSafeTypes() {
	// We consider booleans and numeric values to be always safe for
	// reporting. A log call can opt out by using redact.Unsafe() around
	// a value that would be otherwise considered safe.
	redact.RegisterSafeType(reflect.TypeOf(true)) // bool
	redact.RegisterSafeType(reflect.TypeOf(123))  // int
	redact.RegisterSafeType(reflect.TypeOf(int8(0)))
	redact.RegisterSafeType(reflect.TypeOf(int16(0)))
	redact.RegisterSafeType(reflect.TypeOf(int32(0)))
	redact.RegisterSafeType(reflect.TypeOf(int64(0)))
	redact.RegisterSafeType(reflect.TypeOf(uint8(0)))
	redact.RegisterSafeType(reflect.TypeOf(uint16(0)))
	redact.RegisterSafeType(reflect.TypeOf(uint32(0)))
	redact.RegisterSafeType(reflect.TypeOf(uint64(0)))
	redact.RegisterSafeType(reflect.TypeOf(float32(0)))
	redact.RegisterSafeType(reflect.TypeOf(float64(0)))
	redact.RegisterSafeType(reflect.TypeOf(complex64(0)))
	redact.RegisterSafeType(reflect.TypeOf(complex128(0)))
	// Signal names are also safe for reporting.
	redact.RegisterSafeType(reflect.TypeOf(os.Interrupt))
	// Times and durations too.
	redact.RegisterSafeType(reflect.TypeOf(time.Time{}))
	redact.RegisterSafeType(reflect.TypeOf(time.Duration(0)))
}
