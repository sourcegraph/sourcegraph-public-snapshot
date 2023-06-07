package scout

import (
	"github.com/docker/docker/client"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Config struct {
	Namespace     string
	Pod           string
	Container     string
	Output        string
	Spy           bool
	Docker        bool
	RestConfig    *rest.Config
	K8sClient     *kubernetes.Clientset
	DockerClient  *client.Client
	MetricsClient *metricsv.Clientset
}

type ContainerMetrics struct {
	PodName string
	Limits  map[string]Resources
}

type Resources struct {
	Cpu     *resource.Quantity
	Memory  *resource.Quantity
	Storage *resource.Quantity
}

type UsageStats struct {
	ContainerName string
	CpuCores      *resource.Quantity
	Memory        *resource.Quantity
	Storage       *resource.Quantity
	CpuUsage      float64
	MemoryUsage   float64
	StorageUsage  float64
}
