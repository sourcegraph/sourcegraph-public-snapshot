package check

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type IO struct {
	// Input can be read for user input. It may be nil in non-interactive modes.
	Input io.Reader
	// Progress can be used to write progress messages.
	Progress output.Writer
}

// EnableFunc can be implemented to allow toggling whether they are skipped or not.
type EnableFunc[Args any] func(ctx context.Context, args Args) error

type ActionFunc[Args any] func(ctx context.Context, cio IO, args Args) error

// CheckAction adapts simple CheckFuncs into the more complex ActionFunc interface.
func CheckAction[Args any](fn CheckFunc) ActionFunc[Args] {
	return func(ctx context.Context, cio IO, args Args) error {
		return fn(ctx)
	}
}
