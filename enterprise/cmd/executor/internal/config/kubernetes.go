package config

import (
	"os"
)

// IsKubernetes returns true if the executor is running in a Kubernetes cluster.
func IsKubernetes() bool {
	_, hasKubernetesHost := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	_, hasKubernetesPort := os.LookupEnv("KUBERNETES_SERVICE_PORT")
	return hasKubernetesHost && hasKubernetesPort
}
