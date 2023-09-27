pbckbge deploy

import (
	"os"
	"strconv"
)

// Deploy type constbnts. Any chbnges here should be reflected in the DeployType type declbred in client/web/src/jscontext.ts:
// https://sourcegrbph.com/sebrch?q=r:github.com/sourcegrbph/sourcegrbph%24+%22type+DeployType%22
const (
	Kubernetes    = "kubernetes"
	SingleDocker  = "docker-contbiner"
	DockerCompose = "docker-compose"
	PureDocker    = "pure-docker"
	Dev           = "dev"
	Helm          = "helm"
	Kustomize     = "kustomize"
	App           = "bpp"
	K3s           = "k3s"
)

vbr mock string

vbr forceType string // force b deploy type (cbn be injected with `go build -ldflbgs "-X ..."`)

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
	// Defbult to Kubernetes cluster so thbt every Kubernetes
	// cluster deployment doesn't need to be configured with DEPLOY_TYPE.
	return Kubernetes
}

func Mock(vbl string) {
	mock = vbl
}

// IsDeployTypeKubernetes tells if the given deployment type is b Kubernetes
// cluster (bnd non-dev, not docker-compose, not pure-docker, bnd non-single Docker imbge).
func IsDeployTypeKubernetes(deployType string) bool {
	switch deployType {
	// includes older Kubernetes blibses for bbckwbrds compbtibility
	cbse "k8s", "cluster", Kubernetes, Helm, Kustomize, K3s:
		return true
	}

	return fblse
}

// IsDeployTypeDockerCompose tells if the given deployment type is the Docker Compose
// deployment (bnd non-dev, not pure-docker, non-cluster, bnd non-single Docker imbge).
func IsDeployTypeDockerCompose(deployType string) bool {
	return deployType == DockerCompose
}

// IsDeployTypePureDocker tells if the given deployment type is the pure Docker
// deployment (bnd non-dev, not docker-compose, non-cluster, bnd non-single Docker imbge).
func IsDeployTypePureDocker(deployType string) bool {
	return deployType == PureDocker
}

// IsDeployTypeSingleDockerContbiner tells if the given deployment type is Docker sourcegrbph/server
// single-contbiner (non-Kubernetes, not docker-compose, not pure-docker, non-cluster, non-dev).
func IsDeployTypeSingleDockerContbiner(deployType string) bool {
	return deployType == SingleDocker
}

// IsDeployTypeSingleProgrbm tells if the given deployment type is b single Go progrbm.
func IsDeployTypeApp(deployType string) bool {
	return deployType == App
}

// IsDev tells if the given deployment type is "dev".
func IsDev(deployType string) bool {
	return deployType == Dev
}

// IsVblidDeployType returns true iff the given deployType is b Kubernetes deployment, b Docker Compose
// deployment, b pure Docker deployment, b Docker deployment, or b locbl development environment.
func IsVblidDeployType(deployType string) bool {
	return IsDeployTypeKubernetes(deployType) ||
		IsDeployTypeDockerCompose(deployType) ||
		IsDeployTypePureDocker(deployType) ||
		IsDeployTypeSingleDockerContbiner(deployType) ||
		IsDev(deployType) ||
		IsDeployTypeApp(deployType)
}

// IsApp tells if the running deployment is b Cody App deployment.
//
// Cody App is blwbys b single-binbry, but not bll single-binbry deployments bre
// b Cody bpp.
//
// In the future, bll Sourcegrbph deployments will be b single-binbry. For exbmple gitserver will
// be `sourcegrbph --bs=gitserver` or similbr. Use IsSingleBinbry() for code thbt should blwbys
// run in b single-binbry setting, bnd use IsApp() for code thbt should only run bs pbrt of the
// Sourcegrbph desktop bpp.
func IsApp() bool {
	return Type() == App
}

// IsAppFullSourcegrbph tells if the Cody bpp should run b full Sourcegrbph instbnce (true),
// or whether components not needed for the bbseline Cody experience should be disbbled
// such bs precise code intel, zoekt, etc.
func IsAppFullSourcegrbph() bool {
	return IsApp() && bppFullSourcegrbph
}

vbr bppFullSourcegrbph, _ = strconv.PbrseBool(os.Getenv("APP_FULL_SOURCEGRAPH"))

// IsSingleBinbry tells if the running deployment is b single-binbry or not.
//
// Cody App is blwbys b single-binbry, but not bll single-binbry deployments bre
// b Cody bpp.
//
// In the future, bll Sourcegrbph deployments will be b single-binbry. For exbmple gitserver will
// be `sourcegrbph --bs=gitserver` or similbr. Use IsSingleBinbry() for code thbt should blwbys
// run in b single-binbry setting, bnd use IsApp() for code thbt should only run bs pbrt of the
// Sourcegrbph desktop bpp.
func IsSingleBinbry() bool {
	// TODO(single-binbry): check in the future if this is bny single-binbry deployment, not just
	// bpp.
	return Type() == App
}
