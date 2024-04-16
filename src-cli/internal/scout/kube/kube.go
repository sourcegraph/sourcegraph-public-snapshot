package kube

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout"
	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	metav1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// GetPods fetches all pods in a given namespace.
func GetPods(ctx context.Context, cfg *scout.Config) ([]corev1.Pod, error) {
	podInterface := cfg.K8sClient.CoreV1().Pods(cfg.Namespace)
	podList, err := podInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Pod{}, errors.Wrap(err, "could not list pods")
	}

	if len(podList.Items) == 0 {
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		fmt.Println(msg.Render(`
            No pods exist in this namespace.
            Did you mean to use the --namespace flag?

            If you are attempting to check
            resources for a docker deployment, you
            must use the --docker flag.
            See --help for more info.
            `))
		os.Exit(1)
	}

	return podList.Items, nil
}

// GetPod returns a pod object with the given name from a list of pods.
func GetPod(podName string, pods []corev1.Pod) (corev1.Pod, error) {
	for _, p := range pods {
		if p.Name == podName {
			return p, nil
		}
	}
	return corev1.Pod{}, errors.New("no pod with this name exists in this namespace")
}

// GetPodMetrics fetches metrics for a given pod from the Kubernetes Metrics API.
// It accepts:
// - ctx: The context for the request
// - cfg: The scout config containing Kubernetes clientsets
// - pod: The pod specification
//
// It returns:
// - podMetrics: The PodMetrics object containing metrics for the pod
// - Any error that occurred while fetching the metrics
func GetPodMetrics(ctx context.Context, cfg *scout.Config, pod corev1.Pod) (*metav1beta1.PodMetrics, error) {
	podMetrics, err := cfg.MetricsClient.
		MetricsV1beta1().
		PodMetricses(cfg.Namespace).
		Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pod metrics")
	}

	return podMetrics, nil
}

// GetLimits generates resource limits for containers in a pod.
//
// It accepts:
// - ctx: The context for the request
// - cfg: The scout config containing Kubernetes clientsets
// - pod: The pod specification
// - containerMetrics: A pointer to a ContainerMetrics struct to populate
//
// It populates the containerMetrics.Limits field with a map of container names
// to resource limits (CPU, memory, storage) for each container in the pod.
//
// It returns:
// - Any error that occurred while fetching resource limits
func AddLimits(ctx context.Context, cfg *scout.Config, pod *corev1.Pod, containerMetrics *scout.ContainerMetrics) error {
	for _, container := range pod.Spec.Containers {
		containerName := container.Name
		capacity, err := GetPvcCapacity(ctx, cfg, container, pod)
		if err != nil {
			return errors.Wrap(err, "while getting storage capacity of PV")
		}

		rsrcs := scout.Resources{
			Cpu:     container.Resources.Limits.Cpu().ToDec(),
			Memory:  container.Resources.Limits.Memory().ToDec(),
			Storage: capacity,
		}
		containerMetrics.Limits[containerName] = rsrcs
	}
	return nil
}

// GetUsage generates resource usage statistics for a Kubernetes container.
//
// It accepts:
// - ctx: The context for the request
// - cfg: The scout config containing Kubernetes clientsets
// - metrics: Container resource limits
// - pod: The pod specification
// - container: Container metrics from the Metrics API
//
// It returns:
// - usageStats: A UsageStats struct containing the resource usage info
// - Any error that occurred while generating the usage stats
func GetUsage(
	ctx context.Context,
	cfg *scout.Config,
	metrics scout.ContainerMetrics,
	pod corev1.Pod,
	container metav1beta1.ContainerMetrics,
) (scout.UsageStats, error) {
	var usageStats scout.UsageStats
	usageStats.ContainerName = container.Name

	cpuUsage, err := GetRawUsage(container.Usage, "cpu")
	if err != nil {
		return usageStats, errors.Wrap(err, "failed to get raw CPU usage")
	}

	memUsage, err := GetRawUsage(container.Usage, "memory")
	if err != nil {
		return usageStats, errors.Wrap(err, "failed to get raw memory usage")
	}

	limits := metrics.Limits[container.Name]

	var storageCapacity float64
	var storageUsage float64
	if limits.Storage != nil {
		storageCapacity, storageUsage, err = GetStorageUsage(ctx, cfg, pod.Name, container.Name)
		if err != nil {
			return usageStats, errors.Wrap(err, "failed to get storage usage")
		}
	}

	usageStats.CpuCores = limits.Cpu
	usageStats.CpuUsage = scout.GetPercentage(
		cpuUsage,
		limits.Cpu.AsApproximateFloat64()*scout.ABillion,
	)

	usageStats.Memory = limits.Memory
	usageStats.MemoryUsage = scout.GetPercentage(
		memUsage,
		limits.Memory.AsApproximateFloat64(),
	)

	if limits.Storage == nil {
		storageDec := *inf.NewDec(0, 0)
		usageStats.Storage = resource.NewDecimalQuantity(storageDec, resource.Format("DecimalSI"))
	} else {
		usageStats.Storage = limits.Storage
	}

	usageStats.StorageUsage = scout.GetPercentage(
		storageUsage,
		storageCapacity,
	)

	if metrics.Limits[container.Name].Storage == nil {
		usageStats.Storage = nil
	}

	return usageStats, nil
}

// GetRawUsage returns the raw usage value for a given resource type from a Kubernetes ResourceList.
//
// It accepts:
// - usages: A Kubernetes ResourceList containing usage values
// - targetKey: The resource type to get the usage for (e.g. "cpu" or "memory")
//
// It returns:
// - The raw usage value for the target resource type
// - Any error that occurred while parsing the usage value
func GetRawUsage(usages corev1.ResourceList, targetKey string) (float64, error) {
	var usage *inf.Dec

	for key, val := range usages {
		if key.String() == targetKey {
			usage = val.AsDec().SetScale(0)
		}
	}

	toFloat, err := strconv.ParseFloat(usage.String(), 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert inf.Dec type to float")
	}

	return toFloat, nil
}

// GetPvcCapacity returns the storage capacity of a PersistentVolumeClaim mounted to a container.
//
// It accepts:
// - ctx: The context for the request
// - cfg: The scout config containing Kubernetes clientsets
// - container: The container specification
// - pod: The pod specification
//
// It returns:
// - The storage capacity of the PVC in bytes
// - Any error that occurred while fetching the PVC
//
// If no PVC is mounted to the container, nil is returned for the capacity and no error.
func GetPvcCapacity(ctx context.Context, cfg *scout.Config, container corev1.Container, pod *corev1.Pod) (*resource.Quantity, error) {
	for _, vm := range container.VolumeMounts {
		for _, v := range pod.Spec.Volumes {
			if v.Name == vm.Name && v.PersistentVolumeClaim != nil {
				pvc, err := cfg.K8sClient.
					CoreV1().
					PersistentVolumeClaims(cfg.Namespace).
					Get(
						ctx,
						v.PersistentVolumeClaim.ClaimName,
						metav1.GetOptions{},
					)
				if err != nil {
					return nil, errors.Wrapf(
						err,
						"failed to get PVC %s",
						v.PersistentVolumeClaim.ClaimName,
					)
				}
				return pvc.Status.Capacity.Storage().ToDec(), nil
			}
		}
	}
	return nil, nil
}

// GetStorageUsage returns the storage capacity and usage for a given pod and container.
//
// It accepts:
// - ctx: The context for the request
// - cfg: The scout config containing Kubernetes clientsets
// - podName: The name of the pod
// - containerName: The name of the container
//
// It returns:
// - storageCapacity: The total storage capacity for the container in bytes
// - storageUsage: The used storage for the container in bytes
// - Any error that occurred while fetching the storage usage
func GetStorageUsage(
	ctx context.Context,
	cfg *scout.Config,
	podName string,
	containerName string,
) (float64, float64, error) {
	var storageCapacity float64
	var storageUsage float64

	stateless := []string{
		"cadvisor",
		"pgsql-exporter",
		"executor",
		"dind",
		"github-proxy",
		"jaeger",
		"node-exporter",
		"otel-agent",
		"otel-collector",
		"precise-code-intel-worker",
		"redis-exporter",
		"repo-updater",
		"frontend",
		"syntect-server",
		"worker",
	}

	// if pod is stateless, return 0 for capacity and usage
	if scout.Contains(stateless, containerName) {
		return storageCapacity, storageUsage, nil
	}

	req := cfg.K8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(cfg.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   []string{"df"},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg.RestConfig, "POST", req.URL())
	if err != nil {
		return 0, 0, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines[1 : len(lines)-1] {
		fields := strings.Fields(line)

		if acceptedFileSystem(fields[0]) {
			capacityFloat, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return 0, 0, errors.Wrap(err, "could not convert string to float64")
			}

			usageFloat, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return 0, 0, errors.Wrap(err, "could not convert string to float64")
			}
			return capacityFloat, usageFloat, nil
		}
	}

	return 0, 0, nil
}

// acceptedFileSystem checks if a given file system, represented
// as a string, is an accepted system.
//
// It returns:
// - True if the file system matches the pattern '/dev/sd[a-z]$'
// - False otherwise
func acceptedFileSystem(fileSystem string) bool {
	matched, _ := regexp.MatchString(`/dev/sd[a-z]$`, fileSystem)
	return matched
}
