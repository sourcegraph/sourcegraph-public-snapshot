package runner

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// KubernetesOptions contains options for the Kubernetes runner.
type KubernetesOptions struct {
	Enabled          bool
	ConfigPath       string
	ContainerOptions command.KubernetesContainerOptions
}

type kubernetesRunner struct {
	internalLogger log.Logger
	commandLogger  command.Logger
	cmd            *command.KubernetesCommand
	jobNames       []string
	dir            string
	options        command.KubernetesContainerOptions
	// tmpDir is used to store temporary files used for k8s execution.
	tmpDir string
}

var _ Runner = &kubernetesRunner{}

// NewKubernetesRunner creates a new Kubernetes runner.
func NewKubernetesRunner(
	cmd *command.KubernetesCommand,
	commandLogger command.Logger,
	dir string,
	options command.KubernetesContainerOptions,
) Runner {
	return &kubernetesRunner{
		internalLogger: log.Scoped("kubernetes-runner", ""),
		commandLogger:  commandLogger,
		cmd:            cmd,
		dir:            dir,
		options:        options,
	}
}

func (r *kubernetesRunner) Setup(ctx context.Context) error {
	// Nothing to do here.
	return nil
}

func (r *kubernetesRunner) TempDir() string {
	return ""
}

func (r *kubernetesRunner) Teardown(ctx context.Context) error {
	if !r.options.KeepJobs {
		for _, name := range r.jobNames {
			r.internalLogger.Debug("Deleting kubernetes job", log.String("name", name))
			if err := r.cmd.DeleteJob(ctx, r.options.Namespace, name); err != nil {
				r.internalLogger.Error(
					"Failed to delete kubernetes job",
					log.String("jobName", name),
					log.Error(err),
				)
			}
		}
	}

	return nil
}

func (r *kubernetesRunner) Run(ctx context.Context, spec Spec) error {
	job := command.NewKubernetesJob(
		fmt.Sprintf("sg-executor-job-%s-%d-%s", spec.Queue, spec.JobID, spec.CommandSpec.Key),
		spec.Image,
		spec.CommandSpec,
		r.dir,
		r.options,
	)
	r.internalLogger.Debug("Creating job", log.Int("jobID", spec.JobID))
	if _, err := r.cmd.CreateJob(ctx, r.options.Namespace, job); err != nil {
		return errors.Wrap(err, "creating job")
	}
	r.jobNames = append(r.jobNames, job.Name)

	// Start the log entry for the command.
	logEntry := r.commandLogger.LogEntry(spec.CommandSpec.Key, spec.CommandSpec.Command)
	defer logEntry.Close()

	// Wait for the job to complete before reading the logs. This lets us get also get exit codes.
	r.internalLogger.Debug("Waiting for pod to succeed", log.Int("jobID", spec.JobID), log.String("jobName", job.Name))
	pod, podWaitErr := r.cmd.WaitForPodToSucceed(ctx, r.options.Namespace, job.Name)
	// Handle when the wait failed to do the things.
	if podWaitErr != nil && pod == nil {
		// There is no pod to read the logs of. Finalize the log entry and return the error.
		logEntry.Finalize(1)
		return errors.Wrapf(podWaitErr, "waiting for job %s to complete", job.Name)
	}
	// Always read the logs, even if the job fails.
	r.internalLogger.Debug("Reading logs", log.String("podName", pod.Name))
	readLogErr := r.cmd.ReadLogs(
		ctx,
		r.options.Namespace,
		pod,
		command.KubernetesJobContainerName,
		logEntry,
	)
	// Now handle the wait error.
	if podWaitErr != nil {
		var errMessage string
		if pod.Status.Message != "" {
			errMessage = fmt.Sprintf("job %s failed: %s", job.Name, pod.Status.Message)
		} else {
			errMessage = fmt.Sprintf("job %s failed", job.Name)
		}

		if readLogErr != nil {
			return errors.Wrap(readLogErr, errMessage)
		}
		return errors.New(errMessage)
	}
	r.internalLogger.Debug("Job completed successfully", log.Int("jobID", spec.JobID))
	return readLogErr
}
