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

type Runtime interface {
	PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error)
	NewRunner(ctx context.Context, logger command.Logger, vmName string, path string, job types.Job) (command.Runner, error)
	GetCommands(ws workspace.Workspace, steps []types.DockerStep) ([]command.CommandSpec, error)
}

var runtimes = make(map[types.RuntimeMode]Runtime)
var once = &sync.Once{}

func SetupRuntimes(
	logger log.Logger,
	ops *command.Operations,
	filesStore workspace.FilesStore,
	commandOpts command.Options,
	cloneOpts workspace.CloneOptions,
) {
	once.Do(func() {
		// Docker
		notFoundDockerTools, err := validateDockerRuntime()
		if err != nil {
			logger.Warn("failed to determine if docker tools are configured", log.Error(err))
		} else if len(notFoundDockerTools) > 0 {
			logger.Warn("runtime 'docker' is not supported: missing required tools", log.Strings("dockerTools", notFoundDockerTools))
		} else {
			logger.Info("runtime 'docker' is supported")
			runtimes[types.RuntimeModeDocker] = &dockerRuntime{
				operations:   ops,
				filesStore:   filesStore,
				commandOpts:  commandOpts,
				cloneOptions: cloneOpts,
			}
		}
	})
}

func GetRuntime(mode types.RuntimeMode) (Runtime, error) {
	runtime, ok := runtimes[mode]
	if !ok {
		return nil, errors.Newf("runtime %s is not configured", mode)
	}
	return runtime, nil
}
