package advise

import (
	"context"
	"fmt"
	"time"

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
		Warnings:      false,
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

	if cfg.Output != "" {
		fmt.Printf("writing to %s. This can take a few minutes...", cfg.Output)
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
	var advice []scout.Advice
	usageMetrics, err := getUsageMetrics(ctx, cfg, pod)
	if err != nil {
		return errors.Wrap(err, "could not get usage metrics")
	}
	for _, metrics := range usageMetrics {
		cpuAdvice := CheckUsage(metrics.CpuUsage, "CPU", metrics.ContainerName)
		if cfg.Warnings {
			advice = append(advice, cpuAdvice)
		} else if !cfg.Warnings && cpuAdvice.Kind != scout.WARNING {
			advice = append(advice, cpuAdvice)
		}

		memoryAdvice := CheckUsage(metrics.MemoryUsage, "memory", metrics.ContainerName)
		if cfg.Warnings {
			advice = append(advice, memoryAdvice)
		} else if !cfg.Warnings && memoryAdvice.Kind != scout.WARNING {
			advice = append(advice, memoryAdvice)
		}

		if metrics.Storage != nil {
			storageAdvice := CheckUsage(metrics.StorageUsage, "storage", metrics.ContainerName)
			if cfg.Warnings {
				advice = append(advice, storageAdvice)
			} else if !cfg.Warnings && storageAdvice.Kind != scout.WARNING {
				advice = append(advice, storageAdvice)
			}
		}

		if cfg.Output != "" {
			OutputToFile(ctx, cfg, pod.Name, advice)
		} else {
			fmt.Printf("%s %s: advising...\n", scout.EmojiFingerPointRight, pod.Name)
			time.Sleep(time.Millisecond * 300)
			for _, adv := range advice {
				fmt.Printf("\t%s\n", adv.Msg)
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
