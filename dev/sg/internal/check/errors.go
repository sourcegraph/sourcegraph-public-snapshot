package check

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/std"

// RenderableError can be implemented to render expanded details about the error's results
// to sg output.
//
// This is intended as an equivalent to the old lint.Report-style return values.
//
// TODO check for this type and render the data
type RenderableError interface {
	error

	Render(dst *std.Output)
}

type renderFuncError struct {
	error
	render func(dst *std.Output)
}

var _ RenderableError = &renderFuncError{}

func NewRenderableError(err error, render func(dst *std.Output)) error {
	return &renderFuncError{error: err, render: render}
}

func (r *renderFuncError) Render(dst *std.Output) { r.render(dst) }
