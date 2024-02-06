package shared

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type SyntacticIndexRecord struct {
	ID             int        `json:"id"`
	Commit         string     `json:"commit"`
	QueuedAt       time.Time  `json:"queuedAt"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFailures    int        `json:"numFailures"`
	RepositoryID   int        `json:"repositoryId"`
	RepositoryName string     `json:"repositoryName"`
	ShouldReindex  bool       `json:"shouldReindex"`
	EnqueuerUserID int32      `json:"enqueuerUserID"`
}

func (i SyntacticIndexRecord) RecordID() int {
	return i.ID
}

func (i SyntacticIndexRecord) RecordUID() string {
	return strconv.Itoa(i.ID)
}

func ScanSyntacticIndexRecord(s dbutil.Scanner) (*SyntacticIndexRecord, error) {
	var job SyntacticIndexRecord
	if err := scanSyntacticIndexRecord(&job, s); err != nil {
		return nil, err
	}
	return &job, nil
}

func scanSyntacticIndexRecord(job *SyntacticIndexRecord, s dbutil.Scanner) error {

	// Make sure this is in sync with columnExpressions below...
	if err := s.Scan(
		&job.ID,
		&job.Commit,
		&job.QueuedAt,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.RepositoryID,
		&job.RepositoryName,
		&job.ShouldReindex,
		&job.EnqueuerUserID,
	); err != nil {
		return err
	}

	return nil
}

func NewStore(observationCtx *observation.Context, db *sql.DB) (dbworkerstore.Store[*SyntacticIndexRecord], error) {

	// Make sure this is in sync with the columns of the
	// syntactic_scip_indexes_with_repository_name view
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

	storeOptions := dbworkerstore.Options[*SyntacticIndexRecord]{
		Name:      "syntactic_scip_index_store",
		TableName: "syntactic_scip_indexes",
		ViewName:  "syntactic_scip_indexes_with_repository_name u",
		// Using enqueuer_user_id prioritises manually scheduled indexing
		OrderByExpression: sqlf.Sprintf("(u.enqueuer_user_id > 0) DESC, u.queued_at, u.id"),
		ColumnExpressions: columnExpressions,
		Scan:              dbworkerstore.BuildWorkerScan(ScanSyntacticIndexRecord),
	}

	handle := basestore.NewHandleWithDB(observationCtx.Logger, db, sql.TxOptions{})
	return dbworkerstore.New(observationCtx, handle, storeOptions), nil
}

func mustInitializeDB(observationCtx *observation.Context, name string) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, name)
	if err != nil {
		log.Scoped("init db ("+name+")").Fatal("Failed to connect to frontend database", log.Error(err))
	}

	// This is an internal service, so we rely on the
	// frontend to do authz checks for user requests.
	// Authz checks are enforced by the DB layer
	//
	// This call to SetProviders is here so that calls to GetProviders don't block.
	// Relevant PR: https://github.com/sourcegraph/sourcegraph/pull/15755
	// Relevant issue: https://github.com/sourcegraph/sourcegraph/issues/15962

	ctx := context.Background()
	db := database.NewDB(observationCtx.Logger, sqlDB)
	go func() {
		for range time.NewTicker(providers.RefreshInterval()).C {
			allowAccessByDefault, authzProviders, _, _, _ := providers.ProvidersFromConfig(ctx, conf.Get(), db)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()
	return sqlDB
}
