package own

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/background"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ownRepoIndexingQueue struct{}

func NewOwnRepoIndexingQueue() job.Job {
	return &ownRepoIndexingQueue{}
}

func (o *ownRepoIndexingQueue) Description() string {
	return "Queue used by Sourcegraph Own to index ownership data partitioned per repository"
}

func (o *ownRepoIndexingQueue) Config() []env.Config {
	return nil
}

func (o *ownRepoIndexingQueue) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	var routines []goroutine.BackgroundRoutine
	routines = append(routines, background.NewOwnBackgroundWorker(context.Background(), db, observationCtx)...)
	routines = append(routines, background.GetOwnIndexSchedulerRoutines(db, observationCtx)...)

	return routines, nil
}
