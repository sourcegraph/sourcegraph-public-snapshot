package uploads

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
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

func (s *Service) NewJanitor(
	interval time.Duration,
	uploadTimeout time.Duration,
	auditLogMaxAge time.Duration,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	commitResolverMaximumCommitLag time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.handleCleanup(ctx, janitorConfig{
			uploadTimeout:                  uploadTimeout,
			auditLogMaxAge:                 auditLogMaxAge,
			minimumTimeSinceLastCheck:      minimumTimeSinceLastCheck,
			commitResolverBatchSize:        commitResolverBatchSize,
			commitResolverMaximumCommitLag: commitResolverMaximumCommitLag,
		})
	}))
}

func (s *Service) handleCleanup(ctx context.Context, cfg janitorConfig) (errs error) {
	// Reconciliation and denormalization
	if err := s.handleDeletedRepository(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := s.handleUnknownCommit(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	// Expiration
	if err := s.handleAbandonedUpload(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := s.handleExpiredUploadDeleter(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := s.handleHardDeleter(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := s.handleAuditLog(ctx, cfg); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}

func (s *Service) handleDeletedRepository(ctx context.Context) (err error) {
	uploadsCounts, err := s.store.DeleteUploadsWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteUploadsWithoutRepository")
	}

	for _, counts := range gatherCounts(uploadsCounts) {
		s.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("uploads_count", counts.uploadsCount),
		)

		s.janitorMetrics.numUploadRecordsRemoved.Add(float64(counts.uploadsCount))
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

func (s *Service) handleUnknownCommit(ctx context.Context, cfg janitorConfig) (err error) {
	staleUploads, err := s.store.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, s.clock.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.StaleSourcedCommits")
	}

	for _, sourcedCommits := range staleUploads {
		if err := s.handleSourcedCommits(ctx, sourcedCommits, cfg); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) handleSourcedCommits(ctx context.Context, sc shared.SourcedCommits, cfg janitorConfig) error {
	for _, commit := range sc.Commits {
		if err := s.handleCommit(ctx, sc.RepositoryID, sc.RepositoryName, commit, cfg); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) handleCommit(ctx context.Context, repositoryID int, repositoryName, commit string, cfg janitorConfig) error {
	var shouldDelete bool
	_, err := s.gitserverClient.ResolveRevision(ctx, repositoryID, commit)
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
		_, uploadsDeleted, err := s.store.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag, s.clock.Now())
		if err != nil {
			return errors.Wrap(err, "uploadSvc.DeleteSourcedCommits")
		}
		if uploadsDeleted > 0 {
			// log.Debug("Deleted upload records with unresolvable commits", "count", uploadsDeleted)
			s.janitorMetrics.numUploadRecordsRemoved.Add(float64(uploadsDeleted))
		}

		return nil
	}

	if _, err := s.store.UpdateSourcedCommits(ctx, repositoryID, commit, s.clock.Now()); err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateSourcedCommits")
	}

	return nil
}

// handleAbandonedUpload removes upload records which have not left the uploading state within the given TTL.
func (s *Service) handleAbandonedUpload(ctx context.Context, cfg janitorConfig) error {
	count, err := s.store.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-cfg.uploadTimeout))
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteUploadsStuckUploading")
	}
	if count > 0 {
		s.logger.Debug("Deleted abandoned upload records", log.Int("count", count))
		s.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (s *Service) handleExpiredUploadDeleter(ctx context.Context) error {
	count, err := s.store.SoftDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		s.logger.Info("Deleted expired codeintel uploads", log.Int("count", count))
		s.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (s *Service) handleHardDeleter(ctx context.Context) error {
	count, err := s.hardDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.HardDeleteExpiredUploads")
	}

	s.janitorMetrics.numUploadsPurged.Add(float64(count))
	return nil
}

func (s *Service) hardDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	const uploadsBatchSize = 100
	options := types.GetUploadsOptions{
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
		uploads, totalCount, err := s.store.GetUploads(ctx, options)
		if err != nil {
			return 0, errors.Wrap(err, "store.GetUploads")
		}

		ids := uploadIDs(uploads)
		if err := s.lsifstore.DeleteLsifDataByUploadIds(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "lsifstore.Clear")
		}

		if err := s.store.HardDeleteUploadsByIDs(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "store.HardDeleteUploadsByIDs")
		}

		count += len(uploads)
		if count >= totalCount {
			break
		}
	}

	return count, nil
}

func (s *Service) handleAuditLog(ctx context.Context, cfg janitorConfig) (err error) {
	count, err := s.store.DeleteOldAuditLogs(ctx, cfg.auditLogMaxAge, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteOldAuditLogs")
	}

	s.janitorMetrics.numAuditLogRecordsExpired.Add(float64(count))
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
