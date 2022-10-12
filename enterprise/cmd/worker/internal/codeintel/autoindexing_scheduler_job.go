package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	policies "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type autoindexingScheduler struct {
	observationContext *observation.Context
}

func NewAutoindexingSchedulerJob(observationContext *observation.Context) job.Job {
	return &autoindexingScheduler{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *autoindexingScheduler) Description() string {
	return ""
}

func (j *autoindexingScheduler) Config() []env.Config {
	return []env.Config{
		autoindexing.ConfigIndexingInst,
	}
}

func (j *autoindexingScheduler) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.New(db, observation.ScopedContext("codeintel", "indexScheduler", "gitserver", j.observationContext))

	return autoindexing.NewIndexSchedulers(
		services.UploadsService,
		services.PoliciesService,
		policies.NewMatcher(gitserverClient, policies.IndexingExtractor, false, true),
		services.AutoIndexingService,
		observation.ContextWithLogger(logger, j.observationContext),
	), nil
}
