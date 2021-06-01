package commitgraph

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
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := codeintelshared.InitDBStore()
	if err != nil {
		return nil, err
	}
	gitserverClient, err := codeintelshared.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		NewUpdater(
			dbStore,
			gitserverClient,
			config.CommitGraphUpdateTaskInterval,
			observationContext,
		),
	}, nil
}
