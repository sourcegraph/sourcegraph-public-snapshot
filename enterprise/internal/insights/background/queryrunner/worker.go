package queryrunner

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// This file contains all the methods required to:
//
// 1. Create the query runner worker
// 2. Enqueue jobs for the query runner to execute.
// 3. Dequeue jobs from the query runner.
// 4. Serialize jobs for the query runner into the DB.
//

// NewWorker returns a worker that will execute search queries and insert information about the
// results into the code insights database.
func NewWorker(ctx context.Context, logger log.Logger, workerStore dbworkerstore.Store, insightsStore *store.Store, repoStore discovery.RepoStore, metrics workerutil.WorkerMetrics) *workerutil.Worker {
	numHandlers := conf.Get().InsightsQueryWorkerConcurrency
	if numHandlers <= 0 {
		numHandlers = 1
	}

	options := workerutil.WorkerOptions{
		Name:              "insights_query_runner_worker",
		NumHandlers:       numHandlers,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics,
	}

	defaultRateLimit := rate.Limit(10.0)
	getRateLimit := getRateLimit(defaultRateLimit)

	limiter := rate.NewLimiter(getRateLimit(), 1)

	go conf.Watch(func() {
		val := getRateLimit()
		logger.Info("Updating insights/query-worker rate limit", log.Int("value", int(val)))
		limiter.SetLimit(val)
	})

	sharedCache := make(map[string]*types.InsightSeries)

	prometheus.DefaultRegisterer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_insights_search_queue_total",
		Help: "Total number of jobs in the queued state.",
	}, func() float64 {
		count, err := workerStore.QueuedCount(context.Background(), false, nil)
		if err != nil {
			logger.Error("Failed to get queued job count", log.Error(err))
		}

		return float64(count)
	}))

	return dbworker.NewWorker(ctx, workerStore, &workHandler{
		baseWorkerStore: basestore.NewWithDB(workerStore.Handle().DB(), sql.TxOptions{}),
		insightsStore:   insightsStore,
		repoStore:       repoStore,
		limiter:         limiter,
		metadadataStore: store.NewInsightStore(insightsStore.Handle().DB()),
		seriesCache:     sharedCache,
		search:          query.Search,
		searchStream: func(ctx context.Context, query string) (*streaming.TabulationResult, error) {
			decoder, streamResults := streaming.TabulationDecoder()
			err := streaming.Search(ctx, query, decoder)
			if err != nil {
				return nil, errors.Wrap(err, "streaming.Search")
			}
			return streamResults, nil
		},
		computeSearch: query.ComputeSearch,
		computeSearchStream: func(ctx context.Context, query string) (*streaming.ComputeTabulationResult, error) {
			decoder, streamResults := streaming.ComputeDecoder()
			err := streaming.ComputeMatchContextStream(ctx, query, decoder)
			if err != nil {
				return nil, errors.Wrap(err, "streaming.Compute")
			}
			return streamResults, nil
		},
	}, options)
}

func getRateLimit(defaultValue rate.Limit) func() rate.Limit {
	return func() rate.Limit {
		val := conf.Get().InsightsQueryWorkerRateLimit

		var result rate.Limit
		if val == nil {
			result = defaultValue
		} else {
			result = rate.Limit(*val)
		}

		return result
	}
}

// NewResetter returns a resetter that will reset pending query runner jobs if they take too long
// to complete.
func NewResetter(ctx context.Context, workerStore dbworkerstore.Store, metrics dbworker.ResetterMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "insights_query_runner_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics,
	}
	return dbworker.NewResetter(workerStore, options)
}

// CreateDBWorkerStore creates the dbworker store for the query runner worker.
//
// See internal/workerutil/dbworker for more information about dbworkers.
func CreateDBWorkerStore(s *basestore.Store, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(s.Handle(), dbworkerstore.Options{
		Name:              "insights_query_runner_jobs_store",
		TableName:         "insights_query_runner_jobs",
		ColumnExpressions: jobsColumns,
		Scan:              scanJobs,

		// If you change this, be sure to adjust the interval that work is enqueued in
		// enterprise/internal/insights/background:newInsightEnqueuer.
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        30 * time.Minute,
		MaxNumRetries:     10,
		MaxNumResets:      10,
		OrderByExpression: sqlf.Sprintf("priority, id"),
	}, observationContext)
}

func getDependencies(ctx context.Context, workerBaseStore *basestore.Store, jobID int) (_ []time.Time, err error) {
	q := sqlf.Sprintf(getJobDependencies, jobID)
	return scanDependencies(workerBaseStore.Query(ctx, q))
}

func scanDependencies(rows *sql.Rows, queryErr error) (_ []time.Time, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]time.Time, 0)
	for rows.Next() {
		var temp time.Time
		if err := rows.Scan(&temp); err != nil {
			return nil, err
		}
		results = append(results, temp)
	}
	return results, nil
}

func insertDependencies(ctx context.Context, workerBaseStore *basestore.Store, job *Job) error {
	vals := make([]*sqlf.Query, 0, len(job.DependentFrames))
	for _, frame := range job.DependentFrames {
		vals = append(vals, sqlf.Sprintf("(%s, %s)", job.ID, frame))
	}
	if len(vals) == 0 {
		return nil
	}
	q := sqlf.Sprintf(insertJobDependencies, sqlf.Join(vals, ","))
	if err := workerBaseStore.Exec(ctx, q); err != nil {
		return err
	}
	return nil
}

const getJobDependencies = `
-- source: enterprise/internal/insights/background/queryrunner/worker.go:getDependencies
select recording_time from insights_query_runner_jobs_dependencies where job_id = %s;
`

const insertJobDependencies = `
-- source: enterprise/internal/insights/background/queryrunner/worker.go:insertDependencies
INSERT INTO insights_query_runner_jobs_dependencies (job_id, recording_time) VALUES %s;`

// EnqueueJob enqueues a job for the query runner worker to execute later.
func EnqueueJob(ctx context.Context, workerBaseStore *basestore.Store, job *Job) (id int, err error) {
	tx, err := workerBaseStore.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err = basestore.ScanFirstInt(tx.Query(
		ctx,
		sqlf.Sprintf(
			enqueueJobFmtStr,
			job.SeriesID,
			job.SearchQuery,
			job.RecordTime,
			job.State,
			job.ProcessAfter,
			job.Cost,
			job.Priority,
			job.PersistMode,
		),
	))
	if err != nil {
		return 0, err
	}
	job.ID = id
	if err := insertDependencies(ctx, tx, job); err != nil {
		return 0, nil
	}
	return id, nil
}

const enqueueJobFmtStr = `
-- source: enterprise/internal/insights/background/queryrunner/worker.go:EnqueueJob
INSERT INTO insights_query_runner_jobs (
	series_id,
	search_query,
	record_time,
	state,
	process_after,
	cost,
	priority,
	persist_mode
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id
`

func dequeueJob(ctx context.Context, workerBaseStore *basestore.Store, recordID int) (_ *Job, err error) {
	tx, err := workerBaseStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	rows, err := tx.Query(ctx, sqlf.Sprintf(dequeueJobFmtStr, recordID))
	if err != nil {
		return nil, err
	}
	jobs, err := doScanJobs(rows, nil)
	if err != nil {
		return nil, err
	}
	if len(jobs) != 1 {
		return nil, errors.Errorf("expected 1 job to dequeue, found %v", len(jobs))
	}

	deps, err := getDependencies(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	job := jobs[0]
	job.DependentFrames = deps

	return job, nil
}

const dequeueJobFmtStr = `
-- source: enterprise/internal/insights/background/queryrunner/worker.go:dequeueJob
SELECT
	series_id,
	search_query,
	record_time,
	cost,
	priority,
	persist_mode,
	id,
	state,
	failure_message,
	started_at,
	finished_at,
	process_after,
	num_resets,
	num_failures,
	execution_logs
FROM insights_query_runner_jobs
WHERE id = %s;
`

type JobsStatus struct {
	Queued, Processing uint64
	Completed          uint64
	Errored, Failed    uint64
}

// QueryJobsStatus queries the current status of jobs for the specified series.
func QueryJobsStatus(ctx context.Context, workerBaseStore *basestore.Store, seriesID string) (*JobsStatus, error) {
	var status JobsStatus
	for _, work := range []struct {
		stateName string
		result    *uint64
	}{
		{"queued", &status.Queued},
		{"processing", &status.Processing},
		{"completed", &status.Completed},
		{"errored", &status.Errored},
		{"failed", &status.Failed},
	} {
		value, _, err := basestore.ScanFirstInt(workerBaseStore.Query(
			ctx,
			sqlf.Sprintf(queryJobsStatusFmtStr, seriesID, work.stateName)),
		)
		if err != nil {
			return nil, err
		}
		*work.result = uint64(value)
	}
	return &status, nil
}

const queryJobsStatusFmtStr = `
-- source: enterprise/internal/insights/background/queryrunner/worker.go:JobsStatus
SELECT COUNT(*) FROM insights_query_runner_jobs WHERE series_id=%s AND state=%s
`

func QueryAllSeriesStatus(ctx context.Context, workerBaseStore *basestore.Store) (_ []types.InsightSeriesStatus, err error) {
	q := sqlf.Sprintf(queryAllSeriesStatusSql)
	query, err := workerBaseStore.Query(ctx, q)
	return scanAllSeriesStatusRows(query, err)
}
func scanAllSeriesStatusRows(rows *sql.Rows, queryErr error) (_ []types.InsightSeriesStatus, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []types.InsightSeriesStatus
	for rows.Next() {
		var temp types.InsightSeriesStatus
		if err := rows.Scan(
			&temp.SeriesId,
			&temp.Errored,
			&temp.Processing,
			&temp.Failed,
			&temp.Completed,
			&temp.Queued,
		); err != nil {
			return []types.InsightSeriesStatus{}, err
		}
		results = append(results, temp)
	}
	return results, nil
}

const queryAllSeriesStatusSql = `
select
       series_id,
       sum(case when state = 'errored' then 1 else 0 end) as errored,
       sum(case when state = 'processing' then 1 else 0 end) as processing,
       sum(case when state = 'failed' then 1 else 0 end) as failed,
       sum(case when state = 'completed' then 1 else 0 end) as completed,
       sum(case when state = 'queued' then 1 else 0 end) as queued
from insights_query_runner_jobs
group by series_id
order by series_id;
`

// Job represents a single job for the query runner worker to perform. When enqueued, it is stored
// in the insights_query_runner_jobs table - then the worker dequeues it by reading it from that
// table.
//
// See internal/workerutil/dbworker for more information about dbworkers.
type Job struct {
	// Query runner fields.
	SeriesID    string
	SearchQuery string
	RecordTime  *time.Time // If non-nil, record results at this time instead of the time at which search results were found.
	Cost        int
	Priority    int
	PersistMode string

	DependentFrames []time.Time // This field isn't part of the job table, but maps to a table one-many on this job.

	// Standard/required dbworker fields. If enqueuing a job, these may all be zero values except State.
	//
	// See https://sourcegraph.com/github.com/sourcegraph/sourcegraph@cd0b3904c674ee3568eb2ef5d7953395b6432d20/-/blob/internal/workerutil/dbworker/store/store.go#L114-134
	ID             int
	State          string // If enqueing a job, set to "queued"
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	ExecutionLogs  []workerutil.ExecutionLogEntry
}

// Implements the internal/workerutil.Record interface, used by the work handler to locate the job
// once executing (see work_handler.go:Handle).
func (j *Job) RecordID() int {
	return j.ID
}

func scanJobs(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	records, err := doScanJobs(rows, err)
	if err != nil || len(records) == 0 {
		return &Job{}, false, err
	}
	return records[0], true, nil
}

func doScanJobs(rows *sql.Rows, err error) ([]*Job, error) {
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()
	var jobs []*Job
	for rows.Next() {
		j := &Job{}
		if err := rows.Scan(
			// Query runner fields.
			&j.SeriesID,
			&j.SearchQuery,
			&j.RecordTime,
			&j.Cost,
			&j.Priority,
			&j.PersistMode,

			// Standard/required dbworker fields.
			&j.ID,
			&j.State,
			&j.FailureMessage,
			&j.StartedAt,
			&j.FinishedAt,
			&j.ProcessAfter,
			&j.NumResets,
			&j.NumFailures,
			pq.Array(&j.ExecutionLogs),
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

var jobsColumns = []*sqlf.Query{
	sqlf.Sprintf("insights_query_runner_jobs.series_id"),
	sqlf.Sprintf("insights_query_runner_jobs.search_query"),
	sqlf.Sprintf("insights_query_runner_jobs.record_time"),
	sqlf.Sprintf("insights_query_runner_jobs.cost"),
	sqlf.Sprintf("insights_query_runner_jobs.priority"),
	sqlf.Sprintf("insights_query_runner_jobs.persist_mode"),
	sqlf.Sprintf("id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("execution_logs"),
}

// ToQueueJob converts the query execution into a queueable job with it's relevant dependent times.
func ToQueueJob(q *compression.QueryExecution, seriesID string, query string, cost priority.Cost, jobPriority priority.Priority) *Job {
	return &Job{
		SeriesID:        seriesID,
		SearchQuery:     query,
		RecordTime:      &q.RecordingTime,
		Cost:            int(cost),
		Priority:        int(jobPriority),
		DependentFrames: q.SharedRecordings,
		State:           "queued",
		PersistMode:     string(store.RecordMode),
	}
}
