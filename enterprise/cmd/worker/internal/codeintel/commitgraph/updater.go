package commitgraph

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Updater periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.
type Updater struct {
	dbStore                   DBStore
	locker                    Locker
	gitserverClient           GitserverClient
	maxAgeForNonStaleBranches time.Duration
	maxAgeForNonStaleTags     time.Duration
	operations                *operations
}

var (
	_ goroutine.Handler      = &Updater{}
	_ goroutine.ErrorHandler = &Updater{}
)

// NewUpdater returns a background routine that periodically updates the commit graph
// and visible uploads for each repository marked as dirty.
func NewUpdater(
	dbStore DBStore,
	locker Locker,
	gitserverClient GitserverClient,
	maxAgeForNonStaleBranches time.Duration,
	maxAgeForNonStaleTags time.Duration,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &Updater{
		dbStore:                   dbStore,
		locker:                    locker,
		gitserverClient:           gitserverClient,
		maxAgeForNonStaleBranches: maxAgeForNonStaleBranches,
		maxAgeForNonStaleTags:     maxAgeForNonStaleTags,
		operations:                newOperations(dbStore, observationContext),
	})
}

// Handle checks for dirty repositories and invokes the underlying updater on each one.
func (u *Updater) Handle(ctx context.Context) error {
	repositoryIDs, err := u.dbStore.DirtyRepositories(ctx)
	if err != nil {
		return errors.Wrap(err, "dbstore.DirtyRepositories")
	}

	var updateErr error
	for repositoryID, dirtyFlag := range repositoryIDs {
		if err := u.tryUpdate(ctx, repositoryID, dirtyFlag); err != nil {
			if updateErr == nil {
				updateErr = err
			} else {
				updateErr = errors.Append(updateErr, err)
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
	ok, unlock, err := u.locker.Lock(ctx, int32(repositoryID), false)
	if err != nil || !ok {
		return errors.Wrap(err, "locker.Lock")
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
	ctx, trace, endObservation := u.operations.commitUpdate.With(ctx, &err, observation.Args{
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
	trace.Log(log.Int("numCommitGraphKeys", len(commitGraph.Order())))

	refDescriptions, err := u.gitserverClient.RefDescriptions(ctx, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.RefDescriptions")
	}
	trace.Log(log.Int("numRefDescriptions", len(refDescriptions)))

	// Decorate the commit graph with the set of processed uploads are visible from each commit,
	// then bulk update the denormalized view in Postgres. We call this with an empty graph as well
	// so that we end up clearing the stale data and bulk inserting nothing.
	if err := u.dbStore.CalculateVisibleUploads(ctx, repositoryID, commitGraph, refDescriptions, u.maxAgeForNonStaleBranches, u.maxAgeForNonStaleTags, dirtyToken); err != nil {
		return errors.Wrap(err, "dbstore.CalculateVisibleUploads")
	}

	return nil
}

// getCommitGraph builds a partial commit graph that includes the most recent commits on each branch
// extending back as as the date of the oldest commit for which we have a processed upload for this
// repository.
//
// This optimization is necessary as decorating the commit graph is an operation that scales with
// the size of both the git graph and the number of uploads (multiplicatively). For repositories with
// a very large number of commits or distinct roots (most monorepos) this is a necessary optimization.
//
// The number of commits pulled back here should not grow over time unless the repo is growing at an
// accelerating rate, as we routinely expire old information for active repositories in a janitor
// process.
func (u *Updater) getCommitGraph(ctx context.Context, repositoryID int) (*gitdomain.CommitGraph, error) {
	commitDate, ok, err := u.dbStore.GetOldestCommitDate(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !ok {
		// We either have no uploads or the committed_at fields for this repository are still being
		// backfilled. In the first case, we'll return an empty graph to no-op the update. In the
		// latter case, we'll end up retrying to recalculate the commit graph for this repository
		// again once the migration fills the commit dates for this repository's uploads.
		log15.Warn("No oldest commit date found", "repositoryID", repositoryID)
		return gitdomain.ParseCommitGraph(nil), nil
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
