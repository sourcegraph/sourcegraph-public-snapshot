package codeintel

import (
	"context"

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

type autoindexingScheduler struct{}

func NewAutoindexingSchedulerJob() job.Job {
	return &autoindexingScheduler{}
}

func (j *autoindexingScheduler) Description() string {
	return ""
}

func (j *autoindexingScheduler) Config() []env.Config {
	return []env.Config{
		autoindexing.ConfigIndexingInst,
	}
}

func (j *autoindexingScheduler) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDBWithLogger(observationCtx)
	if err != nil {
		return nil, err
	}

	// TODO: nsc move scope down
	gitserverClient := gitserver.New(db, observation.ScopedContext("codeintel", "indexScheduler", "gitserver", observationCtx))

	return autoindexing.NewIndexSchedulers(
		services.UploadsService,
		services.PoliciesService,
		policies.NewMatcher(gitserverClient, policies.IndexingExtractor, false, true),
		services.AutoIndexingService,
		observationCtx,
	), nil
}
