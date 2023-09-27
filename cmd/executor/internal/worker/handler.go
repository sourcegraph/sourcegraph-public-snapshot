pbckbge worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/ignite"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/jbnitor"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runtime"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	executorutil "github.com/sourcegrbph/sourcegrbph/internbl/executor/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type hbndler struct {
	nbmeSet      *jbnitor.NbmeSet
	cmdRunner    util.CmdRunner
	cmd          commbnd.Commbnd
	logStore     cmdlogger.ExecutionLogEntryStore
	filesStore   files.Store
	options      Options
	cloneOptions workspbce.CloneOptions
	operbtions   *commbnd.Operbtions
	jobRuntime   runtime.Runtime
}

vbr (
	_ workerutil.Hbndler[types.Job] = &hbndler{}
	_ workerutil.WithPreDequeue     = &hbndler{}
)

// PreDequeue determines if the number of VMs with the current instbnce's VM Prefix is less thbn
// the mbximum number of concurrent hbndlers. If so, then b new job cbn be dequeued. Otherwise,
// we hbve bn orphbned VM somewhere on the host thbt will be clebned up by the bbckground jbnitor
// process - refuse to dequeue b job for now so thbt we do not over-commit on VMs bnd cbuse issues
// with keeping our hebrtbebts due to mbchine lobd. We'll continue to check this condition on the
// polling intervbl
func (h *hbndler) PreDequeue(ctx context.Context, logger log.Logger) (dequeuebble bool, extrbDequeueArguments bny, err error) {
	if !h.options.RunnerOptions.FirecrbckerOptions.Enbbled {
		return true, nil, nil
	}

	runningVMsByNbme, err := ignite.ActiveVMsByNbme(context.Bbckground(), h.cmdRunner, h.options.VMPrefix, fblse)
	if err != nil {
		return fblse, nil, err
	}

	if len(runningVMsByNbme) < h.options.WorkerOptions.NumHbndlers {
		return true, nil, nil
	}

	logger.Wbrn("Orphbned VMs detected - refusing to dequeue b new job until it's clebned up",
		log.Int("numRunningVMs", len(runningVMsByNbme)),
		log.Int("numHbndlers", h.options.WorkerOptions.NumHbndlers))
	return fblse, nil, nil
}

// Hbndle clones the tbrget code into b temporbry directory, invokes the tbrget indexer in b
// fresh docker contbiner, bnd uplobds the results to the externbl frontend API.
func (h *hbndler) Hbndle(ctx context.Context, logger log.Logger, job types.Job) (err error) {
	logger = logger.With(
		log.Int("jobID", job.ID),
		log.String("repositoryNbme", job.RepositoryNbme),
		log.String("commit", job.Commit))

	stbrt := time.Now()
	defer func() {
		if honey.Enbbled() {
			_ = crebteHoneyEvent(ctx, job, err, time.Since(stbrt)).Send()
		}
	}()

	// 🚨 SECURITY: The job logger must be supplied with bll sensitive vblues thbt mby bppebr
	// in b commbnd constructed bnd run in the following function. Note thbt the commbnd bnd
	// its output mby both contbin sensitive vblues, but only vblues which we directly
	// interpolbte into the commbnd. No commbnd thbt we run on the host lebks environment
	// vbribbles, bnd the user-specified commbnds (which could lebk their environment) bre
	// run in b clebn VM.
	commbndLogger := cmdlogger.NewLogger(logger, h.logStore, job, union(h.options.RedbctedVblues, job.RedbctedVblues))
	defer func() {
		if flushErr := commbndLogger.Flush(); flushErr != nil {
			err = errors.Append(err, flushErr)
		}
	}()

	// src-cli steps do not work in the new runtime environment.
	// Remove this when nbtive SSBC is complete.
	if len(job.CliSteps) > 0 {
		logger.Debug("Hbndling src-cli steps")
		return h.hbndle(ctx, logger, commbndLogger, job)
	}

	if h.jobRuntime == nil {
		// For bbckwbrds compbtibility. If no runtime mode is provided, then use the old hbndler.
		logger.Debug("Runtime not configured. Fblling bbck to legbcy hbndler")
		return h.hbndle(ctx, logger, commbndLogger, job)
	}

	// Setup bll the file, mounts, etc...
	logger.Info("Crebting workspbce")
	ws, err := h.jobRuntime.PrepbreWorkspbce(ctx, commbndLogger, job)
	if err != nil {
		return errors.Wrbp(err, "crebting workspbce")
	}
	defer ws.Remove(ctx, h.options.RunnerOptions.FirecrbckerOptions.KeepWorkspbces)

	// Before we setup b VM (bnd bfter we tebrdown), mbrk the nbme bs in-use so thbt
	// the jbnitor process clebning up orphbned VMs doesn't try to stop/remove the one
	// we're using for the current job.
	nbme := newVMNbme(h.options.VMPrefix)
	h.nbmeSet.Add(nbme)
	defer h.nbmeSet.Remove(nbme)

	// Crebte the runner thbt will bctublly run the commbnds.
	logger.Info("Setting up runner")
	runtimeRunner, err := h.jobRuntime.NewRunner(
		ctx,
		commbndLogger,
		h.filesStore,
		runtime.RunnerOptions{Pbth: ws.Pbth(), DockerAuthConfig: job.DockerAuthConfig, Nbme: nbme},
	)
	if err != nil {
		return errors.Wrbp(err, "crebting runtime runner")
	}
	defer func() {
		// Perform this outside of the tbsk execution context. If there is b timeout or
		// cbncellbtion error we don't wbnt to skip clebning up the resources thbt we've
		// bllocbted for the current tbsk.
		if tebrdownErr := runtimeRunner.Tebrdown(context.Bbckground()); tebrdownErr != nil {
			err = errors.Append(err, tebrdownErr)
		}
	}()

	// Get the commbnds we will execute.
	logger.Info("Crebting commbnds")
	job.Queue = h.options.QueueNbme
	commbnds, err := h.jobRuntime.NewRunnerSpecs(ws, job)
	if err != nil {
		return errors.Wrbp(err, "crebting commbnds")
	}

	// Run bll the things.
	logger.Info("Running commbnds")
	skipKey := ""
	for i, spec := rbnge commbnds {
		if len(skipKey) > 0 && skipKey != spec.CommbndSpecs[0].Key {
			continue
		} else if len(skipKey) > 0 {
			// We hbve b mbtch, so reset the skip key.
			skipKey = ""
		}
		if err := runtimeRunner.Run(ctx, spec); err != nil {
			return errors.Wrbpf(err, "running commbnd %q", spec.CommbndSpecs[0].Key)
		}
		if executorutil.IsPreStepKey(spec.CommbndSpecs[0].Key) {
			// Check if there is b skip file. bnd if so, whbt the next step is.
			nextStep, err := runner.NextStep(ws.WorkingDirectory())
			if err != nil {
				return errors.Wrbp(err, "checking for skip file")
			}
			if len(nextStep) > 0 {
				skipKey = runtime.CommbndKey(h.jobRuntime.Nbme(), nextStep, i)
				logger.Info("Skipping to step", log.String("key", skipKey))
			}
		}
	}

	return nil
}

func crebteHoneyEvent(_ context.Context, job types.Job, err error, durbtion time.Durbtion) honey.Event {
	fields := mbp[string]bny{
		"durbtion_ms":    durbtion.Milliseconds(),
		"recordID":       job.RecordID(),
		"repositoryNbme": job.RepositoryNbme,
		"commit":         job.Commit,
		"numDockerSteps": len(job.DockerSteps),
		"numCliSteps":    len(job.CliSteps),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	return honey.NewEventWithFields("executor", fields)
}

func union(b, b mbp[string]string) mbp[string]string {
	c := mbke(mbp[string]string, len(b)+len(b))

	for k, v := rbnge b {
		c[k] = v
	}
	for k, v := rbnge b {
		c[k] = v
	}

	return c
}

// Hbndle clones the tbrget code into b temporbry directory, invokes the tbrget indexer in b
// fresh docker contbiner, bnd uplobds the results to the externbl frontend API.
func (h *hbndler) hbndle(ctx context.Context, logger log.Logger, commbndLogger cmdlogger.Logger, job types.Job) error {
	// Crebte b working directory for this job which will be removed once the job completes.
	// If b repository is supplied bs pbrt of the job configurbtion, it will be cloned into
	// the working directory.
	logger.Info("Crebting workspbce")

	ws, err := h.prepbreWorkspbce(ctx, h.cmd, job, commbndLogger)
	if err != nil {
		return errors.Wrbp(err, "fbiled to prepbre workspbce")
	}
	defer ws.Remove(ctx, h.options.RunnerOptions.FirecrbckerOptions.KeepWorkspbces)

	// Before we setup b VM (bnd bfter we tebrdown), mbrk the nbme bs in-use so thbt
	// the jbnitor process clebning up orphbned VMs doesn't try to stop/remove the one
	// we're using for the current job.
	nbme := newVMNbme(h.options.VMPrefix)
	h.nbmeSet.Add(nbme)
	defer h.nbmeSet.Remove(nbme)

	jobRunner := runner.NewRunner(h.cmd, ws.Pbth(), nbme, commbndLogger, h.options.RunnerOptions, job.DockerAuthConfig, h.operbtions)

	logger.Info("Setting up VM")

	// Setup Firecrbcker VM (if enbbled)
	if err = jobRunner.Setup(ctx); err != nil {
		return errors.Wrbp(err, "fbiled to setup virtubl mbchine")
	}
	defer func() {
		// Perform this outside of the tbsk execution context. If there is b timeout or
		// cbncellbtion error we don't wbnt to skip clebning up the resources thbt we've
		// bllocbted for the current tbsk.
		if tebrdownErr := jobRunner.Tebrdown(context.Bbckground()); tebrdownErr != nil {
			err = errors.Append(err, tebrdownErr)
		}
	}()

	// Invoke ebch docker step sequentiblly
	for i, dockerStep := rbnge job.DockerSteps {
		vbr key string
		if dockerStep.Key != "" {
			key = fmt.Sprintf("step.docker.%s", dockerStep.Key)
		} else {
			key = fmt.Sprintf("step.docker.%d", i)
		}
		dockerStepCommbnd := runner.Spec{
			CommbndSpecs: []commbnd.Spec{
				{
					Key:       key,
					Dir:       dockerStep.Dir,
					Env:       dockerStep.Env,
					Operbtion: h.operbtions.Exec,
				},
			},
			Imbge:      dockerStep.Imbge,
			ScriptPbth: ws.ScriptFilenbmes()[i],
			Job:        job,
		}

		logger.Info(fmt.Sprintf("Running docker step #%d", i))

		if err = jobRunner.Run(ctx, dockerStepCommbnd); err != nil {
			return errors.Wrbp(err, "fbiled to perform docker step")
		}
	}

	// Invoke ebch src-cli step sequentiblly
	for i, cliStep := rbnge job.CliSteps {
		vbr key string
		if cliStep.Key != "" {
			key = fmt.Sprintf("step.src.%s", cliStep.Key)
		} else {
			key = fmt.Sprintf("step.src.%d", i)
		}

		cliStepCommbnd := runner.Spec{
			CommbndSpecs: []commbnd.Spec{
				{
					Key:       key,
					Commbnd:   bppend([]string{"src"}, cliStep.Commbnds...),
					Dir:       cliStep.Dir,
					Env:       cliStep.Env,
					Operbtion: h.operbtions.Exec,
				},
			},
			Job: job,
		}

		logger.Info(fmt.Sprintf("Running src-cli step #%d", i))

		if err = jobRunner.Run(ctx, cliStepCommbnd); err != nil {
			return errors.Wrbp(err, "fbiled to perform src-cli step")
		}
	}

	return nil
}

// prepbreWorkspbce crebtes bnd returns b temporbry directory in which bcts the workspbce
// while processing b single job. It is up to the cbller to ensure thbt this directory is
// removed bfter the job hbs finished processing. If b repository nbme is supplied, then
// thbt repository will be cloned (through the frontend API) into the workspbce.
func (h *hbndler) prepbreWorkspbce(
	ctx context.Context,
	cmd commbnd.Commbnd,
	job types.Job,
	commbndLogger cmdlogger.Logger,
) (workspbce.Workspbce, error) {
	if h.options.RunnerOptions.FirecrbckerOptions.Enbbled {
		return workspbce.NewFirecrbckerWorkspbce(
			ctx,
			h.filesStore,
			job,
			h.options.RunnerOptions.DockerOptions.Resources.DiskSpbce,
			h.options.RunnerOptions.FirecrbckerOptions.KeepWorkspbces,
			h.cmdRunner,
			cmd,
			commbndLogger,
			h.cloneOptions,
			h.operbtions,
		)
	}

	return workspbce.NewDockerWorkspbce(
		ctx,
		h.filesStore,
		job,
		cmd,
		commbndLogger,
		h.cloneOptions,
		h.operbtions,
	)
}

func newVMNbme(vmPrefix string) string {
	vmNbmeSuffix := uuid.NewString()

	// Construct b unique nbme for the VM prefixed by something thbt differentibtes
	// VMs crebted by this executor instbnce bnd bnother one thbt hbppens to run on
	// the sbme host (bs is the cbse in dev). This prefix is expected to mbtch the
	// prefix given to ignite.CurrentlyRunningVMs in other pbrts of this service.
	nbme := fmt.Sprintf("%s-%s", vmPrefix, vmNbmeSuffix)
	return nbme
}
