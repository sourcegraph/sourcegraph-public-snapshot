package jobstore

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type SyntacticIndexingJobStore interface {
	DBWorkerStore() dbworkerstore.Store[*SyntacticIndexingJob]
	InsertIndexes(ctx context.Context, indexes []SyntacticIndexingJob) ([]SyntacticIndexingJob, error)
	IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error)
}

type syntacticIndexingJobStoreImpl struct {
	store      dbworkerstore.Store[*SyntacticIndexingJob]
	db         *basestore.Store
	operations *operations
	logger     log.Logger
}

var _ SyntacticIndexingJobStore = &syntacticIndexingJobStoreImpl{}

func (s *syntacticIndexingJobStoreImpl) DBWorkerStore() dbworkerstore.Store[*SyntacticIndexingJob] {
	return s.store
}

func NewStoreWithDB(observationCtx *observation.Context, db *sql.DB) (SyntacticIndexingJobStore, error) {
	// Make sure this is in sync with the columns of the
	// syntactic_scip_indexing_jobs_with_repository_name view
	var columnExpressions = []*sqlf.Query{
		sqlf.Sprintf("u.id"),
		sqlf.Sprintf("u.commit"),
		sqlf.Sprintf("u.queued_at"),
		sqlf.Sprintf("u.state"),
		sqlf.Sprintf("u.failure_message"),
		sqlf.Sprintf("u.started_at"),
		sqlf.Sprintf("u.finished_at"),
		sqlf.Sprintf("u.process_after"),
		sqlf.Sprintf("u.num_resets"),
		sqlf.Sprintf("u.num_failures"),
		sqlf.Sprintf("u.repository_id"),
		sqlf.Sprintf("u.repository_name"),
		sqlf.Sprintf("u.should_reindex"),
		sqlf.Sprintf("u.enqueuer_user_id"),
	}

	storeOptions := dbworkerstore.Options[*SyntacticIndexingJob]{
		Name:      "syntactic_scip_indexing_jobs_store",
		TableName: "syntactic_scip_indexing_jobs",
		ViewName:  "syntactic_scip_indexing_jobs_with_repository_name u",
		// Using enqueuer_user_id prioritises manually scheduled indexing
		OrderByExpression: sqlf.Sprintf("(u.enqueuer_user_id > 0) DESC, u.queued_at, u.id"),
		ColumnExpressions: columnExpressions,
		Scan:              dbworkerstore.BuildWorkerScan(ScanSyntacticIndexRecord),
	}

	handle := basestore.NewHandleWithDB(observationCtx.Logger, db, sql.TxOptions{})
	return &syntacticIndexingJobStoreImpl{
		store:      dbworkerstore.New(observationCtx, handle, storeOptions),
		db:         basestore.NewWithHandle(handle),
		operations: newOperations(observationCtx),
		logger:     observationCtx.Logger.Scoped("syntactic_indexing.store"),
	}, nil
}

func (s *syntacticIndexingJobStoreImpl) InsertIndexes(ctx context.Context, indexes []SyntacticIndexingJob) (jobs []SyntacticIndexingJob, err error) {

	ctx, _, endObservation := s.operations.insertIndexingJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numIndexes", len(indexes)),
	}})
	endObservation(1, observation.Args{})

	if len(indexes) == 0 {
		return nil, nil
	}

	actor := actor.FromContext(ctx)

	values := make([]*sqlf.Query, 0, len(indexes))
	for _, index := range indexes {
		values = append(values, sqlf.Sprintf(
			"(%s, %s, %s, %s)",
			index.State,
			index.Commit,
			index.RepositoryID,
			actor.UID,
		))
	}

	indexes = []SyntacticIndexingJob{}

	err = s.db.WithTransact(ctx, func(tx *basestore.Store) error {
		ids, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(insertIndexQuery, sqlf.Join(values, ","))))
		if err != nil {
			return err
		}
		s.operations.indexingJobsInserted.Add(float64(len(ids)))

		authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
		if err != nil {
			return err
		}

		queries := make([]*sqlf.Query, 0, len(ids))
		for _, id := range ids {
			queries = append(queries, sqlf.Sprintf("%d", id))
		}

		indexes, err = scanIndexes(tx.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(queries, ", "), authzConds)))
		return err
	})

	return indexes, err
}

func (s *syntacticIndexingJobStoreImpl) IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error) {
	isQueued, _, err := basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		isQueuedQuery,
		repositoryID,
		commit,
	)))
	return isQueued, err
}

const insertIndexQuery = `
INSERT INTO syntactic_scip_indexing_jobs (
	state,
	commit,
	repository_id,
	enqueuer_user_id
)
VALUES %s
RETURNING id
`

const isQueuedQuery = `
SELECT EXISTS(
	SELECT queued_at
	FROM syntactic_scip_indexing_jobs
	WHERE
		repository_id  = %s AND
		commit = %s
	ORDER BY queued_at DESC
	LIMIT 1
)
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
	u.repository_name,
	u.should_reindex,
	u.enqueuer_user_id
FROM syntactic_scip_indexing_jobs_with_repository_name u
WHERE u.id IN (%s) and %s
ORDER BY u.id
`

func scanIndex(s dbutil.Scanner) (index SyntacticIndexingJob, err error) {
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
		&index.ShouldReindex,
		&index.EnqueuerUserID,
	); err != nil {
		return index, err
	}

	return index, nil
}

var scanIndexes = basestore.NewSliceScanner(scanIndex)
