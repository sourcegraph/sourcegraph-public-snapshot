package indexing

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexScheduler struct {
	dbStore                DBStore
	policyMatcher          PolicyMatcher
	indexEnqueuer          IndexEnqueuer
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	policyBatchSize        int
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
	policyBatchSize int,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	scheduler := &IndexScheduler{
		dbStore:                dbStore,
		policyMatcher:          policyMatcher,
		indexEnqueuer:          indexEnqueuer,
		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		policyBatchSize:        policyBatchSize,
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

	var repositoryMatchLimit *int
	if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
		repositoryMatchLimit = &val
	}

	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog.
	repositories, err := s.dbStore.SelectRepositoriesForIndexScan(
		ctx,
		s.repositoryProcessDelay,
		conf.CodeIntelAutoIndexingAllowGlobalPolicies(),
		repositoryMatchLimit,
		s.repositoryBatchSize,
	)
	if err != nil {
		return errors.Wrap(err, "dbstore.SelectRepositoriesForIndexScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	now := timeutil.Now()

	for _, repositoryID := range repositories {
		if repositoryErr := s.handleRepository(ctx, repositoryID, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = errors.Append(err, repositoryErr)
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
	now time.Time,
) error {
	offset := 0

	for {
		// Retrieve the set of configuration policies that affect indexing for this repository.
		policies, totalCount, err := s.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
			RepositoryID: repositoryID,
			ForIndexing:  true,
			Limit:        s.policyBatchSize,
			Offset:       offset,
		})
		if err != nil {
			return errors.Wrap(err, "dbstore.GetConfigurationPolicies")
		}
		offset += len(policies)

		// Get the set of commits within this repository that match an indexing policy
		commitMap, err := s.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, policies, now)
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

		if len(policies) == 0 || offset >= totalCount {
			return nil
		}
	}
}
