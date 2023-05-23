package usage

import (
	"github.com/docker/docker/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Option = func(config *Config)
type Config struct {
	namespace     string
	pod           string
	container     string
	spy           bool
	docker        bool
	restConfig    *rest.Config
	k8sClient     *kubernetes.Clientset
	dockerClient  *client.Client
	metricsClient *metricsv.Clientset
}

func WithNamespace(namespace string) Option {
	return func(config *Config) {
		config.namespace = namespace
	}
}

func WithPod(podname string) Option {
	return func(config *Config) {
		config.pod = podname
	}
}

func WithContainer(containerName string) Option {
	return func(config *Config) {
		config.container = containerName
	}
}

func WithSpy(spy bool) Option {
	return func(config *Config) {
		config.spy = true
	}
}

// contains checks if a string slice contains a given value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// getPercentage calculates the percentage of x in relation to y.
func getPercentage(x, y float64) float64 {
	if x == 0 {
		return 0
	}

	if y == 0 {
		return 0
	}

	return x * 100 / y
}
