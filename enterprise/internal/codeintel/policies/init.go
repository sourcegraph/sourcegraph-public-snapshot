package policies

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
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

	svc := newService(store, deps.uploadSvc, deps.gitserver, scopedContext("service"))

	return svc, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "policies", component)
}

func PolicyMatcherJobs(service background.PolicyService, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRepositoryMatcher(
			service, observationContext,
			PolicyMatcherConfigInst.Interval,
			PolicyMatcherConfigInst.ConfigurationPolicyMembershipBatchSize,
		),
	}
}
