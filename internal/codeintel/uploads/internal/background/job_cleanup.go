package background

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitorConfig struct {
	uploadTimeout                  time.Duration
	auditLogMaxAge                 time.Duration
	minimumTimeSinceLastCheck      time.Duration
	commitResolverBatchSize        int
	commitResolverMaximumCommitLag time.Duration
}

func (b backgroundJob) NewJanitor(
	interval time.Duration,
	uploadTimeout time.Duration,
	auditLogMaxAge time.Duration,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	commitResolverMaximumCommitLag time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleCleanup(ctx, janitorConfig{
			uploadTimeout:                  uploadTimeout,
			auditLogMaxAge:                 auditLogMaxAge,
			minimumTimeSinceLastCheck:      minimumTimeSinceLastCheck,
			commitResolverBatchSize:        commitResolverBatchSize,
			commitResolverMaximumCommitLag: commitResolverMaximumCommitLag,
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
	if err := b.handleAbandonedUpload(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.handleExpiredUploadDeleter(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.handleHardDeleter(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.handleAuditLog(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}

func (b backgroundJob) handleDeletedRepository(ctx context.Context) (err error) {
	uploadsCounts, err := b.uploadSvc.DeleteUploadsWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteUploadsWithoutRepository")
	}

	for _, counts := range gatherCounts(uploadsCounts) {
		b.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("uploads_count", counts.uploadsCount),
		)

		b.janitorMetrics.numUploadRecordsRemoved.Add(float64(counts.uploadsCount))
	}

	return nil
}

type recordCount struct {
	repoID       int
	uploadsCount int
}

func gatherCounts(uploadsCounts map[int]int) []recordCount {
	repoIDsMap := map[int]struct{}{}
	for repoID := range uploadsCounts {
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
			uploadsCount: uploadsCounts[repoID],
		})
	}

	return recordCounts
}

func (b backgroundJob) handleUnknownCommit(ctx context.Context, cfg janitorConfig) (err error) {
	staleUploads, err := b.uploadSvc.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, b.clock.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.StaleSourcedCommits")
	}

	for _, sourcedCommits := range staleUploads {
		if err := b.handleSourcedCommits(ctx, sourcedCommits, cfg); err != nil {
			return err
		}
	}

	return nil
}

func (b backgroundJob) handleSourcedCommits(ctx context.Context, sc shared.SourcedCommits, cfg janitorConfig) error {
	for _, commit := range sc.Commits {
		if err := b.handleCommit(ctx, sc.RepositoryID, sc.RepositoryName, commit, cfg); err != nil {
			return err
		}
	}

	return nil
}

func (b backgroundJob) handleCommit(ctx context.Context, repositoryID int, repositoryName, commit string, cfg janitorConfig) error {
	var shouldDelete bool
	_, err := b.gitserverClient.ResolveRevision(ctx, repositoryID, commit)
	if err == nil {
		// If we have no error then the commit is resolvable and we shouldn't touch it.
		shouldDelete = false
	} else if gitdomain.IsRepoNotExist(err) {
		// If we have a repository not found error, then we'll just update the timestamp
		// of the record so we can move on to other data; we deleted records associated
		// with deleted repositories in a separate janitor process.
		shouldDelete = false
	} else if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		// Target condition: repository is resolvable bu the commit is not; was probably
		// force-pushed away and the commit was gc'd after some time or after a re-clone
		// in gitserver.
		shouldDelete = true
	} else {
		// unexpected error
		return errors.Wrap(err, "git.ResolveRevision")
	}

	if shouldDelete {
		_, uploadsDeleted, err := b.uploadSvc.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag, b.clock.Now())
		if err != nil {
			return errors.Wrap(err, "uploadSvc.DeleteSourcedCommits")
		}
		if uploadsDeleted > 0 {
			b.janitorMetrics.numUploadRecordsRemoved.Add(float64(uploadsDeleted))
		}

		return nil
	}

	if _, err := b.uploadSvc.UpdateSourcedCommits(ctx, repositoryID, commit, b.clock.Now()); err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateSourcedCommits")
	}

	return nil
}

// handleAbandonedUpload removes upload records which have not left the uploading state within the given TTL.
func (b backgroundJob) handleAbandonedUpload(ctx context.Context, cfg janitorConfig) error {
	count, err := b.uploadSvc.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-cfg.uploadTimeout))
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteUploadsStuckUploading")
	}
	if count > 0 {
		b.logger.Debug("Deleted abandoned upload records", log.Int("count", count))
		b.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

const expiredUploadsBatchSize = 1000

func (b backgroundJob) handleExpiredUploadDeleter(ctx context.Context) error {
	count, err := b.uploadSvc.SoftDeleteExpiredUploads(ctx, expiredUploadsBatchSize)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		b.logger.Info("Deleted expired codeintel uploads", log.Int("count", count))
		b.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (b backgroundJob) handleHardDeleter(ctx context.Context) error {
	count, err := b.hardDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.HardDeleteExpiredUploads")
	}

	b.janitorMetrics.numUploadsPurged.Add(float64(count))
	return nil
}

func (b backgroundJob) hardDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	const uploadsBatchSize = 100
	options := shared.GetUploadsOptions{
		State:            "deleted",
		Limit:            uploadsBatchSize,
		AllowExpired:     true,
		AllowDeletedRepo: true,
	}

	for {
		// Always request the first page of deleted uploads. If this is not
		// the first iteration of the loop, then the previous iteration has
		// deleted the records that composed the previous page, and the
		// previous "second" page is now the first page.
		uploads, totalCount, err := b.uploadSvc.GetUploads(ctx, options)
		if err != nil {
			return 0, errors.Wrap(err, "store.GetUploads")
		}

		ids := uploadIDs(uploads)
		if err := b.uploadSvc.DeleteLsifDataByUploadIds(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "lsifstore.Clear")
		}

		if err := b.uploadSvc.HardDeleteUploadsByIDs(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "store.HardDeleteUploadsByIDs")
		}

		count += len(uploads)
		if count >= totalCount {
			break
		}
	}

	return count, nil
}

func (b backgroundJob) handleAuditLog(ctx context.Context, cfg janitorConfig) (err error) {
	count, err := b.uploadSvc.DeleteOldAuditLogs(ctx, cfg.auditLogMaxAge, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteOldAuditLogs")
	}

	b.janitorMetrics.numAuditLogRecordsExpired.Add(float64(count))
	return nil
}

func uploadIDs(uploads []types.Upload) []int {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}
	sort.Ints(ids)

	return ids
}
