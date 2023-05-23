package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout/style"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Option = func(config *Config)

type Config struct {
	namespace    string
	docker       bool
	k8sClient    *kubernetes.Clientset
	dockerClient *client.Client
}

func WithNamespace(namespace string) Option {
	return func(config *Config) {
		config.namespace = namespace
	}
}

// K8s prints the CPU and memory resource limits and requests for all pods in the given namespace.
func K8s(ctx context.Context, clientSet *kubernetes.Clientset, restConfig *rest.Config, opts ...Option) error {
	cfg := &Config{
		namespace:    "default",
		docker:       false,
		k8sClient:    clientSet,
		dockerClient: nil,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return listPodResources(ctx, cfg)
}

func listPodResources(ctx context.Context, cfg *Config) error {
	podInterface := cfg.k8sClient.CoreV1().Pods(cfg.namespace)
	podList, err := podInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing pods: ")
	}

	columns := []table.Column{
		{Title: "CONTAINER", Width: 20},
		{Title: "CPU LIMITS", Width: 10},
		{Title: "CPU REQUESTS", Width: 12},
		{Title: "MEM LIMITS", Width: 10},
		{Title: "MEM REQUESTS", Width: 12},
		{Title: "CAPACITY", Width: 8},
	}

	if len(podList.Items) == 0 {
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		fmt.Println(msg.Render(`
        No pods exist in this namespace. 
        Did you mean to use the --namespace flag?
        
        If you are attemptying to check 
        resources for a docker deployment, you 
        must use the --docker flag.
        See --help for more info.
        `))
		os.Exit(1)
	}

	var rows []table.Row
	for _, pod := range podList.Items {
		if pod.GetNamespace() == cfg.namespace {
			for _, container := range pod.Spec.Containers {
				cpuLimits := container.Resources.Limits.Cpu()
				cpuRequests := container.Resources.Requests.Cpu()
				memLimits := container.Resources.Limits.Memory()
				memRequests := container.Resources.Requests.Memory()

				capacity, err := getPVCCapacity(ctx, cfg, container, pod)
				if err != nil {
					return err
				}

				if capacity == "" {
					capacity = "N/A"
				}

				row := table.Row{
					container.Name,
					cpuLimits.String(),
					cpuRequests.String(),
					memLimits.String(),
					memRequests.String(),
					capacity,
				}
				rows = append(rows, row)
			}
		}
	}

	style.ResourceTable(columns, rows)
	return nil
}

func getPVCCapacity(ctx context.Context, cfg *Config, container v1.Container, pod v1.Pod) (string, error) {
	for _, volumeMount := range container.VolumeMounts {
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == volumeMount.Name && volume.PersistentVolumeClaim != nil {
				pvc, err := cfg.k8sClient.CoreV1().PersistentVolumeClaims(cfg.namespace).Get(
					ctx,
					volume.PersistentVolumeClaim.ClaimName,
					metav1.GetOptions{},
				)
				if err != nil {
					return "", errors.Wrapf(err, "error getting PVC %s", volume.PersistentVolumeClaim.ClaimName)
				}
				return pvc.Status.Capacity.Storage().String(), nil
			}
		}
	}
	return "", nil
}

// Docker prints the CPU and memory resource limits and requests for running Docker containers.
func Docker(ctx context.Context, dockerClient client.Client) error {
	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return fmt.Errorf("error listing docker containers: %v", err)
	}

	if len(containers) == 0 {
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		fmt.Println(msg.Render(`
        There are no containers, or the Docker Daemon is not running
        `))
		os.Exit(1)
	}

	columns := []table.Column{
		{Title: "Container", Width: 20},
		{Title: "CPU Cores", Width: 15},
		{Title: "CPU Shares", Width: 15},
		{Title: "Mem Limits", Width: 15},
		{Title: "Mem Reservations", Width: 17},
	}

	var rows []table.Row

	for _, container := range containers {
		containerInfo, err := dockerClient.ContainerInspect(ctx, container.ID)
		if err != nil {
			return fmt.Errorf("error inspecting container %s: %v", container.ID, err)
		}

		row, err := getResourceInfo(&containerInfo, rows)
		if err != nil {
			return errors.Wrap(err, "error while getting resource info from container: ")
		}

		rows = append(rows, row)
	}

	style.ResourceTable(columns, rows)
	return nil
}

func getResourceInfo(container *types.ContainerJSON, rows []table.Row) (table.Row, error) {
	cpuCores := container.HostConfig.NanoCPUs
	cpuShares := container.HostConfig.CPUShares
	memLimits := container.HostConfig.Memory
	memReservations := container.HostConfig.MemoryReservation

	reqUnit, reqVal, err := getMemUnits(memReservations)
	if err != nil {
		return table.Row{}, errors.Wrap(err, "error while getting request units")
	}

	limUnit, limVal, err := getMemUnits(memLimits)
	if err != nil {
		return table.Row{}, errors.Wrap(err, "error while getting limit units")
	}

	row := table.Row{
		container.Name,
		fmt.Sprintf("%v", float64(cpuCores/1e9)),
		fmt.Sprint(cpuShares),
		fmt.Sprintf("%d%s", limVal, limUnit),
		fmt.Sprintf("%d%s", reqVal, reqUnit),
	}

	return row, nil
}

// getMemUnits converts a byte value to the appropriate memory unit.
func getMemUnits(valToConvert int64) (string, int64, error) {
	if valToConvert < 0 {
		return "", valToConvert, fmt.Errorf("invalid memory value: %d", valToConvert)
	}

	var memUnit string
	switch {
	case valToConvert < 1000000:
		memUnit = "KB"
	case valToConvert < 1000000000:
		memUnit = "MB"
		valToConvert = valToConvert / 1000000
	default:
		memUnit = "GB"
		valToConvert = valToConvert / 1000000000
	}

	return memUnit, valToConvert, nil
}
