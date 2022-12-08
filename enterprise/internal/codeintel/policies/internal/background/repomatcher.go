package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewRepositoryMatcher(store store.Store, observationCtx *observation.Context, interval time.Duration, configurationPolicyMembershipBatchSize int) goroutine.BackgroundRoutine {
	repoMatcher := &repoMatcher{
		store: store,
	}
	metrics := newMetrics(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.policies-matcher", "match repositories to autoindexing+retention policies",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return repoMatcher.handleRepositoryMatcherBatch(ctx, configurationPolicyMembershipBatchSize, metrics)
		}),
	)
}

type repoMatcher struct {
	store store.Store
}

func (matcher *repoMatcher) handleRepositoryMatcherBatch(ctx context.Context, batchSize int, metrics *matcherMetrics) error {
	policies, err := matcher.store.SelectPoliciesForRepositoryMembershipUpdate(ctx, batchSize)
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
		if err := matcher.store.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		metrics.numPoliciesUpdated.Inc()
	}

	return nil
}
