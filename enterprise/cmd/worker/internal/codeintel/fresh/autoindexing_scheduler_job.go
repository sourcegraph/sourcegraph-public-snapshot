package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/scheduler"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
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
		scheduler.ConfigInst,
	}
}

func (j *autoindexingScheduler) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel autoindexing scheduling routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize stores
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	databaseDB := database.NewDB(logger, db)

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return nil, err
	}

	// Initialize necessary clients
	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}
	repoUpdater := codeintel.InitRepoUpdaterClient()
	policyMatcher := policies.NewMatcher(gitserverClient, policies.IndexingExtractor, false, true)

	// Initialize services
	uploadSvc := uploads.GetService(databaseDB, database.NewDBWith(logger, lsifStore), gitserverClient)
	autoindexingSvc := autoindexing.GetService(databaseDB, uploadSvc, gitserverClient, repoUpdater)

	return []goroutine.BackgroundRoutine{
		scheduler.NewScheduler(autoindexingSvc, dbStore, policyMatcher, observationContext),
	}, nil
}
