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
		Container:     "",
		Spy:           false,
		Docker:        false,
		RestConfig:    restConfig,
		K8sClient:     k8sClient,
		DockerClient:  nil,
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

func Advise(ctx context.Context, cfg *scout.Config, pod v1.Pod) error {
	var advice []string
	usageMetrics, err := getUsageMetrics(ctx, cfg, pod)
	if err != nil {
		return errors.Wrap(err, "could not get usage metrics")
	}

	for _, metrics := range usageMetrics {
		cpuAdvice := checkUsage(metrics.CpuUsage, "CPU", metrics.ContainerName, pod.Name)
		advice = append(advice, cpuAdvice)

		memoryAdvice := checkUsage(metrics.MemoryUsage, "memory", metrics.ContainerName, pod.Name)
		advice = append(advice, memoryAdvice)

		if metrics.Storage != nil {
			storageAdvice := checkUsage(metrics.StorageUsage, "storage", metrics.ContainerName, pod.Name)
			advice = append(advice, storageAdvice)
		}

		fmt.Println(scout.EmojiFingerPointRight, pod.Name)
		for _, msg := range advice {
			fmt.Println(msg)
		}
	}

	return nil
}

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

func checkUsage(usage float64, resourceType, container, pod string) string {
	var message string

	switch {
	case usage >= 100:
		message = fmt.Sprintf(
			OVER_100,
			scout.FlashingLightEmoji,
			container,
			resourceType,
			usage,
			resourceType,
		)
	case usage >= 80 && usage < 100:
		message = fmt.Sprintf(
			OVER_80,
			scout.WarningSign,
			container,
			resourceType,
			usage,
		)
	case usage >= 40 && usage < 80:
		message = fmt.Sprintf(
			OVER_40,
			scout.SuccessEmoji,
			container,
			resourceType,
			usage,
			resourceType,
		)
	default:
		message = fmt.Sprintf(
			UNDER_40,
			scout.WarningSign,
			container,
			resourceType,
			usage,
		)
	}

	return message
}
