package config

import (
	"os"
)

// IsKubernetes returns true if the executor is running in a Kubernetes cluster.
func IsKubernetes() bool {
	_, isKubernetes := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return isKubernetes
}
