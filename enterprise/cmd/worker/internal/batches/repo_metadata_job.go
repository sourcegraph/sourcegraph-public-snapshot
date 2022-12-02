package batches

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/workers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"go.opentelemetry.io/otel"
)

type repoMetadataJob struct{}

func NewRepoMetadataJob() job.Job {
	return &repoMetadataJob{}
}

func (j *repoMetadataJob) Description() string {
	return ""
}

func (j *repoMetadataJob) Config() []env.Config {
	return []env.Config{}
}

func (j *repoMetadataJob) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "repo metadata job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	workCtx := actor.WithInternalActor(context.Background())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	store, err := InitRepoMetadataWorkerStore()
	if err != nil {
		return nil, err
	}

	worker := workers.NewRepoMetadataWorker(
		workCtx,
		bstore,
		store,
		gitserver.NewClient(bstore.DatabaseDB()),
		observationContext,
	)

	routines := []goroutine.BackgroundRoutine{worker}

	return routines, nil
}
