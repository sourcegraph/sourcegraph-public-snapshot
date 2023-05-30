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
	if _, err := r.cmd.CreateJob(ctx, r.options.Namespace, job); err != nil {
		return errors.Wrap(err, "creating job")
	}
	r.jobNames = append(r.jobNames, job.Name)

	// Wait for the job to complete before reading the logs. This lets us get also get exit codes.
	pod, podWaitErr := r.cmd.WaitForPodToSucceed(ctx, r.options.Namespace, job.Name)
	// Handle when the wait failed to do the things.
	if podWaitErr != nil && pod == nil {
		return errors.Wrapf(podWaitErr, "waiting for job %s to complete", job.Name)
	}
	// Always read the logs, even if the job fails.
	readLogErr := r.cmd.ReadLogs(
		ctx,
		r.options.Namespace,
		pod,
		command.KubernetesJobContainerName,
		r.commandLogger,
		spec.CommandSpec.Key,
		spec.CommandSpec.Command,
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

	return readLogErr
}
