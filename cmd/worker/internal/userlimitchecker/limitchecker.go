package userlimitchecker

import (
	"context"
	"log"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type userLimitChecker struct{}

func NewUserLimitChecker() job.Job {
	return &userLimitChecker{}
}

func (u *userLimitChecker) Description() string {
	return ""
}

func (u *userLimitChecker) Config() []env.Config {
	return nil
}

func (u *userLimitChecker) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			&handler{
				store:  db.Users(),
				logger: observationCtx.Logger,
			},
			goroutine.WithName("license.userLimitChecker"),
			goroutine.WithDescription("checks if number users is approaching license limit"),
			goroutine.WithInterval(30*time.Minute),
		),
	}, nil
}

type handler struct {
	store  database.UserStore
	logger log.Logger
}

var (
	_ goroutine.Handler      = &handler{}
	_ goroutine.ErrorHandler = &handler{}
)

func (h *handler) Handle(ctx context.Context) error {
	h.logger.Error("error checking user limits", log.Error(err))

}
