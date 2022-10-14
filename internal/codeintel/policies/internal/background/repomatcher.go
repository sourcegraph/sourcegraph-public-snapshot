package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b backgroundJob) NewRepositoryMatcher(interval time.Duration, configurationPolicyMembershipBatchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleRepositoryMatcherBatch(ctx, configurationPolicyMembershipBatchSize)
	}))
}

func (b backgroundJob) handleRepositoryMatcherBatch(ctx context.Context, batchSize int) error {
	policies, err := b.policySvc.SelectPoliciesForRepositoryMembershipUpdate(ctx, batchSize)
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
		if err := b.policySvc.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		b.matcherMetrics.numPoliciesUpdated.Inc()
	}

	return nil
}
