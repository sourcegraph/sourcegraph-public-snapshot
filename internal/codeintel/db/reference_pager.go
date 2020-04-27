package db

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// ReferencePager holds state for a reference result in a SQL transaction. Each page
// requested should have a consistent view into the database.
type ReferencePager interface {
	TxCloser

	// PageFromOffset returns the page of package references that starts at the given offset.
	PageFromOffset(offset int) ([]types.PackageReference, error)
}

type referencePager struct {
	*txCloser
	pageFromOffset func(offset int) ([]types.PackageReference, error)
}

// PageFromOffset returns the page of package references that starts at the given offset.
func (rp *referencePager) PageFromOffset(offset int) ([]types.PackageReference, error) {
	return rp.pageFromOffset(offset)
}

func newReferencePager(tx *sql.Tx, pageFromOffset func(offset int) ([]types.PackageReference, error)) ReferencePager {
	return &referencePager{
		txCloser:       &txCloser{tx},
		pageFromOffset: pageFromOffset,
	}
}

func newEmptyReferencePager(tx *sql.Tx) ReferencePager {
	return newReferencePager(tx, func(offset int) ([]types.PackageReference, error) {
		return nil, nil
	})
}
