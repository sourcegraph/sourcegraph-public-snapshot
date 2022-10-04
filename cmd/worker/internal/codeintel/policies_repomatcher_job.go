package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/background/repomatcher"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type policiesRepositoryMatcherJob struct{}

func NewPoliciesRepositoryMatcherJob() job.Job {
	return &policiesRepositoryMatcherJob{}
}

func (j *policiesRepositoryMatcherJob) Description() string {
	return ""
}

func (j *policiesRepositoryMatcherJob) Config() []env.Config {
	return []env.Config{
		repomatcher.ConfigInst,
	}
}

func (j *policiesRepositoryMatcherJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationCtx := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := repomatcher.NewMetrics(observationCtx)

	rawDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, rawDB)

	rawCodeIntelDB, err := codeintel.InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}
	codeIntelDB := database.NewDB(logger, rawCodeIntelDB)

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	uploadSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	policySvc := policies.GetService(db, uploadSvc, gitserverClient)

	return []goroutine.BackgroundRoutine{
		repomatcher.NewMatcher(policySvc, metrics),
	}, nil
}
