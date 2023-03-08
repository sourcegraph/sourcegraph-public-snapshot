package runner

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegraph/log"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type kubernetesRunner struct {
	internalLogger log.Logger
	commandLogger  command.Logger
	cmd            *command.KubernetesCommand
	jobName        string
	// tmpDir is used to store temporary files used for k8s execution.
	tmpDir string
}

var _ Runner = &kubernetesRunner{}

func NewKubernetesRunner(
	cmd *command.KubernetesCommand,
) Runner {
	return &kubernetesRunner{
		internalLogger: log.Scoped("kubernetes-runner", ""),
		cmd:            cmd,
	}
}

func (r *kubernetesRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-kubernetes-runner")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp dir for kubernetes runner")
	}
	r.tmpDir = dir

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
	if err := r.cmd.DeleteJob(ctx, r.jobName); err != nil {
		r.internalLogger.Error(
			"Failed to delete kubernetes job",
			log.String("jobName", r.jobName),
			log.Error(err),
		)
	}

	return nil
}

func (r *kubernetesRunner) Run(ctx context.Context, spec Spec) error {
	job := newJob(fmt.Sprintf("job-%s-%s", spec.Queue, spec.JobID), spec)
	if _, err := r.cmd.CreateJob(ctx, job); err != nil {
		return errors.Wrap(err, "creating job")
	}
	r.jobName = job.Name

	podName, err := r.cmd.WaitForPodToStart(ctx, job.Name)
	if err != nil {
		return errors.Wrap(err, "waiting for pod to start")
	}

	if err = r.cmd.ReadLogs(ctx, podName, r.commandLogger, spec.CommandSpec.Key, spec.CommandSpec.Command); err != nil {
		return errors.Wrap(err, "reading logs")
	}

	return nil
}

func newJob(name string, spec Spec) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "job-container",
							Image:   spec.Image,
							Command: spec.CommandSpec.Command,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "job-volume",
									MountPath: "/job/temp",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "job-volume",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/Users/randell/Documents/dev/k8s-exp/temp",
								},
							},
						},
					},
				},
			},
		},
	}
}
