package check

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type IO struct {
	// Input can be read for user input. It may be nil in non-interactive modes.
	Input io.Reader
	// Writer should be used to write progress messages. When in doubt, prefer to use
	// Verbose() and friends to limit noise in the output.
	output.Writer
}

// EnableFunc can be implemented to allow toggling whether they are skipped or not.
//
// Errors can implement RenderableError to have their output rendered nicely.
type EnableFunc[Args any] func(ctx context.Context, args Args) error

// ActionFunc is the interface used by Checks and Fixes. All output should be written
// from cio, and all input should only be read from cio.
type ActionFunc[Args any] func(ctx context.Context, cio IO, args Args) error

// CheckAction adapts simple CheckFuncs into the more complex ActionFunc interface.
func CheckAction[Args any](fn CheckFunc) ActionFunc[Args] {
	return func(ctx context.Context, cio IO, args Args) error {
		return fn(ctx)
	}
}
