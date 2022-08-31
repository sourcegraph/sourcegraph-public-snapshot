package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/background/repomatcher"
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

func (j *policiesRepositoryMatcherJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationCtx := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := repomatcher.NewMetrics(observationCtx)

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		repomatcher.NewMatcher(dbStore, metrics),
	}, nil
}
