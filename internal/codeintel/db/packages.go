package db

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// GetPackage returns the dump that provides the package with the given scheme, name, and version and a flag indicating its existence.
func (db *dbImpl) GetPackage(ctx context.Context, scheme, name, version string) (Dump, bool, error) {
	return scanFirstDump(db.query(ctx, sqlf.Sprintf(`
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
	`, scheme, name, version)))
}

// UpdatePackages upserts package data tied to the given upload.
func (db *dbImpl) UpdatePackages(ctx context.Context, packages []types.Package) (err error) {
	if len(packages) == 0 {
		return nil
	}

	var values []*sqlf.Query
	for _, p := range packages {
		values = append(values, sqlf.Sprintf("(%s, %s, %s, %s)", p.DumpID, p.Scheme, p.Name, p.Version))
	}

	return db.exec(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_packages (dump_id, scheme, name, version)
		VALUES %s
		ON CONFLICT DO NOTHING
	`, sqlf.Join(values, ",")))
}
