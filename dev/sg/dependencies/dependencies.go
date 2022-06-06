package dependencies

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
)

type CheckArgs struct {
	Input io.Reader
	Out   *std.Output

	Teammate bool
	InRepo   bool

	ConfigFile          string
	ConfigOverwriteFile string
}

type category = check.Category[CheckArgs]

type dependency = check.Check[CheckArgs]

var checkAction = check.CheckAction[CheckArgs]

// commandAction executes the given command as an action in a new user shell.
func commandAction(cmd string) check.ActionFunc[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		// TODO send to cio, and pipe stdin in
		out, err := usershell.CombinedExec(ctx, cmd)
		cio.Write(string(out))
		return err
	}
}

func NewRunner(os string, cio check.IO) *check.Runner[CheckArgs] {
	if os == "darwin" {
		return check.NewRunner(cio, MacOS)
	}
	return check.NewRunner(cio, Ubuntu)
}
