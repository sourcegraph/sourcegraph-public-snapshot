package dbstore

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
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
	Next() (shared.PackageReference, bool, error)

	// Close the underlying row object.
	Close() error
}

type rowScanner struct {
	rows *sql.Rows
}

// packageReferenceScannerFromRows creates a PackageReferenceScanner that feeds the given values.
func packageReferenceScannerFromRows(rows *sql.Rows) PackageReferenceScanner {
	return &rowScanner{
		rows: rows,
	}
}

// Next reads the next package reference value from the database cursor.
func (s *rowScanner) Next() (reference shared.PackageReference, _ bool, _ error) {
	if !s.rows.Next() {
		return shared.PackageReference{}, false, nil
	}

	if err := s.rows.Scan(
		&reference.DumpID,
		&reference.Scheme,
		&reference.Name,
		&reference.Version,
	); err != nil {
		return shared.PackageReference{}, false, err
	}

	return reference, true, nil
}

// Close the underlying row object.
func (s *rowScanner) Close() error {
	return basestore.CloseRows(s.rows, nil)
}

type sliceScanner struct {
	references []shared.PackageReference
}

// PackageReferenceScannerFromSlice creates a PackageReferenceScanner that feeds the given values.
func PackageReferenceScannerFromSlice(references ...shared.PackageReference) PackageReferenceScanner {
	return &sliceScanner{
		references: references,
	}
}

func (s *sliceScanner) Next() (shared.PackageReference, bool, error) {
	if len(s.references) == 0 {
		return shared.PackageReference{}, false, nil
	}

	next := s.references[0]
	s.references = s.references[1:]
	return next, true, nil
}

func (s *sliceScanner) Close() error {
	return nil
}
