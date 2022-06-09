package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type handler struct {
	nameSet       *janitor.NameSet
	store         workerutil.Store
	options       Options
	operations    *command.Operations
	runnerFactory func(dir string, logger *command.Logger, options command.Options, operations *command.Operations) command.Runner
}

var (
	_ workerutil.Handler        = &handler{}
	_ workerutil.WithPreDequeue = &handler{}
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
func (h *handler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	job := record.(executor.Job)
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
	commandLogger := command.NewLogger(h.store, job, record.RecordID(), union(h.options.RedactedValues, job.RedactedValues))
	defer func() {
		flushErr := commandLogger.Flush()
		if flushErr != nil {
			if err != nil {
				err = errors.Append(err, flushErr)
			} else {
				err = flushErr
			}
		}
	}()

	// Create a working directory for this job which will be removed once the job completes.
	// If a repository is supplied as part of the job configuration, it will be cloned into
	// the working directory.
	logger.Info("Creating workspace")

	hostRunner := h.runnerFactory("", commandLogger, command.Options{}, h.operations)
	workingDirectory, err := h.prepareWorkspace(ctx, hostRunner, job.RepositoryName, job.Commit)
	if err != nil {
		return errors.Wrap(err, "failed to prepare workspace")
	}
	defer func() {
		_ = os.RemoveAll(workingDirectory)
	}()

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
		FirecrackerOptions: h.options.FirecrackerOptions,
		ResourceOptions:    h.options.ResourceOptions,
	}
	runner := h.runnerFactory(workingDirectory, commandLogger, options, h.operations)

	// Construct a map from filenames to file content that should be accessible to jobs
	// within the workspace. This consists of files supplied within the job record itself,
	// as well as file-version of each script step.
	workspaceFileContentsByPath := map[string][]byte{}

	for relativePath, content := range job.VirtualMachineFiles {
		path, err := filepath.Abs(filepath.Join(workingDirectory, relativePath))
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, workingDirectory) {
			return errors.Errorf("refusing to write outside of working directory")
		}

		workspaceFileContentsByPath[path] = []byte(content)
	}

	scriptNames := make([]string, 0, len(job.DockerSteps))
	for i, dockerStep := range job.DockerSteps {
		scriptName := scriptNameFromJobStep(job, i)
		scriptNames = append(scriptNames, scriptName)

		path := filepath.Join(workingDirectory, command.ScriptsPath, scriptName)
		workspaceFileContentsByPath[path] = buildScript(dockerStep)
	}

	if err := writeFiles(workspaceFileContentsByPath, commandLogger); err != nil {
		return errors.Wrap(err, "failed to write virtual machine files")
	}

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
		dockerStepCommand := command.CommandSpec{
			Key:        fmt.Sprintf("step.docker.%d", i),
			Image:      dockerStep.Image,
			ScriptPath: scriptNames[i],
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
		logger.Info(fmt.Sprintf("Running src-cli step #%d", i))

		cliStepCommand := command.CommandSpec{
			Key:       fmt.Sprintf("step.src.%d", i),
			Command:   append([]string{"src"}, cliStep.Commands...),
			Dir:       cliStep.Dir,
			Env:       cliStep.Env,
			Operation: h.operations.Exec,
		}

		if err := runner.Run(ctx, cliStepCommand); err != nil {
			return errors.Wrap(err, "failed to perform src-cli step")
		}
	}

	return nil
}

var scriptPreamble = `
set -x
`

func buildScript(dockerStep executor.DockerStep) []byte {
	return []byte(strings.Join(append([]string{scriptPreamble, ""}, dockerStep.Commands...), "\n") + "\n")
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

func scriptNameFromJobStep(job executor.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.ReplaceAll(job.RepositoryName, "/", "_"), job.Commit)
}

// writeFiles writes to the filesystem the content in the given map.
func writeFiles(workspaceFileContentsByPath map[string][]byte, logger *command.Logger) (err error) {
	handle := logger.Log("setup.fs", nil)
	defer func() {
		if err == nil {
			handle.Finalize(0)
		} else {
			handle.Finalize(1)
		}

		handle.Close()
	}()

	for path, content := range workspaceFileContentsByPath {
		if err := os.WriteFile(path, content, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
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
