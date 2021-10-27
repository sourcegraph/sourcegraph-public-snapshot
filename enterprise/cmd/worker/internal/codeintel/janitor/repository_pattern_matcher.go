package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

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

// NewRepositoryPatternMatcher returns a background routine that periodically updates the lookup table
// lsif_configuration_policies_repository_pattern_lookup from the patterns set in the repository_pattern column from repos tables.
//
// The lookup table updates periodically with new patterns set in the repository_pattern column. Should that column be empty we delete
// all the rows with that id in the lookup table.
func NewRepositoryPatternMatcher(dbStore DBStore, lsifStore LSIFStore, interval time.Duration, batchSize int, metrics *metrics) goroutine.BackgroundRoutine {
	interval = time.Second
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &repositoryPatternMatcher{
		dbStore:   dbStore,
		lsifstore: lsifStore,
		metrics:   metrics,
		batchSize: batchSize,
	})
}

func (r *repositoryPatternMatcher) Handle(ctx context.Context) error {
	// Get all policies (nulls first)
	policies, err := r.dbStore.SelectPoliciesForRepositoryMembershipUpdate(ctx, r.batchSize)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		var patterns []string
		if policy.RepositoryPatterns != nil {
			patterns = make([]string, 0, len(*policy.RepositoryPatterns))
			for _, pattern := range patterns {
				patterns = append(patterns, pattern)
			}
		}

		if err := r.dbStore.UpdateReposMatchingPatterns(ctx, patterns, policy.ID); err != nil {
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
