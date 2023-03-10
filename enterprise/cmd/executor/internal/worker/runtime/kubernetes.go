package runtime

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
)

type kubernetesRuntime struct {
	cmd          command.Command
	kubeCmd      *command.KubernetesCommand
	filesStore   workspace.FilesStore
	cloneOptions workspace.CloneOptions
	operations   *command.Operations
}

var _ Runtime = &kubernetesRuntime{}

func (r *kubernetesRuntime) Name() Name {
	return NameKubernetes
}

func (r *kubernetesRuntime) PrepareWorkspace(ctx context.Context, logger command.Logger, job types.Job) (workspace.Workspace, error) {
	return workspace.NewKubernetesWorkspace(
		ctx,
		r.filesStore,
		job,
		r.cmd,
		logger,
		r.cloneOptions,
		r.operations,
	)
}

func (r *kubernetesRuntime) NewRunner(ctx context.Context, logger command.Logger, options RunnerOptions) (runner.Runner, error) {
	jobRunner := runner.NewKubernetesRunner(r.kubeCmd, logger, options.Path)
	if err := jobRunner.Setup(ctx); err != nil {
		return nil, err
	}
	return jobRunner, nil
}

func (r *kubernetesRuntime) NewRunnerSpecs(ws workspace.Workspace, steps []types.DockerStep) ([]runner.Spec, error) {
	runnerSpecs := make([]runner.Spec, len(steps))
	for i, step := range steps {
		var key string
		if len(step.Key) != 0 {
			key = fmt.Sprintf("step.kubernetes.%s", step.Key)
		} else {
			key = fmt.Sprintf("step.kubernetes.%d", i)
		}

		runnerSpecs[i] = runner.Spec{
			CommandSpec: command.Spec{
				Key: key,
				Command: []string{
					"/bin/sh",
					"-c",
					filepath.Join("/data/.sourcegraph-executor", ws.ScriptFilenames()[i]),
				},
				Dir:       step.Dir,
				Env:       step.Env,
				Operation: r.operations.Exec,
			},
			Image: step.Image,
		}
	}

	return runnerSpecs, nil
}
