package autoindexing

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitorConfig struct {
	minimumTimeSinceLastCheck      time.Duration
	commitResolverBatchSize        int
	commitResolverMaximumCommitLag time.Duration
}

func (s *Service) NewJanitor(
	interval time.Duration,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	commitResolverMaximumCommitLag time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.handleCleanup(ctx, janitorConfig{
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

	return errs
}

func (s *Service) handleDeletedRepository(ctx context.Context) (err error) {
	indexesCounts, err := s.store.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.DeleteIndexesWithoutRepository")
	}

	for _, counts := range gatherCounts(indexesCounts) {
		s.logger.Debug(
			"Deleted codeintel records with a deleted repository",
			log.Int("repository_id", counts.repoID),
			log.Int("indexes_count", counts.indexesCount),
		)

		s.janitorMetrics.numIndexRecordsRemoved.Add(float64(counts.indexesCount))
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

func (s *Service) handleUnknownCommit(ctx context.Context, cfg janitorConfig) (err error) {
	staleIndexes, err := s.store.GetStaleSourcedCommits(ctx, cfg.minimumTimeSinceLastCheck, cfg.commitResolverBatchSize, s.clock.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.StaleSourcedCommits")
	}

	for _, sourcedCommits := range staleIndexes {
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
		indexesDeleted, err := s.store.DeleteSourcedCommits(ctx, repositoryID, commit, cfg.commitResolverMaximumCommitLag)
		if err != nil {
			return errors.Wrap(err, "indexSvc.DeleteSourcedCommits")
		}
		if indexesDeleted > 0 {
			// log.Debug("Deleted index records with unresolvable commits", "count", indexesDeleted)
			s.janitorMetrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
		}

		return nil
	}

	if _, err := s.store.UpdateSourcedCommits(ctx, repositoryID, commit, s.clock.Now()); err != nil {
		return errors.Wrap(err, "indexSvc.UpdateSourcedCommits")
	}

	return nil
}
