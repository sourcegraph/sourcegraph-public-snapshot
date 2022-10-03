package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ReferencesForUpload returns the set of import monikers attached to the given upload identifier.
func (s *Store) ReferencesForUpload(ctx context.Context, uploadID int) (_ PackageReferenceScanner, err error) {
	ctx, _, endObservation := s.operations.referencesForUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.Query(ctx, sqlf.Sprintf(referencesForUploadQuery, uploadID))
	if err != nil {
		return nil, err
	}

	return packageReferenceScannerFromRows(rows), nil
}

const referencesForUploadQuery = `
-- source: internal/codeintel/stores/dbstore/xrepo.go:ReferencesForUpload
SELECT r.dump_id, r.scheme, r.name, r.version
FROM lsif_references r
WHERE dump_id = %s
ORDER BY r.scheme, r.name, r.version
`
