package advise

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout"
	"github.com/sourcegraph/src-cli/internal/scout/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func K8s(
	ctx context.Context,
	k8sClient *kubernetes.Clientset,
	metricsClient *metricsv.Clientset,
	restConfig *rest.Config,
	opts ...Option,
) error {
	cfg := &scout.Config{
		Namespace:     "default",
		Pod:           "",
		Output:        "",
		RestConfig:    restConfig,
		K8sClient:     k8sClient,
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
		pod, err := kube.GetPod(cfg.Pod, pods)
		if err != nil {
			return errors.Wrap(err, "could not get pod")
		}

		err = Advise(ctx, cfg, pod)
		if err != nil {
			return errors.Wrap(err, "could not advise")
		}
		return nil
	}

	for _, pod := range pods {
		err = Advise(ctx, cfg, pod)
		if err != nil {
			return errors.Wrap(err, "could not advise")
		}
	}

	return nil
}

// Advise generates resource allocation advice for a Kubernetes pod.
// The function fetches usage metrics for each container in the pod. It then
// checks the usage percentages against thresholds to determine if more or less
// of a resource is needed. Advice is generated and either printed to the console
// or output to a file depending on the cfg.Output field.
func Advise(ctx context.Context, cfg *scout.Config, pod v1.Pod) error {
	var advice []string
	usageMetrics, err := getUsageMetrics(ctx, cfg, pod)
	if err != nil {
		return errors.Wrap(err, "could not get usage metrics")
	}

	for _, metrics := range usageMetrics {
		cpuAdvice := CheckUsage(metrics.CpuUsage, "CPU", metrics.ContainerName)
		advice = append(advice, cpuAdvice)

		memoryAdvice := CheckUsage(metrics.MemoryUsage, "memory", metrics.ContainerName)
		advice = append(advice, memoryAdvice)

		if metrics.Storage != nil {
			storageAdvice := CheckUsage(metrics.StorageUsage, "storage", metrics.ContainerName)
			advice = append(advice, storageAdvice)
		}

		if cfg.Output != "" {
			OutputToFile(ctx, cfg, pod.Name, advice)
		} else {
			for _, msg := range advice {
				fmt.Println(msg)
			}
		}
	}

	return nil
}

// getUsageMetrics generates resource usage statistics for containers in a Kubernetes pod.
func getUsageMetrics(ctx context.Context, cfg *scout.Config, pod v1.Pod) ([]scout.UsageStats, error) {
	var usages []scout.UsageStats
	var usage scout.UsageStats
	podMetrics, err := kube.GetPodMetrics(ctx, cfg, pod)
	if err != nil {
		return usages, errors.Wrap(err, "while attempting to fetch pod metrics")
	}

	containerMetrics := &scout.ContainerMetrics{
		PodName: cfg.Pod,
		Limits:  map[string]scout.Resources{},
	}

	if err = kube.AddLimits(ctx, cfg, &pod, containerMetrics); err != nil {
		return usages, errors.Wrap(err, "failed to get get container metrics")
	}

	for _, container := range podMetrics.Containers {
		usage, err = kube.GetUsage(ctx, cfg, *containerMetrics, pod, container)
		if err != nil {
			return usages, errors.Wrapf(err, "could not compile usages data for row: %s\n", container.Name)
		}
		usages = append(usages, usage)
	}

	return usages, nil
}
