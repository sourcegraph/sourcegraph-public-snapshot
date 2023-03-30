package store

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	autoindexingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// DeleteIndexesWithoutRepository deletes indexes associated with repositories that were deleted at least
// DeletedRepositoryGracePeriod ago. This returns the repository identifier mapped to the number of indexes
// that were removed for that repository.
func (s *store) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (totalCount int, deletedCount int, err error) {
	ctx, trace, endObservation := s.operations.deleteIndexesWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = tx.Done(err) }()

	// TODO(efritz) - this would benefit from an index on repository_id. We currently have
	// a similar one on this index, but only for uploads that are completed or visible at tip.
	totalCount, repositories, err := scanCountsAndTotalCount(tx.Query(ctx, sqlf.Sprintf(deleteIndexesWithoutRepositoryQuery, now.UTC(), DeletedRepositoryGracePeriod/time.Second)))
	if err != nil {
		return 0, 0, err
	}

	count := 0
	for _, numDeleted := range repositories {
		count += numDeleted
	}
	trace.AddEvent("scanCounts",
		attribute.Int("count", count),
		attribute.Int("numRepositories", len(repositories)))

	return totalCount, count, nil
}

const deleteIndexesWithoutRepositoryQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_indexes u ON u.repository_id = r.id
	WHERE
		%s - r.deleted_at >= %s * interval '1 second' OR
		r.blocked IS NOT NULL

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM candidates)
	RETURNING u.id, u.repository_id
)
SELECT (SELECT COUNT(*) FROM candidates), d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

func scanCountsAndTotalCount(rows *sql.Rows, queryErr error) (totalCount int, _ map[int]int, err error) {
	if queryErr != nil {
		return 0, nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&totalCount, &id, &count); err != nil {
			return 0, nil, err
		}

		visibilities[id] = count
	}

	return totalCount, visibilities, nil
}

// GetIndexes returns a list of indexes and the total count of records matching the given conditions.
func (s *store) GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []types.Index, _ int, err error) {
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
		opts.States = append(opts.States, opts.State)
	}
	if len(opts.States) > 0 {
		conds = append(conds, makeIndexStateCondition(opts.States))
	}
	if opts.WithoutUpload {
		conds = append(conds, sqlf.Sprintf("NOT EXISTS (SELECT 1 FROM lsif_uploads u2 WHERE u2.associated_index_id = u.id)"))
	}

	if len(opts.IndexerNames) != 0 {
		var indexerConds []*sqlf.Query
		for _, indexerName := range opts.IndexerNames {
			indexerConds = append(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerName+"%"))
		}

		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, tx.db))
	if err != nil {
		return nil, 0, err
	}
	conds = append(conds, authzConds)

	indexes, err := scanIndexes(tx.db.Query(ctx, sqlf.Sprintf(
		getIndexesSelectQuery,
		sqlf.Join(conds, " AND "),
		opts.Limit,
		opts.Offset,
	)))
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("scanIndexesWithCount",
		attribute.Int("numIndexes", len(indexes)))

	totalCount, _, err := basestore.ScanFirstInt(tx.db.Query(ctx, sqlf.Sprintf(
		getIndexesCountQuery,
		sqlf.Join(conds, " AND "),
	)))
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("scanIndexesWithCount",
		attribute.Int("totalCount", totalCount),
	)

	return indexes, totalCount, nil
}

const getIndexesSelectQuery = `
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
WHERE
	repo.deleted_at IS NULL AND
	repo.blocked IS NULL AND
	%s
ORDER BY queued_at DESC, u.id
LIMIT %d OFFSET %d
`

const getIndexesCountQuery = `
SELECT COUNT(*) AS count
FROM lsif_indexes u
JOIN repo ON repo.id = u.repository_id
WHERE
	repo.deleted_at IS NULL AND
	repo.blocked IS NULL AND
	%s
`

// DeleteIndexes deletes indexes matching the given filter criteria.
func (s *store) DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) (err error) {
	ctx, _, endObservation := s.operations.deleteIndexes.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("states", strings.Join(opts.States, ",")),
		log.String("term", opts.Term),
	}})
	defer endObservation(1, observation.Args{})

	var conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeIndexSearchCondition(opts.Term))
	}
	if len(opts.States) > 0 {
		conds = append(conds, makeStateCondition(opts.States))
	}
	if opts.WithoutUpload {
		conds = append(conds, sqlf.Sprintf("NOT EXISTS (SELECT 1 FROM lsif_uploads u2 WHERE u2.associated_index_id = u.id)"))
	}
	if len(opts.IndexerNames) != 0 {
		var indexerConds []*sqlf.Query
		for _, indexerName := range opts.IndexerNames {
			indexerConds = append(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerName+"%"))
		}

		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return err
	}
	conds = append(conds, authzConds)

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_indexes_audit.reason", "direct delete by filter criteria request")
	defer unset(ctx)

	err = tx.db.Exec(ctx, sqlf.Sprintf(deleteIndexesQuery, sqlf.Join(conds, " AND ")))
	if err != nil {
		return err
	}

	return nil
}

const deleteIndexesQuery = `
DELETE FROM lsif_indexes u
USING repo
WHERE u.repository_id = repo.id AND %s
`

// ReindexIndexes reindexes indexes matching the given filter criteria.
func (s *store) ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) (err error) {
	ctx, _, endObservation := s.operations.reindexIndexes.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("states", strings.Join(opts.States, ",")),
		log.String("term", opts.Term),
	}})
	defer endObservation(1, observation.Args{})

	var conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeIndexSearchCondition(opts.Term))
	}
	if len(opts.States) > 0 {
		conds = append(conds, makeStateCondition(opts.States))
	}
	if opts.WithoutUpload {
		conds = append(conds, sqlf.Sprintf("NOT EXISTS (SELECT 1 FROM lsif_uploads u2 WHERE u2.associated_index_id = u.id)"))
	}
	if len(opts.IndexerNames) != 0 {
		var indexerConds []*sqlf.Query
		for _, indexerName := range opts.IndexerNames {
			indexerConds = append(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerName+"%"))
		}

		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return err
	}
	conds = append(conds, authzConds)

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_indexes_audit.reason", "direct reindex by filter criteria request")
	defer unset(ctx)

	err = tx.db.Exec(ctx, sqlf.Sprintf(reindexIndexesQuery, sqlf.Join(conds, " AND ")))
	if err != nil {
		return err
	}

	return nil
}

const reindexIndexesQuery = `
WITH candidates AS (
    SELECT u.id
	FROM lsif_indexes u
	JOIN repo ON repo.id = u.repository_id
	WHERE %s
    ORDER BY u.id
    FOR UPDATE
)
UPDATE lsif_indexes u
SET should_reindex = true
WHERE u.id IN (SELECT id FROM candidates)
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
func makeIndexStateCondition(states []string) *sqlf.Query {
	stateMap := make(map[string]struct{}, 2)
	for _, state := range states {
		// Treat errored and failed states as equivalent
		if state == "errored" || state == "failed" {
			stateMap["errored"] = struct{}{}
			stateMap["failed"] = struct{}{}
		} else {
			stateMap[state] = struct{}{}
		}
	}

	orderedStates := make([]string, 0, len(stateMap))
	for state := range stateMap {
		orderedStates = append(orderedStates, state)
	}
	sort.Strings(orderedStates)

	if len(orderedStates) == 1 {
		return sqlf.Sprintf("u.state = %s", orderedStates[0])
	}

	return sqlf.Sprintf("u.state = ANY(%s)", pq.Array(orderedStates))
}

// GetIndexByID returns an index by its identifier and boolean flag indicating its existence.
func (s *store) GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return types.Index{}, false, err
	}

	return scanFirstIndex(s.db.Query(ctx, sqlf.Sprintf(getIndexByIDQuery, id, authzConds)))
}

const getIndexByIDQuery = `
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
WHERE repo.deleted_at IS NULL AND u.id = %s AND %s
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
DELETE FROM lsif_indexes WHERE id = %s RETURNING repository_id
`

// ReindexIndexByID reindexes an index by its identifier.
func (s *store) ReindexIndexByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.reindexIndexByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	return tx.db.Exec(ctx, sqlf.Sprintf(reindexIndexByIDQuery, id))
}

const reindexIndexByIDQuery = `
UPDATE lsif_indexes u
SET should_reindex = true
WHERE id = %s
`

// scanIndexes scans a slice of indexes from the return value of `*Store.query`.
var scanIndexes = basestore.NewSliceScanner(scanIndex)

// scanFirstIndex scans a slice of indexes from the return value of `*Store.query` and returns the first.
var scanFirstIndex = basestore.NewFirstScanner(scanIndex)

func scanIndex(s dbutil.Scanner) (index types.Index, err error) {
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
	); err != nil {
		return index, err
	}

	index.ExecutionLogs = append(index.ExecutionLogs, executionLogs...)

	return index, nil
}

// GetRecentIndexesSummary returns the set of "interesting" indexes for the repository with the given identifier.
// The return value is a list of indexes grouped by root and indexer. In each group, the set of indexes should
// include the set of unprocessed records as well as the latest finished record. These values allow users to
// quickly determine if a particular root/indexer pair os up-to-date or having issues processing.
func (s *store) GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []autoindexingshared.IndexesWithRepositoryNamespace, err error) {
	ctx, logger, endObservation := s.operations.getRecentIndexesSummary.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	indexes, err := scanIndexes(s.db.Query(ctx, sqlf.Sprintf(recentIndexesSummaryQuery, repositoryID, repositoryID)))
	if err != nil {
		return nil, err
	}
	logger.AddEvent("scanIndexes", attribute.Int("numIndexes", len(indexes)))

	groupedIndexes := make([]autoindexingshared.IndexesWithRepositoryNamespace, 1, len(indexes)+1)
	for _, index := range indexes {
		if last := groupedIndexes[len(groupedIndexes)-1]; last.Root != index.Root || last.Indexer != index.Indexer {
			groupedIndexes = append(groupedIndexes, autoindexingshared.IndexesWithRepositoryNamespace{
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
