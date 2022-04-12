package errors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

var (
	New    = errors.New
	Newf   = errors.Newf
	Wrap   = errors.Wrap
	Wrapf  = errors.Wrapf
	Errorf = errors.Errorf

	HasType           = errors.HasType
	WithMessage       = errors.WithMessage
	WithStack         = errors.WithStack
	BuildSentryReport = errors.BuildSentryReport
	Safe              = errors.Safe

	// TODO this implementation is complex, but uses the potentially problematic
	// cockroachdb/errors.UnwrapOnce
	IsAny = errors.IsAny
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
