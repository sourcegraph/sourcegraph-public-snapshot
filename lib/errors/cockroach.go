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

	Is        = errors.Is
	IsAny     = errors.IsAny
	As        = errors.As
	HasType   = errors.HasType
	Cause     = errors.Cause
	Unwrap    = errors.Unwrap
	UnwrapAll = errors.UnwrapAll

	BuildSentryReport = errors.BuildSentryReport
)

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
