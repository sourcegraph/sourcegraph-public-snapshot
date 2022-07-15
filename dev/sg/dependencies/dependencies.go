package dependencies

import (
	"io"

	"github.com/acobaugh/osrelease"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CheckArgs struct {
	Teammate bool

	ConfigFile          string
	ConfigOverwriteFile string
}

type category = check.Category[CheckArgs]

type dependency = check.Check[CheckArgs]

var checkAction = check.CheckFuncAction[CheckArgs]

type OS string

const (
	OSMac   OS = "darwin"
	OSLinux OS = "linux"
)

// Setup instantiates a runner that can check and fix setup dependencies.
func Setup(in io.Reader, out *std.Output, os OS) (*check.Runner[CheckArgs], error) {
	if os == OSMac {
		return check.NewRunner(in, out, Mac), nil
	}

	category, err := detectLinuxDistro()
	if err != nil {
		return nil, err
	}
	return check.NewRunner(in, out, category), nil
}

func detectLinuxDistro() ([]check.Category[CheckArgs], error) {
	osr, err := osrelease.Read()
	if err != nil {
		return nil, errors.Wrap(err, "reading os-release")
	}

	switch osr["ID"] {
	case "arch":
		return Arch, nil
	case "debian", "ubuntu":
		return Ubuntu, nil
	}

	return nil, errors.Newf("unknown distribution: %s", osr["ID"])
}
