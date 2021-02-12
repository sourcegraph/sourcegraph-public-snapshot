package dbstore

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// PackageReferenceScanner allows for on-demand scanning of PackageReference values.
//
// A scanner for this type was introduced as a memory optimization. Instead of reading a
// large number of large byte arrays into memory at once, we allow the user to request
// the next filter value when they are ready to process it. This allows us to hold only
// a single bloom filter in memory at any give time during reference requests.
type PackageReferenceScanner interface {
	// Next reads the next package reference value from the database cursor.
	Next() (lsifstore.PackageReference, bool, error)

	// Close the underlying row object.
	Close() error
}

type packageReferenceScanner struct {
	rows *sql.Rows
}

// Next reads the next package reference value from the database cursor.
func (s *packageReferenceScanner) Next() (reference lsifstore.PackageReference, _ bool, _ error) {
	if !s.rows.Next() {
		return lsifstore.PackageReference{}, false, nil
	}

	if err := s.rows.Scan(&reference.DumpID, &reference.Filter); err != nil {
		return lsifstore.PackageReference{}, false, err
	}

	return reference, true, nil
}

// Close the underlying row object.
func (s *packageReferenceScanner) Close() error {
	return basestore.CloseRows(s.rows, nil)
}
