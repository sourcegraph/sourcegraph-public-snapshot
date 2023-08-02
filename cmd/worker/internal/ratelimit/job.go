package ratelimit

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

type rateLimitConfigJob struct{}

func NewRateLimitConfigJob() job.Job {
	return &rateLimitConfigJob{}
}

func (s *rateLimitConfigJob) Description() string {
	return ""
}

func (s *rateLimitConfigJob) Config() []env.Config {
	return nil
}

func (s *rateLimitConfigJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	rlcWorker := makeRateLimitConfigWorker(observationCtx, db.CodeHosts())
	return []goroutine.BackgroundRoutine{rlcWorker}, nil
}

func makeRateLimitConfigWorker(observationCtx *observation.Context, store database.CodeHostStore) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			h := &handler{
				codeHostStore:  store,
				redisKeyPrefix: redispool.TokenBucketGlobalPrefix,
				kv:             redispool.Store,
			}
			return h.Handle(ctx, observationCtx)
		}),
		goroutine.WithName("rate_limit_config_worker"),
		goroutine.WithDescription("copies the rate limit configurations from Postgres to Redis"),
		goroutine.WithInterval(30*time.Second),
	)
}
