package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetPathExists determines if the path exists in the database.
func (s *store) GetPathExists(ctx context.Context, bundleID int, path string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.getExists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	_, exists, err := basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(existsQuery, bundleID, path)))
	return exists, err
}

const existsQuery = `
SELECT path FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`
