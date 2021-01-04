package commitgraph

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	basegitserver "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// Updater periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.
type Updater struct {
	dbStore         DBStore
	gitserverClient GitserverClient
	operations      *operations
}

var _ goroutine.Handler = &Updater{}

// NewUpdater returns a background routine that periodically updates the commit graph
// and visible uploads for each repository marked as dirty.
func NewUpdater(
	dbStore DBStore,
	gitserverClient GitserverClient,
	interval time.Duration,
	operations *operations,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &Updater{
		dbStore:         dbStore,
		gitserverClient: gitserverClient,
		operations:      operations,
	})
}

// Handle checks for dirty repositories and invokes the underlying updater on each one.
func (u *Updater) Handle(ctx context.Context) error {
	repositoryIDs, err := u.dbStore.DirtyRepositories(ctx)
	if err != nil {
		return errors.Wrap(err, "store.DirtyRepositories")
	}

	var updateErr error
	for repositoryID, dirtyFlag := range repositoryIDs {
		if err := u.tryUpdate(ctx, repositoryID, dirtyFlag); err != nil {
			if updateErr == nil {
				updateErr = err
			} else {
				updateErr = multierror.Append(updateErr, err)
			}
		}
	}

	return updateErr
}

func (u *Updater) HandleError(err error) {
	log15.Error("Failed to run update process", "err", err)
}

// tryUpdate will call update while holding an advisory lock to give exclusive access to the
// update procedure for this repository. If the lock is already held, this method will simply
// do nothing.
func (u *Updater) tryUpdate(ctx context.Context, repositoryID, dirtyToken int) (err error) {
	ok, unlock, err := u.dbStore.Lock(ctx, repositoryID, false)
	if err != nil || !ok {
		return errors.Wrap(err, "store.Lock")
	}
	defer func() {
		err = unlock(err)
	}()

	return u.update(ctx, repositoryID, dirtyToken)
}

// update pulls the commit graph for the given repository from gitserver, pulls the set of LSIF
// upload objects for the given repository from Postgres, and correlates them into a visibility
// graph. This graph is then upserted back into Postgres for use by find closest dumps queries.
//
// The user should supply a dirty token that is associated with the given repository so that
// the repository can be unmarked as long as the repository is not marked as dirty again before
// the update completes.
func (u *Updater) update(ctx context.Context, repositoryID, dirtyToken int) (err error) {
	// Enable tracing on the context and trace the operation
	ctx = ot.WithShouldTrace(ctx, true)

	ctx, traceLog, endObservation := u.operations.commitUpdate.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.Int("dirtyToken", dirtyToken),
		},
	})
	defer endObservation(1, observation.Args{})

	// Construct a view of the git graph that we will later decorate with upload information.
	commitGraph, err := u.getCommitGraph(ctx, repositoryID)
	if err != nil {
		return err
	}
	traceLog(log.Int("numCommitGraphKeys", len(commitGraph.Order())))

	tipCommit, err := u.gitserverClient.Head(ctx, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}
	traceLog(log.String("tipCommit", tipCommit))

	// Decorate the commit graph with the set of processed uploads are visible from each commit,
	// then bulk update the denormalized view in Postgres. We call this with an empty graph as well
	// so that we end up clearing the stale data and bulk inserting nothing.
	if err := u.dbStore.CalculateVisibleUploads(ctx, repositoryID, commitGraph, tipCommit, dirtyToken); err != nil {
		return errors.Wrap(err, "store.CalculateVisibleUploads")
	}

	return nil
}

// getCommitGraph builds a partial commit graph that includes the most recent commits on each branch
// extending back as as the date of the commit of the oldest upload processed for this repository.
//i
// This optimization is necessary as decorating the commit graph is an operation that scales with
// the size of both the git graph and the number of uploads (multiplicatively). For repositories with
// a very large number of commits or distinct roots (most monorepos) this is a necessary optimization.
//
// The number of commits pulled back here should not grow over time unless the repo is growing at an
// accelerating rate, as we routinely expire old information for active repositories in a janitor
// process.
func (u *Updater) getCommitGraph(ctx context.Context, repositoryID int) (*gitserver.CommitGraph, error) {
	commitDate, ok, err := u.getOldestCommitDate(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return gitserver.ParseCommitGraph(nil), nil
	}

	// The --since flag for git log is exclusive, but we want to include the commit where the
	// oldest dump is defined. This flag only has second resolution, so we shouldn't be pulling
	// back any more data than we wanted.
	commitDate = commitDate.Add(-time.Second)

	commitGraph, err := u.gitserverClient.CommitGraph(ctx, repositoryID, gitserver.CommitGraphOptions{
		AllRefs: true,
		Since:   &commitDate,
	})
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.CommitGraph")
	}

	return commitGraph, nil
}

// TODO(efritz) - make adjustable
const commitDateBatchSize = 100

// getOldestCommitDate returns the commit date of the oldest processed upload for the given repository.
func (u *Updater) getOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error) {
	uploads, _, err := u.dbStore.GetUploads(ctx, dbstore.GetUploadsOptions{
		RepositoryID: repositoryID,
		State:        "completed",
		OldestFirst:  true,
		Limit:        commitDateBatchSize,
	})
	if err != nil {
		return time.Time{}, false, errors.Wrap(err, "store.GetUploads")
	}

outer:
	for _, upload := range uploads {
		commitDate, err := u.gitserverClient.CommitDate(ctx, repositoryID, upload.Commit)
		if err != nil {
			for ex := err; ex != nil; ex = errors.Unwrap(ex) {
				if basegitserver.IsRevisionNotFound(ex) {
					log15.Warn("Unknown commit", "commit", upload.Commit)
					continue outer
				}
			}

			return time.Time{}, false, errors.Wrap(err, "gitserver.CommitDate")
		}

		return commitDate, true, nil
	}

	return time.Time{}, false, nil
}
