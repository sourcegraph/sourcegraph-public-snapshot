package codeintel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commits"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	if err := initServices(ctx); err != nil {
		return err
	}

	resolver, err := newResolver(ctx)
	if err != nil {
		return err
	}

	uploadHandler, err := newUploadHandler(ctx)
	if err != nil {
		return err
	}

	routines, err := newBackgroundRoutines()
	if err != nil {
		return err
	}

	enterpriseServices.CodeIntelResolver = resolver
	enterpriseServices.NewCodeIntelUploadHandler = uploadHandler

	// TODO(efritz) - return these to the frontend to run.
	// Requires refactoring of the frontend server setup
	// so I'm going to kick that can down the road for a
	// short while.
	//
	// Repo updater is currently doing something similar
	// here and would also be ripe for a refresher of the
	// startup flow.
	go goroutine.MonitorBackgroundRoutines(context.Background(), routines...)

	return nil
}

func newResolver(ctx context.Context) (gql.CodeIntelResolver, error) {
	hunkCache, err := codeintelresolvers.NewHunkCache(config.HunkCacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hunk cache: %s", err)
	}

	resolver := codeintelgqlresolvers.NewResolver(codeintelresolvers.NewResolver(
		services.store,
		services.bundleManagerClient,
		services.api,
		hunkCache,
	))

	return resolver, err
}

func newUploadHandler(ctx context.Context) (func(internal bool) http.Handler, error) {
	internalHandler, err := NewCodeIntelUploadHandler(ctx, true)
	if err != nil {
		return nil, err
	}

	externalHandler, err := NewCodeIntelUploadHandler(ctx, false)
	if err != nil {
		return nil, err
	}

	uploadHandler := func(internal bool) http.Handler {
		if internal {
			return internalHandler
		}

		return externalHandler
	}

	return uploadHandler, nil
}

func newBackgroundRoutines() ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := background.NewMetrics(observationContext.Registerer)

	routines := []goroutine.BackgroundRoutine{
		background.NewAbandonedUploadJanitor(services.store, config.UploadTimeout, config.BackgroundTaskInterval, metrics),
		background.NewDeletedRepositoryJanitor(services.store, config.BackgroundTaskInterval, metrics),
		background.NewHardDeleter(services.store, services.lsifStore, config.BackgroundTaskInterval, metrics),
		background.NewIndexScheduler(services.store, services.gitserverClient, config.IndexBatchSize, config.MinimumTimeSinceLastEnqueue, config.MinimumSearchCount, float64(config.MinimumSearchRatio)/100, config.MinimumPreciseCount, config.BackgroundTaskInterval, metrics),
		background.NewIndexabilityUpdater(services.store, services.gitserverClient, config.MinimumSearchCount, float64(config.MinimumSearchRatio)/100, config.MinimumPreciseCount, config.BackgroundTaskInterval, metrics),
		background.NewRecordExpirer(services.store, config.DataTTL, config.BackgroundTaskInterval, metrics),
		background.NewUploadResetter(services.store, config.BackgroundTaskInterval, metrics),
		background.NewIndexResetter(services.store, config.BackgroundTaskInterval, metrics),
		background.NewCommitUpdater(services.store, commits.NewUpdater(services.store, services.gitserverClient), config.BackgroundTaskInterval),
	}

	return routines, nil
}
