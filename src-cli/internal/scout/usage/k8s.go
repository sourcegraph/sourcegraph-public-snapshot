package usage

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout"
	"github.com/sourcegraph/src-cli/internal/scout/kube"
	"github.com/sourcegraph/src-cli/internal/scout/style"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func K8s(
	ctx context.Context,
	clientSet *kubernetes.Clientset,
	metricsClient *metricsv.Clientset,
	restConfig *restclient.Config,
	opts ...Option,
) error {
	cfg := &scout.Config{
		Namespace:     "default",
		Pod:           "",
		RestConfig:    restConfig,
		K8sClient:     clientSet,
		MetricsClient: metricsClient,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	pods, err := kube.GetPods(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "could not get list of pods")
	}

	if cfg.Pod != "" {
		return renderSinglePodUsageTable(ctx, cfg, pods)
	}

	return renderUsageTable(ctx, cfg, pods)
}

// renderSinglePodUsageStats prints resource usage statistics for a single pod.
func renderSinglePodUsageTable(ctx context.Context, cfg *scout.Config, pods []corev1.Pod) error {
	pod, err := kube.GetPod(cfg.Pod, pods)
	if err != nil {
		return errors.Wrapf(err, "could not get pod with name %s", cfg.Pod)
	}

	containerMetrics := &scout.ContainerMetrics{
		PodName: cfg.Pod,
		Limits:  map[string]scout.Resources{},
	}
	if err = kube.AddLimits(ctx, cfg, &pod, containerMetrics); err != nil {
		return errors.Wrap(err, "failed to add limits to container metrics")
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

	podMetrics, err := kube.GetPodMetrics(ctx, cfg, pod)
	if err != nil {
		return errors.Wrap(err, "while attempting to fetch pod metrics")
	}

	for _, container := range podMetrics.Containers {
		stats, err := kube.GetUsage(ctx, cfg, *containerMetrics, pod, container)
		if err != nil {
			return errors.Wrapf(err, "could not compile usage data for row: %s\n", container.Name)
		}

		row := makeRow(stats)
		rows = append(rows, row)
	}

	style.ResourceTable(columns, rows)
	return nil
}

// renderUsageTable renders a table displaying resource usage for pods.
func renderUsageTable(ctx context.Context, cfg *scout.Config, pods []corev1.Pod) error {
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
		containerMetrics := &scout.ContainerMetrics{
			PodName: pod.Name,
			Limits:  map[string]scout.Resources{},
		}

		if err := kube.AddLimits(ctx, cfg, &pod, containerMetrics); err != nil {
			return errors.Wrap(err, "failed to get get container metrics")
		}

		podMetrics, err := kube.GetPodMetrics(ctx, cfg, pod)
		if err != nil {
			return errors.Wrap(err, "while attempting to fetch pod metrics")
		}

		for _, container := range podMetrics.Containers {
			stats, err := kube.GetUsage(ctx, cfg, *containerMetrics, pod, container)
			if err != nil {
				return errors.Wrapf(err, "could not compile usage data for row %s\n", container.Name)
			}

			row := makeRow(stats)
			rows = append(rows, row)
		}
	}

	style.UsageTable(columns, rows)
	return nil
}

// makeRow generates a table row containing resource usage data for a container.
// It returns:
// - A table.Row containing the resource usage information
// - An error if there was an issue generating the row
func makeRow(usageStats scout.UsageStats) table.Row {
	if usageStats.Storage == nil {
		return table.Row{
			usageStats.ContainerName,
			usageStats.CpuCores.String(),
			fmt.Sprintf("%.2f%%", usageStats.CpuUsage),
			usageStats.Memory.String(),
			fmt.Sprintf("%.2f%%", usageStats.MemoryUsage),
			"-",
			"-",
		}
	}

	return table.Row{
		usageStats.ContainerName,
		usageStats.CpuCores.String(),
		fmt.Sprintf("%.2f%%", usageStats.CpuUsage),
		usageStats.Memory.String(),
		fmt.Sprintf("%.2f%%", usageStats.MemoryUsage),
		usageStats.Storage.String(),
		fmt.Sprintf("%.2f%%", usageStats.StorageUsage),
	}
}
