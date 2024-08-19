package jobstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type SyntacticIndexingJobStore interface {
	DBWorkerStore() dbworkerstore.Store[*SyntacticIndexingJob]
	InsertIndexingJobs(ctx context.Context, indexingJobs []SyntacticIndexingJob) ([]SyntacticIndexingJob, error)
	IsQueued(ctx context.Context, repositoryID api.RepoID, commitID api.CommitID) (bool, error)
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

func NewStoreWithDB(observationCtx *observation.Context, db database.DB) (SyntacticIndexingJobStore, error) {
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

	return &syntacticIndexingJobStoreImpl{
		store:      dbworkerstore.New(observationCtx, db.Handle(), storeOptions),
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationCtx),
		logger:     observationCtx.Logger.Scoped("syntactic_indexing.store"),
	}, nil
}

func (s *syntacticIndexingJobStoreImpl) InsertIndexingJobs(ctx context.Context, indexingJobs []SyntacticIndexingJob) (_ []SyntacticIndexingJob, err error) {
	ctx, _, endObservation := s.operations.insertIndexingJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numIndexingJobs", len(indexingJobs)),
	}})
	endObservation(1, observation.Args{})

	if len(indexingJobs) == 0 {
		return nil, nil
	}

	indexingJobsValues := make([]*sqlf.Query, 0, len(indexingJobs))
	for _, index := range indexingJobs {
		indexingJobsValues = append(indexingJobsValues, sqlf.Sprintf(
			"(%s, %s, %s, %s)",
			index.State,
			index.Commit,
			index.RepositoryID,
			actor.FromContext(ctx).UID,
		))
	}

	indexingJobs = []SyntacticIndexingJob{}
	err = s.db.WithTransact(ctx, func(tx *basestore.Store) error {
		insertedJobIds, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(insertIndexQuery, sqlf.Join(indexingJobsValues, ","))))
		if err != nil {
			return err
		}
		s.operations.indexingJobsInserted.Add(float64(len(insertedJobIds)))

		authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
		if err != nil {
			return err
		}

		jobLookupQueries := make([]*sqlf.Query, 0, len(insertedJobIds))
		for _, id := range insertedJobIds {
			jobLookupQueries = append(jobLookupQueries, sqlf.Sprintf("%d", id))
		}
		indexingJobs, err = scanJobs(tx.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(jobLookupQueries, ", "), authzConds)))
		return err
	})

	return indexingJobs, err
}

func (s *syntacticIndexingJobStoreImpl) IsQueued(ctx context.Context, repositoryID api.RepoID, commitID api.CommitID) (bool, error) {
	isQueued, _, err := basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		isQueuedQuery,
		repositoryID,
		commitID,
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
JOIN repo ON repo.id = u.repository_id AND repo.deleted_at IS NULL AND repo.blocked IS NULL
WHERE u.id IN (%s) and %s
ORDER BY u.id
`

func scanJob(s dbutil.Scanner) (index SyntacticIndexingJob, err error) {
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

var scanJobs = basestore.NewSliceScanner(scanJob)
