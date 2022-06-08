package dependencies

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CheckArgs struct {
	Teammate bool

	ConfigFile          string
	ConfigOverwriteFile string
}

type category = check.Category[CheckArgs]

type dependency = check.Check[CheckArgs]

var checkAction = check.CheckAction[CheckArgs]

// cmdAction executes the given command as an action in a new user shell.
func cmdAction(cmd string) check.ActionFunc[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		// TODO send to cio, and pipe stdin in
		out, err := usershell.CombinedExec(ctx, cmd)
		cio.Write(string(out))
		return err
	}
}

func cmdsAction(cmds ...string) check.ActionFunc[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		for _, cmd := range cmds {
			// TODO send to cio, and pipe stdin in
			out, err := usershell.CombinedExec(ctx, cmd)
			cio.Write(string(out))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func teammatesOnly() check.EnableFunc[CheckArgs] {
	return func(ctx context.Context, args CheckArgs) error {
		if !args.Teammate {
			return errors.New("Disabled if not a Sourcegraph teammate")
		}
		return nil
	}
}

type OS string

const (
	OSMac    OS = "darwin"
	OSUbuntu OS = "ubuntu"
)

// Setup instantiates a runner that can check and fix setup dependencies.
func Setup(in io.Reader, out *std.Output, os OS) *check.Runner[CheckArgs] {
	if os == OSMac {
		return check.NewRunner(in, out, Mac)
	}
	return check.NewRunner(in, out, Ubuntu)
}
