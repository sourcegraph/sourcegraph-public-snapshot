package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	ctx, _, endObservation := s.operations.insertDependencyIndexingJob.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("uploadId", uploadID),
		attribute.String("extSvcKind", externalServiceKind),
	}})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		insertDependencyIndexingJobQuery,
		uploadID,
		externalServiceKind,
		syncTime,
	)))
	return id, err
}

const insertDependencyIndexingJobQuery = `
INSERT INTO lsif_dependency_indexing_jobs (upload_id, external_service_kind, external_service_sync)
VALUES (%s, %s, %s)
RETURNING id
`

func (s *store) QueueRepoRev(ctx context.Context, repositoryID int, rev string) (err error) {
	ctx, _, endObservation := s.operations.queueRepoRev.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		isQueued, err := tx.IsQueued(ctx, repositoryID, rev)
		if err != nil {
			return err
		}
		if isQueued {
			return nil
		}

		return tx.db.Exec(ctx, sqlf.Sprintf(queueRepoRevQuery, repositoryID, rev))
	})
}

const queueRepoRevQuery = `
INSERT INTO codeintel_autoindex_queue (repository_id, rev)
VALUES (%s, %s)
ON CONFLICT DO NOTHING
`
