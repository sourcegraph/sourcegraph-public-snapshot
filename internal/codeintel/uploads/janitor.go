package uploads

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	autoindexingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
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

func (j *Service) handleDeletedRepository(ctx context.Context) (err error) {
	uploadsCounts, err := j.DeleteUploadsWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteUploadsWithoutRepository")
	}

	indexesCounts, err := j.autoIndexingSvc.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.DeleteIndexesWithoutRepository")
	}

	for _, counts := range gatherCounts(uploadsCounts, indexesCounts) {
		j.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("uploads_count", counts.uploadsCount),
			log.Int("indexes_count", counts.indexesCount),
		)

		j.janitorMetrics.numUploadRecordsRemoved.Add(float64(counts.uploadsCount))
		j.janitorMetrics.numIndexRecordsRemoved.Add(float64(counts.indexesCount))
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

func (j *Service) handleUnknownCommit(ctx context.Context, cfg janitorConfig) (err error) {
	staleUploads, err := j.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, j.clock.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.StaleSourcedCommits")
	}

	staleIndexes, err := j.autoIndexingSvc.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, j.clock.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.StaleSourcedCommits")
	}

	batch := mergeSourceCommits(staleUploads, staleIndexes)
	for _, sourcedCommits := range batch {
		if err := j.handleSourcedCommits(ctx, sourcedCommits, cfg); err != nil {
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

func (j *Service) handleSourcedCommits(ctx context.Context, sc SourcedCommits, cfg janitorConfig) error {
	for _, commit := range sc.Commits {
		if err := j.handleCommit(ctx, sc.RepositoryID, sc.RepositoryName, commit, cfg); err != nil {
			return err
		}
	}

	return nil
}

func (j *Service) handleCommit(ctx context.Context, repositoryID int, repositoryName, commit string, cfg janitorConfig) error {
	var shouldDelete bool
	_, err := j.gitserverClient.ResolveRevision(ctx, repositoryID, commit)
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
		_, uploadsDeleted, err := j.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag, j.clock.Now())
		if err != nil {
			return errors.Wrap(err, "uploadSvc.DeleteSourcedCommits")
		}
		if uploadsDeleted > 0 {
			// log.Debug("Deleted upload records with unresolvable commits", "count", uploadsDeleted)
			j.janitorMetrics.numUploadRecordsRemoved.Add(float64(uploadsDeleted))
		}

		indexesDeleted, err := j.autoIndexingSvc.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag, j.clock.Now())
		if err != nil {
			return errors.Wrap(err, "indexSvc.DeleteSourcedCommits")
		}
		if indexesDeleted > 0 {
			// log.Debug("Deleted index records with unresolvable commits", "count", indexesDeleted)
			j.janitorMetrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
		}

		return nil
	}

	if _, err := j.UpdateSourcedCommits(ctx, repositoryID, commit, j.clock.Now()); err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateSourcedCommits")
	}

	if _, err := j.autoIndexingSvc.UpdateSourcedCommits(ctx, repositoryID, commit, j.clock.Now()); err != nil {
		return errors.Wrap(err, "indexSvc.UpdateSourcedCommits")
	}

	return nil
}

// handleAbandonedUpload removes upload records which have not left the uploading state within the given TTL.
func (j *Service) handleAbandonedUpload(ctx context.Context, cfg janitorConfig) error {
	count, err := j.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-cfg.uploadTimeout))
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteUploadsStuckUploading")
	}
	if count > 0 {
		j.logger.Debug("Deleted abandoned upload records", log.Int("count", count))
		j.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (j *Service) handleExpiredUploadDeleter(ctx context.Context) error {
	count, err := j.SoftDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		j.logger.Info("Deleted expired codeintel uploads", log.Int("count", count))
		j.janitorMetrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (j *Service) handleHardDeleter(ctx context.Context) error {
	count, err := j.HardDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.HardDeleteExpiredUploads")
	}

	j.janitorMetrics.numUploadsPurged.Add(float64(count))
	return nil
}

func (j *Service) handleAuditLog(ctx context.Context, cfg janitorConfig) (err error) {
	count, err := j.DeleteOldAuditLogs(ctx, cfg.auditLogMaxAge, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteOldAuditLogs")
	}

	j.janitorMetrics.numAuditLogRecordsExpired.Add(float64(count))
	return nil
}
