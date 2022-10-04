package policies

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized policies service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc UploadService,
	gitserver GitserverClient,
) *Service {
	svcOnce.Do(func() {
		store := store.New(db, scopedContext("store"))

		svc = newService(
			store,
			uploadSvc,
			gitserver,
			scopedContext("service"),
		)
	})

	return svc
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "policies", component)
}
