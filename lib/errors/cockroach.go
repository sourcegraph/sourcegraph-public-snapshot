package errors

import "github.com/cockroachdb/errors"

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

	// TODO - unify with hashicorp/go-multierror
	CombineErrors = errors.CombineErrors
)
