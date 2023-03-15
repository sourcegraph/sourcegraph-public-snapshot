package config

import (
	"os"
	"sync"
)

var kubernetesOnce sync.Once
var isKubernetes = false

// IsKubernetes returns true if the executor is running in a Kubernetes cluster.
func IsKubernetes() bool {
	kubernetesOnce.Do(func() {
		_, isKubernetes = os.LookupEnv("KUBERNETES_SERVICE_HOST")
	})
	return isKubernetes
}
