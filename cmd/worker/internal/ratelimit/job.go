package ratelimit

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

type rateLimitConfigJob struct{}

func NewRateLimitConfigJob() job.Job {
	return &rateLimitConfigJob{}
}

func (s *rateLimitConfigJob) Description() string {
	return "Copies the rate limit configurations from the database to Redis."
}

func (s *rateLimitConfigJob) Config() []env.Config {
	return nil
}

func (s *rateLimitConfigJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	logger := observationCtx.Logger.Scoped("Periodic rate limit config job")
	rlcWorker := makeRateLimitConfigWorker(&handler{
		logger:               logger,
		externalServiceStore: db.ExternalServices(),
		newRateLimiterFunc: func(bucketName string) ratelimit.GlobalLimiter {
			return ratelimit.NewGlobalRateLimiter(logger, bucketName)
		},
	})

	return []goroutine.BackgroundRoutine{rlcWorker}, nil
}

func makeRateLimitConfigWorker(handler *handler) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("rate_limit_config_worker"),
		goroutine.WithDescription("copies the rate limit configurations from the database to Redis"),
		goroutine.WithInterval(30*time.Second),
	)
}
