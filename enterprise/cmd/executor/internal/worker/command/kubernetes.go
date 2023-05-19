package command

import (
	"context"
	"io"
	"path/filepath"
	"strings"

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
	// KubernetesExecutorMountPath is the path where the PersistentVolumeClaim is mounted in the Executor Pod.
	KubernetesExecutorMountPath = "/data"
	// KubernetesJobContainerName is the name of the container in the Job Pod that runs the step from the executor types.Job.
	KubernetesJobContainerName = "sg-executor-job-container"
	// KubernetesJobMountPath is the path where the PersistentVolumeClaim is mounted in the Job Pod.
	KubernetesJobMountPath = "/job"
)

const (
	// kubernetesJobVolumeName is the name of the PersistentVolumeClaim that is mounted in the Job Pod.
	kubernetesJobVolumeName = "sg-executor-job-volume"
	// kubernetesExecutorVolumeMountSubPath is the path where the PersistentVolumeClaim is mounted to in the Executor Pod.
	// The trailing slash is required to properly trim the specified path when creating the subpath in the Job Pod.
	kubernetesExecutorVolumeMountSubPath = "/data/"
)

// KubernetesContainerOptions contains options for the Kubernetes Job containers.
type KubernetesContainerOptions struct {
	Namespace             string
	NodeName              string
	NodeSelector          map[string]string
	RequiredNodeAffinity  KubernetesNodeAffinity
	PodAffinity           []corev1.PodAffinityTerm
	PodAntiAffinity       []corev1.PodAffinityTerm
	Tolerations           []corev1.Toleration
	PersistenceVolumeName string
	ResourceLimit         KubernetesResource
	ResourceRequest       KubernetesResource
	Deadline              *int64
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
		return errors.Wrapf(err, "opening log stream for pod %s", podName)
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
func (c *KubernetesCommand) WaitForJobToComplete(ctx context.Context, namespace string, name string) error {
	watch, err := c.Clientset.BatchV1().Jobs(namespace).Watch(ctx, metav1.ListOptions{Watch: true, FieldSelector: "metadata.name=" + name})
	if err != nil {
		return errors.Wrap(err, "watching job")
	}
	defer watch.Stop()
	// No need to add a timer. If the job exceeds the deadline, it will fail.
	for event := range watch.ResultChan() {
		job, ok := event.Object.(*batchv1.Job)
		if !ok {
			return errors.New("unexpected object type")
		}
		if job.Status.Succeeded > 0 {
			return nil
		}
		if job.Status.Failed > 0 {
			return ErrKubernetesJobFailed
		}
	}
	// Wont happen
	return nil
}

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
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: nil,
						TopologyKey:   "",
					},
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: nil,
						TopologyKey:   "",
					},
				},
			},
		}
	}
	if len(options.PodAffinity) > 0 {
		if affinity == nil {
			affinity = &corev1.Affinity{}
		}
		affinity.PodAffinity = &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: options.PodAffinity,
		}
	}
	if len(options.PodAntiAffinity) > 0 {
		if affinity == nil {
			affinity = &corev1.Affinity{}
		}
		affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: options.PodAntiAffinity,
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
					Affinity:              affinity,
					RestartPolicy:         corev1.RestartPolicyNever,
					Tolerations:           options.Tolerations,
					ActiveDeadlineSeconds: options.Deadline,
					Containers: []corev1.Container{
						{
							Name:       KubernetesJobContainerName,
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
									Name:      kubernetesJobVolumeName,
									MountPath: KubernetesJobMountPath,
									SubPath:   strings.TrimPrefix(path, kubernetesExecutorVolumeMountSubPath),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: kubernetesJobVolumeName,
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
