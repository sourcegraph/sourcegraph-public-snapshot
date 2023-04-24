package background

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexJobType struct {
	Name            string
	Id              int
	IndexInterval   time.Duration
	RefreshInterval time.Duration
}

var IndexJobTypes = []IndexJobType{{
	Name:            "recent-contributors",
	Id:              1,
	IndexInterval:   time.Hour * 24,
	RefreshInterval: time.Minute * 5,
}}

func featureFlagName(jobType IndexJobType) string {
	return fmt.Sprintf("own-background-index-repo-%s", jobType.Name)
}

const tableName = "own_background_jobs"

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
	RepoId          int
	JobType         int
}

func (b *Job) RecordID() int {
	return b.ID
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
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("job_type"),
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
		&job.RepoId,
		&job.JobType,
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func NewOwnBackgroundWorker(ctx context.Context, db database.DB, observationCtx *observation.Context) []goroutine.BackgroundRoutine {
	worker, resetter, _ := makeWorker(ctx, db, observationCtx)
	janitor := background.NewJanitorJob(ctx, background.JanitorOptions{
		Name:        "own-background-jobs-janitor",
		Description: "Janitor for own-background-jobs queue",
		Interval:    time.Minute * 5,
		Metrics:     background.NewJanitorMetrics(observationCtx, "own-background-jobs-janitor", "own-background"),
		CleanupFunc: janitorFunc(db, time.Hour*24*7),
	})
	return []goroutine.BackgroundRoutine{worker, resetter, janitor}
}

func makeWorker(ctx context.Context, db database.DB, observationCtx *observation.Context) (*workerutil.Worker[*Job], *dbworker.Resetter[*Job], dbworkerstore.Store[*Job]) {
	name := "own_background_worker"

	workerStore := dbworkerstore.New(observationCtx, db.Handle(), dbworkerstore.Options[*Job]{
		Name:              fmt.Sprintf("%s_store", name),
		TableName:         tableName,
		ColumnExpressions: jobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanJob),
		OrderByExpression: sqlf.Sprintf("id"), // processes oldest records first
		MaxNumResets:      10,
		StalledMaxAge:     time.Second * 30,
		RetryAfter:        time.Second * 30,
		MaxNumRetries:     3,
	})

	limiter := getRateLimiter()
	conf.Watch(func() {
		setRateLimitConfig(limiter)
	})

	task := handler{
		workerStore: workerStore,
		limiter:     limiter,
	}

	worker := dbworker.NewWorker(ctx, workerStore, workerutil.Handler[*Job](&task), workerutil.WorkerOptions{
		Name:              name,
		Description:       "Sourcegraph own background processing partitioned by repository",
		NumHandlers:       getConcurrencyConfig(),
		Interval:          10 * time.Second,
		HeartbeatInterval: 20 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, name+"_processor"),
	})

	resetter := dbworker.NewResetter(log.Scoped("OwnBackgroundResetter", ""), workerStore, dbworker.ResetterOptions{
		Name:     fmt.Sprintf("%s_resetter", name),
		Interval: time.Second * 20,
		Metrics:  dbworker.NewResetterMetrics(observationCtx, name),
	})

	return worker, resetter, workerStore
}

type handler struct {
	workerStore dbworkerstore.Store[*Job]
	limiter     *ratelimit.InstrumentedLimiter
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *Job) error {
	err := h.limiter.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "limiter.Wait")
	}
	jsonify, _ := json.Marshal(record)
	logger.Info(fmt.Sprintf("Hello from the own background processor: %v", string(jsonify)))
	return nil
}

func janitorFunc(db database.DB, retention time.Duration) func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, err error) {
	return func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, err error) {
		ts := time.Now().Add(-1 * retention)
		result, err := basestore.NewWithHandle(db.Handle()).ExecResult(ctx, sqlf.Sprintf("delete from %s where state not in ('queued', 'processing', 'errored') and finished_at < %s", sqlf.Sprintf(tableName), ts))
		if err != nil {
			return 0, 0, err
		}
		affected, _ := result.RowsAffected()
		return 0, int(affected), nil
	}
}

const (
	DefaultRateLimit      = 20
	DefaultRateBurstLimit = 5
	DefaultMaxConcurrency = 5
)

func getConcurrencyConfig() int {
	val := conf.Get().SiteConfiguration.OwnBackgroundRepoIndexConcurrencyLimit
	if val == 0 {
		val = DefaultMaxConcurrency
	}
	return val
}

func getRateLimitConfig() (rate.Limit, int) {
	limit := conf.Get().SiteConfiguration.OwnBackgroundRepoIndexRateLimit
	if limit == 0 {
		limit = DefaultRateLimit
	}
	burst := conf.Get().SiteConfiguration.OwnBackgroundRepoIndexRateBurstLimit
	if burst == 0 {
		burst = DefaultRateBurstLimit
	}
	return rate.Limit(limit), burst
}

func getRateLimiter() *ratelimit.InstrumentedLimiter {
	limit, burst := getRateLimitConfig()
	return ratelimit.NewInstrumentedLimiter("OwnRepoIndexWorker", rate.NewLimiter(limit, burst))
}

func setRateLimitConfig(limiter *ratelimit.InstrumentedLimiter) {
	limit, burst := getRateLimitConfig()
	limiter.SetLimit(limit)
	limiter.SetBurst(burst)
}
