package command

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	containerName = "job-container"
	namespace     = "default"
	nodeName      = "job-node"
	volumeName    = "job-volume"
	mountPath     = "/data"
)

type KubernetesCommand struct {
	Logger    log.Logger
	Clientset *kubernetes.Clientset
}

// CreateJob creates a Kubernetes job with the given name and command.
func (c *KubernetesCommand) CreateJob(ctx context.Context, job *batchv1.Job) (*batchv1.Job, error) {
	return c.Clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
}

// DeleteJob deletes the Kubernetes job with the given name.
func (c *KubernetesCommand) DeleteJob(ctx context.Context, jobName string) error {
	return c.Clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
}

var propagationPolicy = metav1.DeletePropagationBackground

// ReadLogs reads the logs of the given pod and writes them to the logger.
func (c *KubernetesCommand) ReadLogs(ctx context.Context, podName string, cmdLogger Logger, key string, command []string) error {
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: containerName})
	stream, err := req.Stream(ctx)
	if err != nil {
		return err
	}

	logEntry := cmdLogger.LogEntry(key, command)
	defer logEntry.Close()

	pipeReaderWaitGroup := readProcessPipe(logEntry, stream)

	select {
	case <-ctx.Done():
	case err = <-watchErrGroup(pipeReaderWaitGroup):
		if err != nil {
			return errors.Wrap(err, "reading process pipes")
		}
	}

	logEntry.Finalize(0)

	return nil
}

func readProcessPipe(w io.WriteCloser, stdout io.Reader) *errgroup.Group {
	eg := &errgroup.Group{}

	eg.Go(func() error {
		//_, err := io.Copy(w, stdout)
		//return err
		return readIntoBuffer("stdout", w, stdout)
	})

	return eg
}

// WaitForPodToStart waits for the pod with the given name to start.
func (c *KubernetesCommand) WaitForPodToStart(ctx context.Context, name string) (string, error) {
	var podName string
	return podName, retry.OnError(backoff, func(error) bool {
		return true
	}, func() error {
		var pod *corev1.Pod
		var err error
		if len(podName) == 0 {
			pod, err = c.FindPod(ctx, name)
		} else {
			pod, err = c.getPod(ctx, podName)
		}
		if err != nil {
			return err
		}
		podName = pod.Name
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			return nil
		}
		if pod.Status.Phase == corev1.PodPending && pod.Status.ContainerStatuses != nil {
			// Pod is starting, check container status
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Running != nil {
					// Container has started
					return nil
				} else if containerStatus.State.Waiting != nil {
					// Container is waiting, retry
					return errors.Newf("pod '%s' is waiting to start", name)
				} else {
					// Container is in an unknown state
					return errors.Newf("pod '%s' is in an unknown state '%s'", name, containerStatus.State)
				}
			}
		}
		return errors.Newf("pod '%s' is in an unknown phase '%s'", name, pod.Status.Phase)
	})
}

func (c *KubernetesCommand) WaitForJobToComplete(ctx context.Context, name string) error {
	for {
		job, err := c.getJob(ctx, name)
		if err != nil {
			return errors.Wrap(err, "retrieving job")
		}
		if job.Status.Active == 0 && job.Status.Succeeded > 0 {
			return nil
		} else if job.Status.Failed > 0 {
			return errors.Newf("job %s failed", name)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *KubernetesCommand) getJob(ctx context.Context, name string) (*batchv1.Job, error) {
	job, err := c.Clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (c *KubernetesCommand) FindPod(ctx context.Context, name string) (*corev1.Pod, error) {
	list, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + name})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, errors.Newf("no pods found for job %s", name)
	}
	return &list.Items[0], nil
}

func (c *KubernetesCommand) getPod(ctx context.Context, name string) (*corev1.Pod, error) {
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

// backoff is a slight modification to retry.DefaultBackoff.
var backoff = wait.Backoff{
	Steps:    50,
	Duration: 10 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

// NewKubernetesJob creates a Kubernetes job with the given name, image, volume path, and spec.
func NewKubernetesJob(name string, image string, spec Spec, path string) *batchv1.Job {
	jobEnvs := make([]corev1.EnvVar, len(spec.Env))
	for i, env := range spec.Env {
		parts := strings.SplitN(env, "=", 2)
		jobEnvs[i] = corev1.EnvVar{
			Name:  parts[0],
			Value: parts[1],
		}
	}
	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	}
	volumeSource := corev1.VolumeSource{}
	if util.IsKubernetes() {
		volumeMount.SubPath = path
		volumeSource.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: "executor-pvc",
		}
	} else {
		volumeSource.HostPath = &corev1.HostPathVolumeSource{
			Path: path,
		}
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					//NodeName:      nodeName,
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:       containerName,
							Image:      image,
							Command:    spec.Command,
							WorkingDir: filepath.Join(mountPath, spec.Dir),
							VolumeMounts: []corev1.VolumeMount{
								volumeMount,
							},
							Env: jobEnvs,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name:         volumeName,
							VolumeSource: volumeSource,
						},
					},
				},
			},
		},
	}
}
