package zoektrepos

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

type updater struct{}

var _ job.Job = &updater{}

func NewUpdater() job.Job {
	return &updater{}
}

func (j *updater) Description() string {
	return "zoektrepos.Updater updates the zoekt_repos table periodically to reflect the search-index status of each repository."
}

func (j *updater) Config() []env.Config {
	return nil
}

func (j *updater) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			&handler{
				db:     db,
				logger: observationCtx.Logger,
			},
			goroutine.WithName("search.index-status-reconciler"),
			goroutine.WithDescription("reconciles indexed status between zoekt and postgres"),
			goroutine.WithInterval(1*time.Hour),
		),
	}, nil
}

type handler struct {
	db     database.DB
	logger log.Logger
}

var (
	_ goroutine.Handler      = &handler{}
	_ goroutine.ErrorHandler = &handler{}
)

func (h *handler) Handle(ctx context.Context) error {
	indexed, err := search.ListAllIndexed(ctx, search.Indexed())
	if err != nil {
		return err
	}

	return h.db.ZoektRepos().UpdateIndexStatuses(ctx, indexed.ReposMap)
}

func (h *handler) HandleError(err error) {
	h.logger.Error("error updating zoekt repos", log.Error(err))
}
