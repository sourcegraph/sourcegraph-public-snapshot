package dependencies

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(db database.DB, obsevationContext *observation.Context) *Service {
	return newService(store.New(db, scopedContext("store", obsevationContext)), scopedContext("service", obsevationContext))
}

type serviceDependencies struct {
	db                 database.DB
	observationContext *observation.Context
}

// TestService creates a new dependencies service with noop observation contexts.
func TestService(db database.DB, gitserver GitserverClient) *Service {
	store := store.New(db, &observation.TestContext)

	return newService(store, &observation.TestContext)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "dependencies", component, parent)
}

func CrateSyncerJob(dependenciesSvc background.DependenciesService, gitserverClient background.GitserverClient, extSvcStore background.ExternalServiceStore, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCrateSyncer(dependenciesSvc, gitserverClient, extSvcStore, observationContext),
	}
}
