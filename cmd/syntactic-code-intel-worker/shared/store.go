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
	Outfile        string     `json:"outfile"`
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

	if err := s.Scan(
		&job.ID,
		&job.Commit, &job.QueuedAt, &job.State, &job.FailureMessage, &job.StartedAt, &job.FinishedAt, &job.ProcessAfter, &job.NumResets, &job.NumFailures, &job.RepositoryID, &job.RepositoryName, &job.Outfile, &job.ShouldReindex, &job.EnqueuerUserID,
	); err != nil {
		return err
	}

	return nil
}

func NewStore(observationCtx *observation.Context) (dbworkerstore.Store[*SyntacticIndexRecord], error) {

	// Make sure this is in sync
	var columnExpressions = []*sqlf.Query{
		sqlf.Sprintf("id"),
		sqlf.Sprintf("commit"),
		sqlf.Sprintf("queued_at"),
		sqlf.Sprintf("state"),
		sqlf.Sprintf("failure_message"),
		sqlf.Sprintf("started_at"),
		sqlf.Sprintf("finished_at"),
		sqlf.Sprintf("process_after"),
		sqlf.Sprintf("num_resets"),
		sqlf.Sprintf("num_failures"),
		sqlf.Sprintf("repository_id"),
		sqlf.Sprintf("repository_name"),
		sqlf.Sprintf("outfile"),
		sqlf.Sprintf("should_reindex"),
		sqlf.Sprintf("enqueuer_user_id"),
	}

	storeOptions := dbworkerstore.Options[*SyntacticIndexRecord]{
		Name:              "syntactic_scip_index_store",
		TableName:         "syntactic_scip_indexes",
		ViewName:          "syntactic_scip_indexes_with_repository_name",
		OrderByExpression: sqlf.Sprintf("syntactic_scip_indexes.queued_at"),
		ColumnExpressions: columnExpressions,
		Scan:              dbworkerstore.BuildWorkerScan(ScanSyntacticIndexRecord),
	}

	db := mustInitializeDB(observationCtx)
	handle := basestore.NewHandleWithDB(observationCtx.Logger, db, sql.TxOptions{})
	return dbworkerstore.New(observationCtx, handle, storeOptions), nil
}

func mustInitializeDB(observationCtx *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "precise-code-intel-worker")
	if err != nil {
		log.Scoped("init db").Fatal("Failed to connect to frontend database", log.Error(err))
	}

	//
	// START FLAILING

	ctx := context.Background()
	db := database.NewDB(observationCtx.Logger, sqlDB)
	go func() {
		for range time.NewTicker(providers.RefreshInterval()).C {
			allowAccessByDefault, authzProviders, _, _, _ := providers.ProvidersFromConfig(ctx, conf.Get(), db)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	// END FLAILING
	//

	return sqlDB
}
