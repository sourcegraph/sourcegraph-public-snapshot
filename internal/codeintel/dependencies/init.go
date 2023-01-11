package dependencies

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(observationCtx *observation.Context, db database.DB) *Service {
	return newService(scopedContext("service", observationCtx), store.New(scopedContext("store", observationCtx), db))
}

type serviceDependencies struct {
	db             database.DB
	observationCtx *observation.Context
}

// TestService creates a new dependencies service with noop observation contexts.
func TestService(db database.DB, gitserver GitserverClient) *Service {
	store := store.New(&observation.TestContext, db)

	return newService(&observation.TestContext, store)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "dependencies", component, parent)
}

func CrateSyncerJob(observationCtx *observation.Context, dependenciesSvc background.DependenciesService, gitserverClient background.GitserverClient, extSvcStore background.ExternalServiceStore) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCrateSyncer(observationCtx, dependenciesSvc, gitserverClient, extSvcStore),
	}
}
