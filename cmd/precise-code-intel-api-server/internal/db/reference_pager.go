package db

import "database/sql"

// ReferencePager holds state for a reference result in a SQL transaction. Each page
// requested should have a consistent view into the database.
type ReferencePager interface {
	TxCloser

	// PageFromOffset returns the page of references that starts at the given offset.
	PageFromOffset(offset int) ([]Reference, error)
}

type referencePager struct {
	*txCloser
	pageFromOffset func(offset int) ([]Reference, error)
}

// PageFromOffset returns the page of references that starts at the given offset.
func (rp *referencePager) PageFromOffset(offset int) ([]Reference, error) {
	return rp.pageFromOffset(offset)
}

func newReferencePager(tx *sql.Tx, pageFromOffset func(offset int) ([]Reference, error)) ReferencePager {
	return &referencePager{
		txCloser:       &txCloser{tx},
		pageFromOffset: pageFromOffset,
	}
}

func newEmptyReferencePager(tx *sql.Tx) ReferencePager {
	return newReferencePager(tx, func(offset int) ([]Reference, error) {
		return nil, nil
	})
}
