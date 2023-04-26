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

var _ Runtime = &dockerRuntime{}

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
		runnerSpecs[i] = runner.Spec{
			CommandSpec: command.Spec{
				Key:       dockerKey(step.Key, i),
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

func dockerKey(stepKey string, index int) string {
	if len(stepKey) > 0 {
		return "step.docker." + stepKey
	}
	return fmt.Sprintf("step.docker.%d", index)
}
