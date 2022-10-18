package dependencies

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetService creates or returns an already-initialized dependencies service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(db database.DB, gitserver GitserverClient) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		gitserver,
	})

	return svc
}

type serviceDependencies struct {
	db        database.DB
	gitserver GitserverClient
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	externalServiceStore := deps.db.ExternalServices()
	backgroundJob := background.New(deps.gitserver, externalServiceStore, scopedContext("background"))

	svc := newService(store, backgroundJob, scopedContext("service"))
	backgroundJob.SetDependenciesService(svc)

	return svc, nil
})

// TestService creates a new dependencies service with noop observation contexts.
func TestService(db database.DB, gitserver GitserverClient) *Service {
	store := store.New(db, &observation.TestContext)

	return newService(store, nil, &observation.TestContext)
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "dependencies", component)
}
