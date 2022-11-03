package background

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitorConfig struct {
	minimumTimeSinceLastCheck      time.Duration
	commitResolverBatchSize        int
	commitResolverMaximumCommitLag time.Duration
	failedIndexBatchSize           int
	failedIndexMaxAge              time.Duration
}

func (b backgroundJob) NewJanitor(
	interval time.Duration,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	commitResolverMaximumCommitLag time.Duration,
	failedIndexBatchSize int,
	failedIndexMaxAge time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleCleanup(ctx, janitorConfig{
			minimumTimeSinceLastCheck:      minimumTimeSinceLastCheck,
			commitResolverBatchSize:        commitResolverBatchSize,
			commitResolverMaximumCommitLag: commitResolverMaximumCommitLag,
			failedIndexBatchSize:           failedIndexBatchSize,
			failedIndexMaxAge:              failedIndexMaxAge,
		})
	}))
}

func (b backgroundJob) handleCleanup(ctx context.Context, cfg janitorConfig) (errs error) {
	// Reconciliation and denormalization
	if err := b.handleDeletedRepository(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.handleUnknownCommit(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	// Expiration
	if err := b.handleExpiredRecords(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}

func (b backgroundJob) handleDeletedRepository(ctx context.Context) (err error) {
	indexesCounts, err := b.autoindexingSvc.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "autoindexingSvc.DeleteIndexesWithoutRepository")
	}

	for _, counts := range gatherCounts(indexesCounts) {
		b.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("indexes_count", counts.indexesCount),
		)

		b.janitorMetrics.numIndexRecordsRemoved.Add(float64(counts.indexesCount))
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

func (b backgroundJob) handleUnknownCommit(ctx context.Context, cfg janitorConfig) (err error) {
	indexesDeleted, err := b.autoindexingSvc.ProcessStaleSourcedCommits(
		ctx,
		cfg.minimumTimeSinceLastCheck,
		cfg.commitResolverBatchSize,
		cfg.commitResolverMaximumCommitLag,
		func(ctx context.Context, repositoryID int, commit string) (bool, error) {
			return shouldDeleteUploadsForCommit(ctx, b.gitserverClient, repositoryID, commit)
		},
	)
	if err != nil {
		return err
	}
	if indexesDeleted > 0 {
		b.janitorMetrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
	}

	return nil
}

func (b backgroundJob) handleExpiredRecords(ctx context.Context, cfg janitorConfig) error {
	return b.autoindexingSvc.ExpireFailedRecords(ctx, cfg.failedIndexBatchSize, cfg.failedIndexMaxAge, b.clock.Now())
}

func shouldDeleteUploadsForCommit(ctx context.Context, gitserverClient GitserverClient, repositoryID int, commit string) (bool, error) {
	if _, err := gitserverClient.ResolveRevision(ctx, repositoryID, commit); err != nil {
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
