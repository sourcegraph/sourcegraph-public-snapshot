package runner

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/files"
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
	commandLogger  cmdlogger.Logger
	cmd            *command.KubernetesCommand
	jobNames       []string
	dir            string
	filesStore     files.Store
	options        command.KubernetesContainerOptions
	// tmpDir is used to store temporary files used for k8s execution.
	tmpDir string
}

var _ Runner = &kubernetesRunner{}

// NewKubernetesRunner creates a new Kubernetes runner.
func NewKubernetesRunner(
	cmd *command.KubernetesCommand,
	commandLogger cmdlogger.Logger,
	dir string,
	filesStore files.Store,
	options command.KubernetesContainerOptions,
) Runner {
	return &kubernetesRunner{
		internalLogger: log.Scoped("kubernetes-runner", ""),
		commandLogger:  commandLogger,
		cmd:            cmd,
		dir:            dir,
		filesStore:     filesStore,
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
	var job *batchv1.Job
	if r.options.SingleJobPod {
		workspaceFiles, err := files.GetWorkspaceFiles(ctx, r.commandLogger, r.filesStore, spec.CommandSpec.Job, command.KubernetesJobMountPath)
		if err != nil {
			return err
		}

		job = command.NewKubernetesSingleJob(
			fmt.Sprintf("sg-executor-job-%s-%d", spec.CommandSpec.Job.Queue, spec.CommandSpec.Job.ID),
			spec.CommandSpec,
			workspaceFiles,
			r.options,
		)
	} else {
		job = command.NewKubernetesJob(
			fmt.Sprintf("sg-executor-job-%s-%d-%s", spec.Queue, spec.JobID, spec.CommandSpec.Key),
			spec.Image,
			spec.CommandSpec,
			r.dir,
			r.options,
		)
	}
	r.internalLogger.Debug("Creating job", log.Int("jobID", spec.JobID))
	if _, err := r.cmd.CreateJob(ctx, r.options.Namespace, job); err != nil {
		return errors.Wrap(err, "creating job")
	}
	r.jobNames = append(r.jobNames, job.Name)

	// Start the log entry for the main command.
	logEntry := r.commandLogger.LogEntry(spec.CommandSpec.Key, spec.CommandSpec.Command)
	defer logEntry.Close()

	// Wait for the job to complete before reading the logs. This lets us get also get exit codes.
	r.internalLogger.Debug("Waiting for pod to succeed", log.Int("jobID", spec.JobID), log.String("jobName", job.Name))

	pod, podWaitErr := r.cmd.WaitForPodToSucceed(ctx, r.commandLogger, r.options.Namespace, job.Name, spec.CommandSpec)
	// Handle when the wait failed to do the things.
	if podWaitErr != nil && pod == nil {
		// There is no pod to read the logs of. Finalize the log entry and return the error.
		logEntry.Finalize(1)
		return errors.Wrapf(podWaitErr, "waiting for job %s to complete", job.Name)
	}

	// Now handle the wait error.
	if podWaitErr != nil {
		var errMessage string
		if pod.Status.Message != "" {
			errMessage = fmt.Sprintf("job %s failed: %s", job.Name, pod.Status.Message)
		} else {
			errMessage = fmt.Sprintf("job %s failed", job.Name)
		}
		return errors.New(errMessage)
	}
	r.internalLogger.Debug("Job completed successfully", log.Int("jobID", spec.JobID))
	return nil
}
