package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// GetPackage returns the dump that provides the package with the given scheme, name, and version and a flag indicating its existence.
func (db *dbImpl) GetPackage(ctx context.Context, scheme, name, version string) (Dump, bool, error) {
	query := `
		SELECT
			d.id,
			d.commit,
			d.root,
			d.visible_at_tip,
			d.uploaded_at,
			d.state,
			d.failure_summary,
			d.failure_stacktrace,
			d.started_at,
			d.finished_at,
			d.tracing_context,
			d.repository_id,
			d.indexer
		FROM lsif_packages p
		JOIN lsif_dumps d ON p.dump_id = d.id
		WHERE p.scheme = %s AND p.name = %s AND p.version = %s
		LIMIT 1
	`

	dump, err := scanDump(db.queryRow(ctx, sqlf.Sprintf(query, scheme, name, version)))
	if err != nil {
		return Dump{}, false, ignoreErrNoRows(err)
	}

	return dump, true, nil
}

// UpdatePackages upserts package data tied to the given upload.
func (db *dbImpl) UpdatePackages(ctx context.Context, tx *sql.Tx, packages []types.Package) (err error) {
	if len(packages) == 0 {
		return nil
	}

	if tx == nil {
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			err = closeTx(tx, err)
		}()
	}
	tw := &transactionWrapper{tx}

	if tw == nil {
		tw, err = db.beginTx(ctx)
		if err != nil {
			return err
		}
		defer func() {
			err = closeTx(tw.tx, err)
		}()
	}

	query := `
		INSERT INTO lsif_packages (dump_id, scheme, name, version)
		VALUES %s
		ON CONFLICT DO NOTHING
	`

	var values []*sqlf.Query
	for _, p := range packages {
		values = append(values, sqlf.Sprintf("(%s, %s, %s, %s)", p.DumpID, p.Scheme, p.Name, p.Version))
	}

	_, err = tw.exec(ctx, sqlf.Sprintf(query, sqlf.Join(values, ",")))
	return err
}
