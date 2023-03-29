package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InsertIndexes inserts a new index and returns the hydrated index models.
func (s *store) InsertIndexes(ctx context.Context, indexes []types.Index) (_ []types.Index, err error) {
	ctx, _, endObservation := s.operations.insertIndex.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numIndexes", len(indexes)),
		}})
	}()

	if len(indexes) == 0 {
		return nil, nil
	}

	values := make([]*sqlf.Query, 0, len(indexes))
	for _, index := range indexes {
		if index.DockerSteps == nil {
			index.DockerSteps = []types.DockerStep{}
		}
		if index.IndexerArgs == nil {
			index.IndexerArgs = []string{}
		}
		if index.LocalSteps == nil {
			index.LocalSteps = []string{}
		}

		values = append(values, sqlf.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
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
		))
	}

	tx, err := s.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.db.Done(err) }()

	ids, err := basestore.ScanInts(tx.db.Query(ctx, sqlf.Sprintf(insertIndexQuery, sqlf.Join(values, ","))))
	if err != nil {
		return nil, err
	}

	s.operations.indexesInserted.Add(float64(len(ids)))

	return tx.GetIndexesByIDs(ctx, ids...)
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
	requested_envvars
) VALUES %s
RETURNING id
`

// GetIndexesByIDs returns an index for each of the given identifiers. Not all given ids will necessarily
// have a corresponding element in the returned list.
func (s *store) GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error) {
	ctx, _, endObservation := s.operations.getIndexesByIDs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("ids", intsToString(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil, nil
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}

	queries := make([]*sqlf.Query, 0, len(ids))
	for _, id := range ids {
		queries = append(queries, sqlf.Sprintf("%d", id))
	}

	return scanIndexes(s.db.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(queries, ", "), authzConds)))
}

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
	` + indexAssociatedUploadIDQueryFragment + `,
	u.should_reindex,
	u.requested_envvars
FROM lsif_indexes u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND u.id IN (%s) AND %s
ORDER BY u.id
`

// GetLastIndexScanForRepository returns the last timestamp, if any, that the repository with the given
// identifier was considered for auto-indexing scheduling.
func (s *store) GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := s.operations.getLastIndexScanForRepository.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	t, ok, err := basestore.ScanFirstTime(s.db.Query(ctx, sqlf.Sprintf(lastIndexScanForRepositoryQuery, repositoryID)))
	if !ok {
		return nil, err
	}

	return &t, nil
}

const lastIndexScanForRepositoryQuery = `
SELECT last_index_scan_at FROM lsif_last_index_scan WHERE repository_id = %s
`

// IsQueued returns true if there is an index or an upload for the repository and commit.
func (s *store) IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.isQueued.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
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

// IsQueuedRootIndexer returns true if there is an index or an upload for the given (repository, commit, root, indexer).
func (s *store) IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.isQueued.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("root", root),
		log.String("indexer", indexer),
	}})
	defer endObservation(1, observation.Args{})

	isQueued, _, err := basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf(isQueuedRootIndexerQuery, repositoryID, commit, root, indexer)))
	return isQueued, err
}

const isQueuedRootIndexerQuery = `
SELECT NOT should_reindex
FROM lsif_indexes
WHERE
	repository_id  = %s AND
	commit         = %s AND
	root           = %s AND
	indexer        = %s
ORDER BY queued_at DESC
LIMIT 1
`

// QueueRepoRev enqueues the given repository and rev to be processed by the auto-indexing scheduler.
// This method is ultimately used to index on-demand (with deduplication) from transport layers.
func (s *store) QueueRepoRev(ctx context.Context, repositoryID int, rev string) (err error) {
	ctx, _, endObservation := s.operations.queueRepoRev.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	isQueued, err := tx.IsQueued(ctx, repositoryID, rev)
	if err != nil {
		return err
	}
	if isQueued {
		return nil
	}

	return tx.db.Exec(ctx, sqlf.Sprintf(queueRepoRevQuery, repositoryID, rev))
}

const queueRepoRevQuery = `
INSERT INTO codeintel_autoindex_queue (repository_id, rev)
VALUES (%s, %s)
ON CONFLICT DO NOTHING
`

type RepoRev struct {
	ID           int
	RepositoryID int
	Rev          string
}

// GetQueuedRepoRev selects a batch of repository and revisions to be processed by the auto-indexing
// scheduler. If in a transaction, the seleted records will remain locked until the enclosing transaction
// has been committed or rolled back.
func (s *store) GetQueuedRepoRev(ctx context.Context, batchSize int) (_ []RepoRev, err error) {
	ctx, _, endObservation := s.operations.getQueuedRepoRev.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSize", batchSize),
	}})
	defer endObservation(1, observation.Args{})

	return ScanRepoRevs(s.db.Query(ctx, sqlf.Sprintf(getQueuedRepoRevQuery, batchSize)))
}

const getQueuedRepoRevQuery = `
SELECT id, repository_id, rev
FROM codeintel_autoindex_queue
WHERE processed_at IS NULL
ORDER BY queued_at ASC
FOR UPDATE SKIP LOCKED
LIMIT %s
`

// MarkRepoRevsAsProcessed sets processed_at for each matching record in codeintel_autoindex_queue.
func (s *store) MarkRepoRevsAsProcessed(ctx context.Context, ids []int) (err error) {
	ctx, _, endObservation := s.operations.markRepoRevsAsProcessed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numIDs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(markRepoRevsAsProcessedQuery, pq.Array(ids)))
}

const markRepoRevsAsProcessedQuery = `
UPDATE codeintel_autoindex_queue
SET processed_at = NOW()
WHERE id = ANY(%s)
`

// GetRecentIndexesSummary returns the set of "interesting" indexes for the repository with the given identifier.
// The return value is a list of indexes grouped by root and indexer. In each group, the set of indexes should
// include the set of unprocessed records as well as the latest finished record. These values allow users to
// quickly determine if a particular root/indexer pair os up-to-date or having issues processing.
func (s *store) GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error) {
	ctx, logger, endObservation := s.operations.getRecentIndexesSummary.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	indexes, err := scanIndexes(s.db.Query(ctx, sqlf.Sprintf(recentIndexesSummaryQuery, repositoryID, repositoryID)))
	if err != nil {
		return nil, err
	}
	logger.AddEvent("scanIndexes", attribute.Int("numIndexes", len(indexes)))

	groupedIndexes := make([]shared.IndexesWithRepositoryNamespace, 1, len(indexes)+1)
	for _, index := range indexes {
		if last := groupedIndexes[len(groupedIndexes)-1]; last.Root != index.Root || last.Indexer != index.Indexer {
			groupedIndexes = append(groupedIndexes, shared.IndexesWithRepositoryNamespace{
				Root:    index.Root,
				Indexer: index.Indexer,
			})
		}

		n := len(groupedIndexes)
		groupedIndexes[n-1].Indexes = append(groupedIndexes[n-1].Indexes, index)
	}

	return groupedIndexes[1:], nil
}

const sanitizedIndexerExpression = `
(
    split_part(
        split_part(
            CASE
                -- Strip sourcegraph/ prefix if it exists
                WHEN strpos(indexer, 'sourcegraph/') = 1 THEN substr(indexer, length('sourcegraph/') + 1)
                ELSE indexer
            END,
        '@', 1), -- strip off @sha256:...
    ':', 1) -- strip off tag
)
`

const indexAssociatedUploadIDQueryFragment = `
(
	SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id
) AS associated_upload_id
`

const indexRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.queued_at), r.id) as rank
FROM lsif_indexes_with_repository_name r
WHERE r.state = 'queued'
`

const recentIndexesSummaryQuery = `
WITH ranked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_at,
		RANK() OVER (PARTITION BY root, ` + sanitizedIndexerExpression + ` ORDER BY finished_at DESC) AS rank
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.state NOT IN ('queued', 'processing', 'deleted')
),
latest_indexes AS (
	SELECT u.id, u.root, u.indexer, u.queued_at
	FROM lsif_indexes u
	WHERE
		u.id IN (
			SELECT rc.id
			FROM ranked_completed rc
			WHERE rc.rank = 1
		)
	ORDER BY u.root, u.indexer
),
new_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.state IN ('queued', 'processing') AND
		u.queued_at >= (
			SELECT lu.queued_at
			FROM latest_indexes lu
			WHERE
				lu.root = u.root AND
				lu.indexer = u.indexer
			-- condition passes when latest_indexes is empty
			UNION SELECT u.queued_at LIMIT 1
		)
)
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
	u.repository_name,
	u.docker_steps,
	u.root,
	u.indexer,
	u.indexer_args,
	u.outfile,
	u.execution_logs,
	s.rank,
	u.local_steps,
	` + indexAssociatedUploadIDQueryFragment + `,
	u.should_reindex,
	u.requested_envvars
FROM lsif_indexes_with_repository_name u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
WHERE u.id IN (
	SELECT lu.id FROM latest_indexes lu
	UNION
	SELECT nu.id FROM new_indexes nu
)
ORDER BY u.root, u.indexer
`
