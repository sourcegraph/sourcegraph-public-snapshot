package background

import (
	"context"
	"sort"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type JanitorConfig struct {
	MinimumTimeSinceLastCheck      time.Duration
	CommitResolverBatchSize        int
	CommitResolverMaximumCommitLag time.Duration
	FailedIndexBatchSize           int
	FailedIndexMaxAge              time.Duration
}

type janitorJob struct {
	store           store.Store
	gitserverClient GitserverClient
	metrics         *janitorMetrics
	logger          log.Logger
	clock           glock.Clock
}

func NewJanitor(
	observationCtx *observation.Context,
	interval time.Duration,
	store store.Store,
	gitserverClient GitserverClient,
	clock glock.Clock,
	config JanitorConfig,
) goroutine.BackgroundRoutine {
	metrics := NewJanitorMetrics(observationCtx)
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		job := janitorJob{
			store:           store,
			gitserverClient: gitserverClient,
			metrics:         metrics,
			logger:          log.Scoped("codeintel.janitor.background", ""),
			clock:           clock,
		}

		return job.handleCleanup(ctx, config)
	}))
}

func (j janitorJob) handleCleanup(ctx context.Context, cfg JanitorConfig) (errs error) {
	// Reconciliation and denormalization
	if err := j.handleDeletedRepository(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := j.handleUnknownCommit(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	// Expiration
	if err := j.handleExpiredRecords(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}

func (j janitorJob) handleDeletedRepository(ctx context.Context) (err error) {
	indexesCounts, err := j.store.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "autoindexingSvc.DeleteIndexesWithoutRepository")
	}

	for _, counts := range gatherCounts(indexesCounts) {
		j.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("indexes_count", counts.indexesCount),
		)

		j.metrics.numIndexRecordsRemoved.Add(float64(counts.indexesCount))
	}

	return nil
}

type recordCount struct {
	repoID       int
	indexesCount int
}

func gatherCounts(indexesCounts map[int]int) []recordCount {
	repoIDsMap := map[int]struct{}{}
	for repoID := range indexesCounts {
		repoIDsMap[repoID] = struct{}{}
	}

	var repoIDs []int
	for repoID := range repoIDsMap {
		repoIDs = append(repoIDs, repoID)
	}
	sort.Ints(repoIDs)

	recordCounts := make([]recordCount, 0, len(repoIDs))
	for _, repoID := range repoIDs {
		recordCounts = append(recordCounts, recordCount{
			repoID:       repoID,
			indexesCount: indexesCounts[repoID],
		})
	}

	return recordCounts
}

func (j janitorJob) handleUnknownCommit(ctx context.Context, cfg JanitorConfig) (err error) {
	indexesDeleted, err := j.store.ProcessStaleSourcedCommits(
		ctx,
		cfg.MinimumTimeSinceLastCheck,
		cfg.CommitResolverBatchSize,
		cfg.CommitResolverMaximumCommitLag,
		j.shouldDeleteUploadsForCommit,
	)
	if err != nil {
		return err
	}
	if indexesDeleted > 0 {
		j.metrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
	}

	return nil
}

func (j janitorJob) handleExpiredRecords(ctx context.Context, cfg JanitorConfig) error {
	return j.store.ExpireFailedRecords(ctx, cfg.FailedIndexBatchSize, cfg.FailedIndexMaxAge, j.clock.Now())
}

func (j janitorJob) shouldDeleteUploadsForCommit(ctx context.Context, repositoryID int, commit string) (bool, error) {
	if _, err := j.gitserverClient.ResolveRevision(ctx, repositoryID, commit); err != nil {
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			// Target condition: repository is resolvable but the commit is not; was probably
			// force-pushed away and the commit was gc'd after some time or after a re-clone
			// in gitserver.
			return true, nil
		}

		if !gitdomain.IsRepoNotExist(err) {
			// unexpected error
			return false, errors.Wrap(err, "git.ResolveRevision")
		}
	}

	// We hit this in one of two conditions:
	//   - If we have no error then the commit is resolvable and we shouldn't touch it.
	//   - If we have a repository not found error, then we'll just update the timestamp
	//     of the record so we can move on to other data; we deleted records associated
	//     with deleted repositories in a separate janitor process.
	return false, nil
}
