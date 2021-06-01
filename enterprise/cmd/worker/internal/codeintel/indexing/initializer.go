package indexing

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type initializer struct{}

func NewInitializer() shared.SetupHook {
	return &initializer{}
}

func (i *initializer) Config() []env.Config {
	return []env.Config{
		config,
	}
}

func (i *initializer) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	dbStore, err := codeintelshared.InitDBStore()
	if err != nil {
		return nil, err
	}
	gitserverClient, err := codeintelshared.InitGitserverClient()
	if err != nil {
		return nil, err
	}
	indexEnqueuer, err := codeintelshared.InitIndexEnqueuer()
	if err != nil {
		return nil, err
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return []goroutine.BackgroundRoutine{
		NewIndexScheduler(
			dbStore,
			indexEnqueuer,
			config.IndexBatchSize,
			config.MinimumTimeSinceLastEnqueue,
			config.MinimumSearchCount,
			float64(config.MinimumSearchRatio)/100,
			config.MinimumPreciseCount,
			config.AutoIndexingTaskInterval,
			observationContext,
		),
		NewIndexabilityUpdater(
			dbStore,
			gitserverClient,
			config.MinimumSearchCount,
			float64(config.MinimumSearchRatio)/100,
			config.MinimumPreciseCount,
			config.AutoIndexingSkipManualInterval,
			config.AutoIndexingTaskInterval,
			observationContext,
		),
	}, nil
}
