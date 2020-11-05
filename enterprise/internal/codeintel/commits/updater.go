package commits

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// Updater calculates, denormalizes, and stores the set of uploads visible from every commit
// for a given repository. A repository's commit graph is updated when we receive code intel
// queries for a commit we are unaware of (a commit newer than our latest LSIF upload), and
// after processing an upload for a repository.
type Updater interface {
	// TryUpdate pulls the commit graph for the given repository from gitserver, pulls the set
	// of LSIF upload objects for the given repository from Postgres, and correlates them into a
	// visibility graph. This graph is then upserted back into Postgres for use by find closest
	// dumps queries.
	//
	// This method will attempt to acquire an advisory lock to give exclusive access to the update
	// procedure for this repository. If the lock is already held, this method will simply return
	// early. The user should supply a dirty token that is associated with the given repository so
	// that the repository can be unmarked as long as the repository is not marked as dirty again
	// before the update completes.
	TryUpdate(ctx context.Context, repositoryID, dirtyToken int) error
}

type updater struct {
	store           store.Store
	gitserverClient gitserverClient
}

type gitserverClient interface {
	Head(ctx context.Context, store store.Store, repositoryID int) (string, error)
	CommitGraph(ctx context.Context, store store.Store, repositoryID int, options gitserver.CommitGraphOptions) (map[string][]string, error)
}

func NewUpdater(store store.Store, gitserverClient gitserverClient) Updater {
	return &updater{
		store:           store,
		gitserverClient: gitserverClient,
	}
}

// Try Update pulls the commit graph for the given repository from gitserver, pulls the set
// of LSIF upload objects for the given repository from Postgres, and correlates them into a
// visibility graph. This graph is then upserted back into Postgres for use by find closest
// dumps queries.
//
// This method will attempt to acquire an advisory lock to give exclusive access to the update
// procedure for this repository. If the lock is already held, this method will simply return
// early. The user should supply a dirty token that is associated with the given repository so
// that the repository can be unmarked as long as the repository is not marked as dirty again
// before the update completes.
func (u *updater) TryUpdate(ctx context.Context, repositoryID, dirtyToken int) error {
	ok, unlock, err := u.store.Lock(ctx, repositoryID, false)
	if err != nil || !ok {
		return errors.Wrap(err, "store.Lock")
	}
	defer func() {
		err = unlock(err)
	}()

	return u.update(ctx, repositoryID, dirtyToken)
}

func (u *updater) update(ctx context.Context, repositoryID, dirtyToken int) error {
	graph, err := u.gitserverClient.CommitGraph(ctx, u.store, repositoryID, gitserver.CommitGraphOptions{})
	if err != nil {
		return errors.Wrap(err, "gitserver.CommitGraph")
	}

	tipCommit, err := u.gitserverClient.Head(ctx, u.store, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	if err := u.store.CalculateVisibleUploads(ctx, repositoryID, graph, tipCommit, dirtyToken); err != nil {
		return errors.Wrap(err, "store.CalculateVisibleUploads")
	}

	return nil
}
