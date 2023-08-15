package background

import (
	repomatcher "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/background/repository_matcher"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func PolicyMatcherJobs(observationCtx *observation.Context, store store.Store, config *repomatcher.Config) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		repomatcher.NewRepositoryMatcher(
			store,
			observationCtx,
			config.Interval,
			config.ConfigurationPolicyMembershipBatchSize,
		),
	}
}
