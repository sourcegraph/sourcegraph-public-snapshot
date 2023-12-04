package deploy

import (
	"os"
)

// Deploy type constants. Any changes here should be reflected in the DeployType type declared in client/web/src/jscontext.ts:
// https://sourcegraph.com/search?q=r:github.com/sourcegraph/sourcegraph%24+%22type+DeployType%22
const (
	Kubernetes    = "kubernetes"
	SingleDocker  = "docker-container"
	DockerCompose = "docker-compose"
	PureDocker    = "pure-docker"
	Dev           = "dev"
	Helm          = "helm"
	Kustomize     = "kustomize"
	App           = "app"
	SingleProgram = "single-program"
	K3s           = "k3s"
)

var mock string

var forceType string // force a deploy type (can be injected with `go build -ldflags "-X ..."`)

// Type tells the deployment type.
func Type() string {
	if forceType != "" {
		return forceType
	}
	if mock != "" {
		return mock
	}
	if e := os.Getenv("DEPLOY_TYPE"); e != "" {
		return e
	}
	// Default to Kubernetes cluster so that every Kubernetes
	// cluster deployment doesn't need to be configured with DEPLOY_TYPE.
	return Kubernetes
}

func Mock(val string) {
	mock = val
}

// IsDeployTypeKubernetes tells if the given deployment type is a Kubernetes
// cluster (and non-dev, not docker-compose, not pure-docker, and non-single Docker image).
func IsDeployTypeKubernetes(deployType string) bool {
	switch deployType {
	// includes older Kubernetes aliases for backwards compatibility
	case "k8s", "cluster", Kubernetes, Helm, Kustomize, K3s:
		return true
	}

	return false
}

// IsDeployTypeDockerCompose tells if the given deployment type is the Docker Compose
// deployment (and non-dev, not pure-docker, non-cluster, and non-single Docker image).
func IsDeployTypeDockerCompose(deployType string) bool {
	return deployType == DockerCompose
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

// IsDeployTypeApp tells if the given deployment is Cody App.
func IsDeployTypeApp(deployType string) bool {
	return deployType == App
}

// IsDeployTypeSingleProgram tells if the given deployment is a single program.
func IsDeployTypeSingleProgram(deployType string) bool {
	return deployType == SingleProgram
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
		IsDev(deployType) ||
		IsDeployTypeApp(deployType) ||
		IsDeployTypeSingleProgram(deployType)
}

// IsApp tells if the running deployment is a Cody App deployment.
//
// Cody App is always a single-binary, but not all single-binary deployments are
// a Cody app.
//
// In the future, all Sourcegraph deployments will be a single-binary. For example gitserver will
// be `sourcegraph --as=gitserver` or similar. Use IsSingleBinary() for code that should always
// run in a single-binary setting, and use IsApp() for code that should only run as part of the
// Sourcegraph desktop app.
func IsApp() bool {
	return Type() == App
}

// IsSingleBinary tells if the running deployment is a single-binary or not.
//
// Cody App is always a single-binary, but not all single-binary deployments are
// a Cody app.
//
// In the future, all Sourcegraph deployments will be a single-binary. For example gitserver will
// be `sourcegraph --as=gitserver` or similar. Use IsSingleBinary() for code that should always
// run in a single-binary setting, and use IsApp() for code that should only run as part of the
// Sourcegraph desktop app.
func IsSingleBinary() bool {
	return Type() == App || Type() == SingleProgram
}
