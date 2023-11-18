package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.isQueued.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	isQueued, _, err := basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		isQueuedQuery,
		repositoryID, commit,
		repositoryID, commit,
	)))
	return isQueued, err
}

const isQueuedQuery = `
-- The query has two parts, 'A' UNION 'B', where 'A' is true if there's a manual and
-- reachable upload for a repo/commit pair. This signifies that the user has configured
-- manual indexing on a repo and we shouldn't clobber it with autoindexing. The other
-- query 'B' is true if there's an auto-index record already enqueued for this repo. This
-- signifies that we've already infered jobs for this repo/commit pair so we can skip it
-- (we should infer the same jobs).

-- We added a way to say "you might infer different jobs" for part 'B' by adding the
-- check on u.should_reindex. We're now adding a way to say "the indexer might result
-- in a different output_ for part A, allowing auto-indexing to clobber records that
-- have undergone some possibly lossy transformation (like LSIF -> SCIP conversion in-db).
SELECT
	EXISTS (
		SELECT 1
		FROM lsif_uploads u
		WHERE
			repository_id = %s AND
			commit = %s AND
			state NOT IN ('deleting', 'deleted') AND
			associated_index_id IS NULL AND
			NOT u.should_reindex
	)

	OR

	-- We want IsQueued to return true when there exists auto-indexing job records
	-- and none of them are marked for reindexing. If we have one or more rows and
	-- ALL of them are not marked for re-indexing, we'll block additional indexing
	-- attempts.
	(
		SELECT COALESCE(bool_and(NOT should_reindex), false)
		FROM (
			-- For each distinct (root, indexer) pair, use the most recently queued
			-- index as the authoritative attempt.
			SELECT DISTINCT ON (root, indexer) should_reindex
			FROM lsif_indexes
			WHERE repository_id = %s AND commit = %s
			ORDER BY root, indexer, queued_at DESC
		) _
	)
`

func (s *store) IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.isQueuedRootIndexer.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("commit", commit),
		attribute.String("root", root),
		attribute.String("indexer", indexer),
	}})
	defer endObservation(1, observation.Args{})

	isQueued, _, err := basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		isQueuedRootIndexerQuery,
		repositoryID,
		commit,
		root,
		indexer,
	)))
	return isQueued, err
}

const isQueuedRootIndexerQuery = `
SELECT NOT should_reindex
FROM lsif_indexes
WHERE
	repository_id  = %s AND
	commit = %s AND
	root = %s AND
	indexer = %s
ORDER BY queued_at DESC
LIMIT 1
`

// TODO (ideas):
// - batch insert
// - canonization methods
// - share code with uploads store (should own this?)

func (s *store) InsertIndexes(ctx context.Context, indexes []shared.Index) (_ []shared.Index, err error) {
	ctx, _, endObservation := s.operations.insertIndexes.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numIndexes", len(indexes)),
	}})
	endObservation(1, observation.Args{})

	if len(indexes) == 0 {
		return nil, nil
	}

	actor := actor.FromContext(ctx)

	values := make([]*sqlf.Query, 0, len(indexes))
	for _, index := range indexes {
		if index.DockerSteps == nil {
			index.DockerSteps = []shared.DockerStep{}
		}
		if index.LocalSteps == nil {
			index.LocalSteps = []string{}
		}
		if index.IndexerArgs == nil {
			index.IndexerArgs = []string{}
		}

		values = append(values, sqlf.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
			index.State,
			index.Commit,
			index.RepositoryID,
			pq.Array(index.DockerSteps),
			pq.Array(index.LocalSteps),
			index.Root,
			index.Indexer,
			pq.Array(index.IndexerArgs),
			index.Outfile,
			pq.Array(index.ExecutionLogs),
			pq.Array(index.RequestedEnvVars),
			actor.UID,
		))
	}

	indexes = []shared.Index{}
	err = s.withTransaction(ctx, func(tx *store) error {
		ids, err := basestore.ScanInts(tx.db.Query(ctx, sqlf.Sprintf(insertIndexQuery, sqlf.Join(values, ","))))
		if err != nil {
			return err
		}

		s.operations.indexesInserted.Add(float64(len(ids)))

		authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
		if err != nil {
			return err
		}

		queries := make([]*sqlf.Query, 0, len(ids))
		for _, id := range ids {
			queries = append(queries, sqlf.Sprintf("%d", id))
		}

		indexes, err = scanIndexes(tx.db.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(queries, ", "), authzConds)))
		return err
	})

	return indexes, err
}

const insertIndexQuery = `
INSERT INTO lsif_indexes (
	state,
	commit,
	repository_id,
	docker_steps,
	local_steps,
	root,
	indexer,
	indexer_args,
	outfile,
	execution_logs,
	requested_envvars,
	enqueuer_user_id
)
VALUES %s
RETURNING id
`

const getIndexesByIDsQuery = `
SELECT
	u.id,
	u.commit,
	u.queued_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	repo.name,
	u.docker_steps,
	u.root,
	u.indexer,
	u.indexer_args,
	u.outfile,
	u.execution_logs,
	s.rank,
	u.local_steps,
	(SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id) AS associated_upload_id,
	u.should_reindex,
	u.requested_envvars,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (
	SELECT
		r.id,
		ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.queued_at), r.id) as rank
	FROM lsif_indexes_with_repository_name r
	WHERE r.state = 'queued'
) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND u.id IN (%s) AND %s
ORDER BY u.id
`

//
//

func scanIndex(s dbutil.Scanner) (index shared.Index, err error) {
	var executionLogs []executor.ExecutionLogEntry
	if err := s.Scan(
		&index.ID,
		&index.Commit,
		&index.QueuedAt,
		&index.State,
		&index.FailureMessage,
		&index.StartedAt,
		&index.FinishedAt,
		&index.ProcessAfter,
		&index.NumResets,
		&index.NumFailures,
		&index.RepositoryID,
		&index.RepositoryName,
		pq.Array(&index.DockerSteps),
		&index.Root,
		&index.Indexer,
		pq.Array(&index.IndexerArgs),
		&index.Outfile,
		pq.Array(&executionLogs),
		&index.Rank,
		pq.Array(&index.LocalSteps),
		&index.AssociatedUploadID,
		&index.ShouldReindex,
		pq.Array(&index.RequestedEnvVars),
		&index.EnqueuerUserID,
	); err != nil {
		return index, err
	}

	index.ExecutionLogs = append(index.ExecutionLogs, executionLogs...)
	return index, nil
}

var scanIndexes = basestore.NewSliceScanner(scanIndex)
