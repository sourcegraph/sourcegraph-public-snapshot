package uploads

import (
	"context"
	"sort"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	autoindexingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	uploadsCounts, err := s.DeleteUploadsWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteUploadsWithoutRepository")
	}

	indexesCounts, err := s.autoIndexingSvc.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.DeleteIndexesWithoutRepository")
	}

	for _, counts := range gatherCounts(uploadsCounts, indexesCounts) {
		s.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("uploads_count", counts.uploadsCount),
			log.Int("indexes_count", counts.indexesCount),
		)

		s.janitorMetrics.numUploadRecordsRemoved.Add(float64(counts.uploadsCount))
		s.janitorMetrics.numIndexRecordsRemoved.Add(float64(counts.indexesCount))
	}

	return nil
}

type recordCount struct {
	repoID       int
	uploadsCount int
	indexesCount int
}

func gatherCounts(uploadsCounts, indexesCounts map[int]int) []recordCount {
	repoIDsMap := map[int]struct{}{}
	for repoID := range uploadsCounts {
		repoIDsMap[repoID] = struct{}{}
	}
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
			uploadsCount: uploadsCounts[repoID],
			indexesCount: indexesCounts[repoID],
		})
	}

	return recordCounts
}

func (s *Service) handleUnknownCommit(ctx context.Context, cfg janitorConfig) (err error) {
	staleUploads, err := s.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, s.clock.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.StaleSourcedCommits")
	}

	staleIndexes, err := s.autoIndexingSvc.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, s.clock.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.StaleSourcedCommits")
	}

	batch := mergeSourceCommits(staleUploads, staleIndexes)
	for _, sourcedCommits := range batch {
		if err := s.handleSourcedCommits(ctx, sourcedCommits, cfg); err != nil {
			return err
		}
	}

	return nil
}

func mergeSourceCommits(usc []shared.SourcedCommits, isc []autoindexingshared.SourcedCommits) []SourcedCommits {
	var sourceCommits []SourcedCommits
	for _, uc := range usc {
		sourceCommits = append(sourceCommits, SourcedCommits{
			RepositoryID:   uc.RepositoryID,
			RepositoryName: uc.RepositoryName,
			Commits:        uc.Commits,
		})
	}

	for _, ic := range isc {
		sourceCommits = append(sourceCommits, SourcedCommits{
			RepositoryID:   ic.RepositoryID,
			RepositoryName: ic.RepositoryName,
			Commits:        ic.Commits,
		})
	}

	return sourceCommits
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

func (s *Service) handleSourcedCommits(ctx context.Context, sc SourcedCommits, cfg janitorConfig) error {
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
		_, uploadsDeleted, err := s.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag, s.clock.Now())
		if err != nil {
			return errors.Wrap(err, "uploadSvc.DeleteSourcedCommits")
		}
		if uploadsDeleted > 0 {
			// log.Debug("Deleted upload records with unresolvable commits", "count", uploadsDeleted)
			s.janitorMetrics.numUploadRecordsRemoved.Add(float64(uploadsDeleted))
		}

		indexesDeleted, err := s.autoIndexingSvc.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag, s.clock.Now())
		if err != nil {
			return errors.Wrap(err, "indexSvc.DeleteSourcedCommits")
		}
		if indexesDeleted > 0 {
			// log.Debug("Deleted index records with unresolvable commits", "count", indexesDeleted)
			s.janitorMetrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
		}

		return nil
	}

	if _, err := s.UpdateSourcedCommits(ctx, repositoryID, commit, s.clock.Now()); err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateSourcedCommits")
	}

	if _, err := s.autoIndexingSvc.UpdateSourcedCommits(ctx, repositoryID, commit, s.clock.Now()); err != nil {
		return errors.Wrap(err, "indexSvc.UpdateSourcedCommits")
	}

	return nil
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.Int("repositoryID", repositoryID),
			otlog.String("commit", commit),
			otlog.Int("maximumCommitLag in ms", int(maximumCommitLag.Milliseconds())),
			otlog.String("now", now.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.Int("repositoryID", repositoryID),
			otlog.String("commit", commit),
			otlog.String("now", now.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

// handleAbandonedUpload removes upload records which have not left the uploading state within the given TTL.
func (s *Service) handleAbandonedUpload(ctx context.Context, cfg janitorConfig) error {
	count, err := s.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-cfg.uploadTimeout))
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteUploadsStuckUploading")
	}
	if count > 0 {
		s.logger.Debug("Deleted abandoned upload records", log.Int("count", count))
		s.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (s *Service) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.String("uploadedBefore", uploadedBefore.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsStuckUploading(ctx, uploadedBefore)
}

func (s *Service) handleExpiredUploadDeleter(ctx context.Context) error {
	count, err := s.SoftDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		s.logger.Info("Deleted expired codeintel uploads", log.Int("count", count))
		s.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (s *Service) SoftDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	ctx, _, endObservation := s.operations.softDeleteExpiredUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SoftDeleteExpiredUploads(ctx)
}

func (s *Service) handleHardDeleter(ctx context.Context) error {
	count, err := s.HardDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.HardDeleteExpiredUploads")
	}

	s.janitorMetrics.numUploadsPurged.Add(float64(count))
	return nil
}

func (s *Service) HardDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	ctx, _, endObservation := s.operations.hardDeleteUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

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
	count, err := s.DeleteOldAuditLogs(ctx, cfg.auditLogMaxAge, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteOldAuditLogs")
	}

	s.janitorMetrics.numAuditLogRecordsExpired.Add(float64(count))
	return nil
}

func (s *Service) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.String("maxAge", maxAge.String()),
			otlog.String("now", now.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteOldAuditLogs(ctx, maxAge, now)
}
