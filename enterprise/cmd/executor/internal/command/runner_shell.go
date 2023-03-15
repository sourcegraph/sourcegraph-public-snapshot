//go:build shell

package command

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewRunner creates a new runner with the given options.
func NewRunner(dir string, logger Logger, options Options, operations *Operations) Runner {
	return &shellRunner{
		dir:       dir,
		logger:    log.Scoped("shell-runner", ""),
		cmdLogger: logger,
		options:   options,
	}
}

type shellRunner struct {
	dir       string
	logger    log.Logger
	cmdLogger Logger
	options   Options
	// tmpDir is used to store temporary files used for docker execution.
	tmpDir string
}

var _ Runner = &shellRunner{}

func (r *shellRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-shell-runner")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp dir for shell runner")
	}
	r.tmpDir = dir

	return nil
}

func (r *shellRunner) Teardown(ctx context.Context) error {
	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.logger.Error("Failed to remove shell state tmp dir", log.String("tmpDir", r.tmpDir), log.Error(err))
	}

	return nil
}

func (r *shellRunner) Run(ctx context.Context, command CommandSpec) error {
	return runCommand(ctx, formatShellCommand(command, r.dir, r.options), r.cmdLogger)
}

func formatShellCommand(spec CommandSpec, dir string, options Options) command {
	// TODO - remove this once src-cli is not required anymore for SSBC.
	if spec.Image == "" {
		env := spec.Env
		return command{
			Key:       spec.Key,
			Command:   spec.Command,
			Dir:       filepath.Join(dir, spec.Dir),
			Env:       env,
			Operation: spec.Operation,
		}
	}

	hostDir := dir
	if options.ResourceOptions.DockerHostMountPath != "" {
		hostDir = filepath.Join(options.ResourceOptions.DockerHostMountPath, filepath.Base(dir))
	}

	return command{
		Key: spec.Key,
		Dir: filepath.Join(hostDir, spec.Dir),
		Env: spec.Env,
		Command: flatten(
			"/bin/sh",
			filepath.Join(hostDir, ScriptsPath, spec.ScriptPath),
		),
		Operation: spec.Operation,
	}
}
