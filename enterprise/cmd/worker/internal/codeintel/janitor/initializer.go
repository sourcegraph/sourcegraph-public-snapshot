package janitor

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
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
	lsifStore, err := codeintelshared.InitLSIFStore()
	if err != nil {
		return nil, err
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	janitorMetrics := NewMetrics(observationContext)

	return []goroutine.BackgroundRoutine{
		NewAbandonedUploadJanitor(
			&DBStoreShim{Store: dbStore},
			config.UploadTimeout,
			config.CleanupTaskInterval,
			janitorMetrics,
		),
		NewDeletedRepositoryJanitor(
			&DBStoreShim{Store: dbStore},
			config.CleanupTaskInterval,
			janitorMetrics,
		),
		NewHardDeleter(
			&DBStoreShim{Store: dbStore},
			lsifStore,
			config.CleanupTaskInterval,
			janitorMetrics,
		),
		NewRecordExpirer(
			&DBStoreShim{Store: dbStore},
			config.DataTTL,
			config.CleanupTaskInterval,
			janitorMetrics,
		),
		NewUnknownCommitJanitor(
			&DBStoreShim{Store: dbStore},
			config.CommitResolverMinimumTimeSinceLastCheck,
			config.CommitResolverBatchSize,
			config.CommitResolverTaskInterval,
			janitorMetrics,
		),
		NewUploadResetter(
			dbstore.WorkerutilUploadStore(dbStore, observationContext),
			config.CleanupTaskInterval,
			janitorMetrics,
			observationContext,
		),
		NewIndexResetter(
			dbstore.WorkerutilIndexStore(dbStore, observationContext),
			config.CleanupTaskInterval,
			janitorMetrics,
			observationContext,
		),
	}, nil
}
