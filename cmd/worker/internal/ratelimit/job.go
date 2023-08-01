package ratelimit

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type rateLimitConfigJob struct{}

func NewRateLimitConfigJob() job.Job {
	return &rateLimitConfigJob{}
}

func (s *rateLimitConfigJob) Description() string {
	return ""
}

func (s *rateLimitConfigJob) Config() []env.Config {
	return []env.Config{embeddings.EmbeddingsUploadStoreConfigInst}
}

func (s *rateLimitConfigJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	workCtx := actor.WithInternalActor(context.Background())
	return newRateLimitConfigWorker(workCtx, db, observationCtx), nil
}

func newRateLimitConfigWorker(ctx context.Context, db database.DB, observationCtx *observation.Context) []goroutine.BackgroundRoutine {
	workerStore := makeWorkerStore(db, observationCtx)
	worker := makeWorker(ctx, observationCtx, workerStore, db.CodeHosts())
	resetter := makeResetter(observationCtx, workerStore)
	return []goroutine.BackgroundRoutine{worker, resetter}
}

func makeWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*Job],
	codeHostStore database.CodeHostStore,
) *workerutil.Worker[*Job] {
	handler := &handler{
		codeHostStore: codeHostStore,
	}

	return dbworker.NewWorker[*Job](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "rate_limit_config_job_worker",
		Interval:          10 * time.Second,
		NumHandlers:       1,
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "rate_limit_config_job_worker"),
	})
}

func makeResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*Job]) *dbworker.Resetter[*Job] {
	return dbworker.NewResetter[*Job](observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "rate_limit_config_job_worker_resetter",
		Interval: time.Second * 30,
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "rate_limit_config_job_worker"),
	})
}
