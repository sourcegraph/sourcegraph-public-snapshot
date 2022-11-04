package background

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (b *backgroundJob) NewScheduler(
	interval time.Duration,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	policyBatchSize int,
	inferenceConcurrency int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleScheduler(ctx, repositoryProcessDelay, repositoryBatchSize, policyBatchSize, inferenceConcurrency)
	}), b.operations.handleIndexScheduler)
}

func (b backgroundJob) handleScheduler(
	ctx context.Context,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	policyBatchSize int,
	inferenceConcurrency int,
) error {
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
	repositories, err := b.uploadSvc.GetRepositoriesForIndexScan(
		ctx,
		"lsif_last_index_scan",
		"last_index_scan_at",
		repositoryProcessDelay,
		conf.CodeIntelAutoIndexingAllowGlobalPolicies(),
		repositoryMatchLimit,
		repositoryBatchSize,
		time.Now(),
	)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.GetRepositoriesForIndexScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	now := timeutil.Now()

	// In parallel enqueue all the repos.
	var (
		sema  = semaphore.NewWeighted(int64(inferenceConcurrency))
		errs  error
		errMu sync.Mutex
	)

	for _, repositoryID := range repositories {
		if err := sema.Acquire(ctx, 1); err != nil {
			return err
		}
		go func(repositoryID int) {
			defer sema.Release(1)
			if repositoryErr := b.handleRepository(ctx, repositoryID, policyBatchSize, now); repositoryErr != nil {
				errMu.Lock()
				errs = errors.Append(errs, repositoryErr)
				errMu.Unlock()
			}
		}(repositoryID)
	}

	if err := sema.Acquire(ctx, int64(inferenceConcurrency)); err != nil {
		return errors.Wrap(err, "acquiring semaphore")
	}

	return errs
}

func (b backgroundJob) handleRepository(ctx context.Context, repositoryID, policyBatchSize int, now time.Time) error {
	offset := 0

	for {
		// Retrieve the set of configuration policies that affect indexing for this repository.
		policies, totalCount, err := b.policiesSvc.GetConfigurationPolicies(ctx, policiesshared.GetConfigurationPoliciesOptions{
			RepositoryID: repositoryID,
			ForIndexing:  true,
			Limit:        policyBatchSize,
			Offset:       offset,
		})
		if err != nil {
			return errors.Wrap(err, "policySvc.GetConfigurationPolicies")
		}
		offset += len(policies)

		// Get the set of commits within this repository that match an indexing policy
		commitMap, err := b.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, policies, now)
		if err != nil {
			return errors.Wrap(err, "policies.CommitsDescribedByPolicy")
		}

		for commit, policyMatches := range commitMap {
			if len(policyMatches) == 0 {
				continue
			}

			// Attempt to queue an index if one does not exist for each of the matching commits
			if _, err := b.autoindexingSvc.QueueIndexes(ctx, repositoryID, commit, "", false, false); err != nil {
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

func (b backgroundJob) NewOnDemandScheduler(interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		if !autoIndexingEnabled() {
			return nil
		}

		return b.autoindexingSvc.ProcessRepoRevs(ctx, batchSize)
	}))
}
