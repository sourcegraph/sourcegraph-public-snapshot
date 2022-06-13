package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// UpdatePackageReferences inserts reference data tied to the given upload.
func (s *Store) UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error) {
	ctx, _, endObservation := s.operations.updatePackageReferences.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numReferences", len(references)),
	}})
	defer endObservation(1, observation.Args{})

	if len(references) == 0 {
		return nil
	}

	tx, err := s.transact(ctx)
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
		tx.Handle().DBUtilDB(),
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
-- source: internal/codeintel/stores/dbstore/references.go:UpdatePackageReferences
CREATE TEMPORARY TABLE t_lsif_references (
	scheme text NOT NULL,
	name text NOT NULL,
	version text NOT NULL
) ON COMMIT DROP
`

const updateReferencesInsertQuery = `
-- source: internal/codeintel/stores/dbstore/references.go:UpdatePackageReferences
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
