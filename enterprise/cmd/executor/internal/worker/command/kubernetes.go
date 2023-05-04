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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	kubernetesVolumeName = "sg-executor-job-volume"
)

const (
	// KubernetesContainerName is the name of the Kubernetes container that run the main container.
	KubernetesContainerName = "sg-executor-job-container"
	// KubernetesExecutorMountPath is the path where the Kubernetes volume is mounted in the container.
	KubernetesExecutorMountPath = "/data"
	// KubernetesJobMountPath is the path where the Kubernetes volume is mounted in the container.
	KubernetesJobMountPath = "/job"
	// KubernetesVolumeMountSubPath is the path that is mounted in the Kubernetes pod container.
	// The path is removed from the path that the job data is writt
	KubernetesVolumeMountSubPath = "/data/"
)

// KubernetesContainerOptions contains options for the Kubernetes Job containers.
type KubernetesContainerOptions struct {
	Namespace             string
	NodeName              string
	NodeSelector          map[string]string
	RequiredNodeAffinity  KubernetesNodeAffinity
	PersistenceVolumeName string
	ResourceLimit         KubernetesResource
	ResourceRequest       KubernetesResource
	Retry                 KubernetesRetry
	KeepJobs              bool
	SecurityContext       KubernetesSecurityContext
}

// KubernetesNodeAffinity contains the Kubernetes node affinity for a Job.
type KubernetesNodeAffinity struct {
	MatchExpressions []corev1.NodeSelectorRequirement
	MatchFields      []corev1.NodeSelectorRequirement
}

// KubernetesResource contains the CPU and memory resources for a Kubernetes Job.
type KubernetesResource struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}

// KubernetesRetry contains the retry options for a Kubernetes Job.
type KubernetesRetry struct {
	Attempts int
	Backoff  time.Duration
}

// KubernetesSecurityContext contains the security context options for a Kubernetes Job.
type KubernetesSecurityContext struct {
	RunAsUser  *int64
	RunAsGroup *int64
	FSGroup    *int64
}

// KubernetesCommand interacts with the Kubernetes API.
type KubernetesCommand struct {
	Logger    log.Logger
	Clientset kubernetes.Interface
}

// CreateJob creates a Kubernetes job with the given name and command.
func (c *KubernetesCommand) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	return c.Clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
}

// DeleteJob deletes the Kubernetes job with the given name.
func (c *KubernetesCommand) DeleteJob(ctx context.Context, namespace string, jobName string) error {
	return c.Clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
}

var propagationPolicy = metav1.DeletePropagationBackground

// ReadLogs reads the logs of the given pod and writes them to the logger.
func (c *KubernetesCommand) ReadLogs(ctx context.Context, namespace string, podName string, containerName string, cmdLogger Logger, key string, command []string) error {
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: containerName})
	stream, err := req.Stream(ctx)
	if err != nil {
		return errors.Wrap(err, "opening log stream")
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
		return readIntoBuffer("stdout", w, stdout)
	})

	return eg
}

// FindPod finds the pod for the given job name.
func (c *KubernetesCommand) FindPod(ctx context.Context, namespace string, name string) (*corev1.Pod, error) {
	list, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + name})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, errors.Newf("no pods found for job %s", name)
	}
	return &list.Items[0], nil
}

// WaitForJobToComplete waits for the job with the given name to complete.
func (c *KubernetesCommand) WaitForJobToComplete(ctx context.Context, namespace string, name string, retry KubernetesRetry) error {
	attempts := 0
	for {
		// After 60 seconds, give up
		if attempts > retry.Attempts {
			return ErrKubernetesJobDidNotComplete
		}
		job, err := c.getJob(ctx, namespace, name)
		if err != nil {
			return errors.Wrap(err, "retrieving job")
		}
		if job.Status.Active == 0 && job.Status.Succeeded > 0 {
			return nil
		} else if job.Status.Failed > 0 {
			return ErrKubernetesJobFailed
		} else {
			time.Sleep(retry.Backoff)
			attempts++
		}
	}
}

// ErrKubernetesJobDidNotComplete is returned when a Kubernetes job does not complete.
var ErrKubernetesJobDidNotComplete = errors.New("job did not complete")

// ErrKubernetesJobFailed is returned when a Kubernetes job fails.
var ErrKubernetesJobFailed = errors.New("job failed")

func (c *KubernetesCommand) getJob(ctx context.Context, namespace string, name string) (*batchv1.Job, error) {
	return c.Clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *KubernetesCommand) getPod(ctx context.Context, namespace string, name string) (*corev1.Pod, error) {
	return c.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// NewKubernetesJob creates a Kubernetes job with the given name, image, volume path, and spec.
func NewKubernetesJob(name string, image string, spec Spec, path string, options KubernetesContainerOptions) *batchv1.Job {
	jobEnvs := make([]corev1.EnvVar, len(spec.Env))
	for i, env := range spec.Env {
		parts := strings.SplitN(env, "=", 2)
		jobEnvs[i] = corev1.EnvVar{
			Name:  parts[0],
			Value: parts[1],
		}
	}
	var affinity *corev1.Affinity
	if len(options.RequiredNodeAffinity.MatchExpressions) > 0 || len(options.RequiredNodeAffinity.MatchFields) > 0 {
		affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: options.RequiredNodeAffinity.MatchExpressions,
							MatchFields:      options.RequiredNodeAffinity.MatchFields,
						},
					},
				},
			},
		}
	}

	resourceLimit := corev1.ResourceList{
		corev1.ResourceMemory: options.ResourceLimit.Memory,
	}
	if !options.ResourceLimit.CPU.IsZero() {
		resourceLimit[corev1.ResourceCPU] = options.ResourceLimit.CPU
	}

	resourceRequest := corev1.ResourceList{
		corev1.ResourceMemory: options.ResourceRequest.Memory,
	}
	if !options.ResourceRequest.CPU.IsZero() {
		resourceRequest[corev1.ResourceCPU] = options.ResourceRequest.CPU
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					NodeName:     options.NodeName,
					NodeSelector: options.NodeSelector,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  options.SecurityContext.RunAsUser,
						RunAsGroup: options.SecurityContext.RunAsGroup,
						FSGroup:    options.SecurityContext.FSGroup,
					},
					Affinity:      affinity,
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:       KubernetesContainerName,
							Image:      image,
							Command:    spec.Command,
							WorkingDir: filepath.Join(KubernetesJobMountPath, spec.Dir),
							Env:        jobEnvs,
							Resources: corev1.ResourceRequirements{
								Limits:   resourceLimit,
								Requests: resourceRequest,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      kubernetesVolumeName,
									MountPath: KubernetesJobMountPath,
									SubPath:   strings.TrimPrefix(path, KubernetesVolumeMountSubPath),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: kubernetesVolumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: options.PersistenceVolumeName,
								},
							},
						},
					},
				},
			},
		},
	}
}
