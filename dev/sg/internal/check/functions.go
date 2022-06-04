package check

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

type IO struct {
	// Input can be read for user input. It may be nil in non-interactive modes.
	Input io.Reader
	// Output can be used to write messages.
	*std.Output
}

// EnableFunc can be implemented to allow toggling whether they are skipped or not.
type EnableFunc[Args any] func(ctx context.Context, args Args) error

type ActionFunc[Args any] func(ctx context.Context, cio IO, args Args) error
