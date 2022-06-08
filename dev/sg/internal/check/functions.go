package check

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

// EnableFunc can be implemented to allow toggling whether they are skipped or not.
//
// Errors can implement RenderableError to have their output rendered nicely.
type EnableFunc[Args any] func(ctx context.Context, args Args) error

// CheckAction is the interface used to implement check Checks. All output should be
// written to cio, and no input should ever be required.
type CheckAction[Args any] func(ctx context.Context, out output.Writer, args Args) error

// CheckFuncAction adapts simple CheckFuncs into the more complex ActionFunc interface.
func CheckFuncAction[Args any](fn CheckFunc) CheckAction[Args] {
	return func(ctx context.Context, out output.Writer, args Args) error {
		return fn(ctx)
	}
}

type IO struct {
	// Input can be read for user input. It may be nil in non-interactive modes.
	Input io.Reader
	// Writer should be used to write progress messages. When in doubt, prefer to use
	// Verbose() and friends to limit noise in the output.
	output.Writer
}

// ActionFunc is the interface used to implement check Fixes. All output should be written
// to cio, and all input should only be read from cio (i.e. FixActions can be interactive)
type FixAction[Args any] func(ctx context.Context, cio IO, args Args) error
