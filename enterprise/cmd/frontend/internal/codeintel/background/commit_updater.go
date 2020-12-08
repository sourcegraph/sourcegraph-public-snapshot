package background

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// Updater periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.
type CommitUpdater struct {
	dbStore         DBStore
	gitserverClient GitserverClient
}

var _ goroutine.Handler = &CommitUpdater{}

// NewCommitUpdater returns a background routine that periodically updates the commit graph
// and visible uploads for each repository marked as dirty.
func NewCommitUpdater(dbStore DBStore, gitserverClient GitserverClient, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &CommitUpdater{
		dbStore:         dbStore,
		gitserverClient: gitserverClient,
	})
}

// Handle checks for dirty repositories and invokes the underlying updater on each one.
func (u *CommitUpdater) Handle(ctx context.Context) error {
	repositoryIDs, err := u.dbStore.DirtyRepositories(ctx)
	if err != nil {
		return errors.Wrap(err, "store.DirtyRepositories")
	}

	for repositoryID, dirtyFlag := range repositoryIDs {
		if err := u.tryUpdate(ctx, repositoryID, dirtyFlag); err != nil {
			log15.Warn("Failed to update commit graph", "err", err)
		}
	}

	return nil
}

func (u *CommitUpdater) HandleError(err error) {
	log15.Error("Failed to run update process", "err", err)
}

// tryUpdate pulls the commit graph for the given repository from gitserver, pulls the set
// of LSIF upload objects for the given repository from Postgres, and correlates them into a
// visibility graph. This graph is then upserted back into Postgres for use by find closest
// dumps queries.
//
// This method will attempt to acquire an advisory lock to give exclusive access to the update
// procedure for this repository. If the lock is already held, this method will simply return
// early. The user should supply a dirty token that is associated with the given repository so
// that the repository can be unmarked as long as the repository is not marked as dirty again
// before the update completes.
func (u *CommitUpdater) tryUpdate(ctx context.Context, repositoryID, dirtyToken int) error {
	ok, unlock, err := u.dbStore.Lock(ctx, repositoryID, false)
	if err != nil || !ok {
		return errors.Wrap(err, "store.Lock")
	}
	defer func() {
		err = unlock(err)
	}()

	// Construct a view of the git graph that we will later decorate with upload information.
	// We will only fetch commits that are newer than the oldest commit with LSIF data. This
	// will pull back the smaller set of _relevant_ commits which we need denormalized data
	// for in the query path.
	//
	// Decorating the graph (below) is an operation that scales _multiplicatively_ with the
	// size of the git graph and the number of uploads. This is a necessary optimization for
	// repositories with very large number of commits. The number of commits pulled back here
	// should not grow (unless the repo is growing at an accelerating rate) as we routinely
	// expire old information for active repositories in a janitor process.
	dump, ok, err := u.dbStore.OldestDumpForRepository(ctx, repositoryID)
	if err != nil {
		return errors.Wrap(err, "store.OldestDumpForRepository")
	}

	var graph *gitserver.CommitGraph
	if ok {
		oldestCommitDateWithUpload, err := u.gitserverClient.CommitDate(ctx, repositoryID, dump.Commit)
		if err != nil {
			// TODO(efritz) - handle missing commit error condition. This is probably an extremely
			// unlikely edge case, but it's possible that the oldest commit was force-pushed out of
			// existence in the code host (then subsequently gitserver) before the janitor process
			// removes the orphaned commits. This will cause tryUpdate to fail here and the repo
			// will remain dirty until one of the conditions above changes.
			//
			// It's a bit nasty but self-correcting.
			return errors.Wrap(err, "gitserver.CommitDate")
		}

		// The --since flag for git log is exclusive, but we want to include the commit where the
		// oldest dump is defined. This flag only has second resolution, so we shouldn't be pulling
		// back any more data than we wanted.
		oldestCommitDateWithUpload = oldestCommitDateWithUpload.Add(-time.Second)

		graph, err = u.gitserverClient.CommitGraph(ctx, repositoryID, gitserver.CommitGraphOptions{
			Since: &oldestCommitDateWithUpload,
		})
		if err != nil {
			return errors.Wrap(err, "gitserver.CommitGraph")
		}
	} else {
		graph = gitserver.ParseCommitGraph(nil)
	}

	tipCommit, err := u.gitserverClient.Head(ctx, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	// Decorate the commit graph with the set of processed uploads are visible from each commit,
	// then bulk update the denormalized view in Postgres. We call this with an empty graph as well
	// so that we end up clearing the stale data and bulk inserting nothing.
	if err := u.dbStore.CalculateVisibleUploads(ctx, repositoryID, graph, tipCommit, dirtyToken); err != nil {
		return errors.Wrap(err, "store.CalculateVisibleUploads")
	}

	return nil
}
