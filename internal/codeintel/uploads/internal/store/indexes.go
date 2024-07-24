package store

import (
	"context"
	"sort"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetAutoIndexJobs returns a list of indexes and the total count of records matching the given conditions.
func (s *store) GetAutoIndexJobs(ctx context.Context, opts shared.GetAutoIndexJobsOptions) (_ []shared.AutoIndexJob, _ int, err error) {
	ctx, trace, endObservation := s.operations.getAutoIndexJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", opts.RepositoryID),
		attribute.String("state", opts.State),
		attribute.String("term", opts.Term),
		attribute.Int("limit", opts.Limit),
		attribute.Int("offset", opts.Offset),
	}})
	defer endObservation(1, observation.Args{})

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

	var a []shared.AutoIndexJob
	var b int
	err = s.withTransaction(ctx, func(tx *store) error {
		authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, tx.db))
		if err != nil {
			return err
		}
		conds = append(conds, authzConds)

		indexes, err := scanJobs(tx.db.Query(ctx, sqlf.Sprintf(
			getIndexesSelectQuery,
			sqlf.Join(conds, " AND "),
			opts.Limit,
			opts.Offset,
		)))
		if err != nil {
			return err
		}
		trace.AddEvent("scanJobsWithCount",
			attribute.Int("numIndexes", len(indexes)))

		totalCount, _, err := basestore.ScanFirstInt(tx.db.Query(ctx, sqlf.Sprintf(
			getIndexesCountQuery,
			sqlf.Join(conds, " AND "),
		)))
		if err != nil {
			return err
		}
		trace.AddEvent("scanJobsWithCount",
			attribute.Int("totalCount", totalCount),
		)

		a = indexes
		b = totalCount
		return nil
	})

	return a, b, err
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
	u.requested_envvars,
	u.enqueuer_user_id
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

// scanJobs scans a slice of indexes from the return value of `*Store.query`.
var scanJobs = basestore.NewSliceScanner(scanJob)

// scanFirstIndex scans a slice of indexes from the return value of `*Store.query` and returns the first.
var scanFirstIndex = basestore.NewFirstScanner(scanJob)

func scanJob(s dbutil.Scanner) (index shared.AutoIndexJob, err error) {
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

// GetAutoIndexJobByID returns an index by its identifier and boolean flag indicating its existence.
func (s *store) GetAutoIndexJobByID(ctx context.Context, id int) (_ shared.AutoIndexJob, _ bool, err error) {
	ctx, _, endObservation := s.operations.getAutoIndexJobByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return shared.AutoIndexJob{}, false, err
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
	u.requested_envvars,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND u.id = %s AND %s
`

// GetAutoIndexJobsByIDs returns an index for each of the given identifiers. Not all given ids will necessarily
// have a corresponding element in the returned list.
func (s *store) GetAutoIndexJobsByIDs(ctx context.Context, ids ...int) (_ []shared.AutoIndexJob, err error) {
	ctx, _, endObservation := s.operations.getAutoIndexJobsByIDs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.IntSlice("ids", ids),
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

	return scanJobs(s.db.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(queries, ", "), authzConds)))
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
	u.requested_envvars,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_at IS NULL AND u.id IN (%s) AND %s
ORDER BY u.id
`

// DeleteAutoIndexJobByID deletes an index by its identifier.
func (s *store) DeleteAutoIndexJobByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := s.operations.deleteAutoIndexJobByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	_, exists, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(deleteIndexByIDQuery, id)))
	return exists, err
}

const deleteIndexByIDQuery = `
DELETE FROM lsif_indexes WHERE id = %s RETURNING repository_id
`

// DeleteAutoIndexJobs deletes indexes matching the given filter criteria.
func (s *store) DeleteAutoIndexJobs(ctx context.Context, opts shared.DeleteAutoIndexJobsOptions) (err error) {
	ctx, _, endObservation := s.operations.deleteAutoIndexJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", opts.RepositoryID),
		attribute.StringSlice("states", opts.States),
		attribute.String("term", opts.Term),
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

	return s.withTransaction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_indexes_audit.reason", "direct delete by filter criteria request")
		defer unset(ctx)

		return tx.db.Exec(ctx, sqlf.Sprintf(deleteIndexesQuery, sqlf.Join(conds, " AND ")))
	})
}

const deleteIndexesQuery = `
DELETE FROM lsif_indexes u
USING repo
WHERE u.repository_id = repo.id AND %s
`

// SetRerunAutoIndexJobByID reindexes an index by its identifier.
func (s *store) SetRerunAutoIndexJobByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.setRerunAutoIndexJobByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(setRerunAutoIndexJobByIDQuery, id))
}

const setRerunAutoIndexJobByIDQuery = `
UPDATE lsif_indexes u
SET should_reindex = true
WHERE id = %s
`

// SetRerunAutoIndexJobs reindexes indexes matching the given filter criteria.
func (s *store) SetRerunAutoIndexJobs(ctx context.Context, opts shared.SetRerunAutoIndexJobsOptions) (err error) {
	ctx, _, endObservation := s.operations.setRerunAutoIndexJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", opts.RepositoryID),
		attribute.StringSlice("states", opts.States),
		attribute.String("term", opts.Term),
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

	return s.withTransaction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_indexes_audit.reason", "direct reindex by filter criteria request")
		defer unset(ctx)

		return tx.db.Exec(ctx, sqlf.Sprintf(setRerunAutoIndexJobsByIDsQuery, sqlf.Join(conds, " AND ")))
	})
}

const setRerunAutoIndexJobsByIDsQuery = `
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

//
//

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
