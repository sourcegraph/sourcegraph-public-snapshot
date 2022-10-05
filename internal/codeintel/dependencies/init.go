package dependencies

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized dependencies service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	gitserver GitserverClient,
) *Service {
	svcOnce.Do(func() {
		store := store.New(db, scopedContext("store"))
		externalServiceStore := db.ExternalServices()

		svc = newService(
			store,
			gitserver,
			externalServiceStore,
			scopedContext("service"),
		)
	})

	return svc
}

// TestService creates a new dependencies service with noop observation contexts.
func TestService(
	db database.DB,
	gitserver GitserverClient,
) *Service {
	store := store.New(db, &observation.TestContext)
	externalServiceStore := db.ExternalServices()

	return newService(
		store,
		gitserver,
		externalServiceStore,
		&observation.TestContext,
	)
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "dependencies", component)
}
