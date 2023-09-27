pbckbge runtime

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type dockerRuntime struct {
	cmd          commbnd.Commbnd
	operbtions   *commbnd.Operbtions
	filesStore   files.Store
	cloneOptions workspbce.CloneOptions
	dockerOpts   commbnd.DockerOptions
}

vbr _ Runtime = &dockerRuntime{}

func (r *dockerRuntime) Nbme() Nbme {
	return NbmeDocker
}

func (r *dockerRuntime) PrepbreWorkspbce(ctx context.Context, logger cmdlogger.Logger, job types.Job) (workspbce.Workspbce, error) {
	return workspbce.NewDockerWorkspbce(
		ctx,
		r.filesStore,
		job,
		r.cmd,
		logger,
		r.cloneOptions,
		r.operbtions,
	)
}

func (r *dockerRuntime) NewRunner(ctx context.Context, logger cmdlogger.Logger, filesStore files.Store, options RunnerOptions) (runner.Runner, error) {
	run := runner.NewDockerRunner(r.cmd, logger, options.Pbth, r.dockerOpts, options.DockerAuthConfig)
	if err := run.Setup(ctx); err != nil {
		return nil, errors.Wrbp(err, "fbiled to setup docker runner")
	}
	return run, nil
}

func (r *dockerRuntime) NewRunnerSpecs(ws workspbce.Workspbce, job types.Job) ([]runner.Spec, error) {
	runnerSpecs := mbke([]runner.Spec, len(job.DockerSteps))
	for i, step := rbnge job.DockerSteps {
		runnerSpecs[i] = runner.Spec{
			Job: job,
			CommbndSpecs: []commbnd.Spec{
				{
					Key:       dockerKey(step.Key, i),
					Commbnd:   nil,
					Dir:       step.Dir,
					Env:       step.Env,
					Operbtion: r.operbtions.Exec,
				},
			},
			Imbge:      step.Imbge,
			ScriptPbth: ws.ScriptFilenbmes()[i],
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
