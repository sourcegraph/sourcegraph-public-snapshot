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

func (r *dockerRuntime) Name() Name {
	return NameDocker
}

func (r *dockerRuntime) PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error) {
	return workspace.NewDockerWorkspace(
		ctx,
		r.filesStore,
		job,
		r.cmd,
		logger,
		r.cloneOptions,
		r.operations,
	)
}

func (r *dockerRuntime) NewRunner(ctx context.Context, logger command.Logger, options RunnerOptions) (runner.Runner, error) {
	run := runner.NewDockerRunner(r.cmd, logger, options.Path, r.dockerOpts, options.DockerAuthConfig)
	if err := run.Setup(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to setup docker runner")
	}

	return run, nil
}

func (r *dockerRuntime) NewRunnerSpecs(ws workspace.Workspace, steps []types.DockerStep) ([]runner.Spec, error) {
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
				Operation: r.operations.Exec,
			},
			Image:      step.Image,
			ScriptPath: ws.ScriptFilenames()[i],
		}
	}

	return runnerSpecs, nil
}
