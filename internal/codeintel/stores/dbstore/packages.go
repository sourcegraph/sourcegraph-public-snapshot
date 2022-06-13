package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// UpdatePackages upserts package data tied to the given upload.
func (s *Store) UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) (err error) {
	ctx, _, endObservation := s.operations.updatePackages.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numPackages", len(packages)),
	}})
	defer endObservation(1, observation.Args{})

	if len(packages) == 0 {
		return nil
	}

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_packages without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(updatePackagesTemporaryTableQuery)); err != nil {
		return err
	}

	// Bulk insert all the unique column values into the temporary table
	if err := batch.InsertValues(
		ctx,
		tx.Handle().DBUtilDB(),
		"t_lsif_packages",
		batch.MaxNumPostgresParameters,
		[]string{"scheme", "name", "version"},
		loadPackagesChannel(packages),
	); err != nil {
		return err
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return tx.Exec(ctx, sqlf.Sprintf(updatePackagesInsertQuery, dumpID))
}

const updatePackagesTemporaryTableQuery = `
-- source: internal/codeintel/stores/dbstore/packages.go:UpdatePackages
CREATE TEMPORARY TABLE t_lsif_packages (
	scheme text NOT NULL,
	name text NOT NULL,
	version text NOT NULL
) ON COMMIT DROP
`

const updatePackagesInsertQuery = `
-- source: internal/codeintel/stores/dbstore/packages.go:UpdatePackages
INSERT INTO lsif_packages (dump_id, scheme, name, version)
SELECT %s, source.scheme, source.name, source.version
FROM t_lsif_packages source
`

func loadPackagesChannel(packages []precise.Package) <-chan []any {
	ch := make(chan []any, len(packages))

	go func() {
		defer close(ch)

		for _, p := range packages {
			ch <- []any{p.Scheme, p.Name, p.Version}
		}
	}()

	return ch
}
