package policies

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	db database.DB,
	uploadSvc UploadService,
	gitserver GitserverClient,
	observationContext *observation.Context,
) *Service {
	store := store.New(db, scopedContext("store", observationContext))

	return newService(
		store,
		uploadSvc,
		gitserver,
		observationContext,
	)
}

type serviceDependencies struct {
	db                 database.DB
	uploadSvc          UploadService
	gitserver          GitserverClient
	observationContext *observation.Context
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "policies", component, parent)
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
