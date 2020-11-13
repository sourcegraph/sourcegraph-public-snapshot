package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/db/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetPackage returns the dump that provides the package with the given scheme, name, and version and a flag indicating its existence.
func (s *Store) GetPackage(ctx context.Context, scheme, name, version string) (_ Dump, _ bool, err error) {
	ctx, endObservation := s.operations.getPackage.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", scheme),
		log.String("name", name),
		log.String("version", version),
	}})
	defer endObservation(1, observation.Args{})

	return scanFirstDump(s.Store.Query(ctx, sqlf.Sprintf(`
		SELECT
			d.id,
			d.commit,
			d.root,
			EXISTS (SELECT 1 FROM lsif_uploads_visible_at_tip where repository_id = d.repository_id and upload_id = d.id) AS visible_at_tip,
			d.uploaded_at,
			d.state,
			d.failure_message,
			d.started_at,
			d.finished_at,
			d.process_after,
			d.num_resets,
			d.num_failures,
			d.repository_id,
			d.repository_name,
			d.indexer
		FROM lsif_packages p
		JOIN lsif_dumps_with_repository_name d ON d.id = p.dump_id
		WHERE p.scheme = %s AND p.name = %s AND p.version = %s
		ORDER BY d.uploaded_at DESC
		LIMIT 1
	`, scheme, name, version)))
}

// UpdatePackages upserts package data tied to the given upload.
func (s *Store) UpdatePackages(ctx context.Context, packages []lsifstore.Package) (err error) {
	ctx, endObservation := s.operations.updatePackages.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	if len(packages) == 0 {
		return nil
	}

	inserter := batch.NewBatchInserter(ctx, s.Store.Handle().DB(), "lsif_packages", "dump_id", "scheme", "name", "version")
	for _, p := range packages {
		if err := inserter.Insert(ctx, p.DumpID, p.Scheme, p.Name, p.Version); err != nil {
			return err
		}
	}

	return inserter.Flush(ctx)
}
