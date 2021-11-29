package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type repositoryPatternMatcher struct {
	dbStore   DBStore
	lsifstore LSIFStore
	batchSize int
	metrics   *metrics
}

var _ goroutine.Handler = &repositoryPatternMatcher{}
var _ goroutine.ErrorHandler = &repositoryPatternMatcher{}

// NewRepositoryPatternMatcher returns a background routine that periodically updates the set of
// repositories over which a particular configuration policy applies. This set is stored in the
// table lsif_configuration_policies_repository_pattern_lookup, and uses the set of patterns in
// the repository_pattern field of the policy.
func NewRepositoryPatternMatcher(dbStore DBStore, lsifStore LSIFStore, interval time.Duration, batchSize int, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &repositoryPatternMatcher{
		dbStore:   dbStore,
		lsifstore: lsifStore,
		metrics:   metrics,
		batchSize: batchSize,
	})
}

func (r *repositoryPatternMatcher) Handle(ctx context.Context) error {
	policies, err := r.dbStore.SelectPoliciesForRepositoryMembershipUpdate(ctx, r.batchSize)
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
		if err := r.dbStore.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		r.metrics.numPoliciesUpdated.Inc()
	}

	return nil
}

func (r *repositoryPatternMatcher) HandleError(err error) {
	r.metrics.numErrors.Inc()
	log15.Error("Failed to match pattern for repository", "error", err)
}
