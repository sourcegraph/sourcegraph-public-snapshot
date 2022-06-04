package check

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
)

// CheckAction adapts simple CheckFuncs into the more complex ActionFunc interface.
func CheckAction[Args any](fn CheckFunc) ActionFunc[Args] {
	return func(ctx context.Context, cio IO, args Args) error {
		return fn(ctx)
	}
}

// CommandAction executes the given command as an action.
func CommandAction[Args any](cmd string) ActionFunc[Args] {
	return func(ctx context.Context, cio IO, args Args) error {
		// TODO send to cio, and pipe stdin in
		out, err := usershell.CombinedExec(ctx, `brew install git`)
		cio.Write(string(out))
		return err
	}
}
