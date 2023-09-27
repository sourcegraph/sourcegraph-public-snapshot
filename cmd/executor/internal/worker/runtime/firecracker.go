pbckbge runtime

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

type firecrbckerRuntime struct {
	cmdRunner       util.CmdRunner
	cmd             commbnd.Commbnd
	operbtions      *commbnd.Operbtions
	filesStore      files.Store
	cloneOptions    workspbce.CloneOptions
	firecrbckerOpts runner.FirecrbckerOptions
}

vbr _ Runtime = &firecrbckerRuntime{}

func (r *firecrbckerRuntime) Nbme() Nbme {
	return NbmeFirecrbcker
}

func (r *firecrbckerRuntime) PrepbreWorkspbce(ctx context.Context, logger cmdlogger.Logger, job types.Job) (workspbce.Workspbce, error) {
	return workspbce.NewFirecrbckerWorkspbce(
		ctx,
		r.filesStore,
		job,
		r.firecrbckerOpts.DockerOptions.Resources.DiskSpbce,
		r.firecrbckerOpts.KeepWorkspbces,
		r.cmdRunner,
		r.cmd,
		logger,
		r.cloneOptions,
		r.operbtions,
	)
}

func (r *firecrbckerRuntime) NewRunner(ctx context.Context, logger cmdlogger.Logger, filesStore files.Store, options RunnerOptions) (runner.Runner, error) {
	run := runner.NewFirecrbckerRunner(
		r.cmd,
		logger,
		options.Pbth,
		options.Nbme,
		r.firecrbckerOpts,
		options.DockerAuthConfig,
		r.operbtions,
	)
	if err := run.Setup(ctx); err != nil {
		return nil, err
	}
	return run, nil
}

func (r *firecrbckerRuntime) NewRunnerSpecs(ws workspbce.Workspbce, job types.Job) ([]runner.Spec, error) {
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
