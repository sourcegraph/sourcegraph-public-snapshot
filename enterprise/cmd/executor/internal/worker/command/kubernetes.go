package command

import (
	"context"
	"io"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type KubernetesCommand struct {
	Logger    log.Logger
	Clientset *kubernetes.Clientset
}

func (c *KubernetesCommand) CreateJob(ctx context.Context, job *batchv1.Job) (*batchv1.Job, error) {
	return c.Clientset.BatchV1().Jobs("default").Create(ctx, job, metav1.CreateOptions{})
}

func (c *KubernetesCommand) DeleteJob(ctx context.Context, jobName string) error {
	return c.Clientset.BatchV1().Jobs("default").Delete(ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
}

var propagationPolicy = metav1.DeletePropagationBackground

func (c *KubernetesCommand) ReadLogs(ctx context.Context, podName string, cmdLogger Logger, key string, command []string) error {
	req := c.Clientset.CoreV1().Pods("default").GetLogs(podName, &corev1.PodLogOptions{Container: "job-container"})
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
		return readIntoBuffer("stdout", w, stdout)
	})

	return eg
}

func (c *KubernetesCommand) WaitForPodToStart(ctx context.Context, name string) (string, error) {
	var podName string
	return podName, retry.OnError(backoff, func(error) bool {
		return true
	}, func() error {
		var pod *corev1.Pod
		var err error
		if len(podName) == 0 {
			pod, err = c.findPod(ctx, name)
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

func (c *KubernetesCommand) findPod(ctx context.Context, name string) (*corev1.Pod, error) {
	list, err := c.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + name})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, errors.Newf("no pods found for job %s", name)
	}
	return &list.Items[0], nil
}

func (c *KubernetesCommand) getPod(ctx context.Context, name string) (*corev1.Pod, error) {
	pod, err := c.Clientset.CoreV1().Pods("default").Get(ctx, name, metav1.GetOptions{})
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
