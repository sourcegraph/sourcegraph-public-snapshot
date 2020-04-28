package db

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// Scanner is the common interface shared by *sql.Row and *sql.Rows.
type Scanner interface {
	// Scan copies the values of the current row into the values pointed at by dest.
	Scan(dest ...interface{}) error
}

// scanDump populates a Dump value from the given scanner.
func scanDump(scanner Scanner) (dump Dump, err error) {
	err = scanner.Scan(
		&dump.ID,
		&dump.Commit,
		&dump.Root,
		&dump.VisibleAtTip,
		&dump.UploadedAt,
		&dump.State,
		&dump.FailureSummary,
		&dump.FailureStacktrace,
		&dump.StartedAt,
		&dump.FinishedAt,
		&dump.TracingContext,
		&dump.RepositoryID,
		&dump.Indexer,
	)
	return dump, err
}

// scanDumps reads the given set of dump rows and returns a slice of resulting values.
// This method should be called directly with the return value of `*db.query`.
func scanDumps(rows *sql.Rows, err error) ([]Dump, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dumps []Dump
	for rows.Next() {
		dump, err := scanDump(rows)
		if err != nil {
			return nil, err
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}

// scanUpload populates an Upload value from the given scanner.
func scanUpload(scanner Scanner) (upload Upload, err error) {
	err = scanner.Scan(
		&upload.ID,
		&upload.Commit,
		&upload.Root,
		&upload.VisibleAtTip,
		&upload.UploadedAt,
		&upload.State,
		&upload.FailureSummary,
		&upload.FailureStacktrace,
		&upload.StartedAt,
		&upload.FinishedAt,
		&upload.TracingContext,
		&upload.RepositoryID,
		&upload.Indexer,
		&upload.Rank,
	)
	return upload, err
}

// scanUploads reads the given set of upload rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.query`.
func scanUploads(rows *sql.Rows, err error) ([]Upload, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []Upload
	for rows.Next() {
		upload, err := scanUpload(rows)
		if err != nil {
			return nil, err
		}

		uploads = append(uploads, upload)
	}

	return uploads, nil
}

// scanPackageReference populates a package reference value from the given scanner.
func scanPackageReference(scanner Scanner) (reference types.PackageReference, err error) {
	err = scanner.Scan(&reference.DumpID, &reference.Scheme, &reference.Name, &reference.Version, &reference.Filter)
	return reference, err
}

// scanPackageReferences reads the given set of reference rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.queryRows`.
func scanPackageReferences(rows *sql.Rows, err error) ([]types.PackageReference, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []types.PackageReference
	for rows.Next() {
		reference, err := scanPackageReference(rows)
		if err != nil {
			return nil, err
		}

		references = append(references, reference)
	}

	return references, nil
}

// scanString populates a string value from the given scanner.
func scanString(scanner Scanner) (value string, err error) {
	err = scanner.Scan(&value)
	return value, err
}

// scanInt populates an integer value from the given scanner.
func scanInt(scanner Scanner) (value int, err error) {
	err = scanner.Scan(&value)
	return value, err
}

// scanInts reads the given set of `(int)` rows and returns a slice of resulting values.
// This method should be called directly with the return value of `*db.query`.
func scanInts(rows *sql.Rows, err error) ([]int, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []int
	for rows.Next() {
		value, err := scanInt(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// scanState populates an integer and string from the given scanner.
func scanState(scanner Scanner) (repositoryID int, state string, err error) {
	err = scanner.Scan(&repositoryID, &state)
	return repositoryID, state, err
}

// scanStates reads the given set of `(id, state)` rows and returns a map from id to its
// state. This method should be called directly with the return value of `*db.query`.
func scanStates(rows *sql.Rows, err error) (map[int]string, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := map[int]string{}
	for rows.Next() {
		repositoryID, state, err := scanState(rows)
		if err != nil {
			return nil, err
		}

		states[repositoryID] = state
	}

	return states, nil
}

// scanVisibility populates an integer and boolean from the given scanner.
func scanVisibility(scanner Scanner) (repositoryID int, visibleAtTip bool, err error) {
	err = scanner.Scan(&repositoryID, &visibleAtTip)
	return repositoryID, visibleAtTip, err
}

// scanVisibilities reads the given set of `(id, visible_at_tip)` rows and returns a map
// from id to its visibility. This method should be called directly with the return value
// of `*db.query`.
func scanVisibilities(rows *sql.Rows, err error) (map[int]bool, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	visibilities := map[int]bool{}
	for rows.Next() {
		repositoryID, visibleAtTip, err := scanVisibility(rows)
		if err != nil {
			return nil, err
		}

		visibilities[repositoryID] = visibleAtTip
	}

	return visibilities, nil
}
