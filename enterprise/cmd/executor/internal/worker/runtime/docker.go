package runtime

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type dockerRuntime struct {
	cmd          command.Command
	operations   *command.Operations
	filesStore   workspace.FilesStore
	cloneOptions workspace.CloneOptions
	dockerOpts   command.DockerOptions
}

func (d *dockerRuntime) Name() Name {
	return NameDocker
}

func (d *dockerRuntime) PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error) {
	// We can pass empty options as they are not used in the Docker path.
	return workspace.NewDockerWorkspace(
		ctx,
		d.filesStore,
		job,
		d.cmd,
		logger,
		d.cloneOptions,
		d.operations,
	)
}

func (d *dockerRuntime) NewRunner(ctx context.Context, logger command.Logger, options RunnerOptions) (runner.Runner, error) {
	r := runner.NewDockerRunner(d.cmd, logger, options.Path, d.dockerOpts, options.DockerAuthConfig)
	if err := r.Setup(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to setup docker runner")
	}

	return r, nil
}

func (d *dockerRuntime) NewRunnerSpecs(ws workspace.Workspace, steps []types.DockerStep) ([]runner.Spec, error) {
	runnerSpecs := make([]runner.Spec, len(steps))
	for i, step := range steps {
		var key string
		if len(step.Key) != 0 {
			key = fmt.Sprintf("step.docker.%s", step.Key)
		} else {
			key = fmt.Sprintf("step.docker.%d", i)
		}

		runnerSpecs[i] = runner.Spec{
			CommandSpec: command.Spec{
				Key:       key,
				Command:   nil,
				Dir:       step.Dir,
				Env:       step.Env,
				Operation: d.operations.Exec,
			},
			Image:      step.Image,
			ScriptPath: ws.ScriptFilenames()[i],
		}
	}

	return runnerSpecs, nil
}
