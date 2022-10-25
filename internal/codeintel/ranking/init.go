package ranking

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
)

// GetService creates or returns an already-initialized ranking service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc *uploads.Service,
	gitserverClient GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		uploadSvc,
		gitserverClient,
	})

	return svc
}

type serviceDependencies struct {
	db              database.DB
	uploadsService  *uploads.Service
	gitserverClient GitserverClient
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	return newService(
		store.New(deps.db, scopedContext("store")),
		deps.uploadsService,
		deps.gitserverClient,
		symbols.DefaultClient,
		siteConfigQuerier{},
		scopedContext("service"),
	), nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component)
}
