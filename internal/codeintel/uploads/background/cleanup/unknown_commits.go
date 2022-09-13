package cleanup

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	autoindexing "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (j *janitor) HandleUnknownCommit(ctx context.Context) (err error) {
	staleUploads, err := j.uploadSvc.GetStaleSourcedCommits(ctx, ConfigInst.MinimumTimeSinceLastCheck, ConfigInst.CommitResolverBatchSize, j.clock.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.StaleSourcedCommits")
	}

	staleIndexes, err := j.indexSvc.GetStaleSourcedCommits(ctx, ConfigInst.MinimumTimeSinceLastCheck, ConfigInst.CommitResolverBatchSize, j.clock.Now())
	if err != nil {
		return errors.Wrap(err, "indexSvc.StaleSourcedCommits")
	}

	batch := mergeSourceCommits(staleUploads, staleIndexes)
	for _, sourcedCommits := range batch {
		if err := j.handleSourcedCommits(ctx, sourcedCommits); err != nil {
			return err
		}
	}

	return nil
}

func mergeSourceCommits(usc []uploads.SourcedCommits, isc []autoindexing.SourcedCommits) []SourcedCommits {
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

//	func (j *janitor) HandleError(err error) {
//		j.metrics.numErrors.Inc()
//		log.Error("Failed to delete codeintel records with an unknown commit", "error", err)
//	}
type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

func (j *janitor) handleSourcedCommits(ctx context.Context, sc SourcedCommits) error {
	for _, commit := range sc.Commits {
		if err := j.handleCommit(ctx, sc.RepositoryID, sc.RepositoryName, commit); err != nil {
			return err
		}
	}

	return nil
}

func (j *janitor) handleCommit(ctx context.Context, repositoryID int, repositoryName, commit string) error {
	var shouldDelete bool
	_, err := gitserver.NewClient(database.NewDBWith(j.logger, j.dbStore)).ResolveRevision(ctx, api.RepoName(repositoryName), commit, gitserver.ResolveRevisionOptions{})
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
		_, uploadsDeleted, err := j.uploadSvc.DeleteSourcedCommits(ctx, repositoryID, commit, ConfigInst.CommitResolverMaximumCommitLag, j.clock.Now())
		if err != nil {
			return errors.Wrap(err, "uploadSvc.DeleteSourcedCommits")
		}
		if uploadsDeleted > 0 {
			// log.Debug("Deleted upload records with unresolvable commits", "count", uploadsDeleted)
			j.metrics.numUploadRecordsRemoved.Add(float64(uploadsDeleted))
		}

		indexesDeleted, err := j.indexSvc.DeleteSourcedCommits(ctx, repositoryID, commit, ConfigInst.CommitResolverMaximumCommitLag, j.clock.Now())
		if err != nil {
			return errors.Wrap(err, "indexSvc.DeleteSourcedCommits")
		}
		if indexesDeleted > 0 {
			// log.Debug("Deleted index records with unresolvable commits", "count", indexesDeleted)
			j.metrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
		}

		return nil
	}

	if _, err := j.uploadSvc.UpdateSourcedCommits(ctx, repositoryID, commit, j.clock.Now()); err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateSourcedCommits")
	}

	if _, err := j.indexSvc.UpdateSourcedCommits(ctx, repositoryID, commit, j.clock.Now()); err != nil {
		return errors.Wrap(err, "indexSvc.UpdateSourcedCommits")
	}

	return nil
}
