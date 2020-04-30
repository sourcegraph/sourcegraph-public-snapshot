package db

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// ReferencePager holds state for a reference result in a SQL transaction. Each page
// requested should have a consistent view into the database.
type ReferencePager interface {
	TxCloser

	// PageFromOffset returns the page of package references that starts at the given offset.
	PageFromOffset(ctx context.Context, offset int) ([]types.PackageReference, error)
}

type PageFromOffsetFn func(ctx context.Context, offset int) ([]types.PackageReference, error)

type referencePager struct {
	*txCloser
	pageFromOffset PageFromOffsetFn
}

// PageFromOffset returns the page of package references that starts at the given offset.
func (rp *referencePager) PageFromOffset(ctx context.Context, offset int) ([]types.PackageReference, error) {
	return rp.pageFromOffset(ctx, offset)
}

func newReferencePager(tx *sql.Tx, pageFromOffset PageFromOffsetFn) ReferencePager {
	return &referencePager{
		txCloser:       &txCloser{tx},
		pageFromOffset: pageFromOffset,
	}
}

func newEmptyReferencePager(tx *sql.Tx) ReferencePager {
	return newReferencePager(tx, func(ctx context.Context, offset int) ([]types.PackageReference, error) {
		return nil, nil
	})
}
