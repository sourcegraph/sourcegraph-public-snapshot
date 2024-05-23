package repository_matcher

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewRepositoryMatcher(
	store store.Store,
	observationCtx *observation.Context,
	interval time.Duration,
	configurationPolicyMembershipBatchSize int,
) goroutine.BackgroundRoutine {
	repoMatcher := &repoMatcher{
		store:   store,
		metrics: newMetrics(observationCtx),
	}

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return repoMatcher.handleRepositoryMatcherBatch(ctx, configurationPolicyMembershipBatchSize)
		}),
		goroutine.WithName("codeintel.policies-matcher"),
		goroutine.WithDescription("match repositories to autoindexing+retention policies"),
		goroutine.WithInterval(interval),
	)
}

type repoMatcher struct {
	store   store.Store
	metrics *metrics
}

func (m *repoMatcher) handleRepositoryMatcherBatch(ctx context.Context, batchSize int) error {
	policies, err := m.store.SelectPoliciesForRepositoryMembershipUpdate(ctx, batchSize)
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
		if err := m.store.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		m.metrics.numPoliciesUpdated.Inc()
	}

	return nil
}
