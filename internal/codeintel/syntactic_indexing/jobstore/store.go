package jobstore

import (
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type SyntacticIndexingJobStore interface {
	DBWorkerStore() dbworkerstore.Store[*SyntacticIndexingJob]
}

type syntacticIndexingJobStoreImpl struct {
	store dbworkerstore.Store[*SyntacticIndexingJob]
	db    *basestore.Store
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
		store: dbworkerstore.New(observationCtx, handle, storeOptions),
		db:    basestore.NewWithHandle(handle),
	}, nil
}
