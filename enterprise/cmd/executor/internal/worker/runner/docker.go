package command

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type dockerRunner struct {
	dir       string
	logger    log.Logger
	cmdLogger command.Logger
	options   command.DockerOptions
	// tmpDir is used to store temporary files used for docker execution.
	tmpDir string
}

var _ Runner = &dockerRunner{}

func NewDockerRunner(dir string, logger command.Logger, options command.DockerOptions) Runner {
	return &dockerRunner{
		dir:       dir,
		logger:    log.Scoped("docker-runner", ""),
		cmdLogger: logger,
		options:   options,
	}
}

func (r *dockerRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-docker-runner")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp dir for docker runner")
	}
	r.tmpDir = dir

	// If docker auth config is present, write it.
	if len(r.options.DockerAuthConfig.Auths) > 0 {
		d, err := json.Marshal(r.options.DockerAuthConfig)
		if err != nil {
			return err
		}
		dockerConfigPath, err := os.MkdirTemp(r.tmpDir, "docker_auth")
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dockerConfigPath, "config.json"), d, os.ModePerm); err != nil {
			return err
		}
		r.options.ConfigPath = dockerConfigPath
	}

	return nil
}

func (r *dockerRunner) Teardown(ctx context.Context) error {
	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.logger.Error("Failed to remove docker state tmp dir", log.String("tmpDir", r.tmpDir), log.Error(err))
	}

	return nil
}

func (r *dockerRunner) Run(ctx context.Context) error {
	dockerCommand := command.NewDockerCommand(r.cmdLogger, nil, r.dir, r.options)
	return dockerCommand.Run(ctx)
}
