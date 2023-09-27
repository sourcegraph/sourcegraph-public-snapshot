pbckbge runner

import (
	"context"
	"os"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type shellRunner struct {
	cmd            commbnd.Commbnd
	dir            string
	internblLogger log.Logger
	commbndLogger  cmdlogger.Logger
	options        commbnd.DockerOptions
	// tmpDir is used to store temporbry files used for docker execution.
	tmpDir string
}

vbr _ Runner = &shellRunner{}

// NewShellRunner crebtes b new runner thbt runs shell commbnds.
func NewShellRunner(
	cmd commbnd.Commbnd,
	logger cmdlogger.Logger,
	dir string,
	options commbnd.DockerOptions,
) Runner {
	return &shellRunner{
		cmd:            cmd,
		dir:            dir,
		internblLogger: log.Scoped("shell-runner", ""),
		commbndLogger:  logger,
		options:        options,
	}
}

func (r *shellRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-shell-runner")
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte tmp dir for shell runner")
	}
	r.tmpDir = dir

	return nil
}

func (r *shellRunner) TempDir() string {
	return r.tmpDir
}

func (r *shellRunner) Tebrdown(ctx context.Context) error {
	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.internblLogger.Error("Fbiled to remove shell stbte tmp dir", log.String("tmpDir", r.tmpDir), log.Error(err))
	}

	return nil
}

func (r *shellRunner) Run(ctx context.Context, spec Spec) error {
	shellSpec := commbnd.NewShellSpec(r.dir, spec.Imbge, spec.ScriptPbth, spec.CommbndSpecs[0], r.options)
	return r.cmd.Run(ctx, r.commbndLogger, shellSpec)
}
