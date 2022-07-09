package repos

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	workerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type WhBuildOptions struct {
	NumHandlers            int
	WorkerInterval         time.Duration
	PrometheusRegisterer   prometheus.Registerer
	CleanupOldJobs         bool
	CleanupOldJobsInterval time.Duration
}

func NewWhBuildWorker(
	ctx context.Context,
	dbHandle basestore.TransactableHandle,
	handler workerutil.Handler,
	opts WhBuildOptions,
) (*workerutil.Worker, *dbworker.Resetter) {
	if opts.NumHandlers == 0 {
		opts.NumHandlers = 3
	}
	if opts.WorkerInterval == 0 {
		opts.WorkerInterval = 10 * time.Second
	}
	if opts.CleanupOldJobsInterval == 0 {
		opts.CleanupOldJobsInterval = time.Hour
	}

	whBuildJobColumns := []*sqlf.Query{
		sqlf.Sprintf("id"),
		sqlf.Sprintf("state"),
		sqlf.Sprintf("failure_message"),
		sqlf.Sprintf("started_at"),
		sqlf.Sprintf("finished_at"),
		sqlf.Sprintf("process_after"),
		sqlf.Sprintf("num_resets"),
		sqlf.Sprintf("num_failures"),
		sqlf.Sprintf("execution_logs"),
		sqlf.Sprintf("repo_id"),
		sqlf.Sprintf("repo_name"),
		sqlf.Sprintf("extsvc_kind"),
		sqlf.Sprintf("token"),
		sqlf.Sprintf("queued_at"),
	}

	store := workerstore.New(dbHandle, workerstore.Options{
		Name:      "webhook_build_worker_store",
		TableName: "webhook_build_jobs",
		// ViewName:          "webhook_build_jobs_with_next_in_queue",
		Scan:              scanWhBuildJob,
		OrderByExpression: sqlf.Sprintf("webhook_build_jobs.queued_at"),
		ColumnExpressions: whBuildJobColumns,
		StalledMaxAge:     30 * time.Second,
		MaxNumResets:      5,
		MaxNumRetries:     0,
	})

	worker := dbworker.NewWorker(ctx, store, handler, workerutil.WorkerOptions{
		Name:              "webhook_build_worker",
		NumHandlers:       opts.NumHandlers,
		Interval:          opts.WorkerInterval,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           newWorkerMetrics(opts.PrometheusRegisterer), // move to central pacckage
	})

	resetter := dbworker.NewResetter(store, dbworker.ResetterOptions{
		Name:     "webhook_build_resetter",
		Interval: 5 * time.Minute,
		Metrics:  newResetterMetrics(opts.PrometheusRegisterer), // move to central package
	})

	if opts.CleanupOldJobs {
		go runJobCleaner(ctx, dbHandle, opts.CleanupOldJobsInterval) // move to central package
	}

	return worker, resetter
}

func scanWhBuildJob(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	if err != nil {
		return nil, false, err
	}

	jobs, err := scanWhBuildJobs(rows)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}

	return &jobs[0], true, nil
}

type WhBuildJob struct {
	ID             int
	State          string
	FailureMessage sql.NullString
	StartedAt      sql.NullTime
	FinishedAt     sql.NullTime
	ProcessAfter   sql.NullTime
	NumResets      int
	NumFailures    int
	RepoID         int64
	RepoName       string
	ExtsvcKind     string
	Token          string
	QueuedAt       sql.NullTime
}

func (cw *WhBuildJob) RecordID() int {
	return cw.ID
}

type SyncWebhookWorker struct {
	syncRequestQueue *syncRequestQueue
}

func NewWebhookCreator(ctx context.Context) SyncWebhookWorker {
	syncRequestQueue := syncRequestQueue{queue: make([]*syncRequest, 0)}
	worker := SyncWebhookWorker{syncRequestQueue: &syncRequestQueue}
	return worker
}

func (worker *SyncWebhookWorker) Enqueue(repo *types.Repo) error {
	syncRequest := syncRequest{
		repo:   repo,
		secret: "secret",
		token:  "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",
	}
	ok := worker.syncRequestQueue.enqueue(syncRequest)
	if !ok {
		return errors.New("error enqueuing")
	} else {
		return nil
	}
}

type syncRequest struct {
	repo   *types.Repo
	secret string
	token  string
}

type syncRequestQueue struct {
	mu            sync.Mutex
	queue         []*syncRequest
	notifyEnqueue chan struct{}
}

func (sq *syncRequestQueue) enqueue(syncReq syncRequest) bool {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	sq.queue = append(sq.queue, &syncReq)
	return true
}

func (sq *syncRequestQueue) dequeue() (syncRequest, bool) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	syncReq := sq.queue[0]
	sq.queue = sq.queue[1:]
	return *syncReq, true
}

func (sq *syncRequestQueue) len() int {
	return len(sq.queue)
}
