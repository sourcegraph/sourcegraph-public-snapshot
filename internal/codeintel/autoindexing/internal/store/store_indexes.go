package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InsertIndexes inserts a new index and returns the hydrated index models.
func (s *store) InsertIndexes(ctx context.Context, indexes []shared.Index) (_ []shared.Index, err error) {
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
			index.DockerSteps = []shared.DockerStep{}
		}
		if index.IndexerArgs == nil {
			index.IndexerArgs = []string{}
		}
		if index.LocalSteps == nil {
			index.LocalSteps = []string{}
		}

		values = append(values, sqlf.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
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

	return tx.GetIndexesByIDs(ctx, ids...)
}

const insertIndexQuery = `
-- source: internal/codeintel/stores/dbstore/indexes.go:InsertIndex
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
	execution_logs
) VALUES %s
RETURNING id
`

// GetIndexes returns a list of indexes and the total count of records matching the given conditions.
func (s *store) GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []shared.Index, _ int, err error) {
	ctx, trace, endObservation := s.operations.getIndexes.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("state", opts.State),
		log.String("term", opts.Term),
		log.Int("limit", opts.Limit),
		log.Int("offset", opts.Offset),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.db.Done(err) }()

	var conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeIndexSearchCondition(opts.Term))
	}
	if opts.State != "" {
		conds = append(conds, makeStateCondition(opts.State))
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, tx.db))
	if err != nil {
		return nil, 0, err
	}
	conds = append(conds, authzConds)

	indexes, totalCount, err := scanIndexesWithCount(tx.db.Query(ctx, sqlf.Sprintf(getIndexesQuery, sqlf.Join(conds, " AND "), opts.Limit, opts.Offset)))
	if err != nil {
		return nil, 0, err
	}
	trace.Log(
		log.Int("totalCount", totalCount),
		log.Int("numIndexes", len(indexes)),
	)

	return indexes, totalCount, nil
}

const getIndexesQuery = `
-- source: internal/codeintel/stores/dbstore/indexes.go:GetIndexes
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
	COUNT(*) OVER() AS count
FROM lsif_indexes u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND %s ORDER BY queued_at DESC, u.id LIMIT %d OFFSET %d
`

// makeIndexSearchCondition returns a disjunction of LIKE clauses against all searchable columns of an index.
func makeIndexSearchCondition(term string) *sqlf.Query {
	searchableColumns := []string{
		"u.commit",
		"(u.state)::text",
		"u.failure_message",
		`repo.name`,
		"u.root",
		"u.indexer",
	}

	var termConds []*sqlf.Query
	for _, column := range searchableColumns {
		termConds = append(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

// makeStateCondition returns a disjunction of clauses comparing the upload against the target state.
func makeStateCondition(state string) *sqlf.Query {
	states := make([]string, 0, 2)
	if state == "errored" || state == "failed" {
		// Treat errored and failed states as equivalent
		states = append(states, "errored", "failed")
	} else {
		states = append(states, state)
	}

	queries := make([]*sqlf.Query, 0, len(states))
	for _, state := range states {
		queries = append(queries, sqlf.Sprintf("u.state = %s", state))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(queries, " OR "))
}

// GetIndexByID returns an index by its identifier and boolean flag indicating its existence.
func (s *store) GetIndexByID(ctx context.Context, id int) (_ shared.Index, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return shared.Index{}, false, err
	}

	return scanFirstIndex(s.db.Query(ctx, sqlf.Sprintf(getIndexByIDQuery, id, authzConds)))
}

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

const getIndexByIDQuery = `
-- source: internal/codeintel/stores/dbstore/indexes.go:GetIndexByID
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
	` + indexAssociatedUploadIDQueryFragment + `
FROM lsif_indexes u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND u.id = %s AND %s
`

// GetIndexesByIDs returns an index for each of the given identifiers. Not all given ids will necessarily
// have a corresponding element in the returned list.
func (s *store) GetIndexesByIDs(ctx context.Context, ids ...int) (_ []shared.Index, err error) {
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
-- source: internal/codeintel/stores/dbstore/indexes.go:GetIndexesByIDs
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
	` + indexAssociatedUploadIDQueryFragment + `
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
-- source: internal/codeintel/stores/dbstore/indexes.go:LastIndexScanForRepository
SELECT last_index_scan_at FROM lsif_last_index_scan WHERE repository_id = %s
`

// DeleteIndexByID deletes an index by its identifier.
func (s *store) DeleteIndexByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := s.operations.deleteIndexByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return false, err
	}
	defer func() { err = tx.db.Done(err) }()

	_, exists, err := basestore.ScanFirstInt(tx.db.Query(ctx, sqlf.Sprintf(deleteIndexByIDQuery, id)))
	return exists, err
}

const deleteIndexByIDQuery = `
-- source: internal/codeintel/stores/dbstore/indexes.go:DeleteIndexByID
DELETE FROM lsif_indexes WHERE id = %s RETURNING repository_id
`

// DeletedRepositoryGracePeriod is the minimum allowable duration between
// a repo deletion and index records for that repository being deleted.
const DeletedRepositoryGracePeriod = time.Minute * 30

// DeleteIndexesWithoutRepository deletes indexes associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of indexes
// that were removed for that repository.
func (s *store) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, trace, endObservation := s.operations.deleteIndexesWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// TODO(efritz) - this would benefit from an index on repository_id. We currently have
	// a similar one on this index, but only for uploads that are completed or visible at tip.
	repositories, err := scanCounts(tx.Query(ctx, sqlf.Sprintf(deleteIndexesWithoutRepositoryQuery, now.UTC(), DeletedRepositoryGracePeriod/time.Second)))
	if err != nil {
		return nil, err
	}

	count := 0
	for _, numDeleted := range repositories {
		count += numDeleted
	}
	trace.Log(
		log.Int("count", count),
		log.Int("numRepositories", len(repositories)),
	)

	return repositories, nil
}

const deleteIndexesWithoutRepositoryQuery = `
-- source: internal/codeintel/autoindexing/internal/store/store_indexes.go:DeleteIndexesWithoutRepository
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_indexes u ON u.repository_id = r.id
	WHERE %s - r.deleted_at >= %s * interval '1 second'

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// IsQueued returns true if there is an index or an upload for the repository and commit.
func (s *store) IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservation := s.operations.isQueued.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(isQueuedQuery, repositoryID, commit, repositoryID, commit)))
	return count > 0, err
}

const isQueuedQuery = `
-- source: internal/codeintel/stores/dbstore/indexes.go:IsQueued
SELECT COUNT(*) WHERE EXISTS (
	SELECT id FROM lsif_uploads_with_repository_name WHERE repository_id = %s AND commit = %s AND state NOT IN ('deleted', 'deleting')
	UNION
	SELECT id FROM lsif_indexes_with_repository_name WHERE repository_id = %s AND commit = %s
)
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
	logger.Log(log.Int("numIndexes", len(indexes)))

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

const recentIndexesSummaryQuery = `
-- source: internal/codeintel/stores/dbstore/indexes.go:RecentIndexesSummary
WITH ranked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_at,
		RANK() OVER (PARTITION BY root, indexer ORDER BY finished_at DESC) AS rank
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
	` + indexAssociatedUploadIDQueryFragment + `
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
