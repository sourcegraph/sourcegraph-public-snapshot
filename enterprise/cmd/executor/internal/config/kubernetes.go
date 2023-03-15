package config

import (
	"os"
	"sync"
)

var kubernetesOnce sync.Once
var isKubernetes = false

func IsKubernetes() bool {
	kubernetesOnce.Do(func() {
		_, isKubernetes = os.LookupEnv("KUBERNETES_SERVICE_HOST")
	})
	return isKubernetes
}
