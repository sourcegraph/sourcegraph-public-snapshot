package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	ctx, _, endObservation := s.operations.insertDependencyIndexingJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadId", uploadID),
		log.String("extSvcKind", externalServiceKind),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(insertDependencyIndexingJobQuery, uploadID, externalServiceKind, syncTime)))
	return id, err
}

const insertDependencyIndexingJobQuery = `
INSERT INTO lsif_dependency_indexing_jobs (upload_id, external_service_kind, external_service_sync)
VALUES (%s, %s, %s)
RETURNING id
`
