package uploads

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized uploads service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db, codeIntelDB database.DB,
	gsc GitserverClient,
) *Service {
	svcOnce.Do(func() {
		store := store.New(db, scopedContext("store"))
		repoStore := backend.NewRepos(scopedContext("repos").Logger, db)
		lsifStore := lsifstore.New(codeIntelDB, scopedContext("lsifstore"))
		locker := locker.NewWith(db, "codeintel")

		svc = newService(
			store,
			repoStore,
			lsifStore,
			gsc,
			locker,
			scopedContext("service"),
		)
	})

	return svc
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "uploads", component)
}
