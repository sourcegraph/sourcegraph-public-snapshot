package dependencies

import (
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

type CheckArgs struct {
	ConfigFile          string
	ConfigOverwriteFile string
	DisableOverwrite    bool
	DisablePreCommits   bool
}

type category = check.Category[CheckArgs]

type dependency = check.Check[CheckArgs]

var checkAction = check.CheckFuncAction[CheckArgs]

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
