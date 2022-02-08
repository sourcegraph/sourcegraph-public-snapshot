package errors

import "github.com/hashicorp/go-multierror"

type (
	MultiError = multierror.Error
	Group      = multierror.Group
)

var (
	Append         = multierror.Append
	ListFormatFunc = multierror.ListFormatFunc
)
