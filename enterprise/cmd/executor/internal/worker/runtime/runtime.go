package runtime

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Runtime describe how to run a job in a specific runtime environment.
type Runtime interface {
	// Name returns the name of the runtime.
	Name() string
	// PrepareWorkspace sets up the workspace for the Job.
	PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error)
	// NewRunner creates a runner that will execute the steps.
	NewRunner(ctx context.Context, logger command.Logger, vmName string, path string, job types.Job) (command.Runner, error)
	// GetCommands builds and returns the commands that the runner will execute.
	GetCommands(ws workspace.Workspace, steps []types.DockerStep) ([]command.Spec, error)
}

// New creates the runtime based on the configured environment.
func New(
	logger log.Logger,
	ops *command.Operations,
	filesStore workspace.FilesStore,
	commandOpts command.Options,
	cloneOpts workspace.CloneOptions,
	runner util.CmdRunner,
) (Runtime, error) {
	err := util.ValidateDockerTools(runner)
	if err != nil {
		var errMissingTools *util.ErrMissingTools
		if errors.As(err, &errMissingTools) {
			logger.Warn("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", errMissingTools.Tools))
		} else {
			logger.Warn("failed to determine if docker tools are configured", log.Error(err))
		}
	} else {
		logger.Info("runtime 'docker' is supported")
		return &dockerRuntime{
			operations:   ops,
			filesStore:   filesStore,
			commandOpts:  commandOpts,
			cloneOptions: cloneOpts,
		}, nil
	}
	return nil, ErrNoRuntime
}

// ErrNoRuntime is the error when there is no runtime configured.
var ErrNoRuntime = errors.New("runtime is not configured: use SetupRuntime to configure the runtime")
