package runner

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// KubernetesOptions contains options for the Kubernetes runner.
type KubernetesOptions struct {
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
	dir, err := os.MkdirTemp("", "executor-kubernetes-runner")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp dir for kubernetes runner")
	}
	r.tmpDir = dir

	// TODO: figure out private registry stuff
	// If docker auth config is present, write it.
	//if len(r.dockerAuthConfig.Auths) > 0 {
	//	d, err := json.Marshal(r.dockerAuthConfig)
	//	if err != nil {
	//		return err
	//	}
	//
	//	dockerConfigPath, err := os.MkdirTemp(r.tmpDir, "docker_auth")
	//	if err != nil {
	//		return err
	//	}
	//	r.options.ConfigPath = dockerConfigPath
	//
	//	if err = os.WriteFile(filepath.Join(r.options.ConfigPath, "config.json"), d, os.ModePerm); err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (r *kubernetesRunner) TempDir() string {
	return r.tmpDir
}

func (r *kubernetesRunner) Teardown(ctx context.Context) error {
	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.internalLogger.Error(
			"Failed to remove kubernetes state tmp dir",
			log.String("tmpDir", r.tmpDir),
			log.Error(err),
		)
	}
	for _, name := range r.jobNames {
		if err := r.cmd.DeleteJob(ctx, name, r.options.Namespace); err != nil {
			r.internalLogger.Error(
				"Failed to delete kubernetes job",
				log.String("jobName", name),
				log.Error(err),
			)
		}
	}

	return nil
}

func (r *kubernetesRunner) Run(ctx context.Context, spec Spec) error {
	job := command.NewKubernetesJob(
		fmt.Sprintf("job-%s-%d-%s", spec.Queue, spec.JobID, spec.CommandSpec.Key),
		spec.Image,
		spec.CommandSpec,
		r.dir,
		r.options,
	)
	if _, err := r.cmd.CreateJob(ctx, job, r.options.Namespace); err != nil {
		return errors.Wrap(err, "creating job")
	}
	r.jobNames = append(r.jobNames, job.Name)

	if err := r.cmd.WaitForJobToComplete(ctx, job.Name, r.options.Namespace); err != nil {
		return errors.Wrap(err, "waiting for job to complete")
	}

	pod, err := r.cmd.FindPod(ctx, job.Name, r.options.Namespace)
	if err != nil {
		return errors.Wrap(err, "finding pod")
	}

	if err = r.cmd.ReadLogs(ctx, pod.Name, r.commandLogger, spec.CommandSpec.Key, spec.CommandSpec.Command, r.options.Namespace); err != nil {
		return errors.Wrap(err, "reading logs")
	}

	return nil
}
