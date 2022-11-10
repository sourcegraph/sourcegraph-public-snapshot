package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewRepositoryMatcher(policiesService PolicyService, observationContext *observation.Context, interval time.Duration, configurationPolicyMembershipBatchSize int) goroutine.BackgroundRoutine {
	metrics := newMetrics(observationContext)
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return handleRepositoryMatcherBatch(ctx, policiesService, configurationPolicyMembershipBatchSize, metrics)
	}))
}

func handleRepositoryMatcherBatch(ctx context.Context, service PolicyService, batchSize int, metrics *matcherMetrics) error {
	policies, err := service.SelectPoliciesForRepositoryMembershipUpdate(ctx, batchSize)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		var patterns []string
		if policy.RepositoryPatterns != nil {
			patterns = *policy.RepositoryPatterns
		}

		var repositoryMatchLimit *int
		if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
			repositoryMatchLimit = &val
		}

		// Always call this even if patterns are not supplied. Otherwise we run into the
		// situation where we have deleted all of the patterns associated with a policy
		// but it still has entries in the lookup table.
		if err := service.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		metrics.numPoliciesUpdated.Inc()
	}

	return nil
}
