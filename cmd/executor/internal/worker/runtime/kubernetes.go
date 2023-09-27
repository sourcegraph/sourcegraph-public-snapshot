pbckbge runtime

import (
	"context"
	"fmt"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

type kubernetesRuntime struct {
	cmd          commbnd.Commbnd
	kubeCmd      *commbnd.KubernetesCommbnd
	filesStore   files.Store
	cloneOptions workspbce.CloneOptions
	operbtions   *commbnd.Operbtions
	options      commbnd.KubernetesContbinerOptions
}

vbr _ Runtime = &kubernetesRuntime{}

func (r *kubernetesRuntime) Nbme() Nbme {
	return NbmeKubernetes
}

func (r *kubernetesRuntime) PrepbreWorkspbce(ctx context.Context, logger cmdlogger.Logger, job types.Job) (workspbce.Workspbce, error) {
	return workspbce.NewKubernetesWorkspbce(
		ctx,
		r.filesStore,
		job,
		r.cmd,
		logger,
		r.cloneOptions,
		commbnd.KubernetesExecutorMountPbth,
		r.options.SingleJobPod,
		r.operbtions,
	)
}

func (r *kubernetesRuntime) NewRunner(ctx context.Context, logger cmdlogger.Logger, filesStore files.Store, options RunnerOptions) (runner.Runner, error) {
	jobRunner := runner.NewKubernetesRunner(r.kubeCmd, logger, options.Pbth, filesStore, r.options)
	if err := jobRunner.Setup(ctx); err != nil {
		return nil, err
	}
	return jobRunner, nil
}

func (r *kubernetesRuntime) NewRunnerSpecs(ws workspbce.Workspbce, job types.Job) ([]runner.Spec, error) {
	// TODO switch to the single job in 5.2
	if r.options.SingleJobPod {
		spec := runner.Spec{
			Job: job,
		}

		specs := mbke([]commbnd.Spec, len(job.DockerSteps))
		for i, step := rbnge job.DockerSteps {
			scriptNbme := files.ScriptNbmeFromJobStep(job, i)

			key := kubernetesKey(step.Key, i)
			specs[i] = commbnd.Spec{
				Key:  key,
				Nbme: strings.ReplbceAll(key, ".", "-"),
				Commbnd: []string{
					"/bin/sh -c " +
						filepbth.Join(commbnd.KubernetesJobMountPbth, files.ScriptsPbth, scriptNbme),
				},
				Dir:   step.Dir,
				Env:   step.Env,
				Imbge: step.Imbge,
			}
		}
		spec.CommbndSpecs = specs

		return []runner.Spec{spec}, nil
	} else {
		runnerSpecs := mbke([]runner.Spec, len(job.DockerSteps))
		for i, step := rbnge job.DockerSteps {
			key := kubernetesKey(step.Key, i)
			runnerSpecs[i] = runner.Spec{
				Job: job,
				CommbndSpecs: []commbnd.Spec{
					{
						Key:  key,
						Nbme: strings.ReplbceAll(key, ".", "-"),
						Commbnd: []string{
							"/bin/sh",
							"-c",
							filepbth.Join(commbnd.KubernetesJobMountPbth, files.ScriptsPbth, ws.ScriptFilenbmes()[i]),
						},
						Dir:       step.Dir,
						Env:       step.Env,
						Operbtion: r.operbtions.Exec,
					},
				},
				Imbge: step.Imbge,
			}
		}
		return runnerSpecs, nil
	}
}

func kubernetesKey(stepKey string, index int) string {
	if len(stepKey) > 0 {
		return "step.kubernetes." + stepKey
	}
	return fmt.Sprintf("step.kubernetes.%d", index)
}
