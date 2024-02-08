package shared

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type recordState string

const (
	Queued     recordState = "queued"
	Errored    recordState = "errored"
	Processing recordState = "processing"
	Completed  recordState = "completed"
)

// Unless marked otherwise, the columns in this
// record have a special meaning assigned to them by
// the queries dbworker performs. You can read more
// about the different fields and what they do here:
// https://sourcegraph.com/docs/dev/background-information/workers#database-backed-stores
type SyntacticIndexingJob struct {
	ID             int         `json:"id"`
	State          recordState `json:"state"`
	QueuedAt       time.Time   `json:"queuedAt"`
	StartedAt      *time.Time  `json:"startedAt"`
	FinishedAt     *time.Time  `json:"finishedAt"`
	ProcessAfter   *time.Time  `json:"processAfter"`
	NumResets      int         `json:"numResets"`
	NumFailures    int         `json:"numFailures"`
	FailureMessage *string     `json:"failureMessage"`
	ShouldReindex  bool        `json:"shouldReindex"`

	// The fields below are not part of the standard dbworker fields

	// Which commit to index
	Commit string `json:"commit"`
	// Which repository id to index
	RepositoryID int `json:"repositoryId"`
	// Name of repository being indexed
	RepositoryName string `json:"repositoryName"`
	// Which user scheduled this job
	EnqueuerUserID int32 `json:"enqueuerUserID"`
}

var _ workerutil.Record = SyntacticIndexingJob{}

func (i SyntacticIndexingJob) RecordID() int {
	return i.ID
}

func (i SyntacticIndexingJob) RecordUID() string {
	return strconv.Itoa(i.ID)
}

func ScanSyntacticIndexRecord(s dbutil.Scanner) (*SyntacticIndexingJob, error) {
	var job SyntacticIndexingJob
	if err := scanSyntacticIndexRecord(&job, s); err != nil {
		return nil, err
	}
	return &job, nil
}

func scanSyntacticIndexRecord(job *SyntacticIndexingJob, s dbutil.Scanner) error {

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

func NewStore(observationCtx *observation.Context, db *sql.DB) (dbworkerstore.Store[*SyntacticIndexingJob], error) {

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
	return dbworkerstore.New(observationCtx, handle, storeOptions), nil
}

func mustInitializeDB(observationCtx *observation.Context, name string) *sql.DB {
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
