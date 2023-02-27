package policies

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/background"
	policiesstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	uploadSvc UploadService,
	gitserver GitserverClient,
) *Service {
	store := policiesstore.New(scopedContext("store", observationCtx), db)

	return newService(
		observationCtx,
		store,
		uploadSvc,
		gitserver,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "policies", component, parent)
}

func PolicyMatcherJobs(observationCtx *observation.Context, service *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRepositoryMatcher(
			service.store, observationCtx,
			PolicyMatcherConfigInst.Interval,
			PolicyMatcherConfigInst.ConfigurationPolicyMembershipBatchSize,
		),
	}
}
