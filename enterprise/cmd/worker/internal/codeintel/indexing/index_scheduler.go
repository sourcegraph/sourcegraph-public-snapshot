package indexing

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

type IndexScheduler struct {
	dbStore                DBStore
	policyMatcher          PolicyMatcher
	indexEnqueuer          IndexEnqueuer
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	operations             *schedulerOperations
}

var (
	_ goroutine.Handler      = &IndexScheduler{}
	_ goroutine.ErrorHandler = &IndexScheduler{}
)

func NewIndexScheduler(
	dbStore DBStore,
	policyMatcher PolicyMatcher,
	indexEnqueuer IndexEnqueuer,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	scheduler := &IndexScheduler{
		dbStore:                dbStore,
		policyMatcher:          policyMatcher,
		indexEnqueuer:          indexEnqueuer,
		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		operations:             newOperations(observationContext),
	}

	return goroutine.NewPeriodicGoroutineWithMetrics(
		context.Background(),
		interval,
		scheduler,
		scheduler.operations.HandleIndexScheduler,
	)
}

// For mocking in tests
var autoIndexingEnabled = conf.CodeIntelAutoIndexingEnabled

func (s *IndexScheduler) Handle(ctx context.Context) (err error) {
	if !autoIndexingEnabled() {
		return nil
	}

	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog.
	repositories, err := s.dbStore.SelectRepositoriesForIndexScan(ctx, s.repositoryProcessDelay, s.repositoryBatchSize)
	if err != nil {
		return errors.Wrap(err, "dbstore.SelectRepositoriesForIndexScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	// Retrieve the set of global configuration policies that affect indexing. These policies are applied
	// to all repositories.
	globalPolicies, err := s.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
		ForIndexing: true,
	})
	if err != nil {
		return errors.Wrap(err, "dbstore.GetConfigurationPolicies")
	}

	now := timeutil.Now()

	for _, repositoryID := range repositories {
		if repositoryErr := s.handleRepository(ctx, repositoryID, globalPolicies, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = multierror.Append(err, repositoryErr)
			}
		}
	}

	return err
}

func (s *IndexScheduler) HandleError(err error) {
	log15.Error("Failed to schedule index jobs", "err", err)
}

func (s *IndexScheduler) handleRepository(
	ctx context.Context,
	repositoryID int,
	globalPolicies []dbstore.ConfigurationPolicy,
	now time.Time,
) error {
	// Retrieve the set of configuration policies that affect indexing. These policies are applied
	// only to this repository.
	repositoryPolicies, err := s.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
		RepositoryID: repositoryID,
		ForIndexing:  true,
	})
	if err != nil {
		return errors.Wrap(err, "dbstore.GetConfigurationPolicies")
	}

	// Combine global and repository-specific policies. The resulting slice may be empty, but that
	// condition is short-circuited in the call to CommitsDescribedByPolicy below.
	combinedPolicies := make([]dbstore.ConfigurationPolicy, 0, len(globalPolicies)+len(repositoryPolicies))
	combinedPolicies = append(combinedPolicies, globalPolicies...)
	combinedPolicies = append(combinedPolicies, repositoryPolicies...)

	// Get the set of commits within this repository that match an indexing policy
	commitMap, err := s.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, combinedPolicies, now)
	if err != nil {
		return errors.Wrap(err, "policies.CommitsDescribedByPolicy")
	}

	for commit, policyMatches := range commitMap {
		if len(policyMatches) == 0 {
			continue
		}

		// Attempt to queue an index if one does not exist for each of the matching commits
		if _, err := s.indexEnqueuer.QueueIndexes(ctx, repositoryID, commit, "", false); err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
				continue
			}

			return errors.Wrap(err, "indexEnqueuer.QueueIndexes")
		}
	}

	return nil
}
