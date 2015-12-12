package local

import (
	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var Changesets sourcegraph.ChangesetsServer = &changesets{}

var _ sourcegraph.ChangesetsServer = (*changesets)(nil)

type changesets struct{}

func (s *changesets) Create(ctx context.Context, op *sourcegraph.ChangesetCreateOp) (*sourcegraph.Changeset, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.Create"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	{
		// TODO(x): Do this after pushing any branch instead?
		// Enqueue builds (if they don't yet exist) for the newly
		// created changeset's base and head.
		//
		// Do this before creating the changeset in case this step
		// fails.
		enqueueBuild := func(rev sourcegraph.RepoRevSpec) error {
			// Builds.Create requires a fully resolved RepoRevSpec
			commit, err := svc.Repos(ctx).GetCommit(ctx, &rev)
			if err != nil {
				return err
			}
			rev.CommitID = string(commit.ID)
			_, err = svc.Builds(ctx).Create(ctx, &sourcegraph.BuildsCreateOp{
				RepoRev: rev,
				Opt:     &sourcegraph.BuildCreateOptions{BuildConfig: sourcegraph.BuildConfig{Queue: true}},
			})
			return err
		}
		par := parallel.NewRun(2)
		par.Do(func() error { return enqueueBuild(op.Changeset.DeltaSpec.Base) })
		par.Do(func() error { return enqueueBuild(op.Changeset.DeltaSpec.Head) })
		err := par.Wait()
		if err != nil {
			return nil, err
		}
	}

	if err := store.ChangesetsFromContext(ctx).Create(ctx, op.Repo.URI, op.Changeset); err != nil {
		return nil, err
	}

	events.Publish(events.ChangesetCreateEvent, events.ChangesetPayload{
		Actor:     authpkg.UserSpecFromContext(ctx),
		ID:        op.Changeset.ID,
		Repo:      op.Repo.URI,
		Title:     op.Changeset.Title,
		Changeset: op.Changeset,
	})

	return op.Changeset, nil
}

func (s *changesets) Get(ctx context.Context, op *sourcegraph.ChangesetSpec) (*sourcegraph.Changeset, error) {
	return store.ChangesetsFromContext(ctx).Get(ctx, op.Repo.URI, op.ID)
}

func (s *changesets) CreateReview(ctx context.Context, op *sourcegraph.ChangesetCreateReviewOp) (*sourcegraph.ChangesetReview, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.CreateReview"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	review, err := store.ChangesetsFromContext(ctx).CreateReview(ctx, op.Repo.URI, op.ChangesetID, op.Review)
	if err != nil {
		return nil, err
	}

	events.Publish(events.ChangesetReviewEvent, events.ChangesetPayload{
		Actor:  authpkg.UserSpecFromContext(ctx),
		ID:     op.ChangesetID,
		Repo:   op.Repo.URI,
		Review: review,
	})

	return review, err
}

func (s *changesets) ListReviews(ctx context.Context, op *sourcegraph.ChangesetListReviewsOp) (*sourcegraph.ChangesetReviewList, error) {
	return store.ChangesetsFromContext(ctx).ListReviews(ctx, op.Repo.URI, op.ChangesetID)
}

func (s *changesets) Merge(ctx context.Context, op *sourcegraph.ChangesetMergeOp) (*sourcegraph.ChangesetEvent, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.Merge"); err != nil {
		return nil, err
	}

	err := store.ChangesetsFromContext(ctx).Merge(ctx, op)
	if err != nil {
		return nil, err
	}

	// TODO(slimsag): use pbtypes.Void instead
	return &sourcegraph.ChangesetEvent{}, nil
}

func (s *changesets) List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
	return store.ChangesetsFromContext(ctx).List(ctx, op)
}

func (s *changesets) ListEvents(ctx context.Context, spec *sourcegraph.ChangesetSpec) (*sourcegraph.ChangesetEventList, error) {
	return store.ChangesetsFromContext(ctx).ListEvents(ctx, spec)
}
