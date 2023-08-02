package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tableName = "rate_limit_config_jobs"

type Job struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool
	CodeHostURL     string
}

func (b *Job) RecordID() int {
	return b.ID
}

func (b *Job) RecordUID() string {
	return strconv.Itoa(b.ID)
}

var jobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("queued_at"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("last_heartbeat_at"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostname"),
	sqlf.Sprintf("cancel"),
	sqlf.Sprintf("code_host_url"),
}

func scanJob(s dbutil.Scanner) (*Job, error) {
	var job Job
	var executionLogs []executor.ExecutionLogEntry

	if err := s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.QueuedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.LastHeartbeatAt,
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.Cancel,
		&job.CodeHostURL,
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func makeWorkerStore(db database.DB, observationCtx *observation.Context) dbworkerstore.Store[*Job] {
	return dbworkerstore.New(observationCtx, db.Handle(), dbworkerstore.Options[*Job]{
		Name:              "rate_limit_config_worker_store",
		TableName:         tableName,
		ColumnExpressions: jobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanJob),
		OrderByExpression: sqlf.Sprintf("id"), // processes oldest records first
		MaxNumResets:      10,
		StalledMaxAge:     time.Second * 30,
		RetryAfter:        time.Second * 30,
		MaxNumRetries:     3,
	})
}

type handler struct {
	codeHostStore database.CodeHostStore
}

var _ workerutil.Handler[*Job] = &handler{}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *Job) error {
	logger.Info("STARTED THE STUFFS")
	return h.process(ctx, record.CodeHostURL, logger)
}

func (h *handler) process(ctx context.Context, codeHostURL string, _ log.Logger) error {
	// Retrieve all the code host rate limit config keys in Redis.
	apiCapKey, apiReplenishmentKey, gitCapKey, gitReplenishmentKey := redispool.GetCodeHostRateLimiterConfigKeys(redispool.TokenBucketGlobalPrefix, codeHostURL)

	// Retrieve the actual rate limit values from the source of truth (Postgres).
	ch, err := h.codeHostStore.GetByURL(ctx, codeHostURL)
	if err != nil {
		return errors.Wrapf(err, "rate limit config worker unable to get code host by URL: %s", codeHostURL)
	}

	// Set all of the rate limit config options in Redis.
	if ch.APIRateLimitQuota != nil {
		err = redispool.Store.Set(apiCapKey, *ch.APIRateLimitQuota)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", apiCapKey, codeHostURL)

		}
	}
	if ch.APIRateLimitIntervalSeconds != nil {
		err = redispool.Store.Set(apiReplenishmentKey, *ch.APIRateLimitIntervalSeconds)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", apiReplenishmentKey, codeHostURL)
		}
	}
	if ch.GitRateLimitQuota != nil {
		err = redispool.Store.Set(gitCapKey, *ch.GitRateLimitQuota)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", gitCapKey, codeHostURL)
		}
	}
	if ch.GitRateLimitIntervalSeconds != nil {
		err = redispool.Store.Set(gitReplenishmentKey, *ch.GitRateLimitIntervalSeconds)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", gitReplenishmentKey, codeHostURL)
		}
	}
	return nil
}
