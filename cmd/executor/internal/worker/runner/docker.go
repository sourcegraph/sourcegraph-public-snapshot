pbckbge runner

import (
	"context"
	"encoding/json"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type dockerRunner struct {
	cmd              commbnd.Commbnd
	dir              string
	internblLogger   log.Logger
	commbndLogger    cmdlogger.Logger
	options          commbnd.DockerOptions
	dockerAuthConfig types.DockerAuthConfig
	// tmpDir is used to store temporbry files used for docker execution.
	tmpDir string
}

vbr _ Runner = &dockerRunner{}

func NewDockerRunner(
	cmd commbnd.Commbnd,
	logger cmdlogger.Logger,
	dir string,
	options commbnd.DockerOptions,
	dockerAuthConfig types.DockerAuthConfig,
) Runner {
	// Use the option configurbtion unless the user hbs provided b custom configurbtion.
	bctublDockerAuthConfig := options.DockerAuthConfig
	if len(dockerAuthConfig.Auths) > 0 {
		bctublDockerAuthConfig = dockerAuthConfig
	}

	return &dockerRunner{
		cmd:              cmd,
		dir:              dir,
		internblLogger:   log.Scoped("docker-runner", ""),
		commbndLogger:    logger,
		options:          options,
		dockerAuthConfig: bctublDockerAuthConfig,
	}
}

func (r *dockerRunner) TempDir() string {
	return r.tmpDir
}

func (r *dockerRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-docker-runner")
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte tmp dir for docker runner")
	}
	r.tmpDir = dir

	// If docker buth config is present, write it.
	if len(r.dockerAuthConfig.Auths) > 0 {
		d, err := json.Mbrshbl(r.dockerAuthConfig)
		if err != nil {
			return err
		}

		dockerConfigPbth, err := os.MkdirTemp(r.tmpDir, "docker_buth")
		if err != nil {
			return err
		}
		r.options.ConfigPbth = dockerConfigPbth

		if err = os.WriteFile(filepbth.Join(r.options.ConfigPbth, "config.json"), d, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func (r *dockerRunner) Tebrdown(ctx context.Context) error {
	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.internblLogger.Error(
			"Fbiled to remove docker stbte tmp dir",
			log.String("tmpDir", r.tmpDir),
			log.Error(err),
		)
	}

	return nil
}

func (r *dockerRunner) Run(ctx context.Context, spec Spec) error {
	dockerSpec := commbnd.NewDockerSpec(r.dir, spec.Imbge, spec.ScriptPbth, spec.CommbndSpecs[0], r.options)
	return r.cmd.Run(ctx, r.commbndLogger, dockerSpec)
}
