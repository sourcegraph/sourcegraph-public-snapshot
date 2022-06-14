package deploy

import "os"

// Deploy type constants. Changes here should be reflected in several other places. Mirror
// the changes from this PR to make sure these places are consistent:
// https://github.com/sourcegraph/sourcegraph/pull/32407/files
const (
	Kubernetes           = "kubernetes"
	SingleDocker         = "docker-container"
	DockerCompose        = "docker-compose"
	ManagedDockerCompose = "managed-docker-compose"
	PureDocker           = "pure-docker"
	Dev                  = "dev"
	Helm                 = "helm"
)

// Type tells the deployment type.
func Type() string {
	if e := os.Getenv("DEPLOY_TYPE"); e != "" {
		return e
	}
	// Default to Kubernetes cluster so that every Kubernetes
	// cluster deployment doesn't need to be configured with DEPLOY_TYPE.
	return Kubernetes
}

// IsDeployTypeKubernetes tells if the given deployment type is a Kubernetes
// cluster (and non-dev, not docker-compose, not pure-docker, and non-single Docker image).
func IsDeployTypeKubernetes(deployType string) bool {
	switch deployType {
	// includes older Kubernetes aliases for backwards compatibility
	case "k8s", "cluster", Kubernetes, Helm:
		return true
	}

	return false
}

// IsDeployTypeDockerCompose tells if the given deployment type is the Docker Compose
// deployment (and non-dev, not pure-docker, non-cluster, and non-single Docker image).
func IsDeployTypeDockerCompose(deployType string) bool {
	switch deployType {
	case DockerCompose, ManagedDockerCompose:
		return true
	}

	return false
}

// IsDeployTypePureDocker tells if the given deployment type is the pure Docker
// deployment (and non-dev, not docker-compose, non-cluster, and non-single Docker image).
func IsDeployTypePureDocker(deployType string) bool {
	return deployType == PureDocker
}

// IsDeployTypeSingleDockerContainer tells if the given deployment type is Docker sourcegraph/server
// single-container (non-Kubernetes, not docker-compose, not pure-docker, non-cluster, non-dev).
func IsDeployTypeSingleDockerContainer(deployType string) bool {
	return deployType == SingleDocker
}

// IsDev tells if the given deployment type is "dev".
func IsDev(deployType string) bool {
	return deployType == Dev
}

// IsValidDeployType returns true iff the given deployType is a Kubernetes deployment, a Docker Compose
// deployment, a pure Docker deployment, a Docker deployment, or a local development environment.
func IsValidDeployType(deployType string) bool {
	return IsDeployTypeKubernetes(deployType) ||
		IsDeployTypeDockerCompose(deployType) ||
		IsDeployTypePureDocker(deployType) ||
		IsDeployTypeSingleDockerContainer(deployType) ||
		IsDev(deployType)
}
