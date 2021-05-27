package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InsertDependencyIndexingJob inserts a new dependency indexing job and returns its identifier.
func (s *Store) InsertDependencyIndexingJob(ctx context.Context, uploadID int) (id int, err error) {
	ctx, endObservation := s.operations.insertDependencyIndexingJob.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(insertDependencyIndexingJobQuery, uploadID)))
	return id, err
}

const insertDependencyIndexingJobQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/dependency_index.go:InsertDependencyIndexingJob
INSERT INTO lsif_dependency_indexing_jobs (upload_id) VALUES (%s)
RETURNING id
`
