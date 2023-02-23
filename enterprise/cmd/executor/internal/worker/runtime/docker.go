package runtime

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type dockerRuntime struct {
	operations   *command.Operations
	filesStore   workspace.FilesStore
	commandOpts  command.Options
	cloneOptions workspace.CloneOptions
}

func (d *dockerRuntime) Name() Name {
	return NameDocker
}

func (d *dockerRuntime) PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error) {
	// We can pass empty options as they are not used in the Docker path.
	hostRunner := command.NewRunner("", logger, command.Options{}, nil)
	return workspace.NewDockerWorkspace(
		ctx,
		d.filesStore,
		job,
		hostRunner,
		logger,
		d.cloneOptions,
		d.operations,
	)
}

func (d *dockerRuntime) NewRunner(ctx context.Context, logger command.Logger, vmName string, path string, job types.Job) (command.Runner, error) {
	options := command.Options{
		DockerOptions:   d.commandOpts.DockerOptions,
		ResourceOptions: d.commandOpts.ResourceOptions,
	}
	// If the job has docker auth config set, prioritize that over the env var.
	if len(job.DockerAuthConfig.Auths) > 0 {
		options.DockerOptions.DockerAuthConfig = job.DockerAuthConfig
	}
	runner := command.NewRunner(path, logger, options, d.operations)
	if err := runner.Setup(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to setup docker runner")
	}

	return runner, nil
}

func (d *dockerRuntime) GetCommands(ws workspace.Workspace, steps []types.DockerStep) ([]command.Spec, error) {
	commandSpecs := make([]command.Spec, len(steps))
	for i, step := range steps {
		var key string
		if len(step.Key) != 0 {
			key = fmt.Sprintf("step.docker.%s", step.Key)
		} else {
			key = fmt.Sprintf("step.docker.%d", i)
		}

		commandSpecs[i] = command.Spec{
			Key:        key,
			Image:      step.Image,
			ScriptPath: ws.ScriptFilenames()[i],
			Command:    nil,
			Dir:        step.Dir,
			Env:        step.Env,
			Operation:  d.operations.Exec,
		}
	}

	return commandSpecs, nil
}
