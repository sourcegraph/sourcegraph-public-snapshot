package background

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type BackgroundJob interface {
	NewCrateSyncer() goroutine.BackgroundRoutine
	SetDependenciesService(dependenciesSvc DependenciesService)
}

type backgroundJob struct {
	dependenciesSvc DependenciesService
	gitClient       GitserverClient
	extSvcStore     ExternalServiceStore
	operations      *operations
}

func New(gitClient GitserverClient, extSvcStore ExternalServiceStore, observationContext *observation.Context) BackgroundJob {
	return &backgroundJob{
		gitClient:   gitClient,
		extSvcStore: extSvcStore,
		operations:  newOperations(observationContext),
	}
}

func (b *backgroundJob) SetDependenciesService(dependenciesSvc DependenciesService) {
	b.dependenciesSvc = dependenciesSvc
}
