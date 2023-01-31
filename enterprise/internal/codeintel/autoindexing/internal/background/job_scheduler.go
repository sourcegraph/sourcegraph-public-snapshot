package background

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexSchedulerConfig struct {
	RepositoryProcessDelay time.Duration
	RepositoryBatchSize    int
	PolicyBatchSize        int
	InferenceConcurrency   int
}

type indexSchedulerJob struct {
	uploadSvc     UploadService
	policiesSvc   PoliciesService
	policyMatcher PolicyMatcher
	indexEnqueuer IndexEnqueuer
}

var m = new(metrics.SingletonREDMetrics)

func NewScheduler(
	observationCtx *observation.Context,
	uploadSvc UploadService,
	policiesSvc PoliciesService,
	policyMatcher PolicyMatcher,
	indexEnqueuer IndexEnqueuer,
	interval time.Duration,
	config IndexSchedulerConfig,
) goroutine.BackgroundRoutine {
	job := indexSchedulerJob{
		uploadSvc:     uploadSvc,
		policiesSvc:   policiesSvc,
		policyMatcher: policyMatcher,
		indexEnqueuer: indexEnqueuer,
	}

	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_autoindexing_background",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	return goroutine.NewPeriodicGoroutineWithMetrics(
		context.Background(),
		"codeintel.autoindexing-background-scheduler", "schedule autoindexing jobs in the background using defined or inferred configurations",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return job.handleScheduler(ctx, config.RepositoryProcessDelay, config.RepositoryBatchSize, config.PolicyBatchSize, config.InferenceConcurrency)
		}),
		observationCtx.Operation(observation.Op{
			Name:              "codeintel.indexing.HandleIndexSchedule",
			MetricLabelValues: []string{"HandleIndexSchedule"},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if errors.As(err, &inference.LimitError{}) {
					return observation.EmitForDefault.Without(observation.EmitForMetrics)
				}
				return observation.EmitForDefault
			},
		}),
	)
}

func (b indexSchedulerJob) handleScheduler(
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

func (b indexSchedulerJob) handleRepository(ctx context.Context, repositoryID, policyBatchSize int, now time.Time) error {
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
			if _, err := b.indexEnqueuer.QueueIndexes(ctx, repositoryID, commit, "", false, false); err != nil {
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

func NewOnDemandScheduler(store store.Store, indexEnqueuer IndexEnqueuer, interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.autoindexing-ondemand-scheduler", "schedule autoindexing jobs for explicitly requested repo+revhash combinations",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if !autoIndexingEnabled() {
				return nil
			}

			tx, err := store.Transact(ctx)
			if err != nil {
				return err
			}
			defer func() { err = tx.Done(err) }()

			repoRevs, err := tx.GetQueuedRepoRev(ctx, batchSize)
			if err != nil {
				return err
			}

			ids := make([]int, 0, len(repoRevs))
			for _, repoRev := range repoRevs {
				if _, err := indexEnqueuer.QueueIndexes(ctx, repoRev.RepositoryID, repoRev.Rev, "", false, false); err != nil {
					return err
				}

				ids = append(ids, repoRev.ID)
			}

			return tx.MarkRepoRevsAsProcessed(ctx, ids)
		}),
	)
}
