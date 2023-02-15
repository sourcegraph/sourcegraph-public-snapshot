package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	nameSet       *janitor.NameSet
	logStore      command.ExecutionLogEntryStore
	filesStore    workspace.FilesStore
	options       Options
	operations    *command.Operations
	runnerFactory func(dir string, logger command.Logger, options command.Options, operations *command.Operations) command.Runner
}

var (
	_ workerutil.Handler[executor.Job] = &handler{}
	_ workerutil.WithPreDequeue        = &handler{}
)

// PreDequeue determines if the number of VMs with the current instance's VM Prefix is less than
// the maximum number of concurrent handlers. If so, then a new job can be dequeued. Otherwise,
// we have an orphaned VM somewhere on the host that will be cleaned up by the background janitor
// process - refuse to dequeue a job for now so that we do not over-commit on VMs and cause issues
// with keeping our heartbeats due to machine load. We'll continue to check this condition on the
// polling interval
func (h *handler) PreDequeue(ctx context.Context, logger log.Logger) (dequeueable bool, extraDequeueArguments any, err error) {
	if !h.options.FirecrackerOptions.Enabled {
		return true, nil, nil
	}

	runningVMsByName, err := ignite.ActiveVMsByName(context.Background(), h.options.VMPrefix, false)
	if err != nil {
		return false, nil, err
	}

	if len(runningVMsByName) < h.options.WorkerOptions.NumHandlers {
		return true, nil, nil
	}

	logger.Warn("Orphaned VMs detected - refusing to dequeue a new job until it's cleaned up",
		log.Int("numRunningVMs", len(runningVMsByName)),
		log.Int("numHandlers", h.options.WorkerOptions.NumHandlers))
	return false, nil, nil
}

// Handle clones the target code into a temporary directory, invokes the target indexer in a
// fresh docker container, and uploads the results to the external frontend API.
func (h *handler) Handle(ctx context.Context, logger log.Logger, job executor.Job) (err error) {
	logger = logger.With(
		log.Int("jobID", job.ID),
		log.String("repositoryName", job.RepositoryName),
		log.String("commit", job.Commit))

	start := time.Now()
	defer func() {
		if honey.Enabled() {
			_ = createHoneyEvent(ctx, job, err, time.Since(start)).Send()
		}
	}()

	// 🚨 SECURITY: The job logger must be supplied with all sensitive values that may appear
	// in a command constructed and run in the following function. Note that the command and
	// its output may both contain sensitive values, but only values which we directly
	// interpolate into the command. No command that we run on the host leaks environment
	// variables, and the user-specified commands (which could leak their environment) are
	// run in a clean VM.
	commandLogger := command.NewLogger(h.logStore, job, job.RecordID(), union(h.options.RedactedValues, job.RedactedValues))
	defer func() {
		if flushErr := commandLogger.Flush(); flushErr != nil {
			err = errors.Append(err, flushErr)
		}
	}()

	// Create a working directory for this job which will be removed once the job completes.
	// If a repository is supplied as part of the job configuration, it will be cloned into
	// the working directory.
	logger.Info("Creating workspace")

	hostRunner := h.runnerFactory("", commandLogger, command.Options{}, h.operations)
	ws, err := h.prepareWorkspace(ctx, hostRunner, job, commandLogger)
	if err != nil {
		return errors.Wrap(err, "failed to prepare workspace")
	}
	defer ws.Remove(ctx, h.options.KeepWorkspaces)

	vmNameSuffix, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// Construct a unique name for the VM prefixed by something that differentiates
	// VMs created by this executor instance and another one that happens to run on
	// the same host (as is the case in dev). This prefix is expected to match the
	// prefix given to ignite.CurrentlyRunningVMs in other parts of this service.
	name := fmt.Sprintf("%s-%s", h.options.VMPrefix, vmNameSuffix.String())

	// Before we setup a VM (and after we teardown), mark the name as in-use so that
	// the janitor process cleaning up orphaned VMs doesn't try to stop/remove the one
	// we're using for the current job.
	h.nameSet.Add(name)
	defer h.nameSet.Remove(name)

	options := command.Options{
		ExecutorName:       name,
		DockerOptions:      h.options.DockerOptions,
		FirecrackerOptions: h.options.FirecrackerOptions,
		ResourceOptions:    h.options.ResourceOptions,
	}
	// If the job has docker auth config set, prioritize that over the env var.
	if len(job.DockerAuthConfig.Auths) > 0 {
		options.DockerOptions.DockerAuthConfig = job.DockerAuthConfig
	}
	runner := h.runnerFactory(ws.Path(), commandLogger, options, h.operations)

	logger.Info("Setting up VM")

	// Setup Firecracker VM (if enabled)
	if err := runner.Setup(ctx); err != nil {
		return errors.Wrap(err, "failed to setup virtual machine")
	}
	defer func() {
		// Perform this outside of the task execution context. If there is a timeout or
		// cancellation error we don't want to skip cleaning up the resources that we've
		// allocated for the current task.
		if teardownErr := runner.Teardown(context.Background()); teardownErr != nil {
			err = errors.Append(err, teardownErr)
		}
	}()

	// Invoke each docker step sequentially
	for i, dockerStep := range job.DockerSteps {
		var key string
		if dockerStep.Key != "" {
			key = fmt.Sprintf("step.docker.%s", dockerStep.Key)
		} else {
			key = fmt.Sprintf("step.docker.%d", i)
		}
		dockerStepCommand := command.CommandSpec{
			Key:        key,
			Image:      dockerStep.Image,
			ScriptPath: ws.ScriptFilenames()[i],
			Dir:        dockerStep.Dir,
			Env:        dockerStep.Env,
			Operation:  h.operations.Exec,
		}

		logger.Info(fmt.Sprintf("Running docker step #%d", i))

		if err := runner.Run(ctx, dockerStepCommand); err != nil {
			return errors.Wrap(err, "failed to perform docker step")
		}
	}

	// Invoke each src-cli step sequentially
	for i, cliStep := range job.CliSteps {
		var key string
		if cliStep.Key != "" {
			key = fmt.Sprintf("step.src.%s", cliStep.Key)
		} else {
			key = fmt.Sprintf("step.src.%d", i)
		}

		cliStepCommand := command.CommandSpec{
			Key:       key,
			Command:   append([]string{"src"}, cliStep.Commands...),
			Dir:       cliStep.Dir,
			Env:       cliStep.Env,
			Operation: h.operations.Exec,
		}

		logger.Info(fmt.Sprintf("Running src-cli step #%d", i))

		if err := runner.Run(ctx, cliStepCommand); err != nil {
			return errors.Wrap(err, "failed to perform src-cli step")
		}
	}

	return nil
}

func union(a, b map[string]string) map[string]string {
	c := make(map[string]string, len(a)+len(b))

	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}

	return c
}

func createHoneyEvent(_ context.Context, job executor.Job, err error, duration time.Duration) honey.Event {
	fields := map[string]any{
		"duration_ms":    duration.Milliseconds(),
		"recordID":       job.RecordID(),
		"repositoryName": job.RepositoryName,
		"commit":         job.Commit,
		"numDockerSteps": len(job.DockerSteps),
		"numCliSteps":    len(job.CliSteps),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	return honey.NewEventWithFields("executor", fields)
}
