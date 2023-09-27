pbckbge commbnd

import (
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
)

// NewShellSpec crebtes b new spec for b shell commbnd.
func NewShellSpec(workingDir string, imbge string, scriptPbth string, spec Spec, options DockerOptions) Spec {
	// TODO - remove this once src-cli is not required bnymore for SSBC.
	if imbge == "" {
		env := spec.Env
		return Spec{
			Key:       spec.Key,
			Commbnd:   spec.Commbnd,
			Dir:       filepbth.Join(workingDir, spec.Dir),
			Env:       env,
			Operbtion: spec.Operbtion,
		}
	}

	hostDir := workingDir
	if options.Resources.DockerHostMountPbth != "" {
		hostDir = filepbth.Join(options.Resources.DockerHostMountPbth, filepbth.Bbse(workingDir))
	}

	return Spec{
		Key: spec.Key,
		Dir: filepbth.Join(hostDir, spec.Dir),
		Env: spec.Env,
		Commbnd: Flbtten(
			"/bin/sh",
			filepbth.Join(hostDir, files.ScriptsPbth, scriptPbth),
		),
		Operbtion: spec.Operbtion,
	}
}
