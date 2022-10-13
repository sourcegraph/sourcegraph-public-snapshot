package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InsertDependencySyncingJob inserts a new dependency syncing job and returns its identifier.
func (s *store) InsertDependencySyncingJob(ctx context.Context, uploadID int) (id int, err error) {
	ctx, _, endObservation := s.operations.insertDependencySyncingJob.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(insertDependencySyncingJobQuery, uploadID)))
	return id, err
}

const insertDependencySyncingJobQuery = `
INSERT INTO lsif_dependency_syncing_jobs (upload_id) VALUES (%s)
RETURNING id
`
