package policies

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetService creates or returns an already-initialized policies service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc UploadService,
	gitserver GitserverClient,
) *Service {
	store := store.New(db, scopedContext("store"))

	return newService(
		store,
		uploadSvc,
		gitserver,
		scopedContext("service"),
	)
}

type serviceDependencies struct {
	db        database.DB
	uploadSvc UploadService
	gitserver GitserverClient
}

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "policies", component)
}

func PolicyMatcherJobs(service *Service, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRepositoryMatcher(
			service.store, observationContext,
			PolicyMatcherConfigInst.Interval,
			PolicyMatcherConfigInst.ConfigurationPolicyMembershipBatchSize,
		),
	}
}
