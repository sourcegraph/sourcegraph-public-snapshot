package janitor

import (
	"context"
	"time"

	"github.com/derision-test/glock"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type unknownCommitJanitor struct {
	dbStore                   DBStore
	metrics                   *metrics
	minimumTimeSinceLastCheck time.Duration
	batchSize                 int
	maximumCommitLag          time.Duration
	clock                     glock.Clock
}

var _ goroutine.Handler = &unknownCommitJanitor{}
var _ goroutine.ErrorHandler = &unknownCommitJanitor{}

// NewUnknownCommitJanitor returns a background routine that periodically resolves each
// commit known by code intelligence data via gitserver to ensure that it has not been
// force-pushed away or pruned from the gitserver clone.
//
// Note that we're rather cautious about the order that we process the batch. We do this
// so that two frontends performing the same migration will not try to update two of the
// same records in the opposite order. If we rely on map iteration order we tend to see a
// lot of Postgres deadlock conditions and very slow progress.
func NewUnknownCommitJanitor(
	dbStore DBStore,
	minimumTimeSinceLastCheck time.Duration,
	batchSize int,
	maximumCommitLag time.Duration,
	interval time.Duration,
	metrics *metrics,
) goroutine.BackgroundRoutine {
	janitor := newJanitor(
		dbStore,
		minimumTimeSinceLastCheck,
		batchSize,
		maximumCommitLag,
		metrics,
		glock.NewRealClock(),
	)

	return goroutine.NewPeriodicGoroutine(context.Background(), interval, janitor)
}

func newJanitor(
	dbStore DBStore,
	minimumTimeSinceLastCheck time.Duration,
	batchSize int,
	maximumCommitLag time.Duration,
	metrics *metrics,
	clock glock.Clock,
) *unknownCommitJanitor {
	return &unknownCommitJanitor{
		dbStore:                   dbStore,
		metrics:                   metrics,
		minimumTimeSinceLastCheck: minimumTimeSinceLastCheck,
		batchSize:                 batchSize,
		maximumCommitLag:          maximumCommitLag,
		clock:                     clock,
	}
}

func (j *unknownCommitJanitor) Handle(ctx context.Context) (err error) {
	tx, err := j.dbStore.Transact(ctx)
	defer func() {
		err = tx.Done(err)
	}()

	batch, err := tx.StaleSourcedCommits(ctx, j.minimumTimeSinceLastCheck, j.batchSize, j.clock.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.StaleSourcedCommits")
	}

	for _, sourcedCommits := range batch {
		if err := j.handleSourcedCommits(ctx, tx, sourcedCommits); err != nil {
			return err
		}
	}

	return nil
}

func (j *unknownCommitJanitor) HandleError(err error) {
	j.metrics.numErrors.Inc()
	log15.Error("Failed to delete codeintel records with an unknown commit", "error", err)
}

func (j *unknownCommitJanitor) handleSourcedCommits(ctx context.Context, tx DBStore, sourcedCommits dbstore.SourcedCommits) error {
	for _, commit := range sourcedCommits.Commits {
		if err := j.handleCommit(ctx, tx, sourcedCommits.RepositoryID, sourcedCommits.RepositoryName, commit); err != nil {
			return err
		}
	}

	return nil
}

func (j *unknownCommitJanitor) handleCommit(ctx context.Context, tx DBStore, repositoryID int, repositoryName, commit string) error {
	var shouldDelete bool
	_, err := gitserver.ResolveRevision(ctx, database.NewDBWith(tx), api.RepoName(repositoryName), commit, gitserver.ResolveRevisionOptions{})
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
		_, uploadsDeleted, indexesDeleted, err := tx.DeleteSourcedCommits(ctx, repositoryID, commit, j.maximumCommitLag, j.clock.Now())
		if err != nil {
			return errors.Wrap(err, "dbstore.DeleteSourcedCommits")
		}

		if uploadsDeleted > 0 {
			log15.Debug("Deleted upload records with unresolvable commits", "count", uploadsDeleted)
			j.metrics.numUploadRecordsRemoved.Add(float64(uploadsDeleted))
		}
		if indexesDeleted > 0 {
			log15.Debug("Deleted index records with unresolvable commits", "count", indexesDeleted)
			j.metrics.numIndexRecordsRemoved.Add(float64(indexesDeleted))
		}

		return nil
	}

	if _, _, err := tx.UpdateSourcedCommits(ctx, repositoryID, commit, j.clock.Now()); err != nil {
		return errors.Wrap(err, "dbstore.UpdateSourcedCommits")
	}

	return nil
}
