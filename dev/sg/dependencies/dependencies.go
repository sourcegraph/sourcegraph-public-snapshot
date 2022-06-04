package dependencies

import (
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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

var commandAction = check.CommandAction[CheckArgs]

func NewRunner(os string, cio check.IO) *check.Runner[CheckArgs] {
	if os == "darwin" {
		return check.NewRunner(cio, MacOS)
	}
	return check.NewRunner(cio, Ubuntu)
}
