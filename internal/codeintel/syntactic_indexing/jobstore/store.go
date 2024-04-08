package jobstore

import (
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
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

func initDB(observationCtx *observation.Context, name string) *sql.DB {
	// This is an internal service, so we rely on the
	// frontend to do authz checks for user requests.
	// Authz checks are enforced by the DB layer
	//
	// This call to SetProviders is here so that calls to GetProviders don't block.
	// Relevant PR: https://github.com/sourcegraph/sourcegraph/pull/15755
	// Relevant issue: https://github.com/sourcegraph/sourcegraph/issues/15962

	authz.SetProviders(true, []authz.Provider{})

	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, name)

	if err != nil {
		log.Scoped("init db ("+name+")").Fatal("Failed to connect to frontend database", log.Error(err))
	}

	return sqlDB
}

func NewStore(observationCtx *observation.Context, name string) (SyntacticIndexingJobStore, error) {
	db := initDB(observationCtx, name)

	return NewStoreWithDB(observationCtx, db)
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
