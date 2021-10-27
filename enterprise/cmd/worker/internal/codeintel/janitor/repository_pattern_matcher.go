package janitor

import (
	"context"
	"strings"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type repositoryPatternMatcher struct {
	dbStore DBStore
	lsifstore LSIFStore
	metrics *metrics
}

var _ goroutine.Handler = &repositoryPatternMatcher{}
var _ goroutine.ErrorHandler = &repositoryPatternMatcher{}

// TODO: add readme notes
func NewRepositoryPatternMatcher(dbStore DBStore, lsifStore LSIFStore, interval time.Duration, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &repositoryPatternMatcher{
		dbStore: dbStore,
		lsifstore: lsifStore,
		metrics: metrics,
	})
}

func (r *repositoryPatternMatcher) Handle(ctx context.Context) error {
	// Get all policies (nulls first)
	batchSize := 100
	policies, err := r.dbStore.GetAllConfigurationPolicies(ctx, batchSize)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if policy.RepositoryPatterns != nil {
			patterns := make([]string, 0, len(*policy.RepositoryPatterns))
			for _, pattern := range *policy.RepositoryPatterns {
				patterns = append(patterns, strings.ReplaceAll(pattern, "*", "%"))
			}

			if err := r.dbStore.UpdateReposMatchingPatterns(ctx, patterns, policy.ID); err != nil {
				return err
			}

		}
	}

	return nil
}

func (r *repositoryPatternMatcher) HandleError(err error) {
	r.metrics.numErrors.Inc()
	log15.Error("Failed to match pattern for repository", "error", err)
}
