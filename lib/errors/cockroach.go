package errors

import (
	"fmt"

	"github.com/cockroachdb/errors" //nolint:depguard
)

var (
	New               = errors.New
	Newf              = errors.Newf
	Wrap              = errors.Wrap
	Wrapf             = errors.Wrapf
	Errorf            = errors.Errorf
	Is                = errors.Is
	As                = errors.As
	HasType           = errors.HasType
	WithMessage       = errors.WithMessage
	Cause             = errors.Cause
	Unwrap            = errors.Unwrap
	UnwrapAll         = errors.UnwrapAll
	WithStack         = errors.WithStack
	BuildSentryReport = errors.BuildSentryReport
	Safe              = errors.Safe
	IsAny             = errors.IsAny
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
