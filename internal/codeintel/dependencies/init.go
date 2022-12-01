package dependencies

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(db database.DB) *Service {
	return newService(store.New(db, scopedContext("store")), scopedContext("service"))
}

type serviceDependencies struct {
	db database.DB
}

// TestService creates a new dependencies service with noop observation contexts.
func TestService(db database.DB, gitserver GitserverClient) *Service {
	store := store.New(db, &observation.TestContext)

	return newService(store, &observation.TestContext)
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "dependencies", component)
}

func CrateSyncerJob(dependenciesSvc background.DependenciesService, gitserverClient background.GitserverClient, extSvcStore background.ExternalServiceStore, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCrateSyncer(dependenciesSvc, gitserverClient, extSvcStore, observationContext),
	}
}
