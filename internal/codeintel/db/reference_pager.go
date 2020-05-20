package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// ReferencePager holds state for a reference result within a SQL transaction so that
// each page requested has a consistent view into the database.
type ReferencePager interface {
	// PageFromOffset returns the page of package references that starts at the given offset.
	PageFromOffset(ctx context.Context, offset int) ([]types.PackageReference, error)

	// Done closes the underlying transaction. If the reference pager was called on a
	// Database instance that was already in a transaction, this method does nothing.
	Done(err error) error
}

// PageFromOffsetFn is the function type of ReferencePager's PageFromOffset method.
type PageFromOffsetFn func(ctx context.Context, offset int) ([]types.PackageReference, error)

// noopPageFromOffsetFn is a behaviorless PageFromOffsetFn.
func noopPageFromOffsetFn(ctx context.Context, offset int) ([]types.PackageReference, error) {
	return nil, nil
}

// referencePager is a small struct that conforms to the ReferencePager interface.
type referencePager struct {
	pageFromOffset PageFromOffsetFn
	done           DoneFn
}

// PageFromOffset returns the page of package references that starts at the given offset.
func (rp *referencePager) PageFromOffset(ctx context.Context, offset int) ([]types.PackageReference, error) {
	return rp.pageFromOffset(ctx, offset)
}

// PageFromOffset returns the page of package references that starts at the given offset.
func (rp *referencePager) Done(err error) error {
	return rp.done(err)
}

func newReferencePager(pageFromOffset PageFromOffsetFn, done DoneFn) ReferencePager {
	return &referencePager{pageFromOffset: pageFromOffset, done: done}
}
