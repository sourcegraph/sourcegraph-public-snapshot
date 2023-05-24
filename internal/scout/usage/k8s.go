package usage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout/style"

	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metav1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ContainerMetrics struct {
	podName string
	limits  map[string]Resources
}

type Resources struct {
	cpu     *resource.Quantity
	memory  *resource.Quantity
	storage *resource.Quantity
}

type UsageStats struct {
	containerName string
	cpuCores      *resource.Quantity
	memory        *resource.Quantity
	storage       *resource.Quantity
	cpuUsage      float64
	memoryUsage   float64
	storageUsage  float64
}

const (
	ABillion float64 = 1000000000
)

func K8s(
	ctx context.Context,
	clientSet *kubernetes.Clientset,
	metricsClient *metricsv.Clientset,
	restConfig *restclient.Config,
	opts ...Option,
) error {
	cfg := &Config{
		namespace:     "default",
		docker:        false,
		pod:           "",
		container:     "",
		spy:           false,
		restConfig:    restConfig,
		k8sClient:     clientSet,
		dockerClient:  nil,
		metricsClient: metricsClient,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	pods, err := getPods(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "could not get list of pods")
	}

	if cfg.pod != "" {
		return renderSinglePodUsageTable(ctx, cfg, pods)
	}

	return renderUsageTable(ctx, cfg, pods)
}

// renderSinglePodUsageStats prints resource usage statistics for a single pod.
func renderSinglePodUsageTable(ctx context.Context, cfg *Config, pods []corev1.Pod) error {
	pod, err := getPod(cfg, pods)
	if err != nil {
		return errors.Wrapf(err, "could not get pod with name %s", cfg.pod)
	}

	podMetrics, err := getPodMetrics(ctx, cfg, pod)
	if err != nil {
		return errors.Wrap(err, "while attempting to fetch pod metrics")
	}

	containerMetrics := &ContainerMetrics{cfg.pod, map[string]Resources{}}
	if err = getLimits(ctx, cfg, &pod, containerMetrics); err != nil {
		return errors.Wrap(err, "failed to get get container metrics")
	}

	columns := []table.Column{
		{Title: "Container", Width: 20},
		{Title: "Cores", Width: 10},
		{Title: "Usage(%)", Width: 10},
		{Title: "Memory", Width: 10},
		{Title: "Usage(%)", Width: 10},
		{Title: "Storage", Width: 10},
		{Title: "Usage(%)", Width: 10},
	}
	var rows []table.Row

	for _, container := range podMetrics.Containers {
		stats, err := getContainerUsageStats(ctx, cfg, *containerMetrics, pod, container)
		if err != nil {
			return errors.Wrapf(err, "could not compile usage data for row: %s\n", container.Name)
		}

		row := makeRow(stats)
		rows = append(rows, row)
	}

	style.ResourceTable(columns, rows)
	return nil
}

// getPod returns a Pod object with the given name.
//
// If a Pod with the given name exists in pods, it is returned.
// Otherwise, an error is returned indicating no Pod with that name exists.
func getPod(cfg *Config, pods []corev1.Pod) (corev1.Pod, error) {
	for _, p := range pods {
		if p.Name == cfg.pod {
			return p, nil
		}
	}
	return corev1.Pod{}, errors.New("no pod with this name exists in this namespace")
}

// getPodMetrics retrieves metrics for a given pod.
//
// It returns:
// - podMetrics: The PodMetrics object containing metrics for the pod
// - err: Any error that occurred while getting the pod metrics
func getPodMetrics(ctx context.Context, cfg *Config, pod corev1.Pod) (*metav1beta1.PodMetrics, error) {
	podMetrics, err := cfg.metricsClient.
		MetricsV1beta1().
		PodMetricses(cfg.namespace).
		Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pod metrics")
	}

	return podMetrics, nil
}

// getLimits extracts resource limits for containers in a pod and stores
// them in a ContainerMetrics struct.
//
// It populates the containerMetrics struct with:
// - The name of each container
// - The CPU, memory, ephemeral storage, and storage resource limits for each container
// - A print method to print the resource limits for each container
func getLimits(ctx context.Context, cfg *Config, pod *corev1.Pod, containerMetrics *ContainerMetrics) error {
	for _, container := range pod.Spec.Containers {
		containerName := container.Name
		capacity, err := getPvcCapacity(ctx, cfg, container, pod)
		if err != nil {
			return errors.Wrap(err, "while getting storage capacity of PV")
		}

		rsrcs := Resources{
			cpu:     container.Resources.Limits.Cpu().ToDec(),
			memory:  container.Resources.Limits.Memory().ToDec(),
			storage: capacity,
		}
		containerMetrics.limits[containerName] = rsrcs
	}

	return nil
}

// renderUsageTable renders a table displaying resource usage for pods.

// It returns:
// - Any error that occurred while rendering the table
func renderUsageTable(ctx context.Context, cfg *Config, pods []corev1.Pod) error {
	columns := []table.Column{
		{Title: "Container", Width: 20},
		{Title: "Cores", Width: 10},
		{Title: "Usage(%)", Width: 10},
		{Title: "Memory", Width: 10},
		{Title: "Usage(%)", Width: 10},
		{Title: "Storage", Width: 10},
		{Title: "Usage(%)", Width: 10},
	}
	var rows []table.Row

	for _, pod := range pods {
		containerMetrics := &ContainerMetrics{pod.Name, map[string]Resources{}}
		podMetrics, err := getPodMetrics(ctx, cfg, pod)
		if err != nil {
			return errors.Wrap(err, "while attempting to fetch pod metrics")
		}

		if err = getLimits(ctx, cfg, &pod, containerMetrics); err != nil {
			return errors.Wrap(err, "failed to get get container metrics")
		}

		for _, container := range podMetrics.Containers {
			stats, err := getContainerUsageStats(ctx, cfg, *containerMetrics, pod, container)
			if err != nil {
				return errors.Wrapf(err, "could not compile usage data for row %s\n", container.Name)
			}

			row := makeRow(stats)
			rows = append(rows, row)
		}
	}

	style.ResourceTable(columns, rows)
	return nil
}

// makeRow generates a table row containing resource usage data for a container.
// It returns:
// - A table.Row containing the resource usage information
// - An error if there was an issue generating the row
func makeRow(usageStats UsageStats) table.Row {
	if usageStats.storage == nil {
		return table.Row{
			usageStats.containerName,
			usageStats.cpuCores.String(),
			fmt.Sprintf("%.2f%%", usageStats.cpuUsage),
			usageStats.memory.String(),
			fmt.Sprintf("%.2f%%", usageStats.memoryUsage),
			"-",
			"-",
		}
	}

	return table.Row{
		usageStats.containerName,
		usageStats.cpuCores.String(),
		fmt.Sprintf("%.2f%%", usageStats.cpuUsage),
		usageStats.memory.String(),
		fmt.Sprintf("%.2f%%", usageStats.memoryUsage),
		usageStats.storage.String(),
		fmt.Sprintf("%.2f%%", usageStats.storageUsage),
	}
}

// getPods returns a list of pods in the given namespace.
// It returns:
// - []v1.Pod: A list of pods in the given namespace
// - error: Any error that occurred while listing the pods
//
// If no pods are found in the given namespace, it will print an error message and exit.
func getPods(ctx context.Context, cfg *Config) ([]corev1.Pod, error) {
	podInterface := cfg.k8sClient.CoreV1().Pods(cfg.namespace)
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

// makeUsageRow generates a table row containing resource usage data for a container.
//
// It returns:
// - A table.Row containing the resource usage information
// - An error if there was an issue generating the row
func getContainerUsageStats(
	ctx context.Context,
	cfg *Config,
	metrics ContainerMetrics,
	pod corev1.Pod,
	container metav1beta1.ContainerMetrics,
) (UsageStats, error) {
	var usageStats UsageStats
	usageStats.containerName = container.Name

	cpuUsage, err := getRawUsage(container.Usage, "cpu")
	if err != nil {
		return UsageStats{}, errors.Wrap(err, "failed to get raw CPU usage")
	}

	memUsage, err := getRawUsage(container.Usage, "memory")
	if err != nil {
		return UsageStats{}, errors.Wrap(err, "failed to get raw memory usage")
	}

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

	if contains(stateless, container.Name) {
		storageUsage = 0
		storageCapacity = 0
	} else {
		storageCapacity, storageUsage, err = getStorageUsage(ctx, cfg, pod.Name, container.Name)
		if err != nil {
			return UsageStats{}, errors.Wrap(err, "failed to get storage usage")
		}
	}

	limits := metrics.limits[container.Name]

	usageStats.cpuCores = limits.cpu
	usageStats.cpuUsage = getPercentage(
		cpuUsage,
		limits.cpu.AsApproximateFloat64()*ABillion,
	)

	usageStats.memory = limits.memory
	usageStats.memoryUsage = getPercentage(
		memUsage,
		limits.memory.AsApproximateFloat64(),
	)

	if limits.storage == nil {
		storageDec := *inf.NewDec(0, 0)
		usageStats.storage = resource.NewDecimalQuantity(storageDec, resource.Format("DecimalSI"))
	} else {
		usageStats.storage = limits.storage
	}

	usageStats.storageUsage = getPercentage(
		storageUsage,
		storageCapacity,
	)

	if metrics.limits[container.Name].storage == nil {
		usageStats.storage = nil
	}

	return usageStats, nil
}

// getRawUsage converts a Kubernetes ResourceList (map[ResourceName]Quantity)
// into a raw float64 usage value for a given resource.
//
// It returns:
// - The raw float64 usage value for the target resource
// - Any error that occurred during conversion
func getRawUsage(usages corev1.ResourceList, targetKey string) (float64, error) {
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

// getPvcCapacity retrieves the storage capacity of a PersistentVolumeClaim
// mounted as a volume by a container.
//
// It returns:
// - The capacity Quantity of the PVC if a matching PVC mount is found
// - nil if no PVC mount is found
// - Any error that occurred while getting the PVC
func getPvcCapacity(ctx context.Context, cfg *Config, container corev1.Container, pod *corev1.Pod) (*resource.Quantity, error) {
	for _, volumeMount := range container.VolumeMounts {
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == volumeMount.Name && volume.PersistentVolumeClaim != nil {
				pvc, err := cfg.k8sClient.
					CoreV1().
					PersistentVolumeClaims(cfg.namespace).
					Get(
						ctx,
						volume.PersistentVolumeClaim.ClaimName,
						metav1.GetOptions{},
					)
				if err != nil {
					return nil, errors.Wrapf(
						err,
						"failed to get PVC %s",
						volume.PersistentVolumeClaim.ClaimName,
					)
				}
				return pvc.Status.Capacity.Storage().ToDec(), nil
			}
		}
	}
	return nil, nil
}

// getStorageUsage executes the df -h command in a container and parses the
// output to get the storage usage percentage for ephemeral storage volumes.
//
// It returns:
// - The storage usage percentage for storage volumes
// - "-" if no storage volumes are found
// - Any error that occurred while executing the df -h command or parsing the output
func getStorageUsage(ctx context.Context, cfg *Config, podName, containerName string) (float64, float64, error) {
	req := cfg.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(cfg.namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   []string{"df"},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg.restConfig, "POST", req.URL())
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
