package dependencies

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/background"
	dependenciesstore "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(observationCtx *observation.Context, db database.DB) *Service {
	return newService(scopedContext("service", observationCtx), dependenciesstore.New(scopedContext("store", observationCtx), db))
}

// TestService creates a new dependencies service with noop observation contexts.
func TestService(db database.DB) *Service {
	store := dependenciesstore.New(&observation.TestContext, db)

	return newService(&observation.TestContext, store)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "dependencies", component, parent)
}

func CrateSyncerJob(
	observationCtx *observation.Context,
	autoindexingSvc background.AutoIndexingService,
	dependenciesSvc background.DependenciesService,
	gitserverClient gitserver.Client,
	extSvcStore background.ExternalServiceStore,
) goroutine.CombinedRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCrateSyncer(observationCtx, autoindexingSvc, dependenciesSvc, gitserverClient, extSvcStore),
	}
}

func PackageFiltersJob(
	obsctx *observation.Context,
	db database.DB,
) goroutine.CombinedRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewPackagesFilterApplicator(obsctx, db),
	}
}
