package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// ChangesetUpdateOp contains the configuration needed for a store update
// operation. It accepts the public API's update options, as well as two fields
// not for external use.
//
// Internally, the server keeps track of the base and head revisions via git
// post-push hooks and updates the store. The store will create git refs towards
// these commits in order to persist history after a changeset is closed, merged
// or deleted.
//
// Base and Head are internal states and should not be part of the public API.
type ChangesetUpdateOp struct {
	// Op is the update operation that will be applied to a changeset's public
	// properties.
	Op *sourcegraph.ChangesetUpdateOp

	// Base, when set, will update the Changeset's base revision CommitID to this
	// value. From this point forward, the changeset will use this value when
	// computing diffs.
	Base string

	// Head, when set, will update the Changeset's head revision CommitID to this
	// value. From this point forward, the changeset will use this value when
	// computing diffs.
	Head string
}

type Changesets interface {
	// Create creates a new changeset within the given repository. It will
	// alter the Changeset in the parameters by updating its fields (ID, CreatedAt).
	Create(ctx context.Context, repo string, cs *sourcegraph.Changeset) error

	// Get returns the changeset from within the specified repository path
	// having the given ID.
	Get(ctx context.Context, repo string, ID int64) (*sourcegraph.Changeset, error)

	// List lists all changesets for a repository.
	List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error)

	// CreateReview creates a new review on the given changeset. It returns it
	// with the ID field updated.
	CreateReview(ctx context.Context, repo string, changesetID int64, newReview *sourcegraph.ChangesetReview) (*sourcegraph.ChangesetReview, error)

	// ListReviews lists all reviews for a given changeset.
	ListReviews(ctx context.Context, repo string, changesetID int64) (*sourcegraph.ChangesetReviewList, error)

	// Update updates the changeset's properties. Internally, the tracked revisions of
	// head and branch may be updated.
	Update(ctx context.Context, op *ChangesetUpdateOp) (*sourcegraph.ChangesetEvent, error)

	// ListChangesetEvents lists the events in a changeset
	ListEvents(ctx context.Context, spec *sourcegraph.ChangesetSpec) (*sourcegraph.ChangesetEventList, error)
}
