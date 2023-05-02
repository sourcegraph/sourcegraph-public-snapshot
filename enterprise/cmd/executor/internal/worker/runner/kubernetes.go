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

	if err := r.cmd.WaitForJobToComplete(ctx, r.options.Namespace, job.Name, r.options.Retry); err != nil {
		return errors.Wrap(err, "waiting for job to complete")
	}

	pod, err := r.cmd.FindPod(ctx, r.options.Namespace, job.Name)
	if err != nil {
		return errors.Wrap(err, "finding pod")
	}

	return r.cmd.ReadLogs(ctx, r.options.Namespace, pod.Name, r.commandLogger, spec.CommandSpec.Key, spec.CommandSpec.Command)
}
