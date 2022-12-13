package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewCommittedAtBackfiller(store store.Store, gitserverClient GitserverClient, interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	backfiller := &backfiller{
		store:           store,
		gitserverClient: gitserverClient,
		batchSize:       batchSize,
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.committed-at-backfiller", "backfills the committed_at column for code-intel uploads",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return backfiller.BackfillCommittedAtBatch(ctx, batchSize)
		}),
	)
}

type backfiller struct {
	batchSize       int
	store           store.Store
	gitserverClient GitserverClient
}

// BackfillCommittedAtBatch calculates the committed_at value for a batch of upload records that do not have
// this value set. This method is used to backfill old upload records prior to this value being reliably set
// during processing.
func (s *backfiller) BackfillCommittedAtBatch(ctx context.Context, batchSize int) (err error) {
	tx, err := s.store.Transact(ctx)
	defer func() {
		err = tx.Done(err)
	}()

	batch, err := tx.SourcedCommitsWithoutCommittedAt(ctx, batchSize)
	if err != nil {
		return errors.Wrap(err, "store.SourcedCommitsWithoutCommittedAt")
	}

	for _, sourcedCommits := range batch {
		for _, commit := range sourcedCommits.Commits {
			commitDateString, err := s.getCommitDate(ctx, sourcedCommits.RepositoryID, commit)
			if err != nil {
				return err
			}

			// Update commit date of all uploads attached to this this repository and commit
			if err := tx.UpdateCommittedAt(ctx, sourcedCommits.RepositoryID, commit, commitDateString); err != nil {
				return errors.Wrap(err, "store.UpdateCommittedAt")
			}
		}

		// Mark repository as dirty so the commit graph is recalculated with fresh data
		if err := tx.SetRepositoryAsDirty(ctx, sourcedCommits.RepositoryID); err != nil {
			return errors.Wrap(err, "store.SetRepositoryAsDirty")
		}
	}

	return nil
}

func (s *backfiller) getCommitDate(ctx context.Context, repositoryID int, commit string) (string, error) {
	_, commitDate, revisionExists, err := s.gitserverClient.CommitDate(ctx, repositoryID, commit)
	if err != nil {
		return "", errors.Wrap(err, "gitserver.CommitDate")
	}

	var commitDateString string
	if revisionExists {
		commitDateString = commitDate.Format(time.RFC3339)
	} else {
		// Set a value here that we'll filter out on the query side so that we don't
		// reprocess the same failing batch infinitely. We could alternatively soft
		// delete the record, but it would be better to keep record deletion behavior
		// together in the same place (so we have unified metrics on that event).
		commitDateString = "-infinity"
	}

	return commitDateString, nil
}
