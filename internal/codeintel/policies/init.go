package policies

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetService creates or returns an already-initialized policies service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc UploadService,
	gitserver GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		uploadSvc,
		gitserver,
	})

	return svc
}

type serviceDependencies struct {
	db        database.DB
	uploadSvc UploadService
	gitserver GitserverClient
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	backgroundJob := background.New(scopedContext("background"))

	svc := newService(store, deps.uploadSvc, deps.gitserver, backgroundJob, scopedContext("service"))

	return svc, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "policies", component)
}
