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
	runnerFactory func(dir string, logger *command.Logger, options command.Options) command.Runner
}

var _ workerutil.Handler = &handler{}

// Handle clones the target code into a temporary directory, invokes the target indexer in a
// fresh docker container, and uploads the results to the external frontend API.
func (h *handler) Handle(ctx context.Context, s workerutil.Store, record workerutil.Record) error {
	// 🚨 SECURITY: The job logger must be supplied with all sensitive values that may appear
	// in a command constructed and run in the following function. Note that the command and
	// its output may both contain sensitive values, but only values which we directly
	// interpolate into the command. No command that we run on the host leaks environment
	// variables, and the user-specified commands (which could leak their environment) are
	// run in a clean VM.
	logger := command.NewLogger(h.options.RedactedValues...)

	defer func() {
		for _, entry := range logger.Entries() {
			if err := s.AddExecutionLogEntry(ctx, record.RecordID(), entry); err != nil {
				log15.Warn("Failed to upload executor log entry for job", "id", record.RecordID(), "err", err)
			}
		}
	}()

	job := record.(apiclient.Job)

	h.idSet.Add(job.ID)
	defer h.idSet.Remove(job.ID)

	// Create a working directory for this job which will be removed once the job completes.
	// If a repository is supplied as part of the job configuration, it will be cloned into
	// the working directory.

	hostRunner := h.runnerFactory("", logger, command.Options{})
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

	runner := h.runnerFactory(workingDirectory, logger, command.Options{
		ExecutorName:       name.String(),
		FirecrackerOptions: h.options.FirecrackerOptions,
		ResourceOptions:    h.options.ResourceOptions,
	})

	imageMap := map[string]struct{}{}
	for _, dockerStep := range job.DockerSteps {
		imageMap[dockerStep.Image] = struct{}{}
	}

	images := make([]string, 0, len(imageMap))
	for image := range imageMap {
		images = append(images, image)
	}
	sort.Strings(images)

	// Setup Firecracker VM (if enabled)
	if err := runner.Setup(ctx, images); err != nil {
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
			Key:      fmt.Sprintf("step.docker.%d", i),
			Image:    dockerStep.Image,
			Commands: dockerStep.Commands,
			Dir:      dockerStep.Dir,
			Env:      dockerStep.Env,
		}

		if err := runner.Run(ctx, dockerStepCommand); err != nil {
			return errors.Wrap(err, "failed to perform docker step")
		}
	}

	// Invoke each src-cli step sequentially
	for i, cliStep := range job.CliSteps {
		cliStepCommand := command.CommandSpec{
			Key:      fmt.Sprintf("step.src.%d", i),
			Commands: append([]string{"src"}, cliStep.Commands...),
			Dir:      cliStep.Dir,
			Env:      cliStep.Env,
		}

		if err := runner.Run(ctx, cliStepCommand); err != nil {
			return errors.Wrap(err, "failed to perform src-cli step")
		}
	}

	return nil
}
