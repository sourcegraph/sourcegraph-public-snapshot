package codeintel

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type outlineIndexingScheduler struct{}

func NewOutlineIndexingSchedulerJob() job.Job {
	return &outlineIndexingScheduler{}
}

func (j *outlineIndexingScheduler) Description() string {
	return ""
}

func (j *outlineIndexingScheduler) Config() []env.Config {
	return []env.Config{
		autoindexing.SchedulerConfigInst,
	}
}

func (j *outlineIndexingScheduler) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	fmt.Println("start the scheduling")
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	matcher := policies.NewMatcher(
		services.GitserverClient,
		policies.IndexingExtractor,
		false,
		true,
	)

	return autoindexing.NewIndexSchedulers(
		observationCtx,
		services.UploadsService,
		services.PoliciesService,
		matcher,
		services.AutoIndexingService,
		db.Repos(),
	), nil
}
