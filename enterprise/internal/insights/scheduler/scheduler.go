package scheduler

import (
	"context"
	"fmt"
	"time"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"

	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/lib/pq"

	log "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type BaseJob struct {
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
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool
	backfillId      int
}

func (b *BaseJob) RecordID() int {
	return b.ID
}

var baseJobColumns = []*sqlf.Query{
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
	sqlf.Sprintf("backfill_id"),
}

func scanBaseJob(s dbutil.Scanner) (*BaseJob, error) {
	var job BaseJob
	var executionLogs []dbworkerstore.ExecutionLogEntry

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
		&dbutil.NullInt{N: &job.backfillId},
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return &job, nil
}

type Scheduler struct {
	inProgressWorker   *workerutil.Worker
	inProgressResetter *dbworker.Resetter

	newBackfillWorker   *workerutil.Worker
	newBackfillResetter *dbworker.Resetter

	store *basestore.Store
}

func NewScheduler(ctx context.Context, db edb.InsightsDB, obsContext *observation.Context) *Scheduler {
	inProgressWorker, inProgressResetter := makeInProgressWorker(ctx, db, obsContext)
	newBackfillWorker, newBackfillResetter := makeNewBackfillWorker(ctx, db, obsContext)

	return &Scheduler{
		inProgressWorker:    inProgressWorker,
		inProgressResetter:  inProgressResetter,
		newBackfillWorker:   newBackfillWorker,
		newBackfillResetter: newBackfillResetter,
		store:               basestore.NewWithHandle(db.Handle()),
	}
}

func (s *Scheduler) Routines() []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		s.inProgressWorker,
		s.inProgressResetter,
		s.newBackfillWorker,
		s.newBackfillResetter,
	}
}

func makeInProgressWorker(ctx context.Context, db edb.InsightsDB, obsContext *observation.Context) (*workerutil.Worker, *dbworker.Resetter) {
	backfillStore := newBackfillStore(db)

	name := "backfill_in_progress_worker"

	workerStore := dbworkerstore.NewWithMetrics(db.Handle(), dbworkerstore.Options{
		Name:              fmt.Sprintf("%s_store", name),
		TableName:         "insights_background_jobs",
		ViewName:          "insights_jobs_backfill_in_progress",
		ColumnExpressions: baseJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanBaseJob),
		OrderByExpression: sqlf.Sprintf("id"), // todo
		MaxNumResets:      100,
		StalledMaxAge:     time.Second * 30,
	}, obsContext)

	task := inProgressHandler{
		workerStore:   workerStore,
		backfillStore: backfillStore,
	}

	worker := dbworker.NewWorker(ctx, workerStore, &task, workerutil.WorkerOptions{
		Name:        name,
		NumHandlers: 1,
		Interval:    5 * time.Second,
		Metrics:     workerutil.NewMetrics(obsContext, name),
	})

	resetter := dbworker.NewResetter(log.Scoped("", ""), workerStore, dbworker.ResetterOptions{
		Name:     fmt.Sprintf("%s_resetter", name),
		Interval: time.Second * 20,
		Metrics:  *dbworker.NewMetrics(obsContext, name),
	})

	return worker, resetter
}

func makeNewBackfillWorker(ctx context.Context, db edb.InsightsDB, obsContext *observation.Context) (*workerutil.Worker, *dbworker.Resetter) {
	backfillStore := newBackfillStore(db)

	name := "backfill_new_backfill_worker"

	workerStore := dbworkerstore.NewWithMetrics(db.Handle(), dbworkerstore.Options{
		Name:              fmt.Sprintf("%s_store", name),
		TableName:         "insights_background_jobs",
		ViewName:          "insights_jobs_backfill_new",
		ColumnExpressions: baseJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanBaseJob),
		OrderByExpression: sqlf.Sprintf("id"), // todo
		MaxNumResets:      100,
		StalledMaxAge:     time.Second * 30,
	}, obsContext)

	task := newBackfillHandler{
		workerStore:   workerStore,
		backfillStore: backfillStore,
	}

	worker := dbworker.NewWorker(ctx, workerStore, &task, workerutil.WorkerOptions{
		Name:        name,
		NumHandlers: 1,
		Interval:    5 * time.Second,
		Metrics:     workerutil.NewMetrics(obsContext, name),
	})

	resetter := dbworker.NewResetter(log.Scoped("", ""), workerStore, dbworker.ResetterOptions{
		Name:     fmt.Sprintf("%s_resetter", name),
		Interval: time.Second * 20,
		Metrics:  *dbworker.NewMetrics(obsContext, name),
	})

	return worker, resetter
}

type inProgressHandler struct {
	workerStore   dbworkerstore.Store
	backfillStore *backfillStore
}

var _ workerutil.Handler = &inProgressHandler{}

func (h *inProgressHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	logger.Info("inProgressHandler called", log.Int("recordId", record.RecordID()))

	job := record.(*BaseJob)

	backfill, err := loadBackfill(ctx, h.backfillStore, job.backfillId)
	if err != nil {
		return errors.Wrap(err, "loadBackfill")
	}

	itr, err := backfill.repoIterator(ctx, h.backfillStore)
	if err != nil {
		return errors.Wrap(err, "repoIterator")
	}

	for true {
		repoId, more, finish := itr.NextWithFinish()
		if !more {
			break
		}

		// todo do backfilling work
		logger.Info("doing iteration work", log.Int("repo_id", int(repoId)))

		err = finish(ctx, h.backfillStore.Store, nil)
		if err != nil {
			return err
		}
	}

	// todo handle errors down here after the main loop https://github.com/sourcegraph/sourcegraph/issues/42724

	return nil
}

type newBackfillHandler struct {
	workerStore   dbworkerstore.Store
	backfillStore *backfillStore
}

var _ workerutil.Handler = &newBackfillHandler{}

func (h *newBackfillHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	logger.Info("newBackfillHandler called", log.Int("recordId", record.RecordID()))

	// load repos
	// estimate cost
	// update backfill record (implies creating repo iterator)
	// requeue

	return nil
}

func (s *Scheduler) EnqueueBackfill(ctx context.Context, backfill *SeriesBackfill) error {
	if backfill == nil || backfill.Id == 0 {
		return errors.New("invalid series backfill")
	}
	return s.store.Exec(ctx, sqlf.Sprintf("insert into insights_background_jobs (backfill_id) VALUES (%s)", backfill.Id))
}
