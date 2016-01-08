package local

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
)

var Changesets sourcegraph.ChangesetsServer = &changesets{}

var _ sourcegraph.ChangesetsServer = (*changesets)(nil)

type changesets struct{}

func (s *changesets) Create(ctx context.Context, op *sourcegraph.ChangesetCreateOp) (*sourcegraph.Changeset, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.Create"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

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

	eventsutil.LogCreateChangeset(ctx)

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

	// The git server has fired off hook events, wait to handle them in
	// changesets_update.go and mark the CS as merged, this way we can return the
	// new CS to the user (which the frontend renders for immediate feedback).
	timeout := time.After(10 * time.Second)
	for {
		// List the events.
		events, err := store.ChangesetsFromContext(ctx).ListEvents(ctx, &sourcegraph.ChangesetSpec{
			ID:   op.ID,
			Repo: op.Repo,
		})
		if err != nil {
			return nil, err
		}
		for _, ev := range events.Events {
			if ev.After.Merged {
				return ev, nil
			}
		}

		select {
		case <-timeout:
			return nil, errors.New("timeout while waiting for changeset merged event")
		default:
			// Wait while the event is handled.
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (s *changesets) List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
	return store.ChangesetsFromContext(ctx).List(ctx, op)
}

func (s *changesets) ListEvents(ctx context.Context, spec *sourcegraph.ChangesetSpec) (*sourcegraph.ChangesetEventList, error) {
	return store.ChangesetsFromContext(ctx).ListEvents(ctx, spec)
}
