package apiworker

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/command"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	idSet         *IDSet
	options       Options
	operations    *command.Operations
	runnerFactory func(dir string, logger *command.Logger, options command.Options, operations *command.Operations) command.Runner
}

var _ workerutil.Handler = &handler{}

// Handle clones the target code into a temporary directory, invokes the target indexer in a
// fresh docker container, and uploads the results to the external frontend API.
func (h *handler) Handle(ctx context.Context, s workerutil.Store, record workerutil.Record) error {
	job := record.(apiclient.Job)

	h.idSet.Add(job.ID)
	defer h.idSet.Remove(job.ID)

	// 🚨 SECURITY: The job logger must be supplied with all sensitive values that may appear
	// in a command constructed and run in the following function. Note that the command and
	// its output may both contain sensitive values, but only values which we directly
	// interpolate into the command. No command that we run on the host leaks environment
	// variables, and the user-specified commands (which could leak their environment) are
	// run in a clean VM.
	logger := command.NewLogger(union(h.options.RedactedValues, job.RedactedValues))

	defer func() {
		for _, entry := range logger.Entries() {
			if err := s.AddExecutionLogEntry(ctx, record.RecordID(), entry); err != nil {
				log15.Warn("Failed to upload executor log entry for job", "id", record.RecordID(), "err", err)
			}
		}
	}()

	// Create a working directory for this job which will be removed once the job completes.
	// If a repository is supplied as part of the job configuration, it will be cloned into
	// the working directory.

	hostRunner := h.runnerFactory("", logger, command.Options{}, h.operations)
	workingDirectory, err := h.prepareWorkspace(ctx, hostRunner, job.RepositoryName, job.Commit)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(workingDirectory)
	}()

	// Copy the file contents from the job record into the working directory
	for relativePath, content := range job.VirtualMachineFiles {
		path, err := filepath.Abs(filepath.Join(workingDirectory, relativePath))
		if err != nil {
			return err
		}

		if !strings.HasPrefix(path, workingDirectory) {
			return fmt.Errorf("refusing to write outside of working directory")
		}

		if err := ioutil.WriteFile(path, []byte(content), os.ModePerm); err != nil {
			return err
		}
	}

	name, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	options := command.Options{
		ExecutorName:       name.String(),
		FirecrackerOptions: h.options.FirecrackerOptions,
		ResourceOptions:    h.options.ResourceOptions,
	}
	runner := h.runnerFactory(workingDirectory, logger, options, h.operations)

	// Deduplicate and sort (for testing)
	imageMap := map[string]struct{}{}
	for _, dockerStep := range job.DockerSteps {
		imageMap[dockerStep.Image] = struct{}{}
	}

	imageNames := make([]string, 0, len(imageMap))
	for image := range imageMap {
		imageNames = append(imageNames, image)
	}
	sort.Strings(imageNames)

	// Create temp directory to store scripts in before they get copied into firecracker.
	// Script content is passed in from the job
	scriptsDir, err := makeTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(scriptsDir)
	}()

	scriptPaths := make([]string, 0, len(job.DockerSteps))
	for i, dockerStep := range job.DockerSteps {
		scriptPath := filepath.Join(scriptsDir, scriptNameFromJobStep(job, i))

		if err := ioutil.WriteFile(scriptPath, buildScript(dockerStep), os.ModePerm); err != nil {
			return err
		}

		scriptPaths = append(scriptPaths, scriptPath)
	}

	// Setup Firecracker VM (if enabled)
	if err := runner.Setup(ctx, imageNames, scriptPaths); err != nil {
		return err
	}
	defer func() {
		if teardownErr := runner.Teardown(ctx); teardownErr != nil {
			err = multierror.Append(err, teardownErr)
		}
	}()

	// Invoke each docker step sequentially
	for i, dockerStep := range job.DockerSteps {
		dockerStepCommand := command.CommandSpec{
			Key:        fmt.Sprintf("step.docker.%d", i),
			Image:      dockerStep.Image,
			ScriptPath: scriptPaths[i],
			Dir:        dockerStep.Dir,
			Env:        dockerStep.Env,
			Operation:  h.operations.Exec,
		}

		if err := runner.Run(ctx, dockerStepCommand); err != nil {
			return errors.Wrap(err, "failed to perform docker step")
		}
	}

	// Invoke each src-cli step sequentially
	for i, cliStep := range job.CliSteps {
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

func buildScript(dockerStep apiclient.DockerStep) []byte {
	return []byte(strings.Join(dockerStep.Commands, "\n"))
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

func scriptNameFromJobStep(job apiclient.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.Replace(job.RepositoryName, "/", "_", -1), job.Commit)
}
