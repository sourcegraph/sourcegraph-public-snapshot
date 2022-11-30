package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// UpdatePackageReferences inserts reference data tied to the given upload.
func (s *store) UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error) {
	ctx, _, endObservation := s.operations.updatePackageReferences.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numReferences", len(references)),
	}})
	defer endObservation(1, observation.Args{})

	if len(references) == 0 {
		return nil
	}

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_references without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(updateReferencesTemporaryTableQuery)); err != nil {
		return err
	}

	// Bulk insert all the unique column values into the temporary table
	if err := batch.InsertValues(
		ctx,
		tx.Handle(),
		"t_lsif_references",
		batch.MaxNumPostgresParameters,
		[]string{"scheme", "name", "version"},
		loadReferencesChannel(references),
	); err != nil {
		return err
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized idump id here since it is the same for all rows in this operation.
	return tx.Exec(ctx, sqlf.Sprintf(updateReferencesInsertQuery, dumpID))
}

const updateReferencesTemporaryTableQuery = `
CREATE TEMPORARY TABLE t_lsif_references (
	scheme text NOT NULL,
	name text NOT NULL,
	version text NOT NULL
) ON COMMIT DROP
`

const updateReferencesInsertQuery = `
INSERT INTO lsif_references (dump_id, scheme, name, version)
SELECT %s, source.scheme, source.name, source.version
FROM t_lsif_references source
`

func loadReferencesChannel(references []precise.PackageReference) <-chan []any {
	ch := make(chan []any, len(references))

	go func() {
		defer close(ch)

		for _, r := range references {
			ch <- []any{r.Scheme, r.Name, r.Version}
		}
	}()

	return ch
}

// ReferencesForUpload returns the set of import monikers attached to the given upload identifier.
func (s *store) ReferencesForUpload(ctx context.Context, uploadID int) (_ shared.PackageReferenceScanner, err error) {
	ctx, _, endObservation := s.operations.referencesForUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(referencesForUploadQuery, uploadID))
	if err != nil {
		return nil, err
	}

	return shared.PackageReferenceScannerFromRows(rows), nil
}

const referencesForUploadQuery = `
SELECT r.dump_id, r.scheme, r.name, r.version
FROM lsif_references r
WHERE dump_id = %s
ORDER BY r.scheme, r.name, r.version
`
