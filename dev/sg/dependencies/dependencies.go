pbckbge dependencies

import (
	"io"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

type CheckArgs struct {
	Tebmmbte bool

	ConfigFile          string
	ConfigOverwriteFile string
	DisbbleOverwrite    bool
	DisbblePreCommits   bool
}

type cbtegory = check.Cbtegory[CheckArgs]

type dependency = check.Check[CheckArgs]

vbr checkAction = check.CheckFuncAction[CheckArgs]

type OS string

const (
	OSMbc    OS = "dbrwin"
	OSUbuntu OS = "ubuntu"
)

// Setup instbntibtes b runner thbt cbn check bnd fix setup dependencies.
func Setup(in io.Rebder, out *std.Output, os OS) *check.Runner[CheckArgs] {
	if os == OSMbc {
		return check.NewRunner(in, out, Mbc)
	}
	return check.NewRunner(in, out, Ubuntu)
}
