package runtime

import (
	"context"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
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
	GetCommands(ws workspace.Workspace, steps []types.DockerStep) ([]command.CommandSpec, error)
}

var runtime Runtime
var once = &sync.Once{}

// SetupRuntime creates the runtime based on the configured environment.
func SetupRuntime(
	logger log.Logger,
	ops *command.Operations,
	filesStore workspace.FilesStore,
	commandOpts command.Options,
	cloneOpts workspace.CloneOptions,
) (setupErr error) {
	once.Do(func() {
		// Docker
		notFoundDockerTools, err := validateDockerRuntime()
		if err != nil {
			logger.Warn("failed to determine if docker tools are configured", log.Error(err))
		} else if len(notFoundDockerTools) > 0 {
			logger.Warn("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", notFoundDockerTools))
		} else {
			logger.Info("runtime 'docker' is supported")
			runtime = &dockerRuntime{
				operations:   ops,
				filesStore:   filesStore,
				commandOpts:  commandOpts,
				cloneOptions: cloneOpts,
			}
		}
		if runtime == nil {
			setupErr = ErrNoRuntime
		}
	})
	return setupErr
}

// GetRuntime returns the runtime that has been configured.
func GetRuntime() (Runtime, error) {
	if runtime == nil {
		return nil, ErrNoRuntime
	}
	return runtime, nil
}

// ErrNoRuntime is the error when there is no runtime configured.
var ErrNoRuntime = errors.New("runtime is not configured: use SetupRuntime to configure the runtime")
