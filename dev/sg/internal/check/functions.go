package check

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

// EnableFunc can be implemented to allow toggling whether they are skipped or not.
//
// Errors can implement RenderableError to have their output rendered nicely.
type EnableFunc[Args any] func(ctx context.Context, args Args) error

// CheckAction is the interface used to implement check Checks. All output should be
// written to cio, and no input should ever be required.
type CheckAction[Args any] func(ctx context.Context, out *std.Output, args Args) error

// CheckFuncAction adapts simple CheckFuncs into the more complex ActionFunc interface.
func CheckFuncAction[Args any](fn CheckFunc) CheckAction[Args] {
	return func(ctx context.Context, out *std.Output, args Args) error {
		return fn(ctx)
	}
}

type IO struct {
	// Input can be read for user input. It may be nil in non-interactive modes.
	Input io.Reader
	// Output should be used to write progress messages. When in doubt, prefer to use
	// Verbose() and friends to limit noise in the output.
	*std.Output
}

// ActionFunc is the interface used to implement check Fixes. All output should be written
// to cio, and all input should only be read from cio (i.e. FixActions can be interactive)
type FixAction[Args any] func(ctx context.Context, cio IO, args Args) error

func CombineFix[CheckArgs any](fixes ...FixAction[CheckArgs]) FixAction[CheckArgs] {
	return func(ctx context.Context, cio IO, args CheckArgs) error {
		for _, fix := range fixes {
			if err := fix(ctx, cio, args); err != nil {
				return err
			}
		}
		return nil
	}
}
