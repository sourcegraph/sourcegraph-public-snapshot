package runtime

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type shellRuntime struct {
	cmd          command.Command
	operations   *command.Operations
	filesStore   workspace.FilesStore
	cloneOptions workspace.CloneOptions
	dockerOpts   command.DockerOptions
}

var _ Runtime = &shellRuntime{}

func (r *shellRuntime) Name() Name {
	return NameShell
}

func (r *shellRuntime) PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error) {
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

func (r *shellRuntime) NewRunner(ctx context.Context, logger command.Logger, options RunnerOptions) (runner.Runner, error) {
	run := runner.NewShellRunner(r.cmd, logger, options.Path, r.dockerOpts)
	if err := run.Setup(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to setup shell runner")
	}

	return run, nil
}

func (r *shellRuntime) NewRunnerSpecs(ws workspace.Workspace, steps []types.DockerStep) ([]runner.Spec, error) {
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
