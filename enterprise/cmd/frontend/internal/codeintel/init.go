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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	if err := initServices(ctx); err != nil {
		return err
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	resolver, err := newResolver(ctx, observationContext)
	if err != nil {
		return err
	}

	uploadHandler, err := newUploadHandler(ctx)
	if err != nil {
		return err
	}

	routines, err := newBackgroundRoutines(observationContext)
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

func newResolver(ctx context.Context, observationContext *observation.Context) (gql.CodeIntelResolver, error) {
	hunkCache, err := codeintelresolvers.NewHunkCache(config.HunkCacheSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize hunk cache: %s", err)
	}

	innerResolver := codeintelresolvers.NewResolver(
		&resolvers.DBStoreShim{services.dbStore},
		services.lsifStore,
		services.api,
		hunkCache,
		observationContext,
	)
	resolver := codeintelgqlresolvers.NewResolver(innerResolver)

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

func newBackgroundRoutines(observationContext *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	dbStoreShim := &background.DBStoreShim{services.dbStore}
	lsifStoreShim := services.lsifStore
	gitserverClient := services.gitserverClient
	metrics := background.NewMetrics(observationContext.Registerer)

	routines := []goroutine.BackgroundRoutine{
		// Commit graph updater
		background.NewCommitUpdater(dbStoreShim, gitserverClient, config.CommitGraphUpdateTaskInterval),

		// Cleanup
		background.NewAbandonedUploadJanitor(dbStoreShim, config.UploadTimeout, config.CleanupTaskInterval, metrics),
		background.NewDeletedRepositoryJanitor(dbStoreShim, config.CleanupTaskInterval, metrics),
		background.NewHardDeleter(dbStoreShim, lsifStoreShim, config.CleanupTaskInterval, metrics),
		background.NewRecordExpirer(dbStoreShim, config.DataTTL, config.CleanupTaskInterval, metrics),
		background.NewUploadResetter(dbStoreShim, config.CleanupTaskInterval, metrics),
		background.NewIndexResetter(dbStoreShim, config.CleanupTaskInterval, metrics),
	}

	if config.EnableAutoIndexing {
		// Auto indexing
		routines = append(
			routines,
			background.NewIndexScheduler(dbStoreShim, gitserverClient, config.IndexBatchSize, config.MinimumTimeSinceLastEnqueue, config.MinimumSearchCount, float64(config.MinimumSearchRatio)/100, config.MinimumPreciseCount, config.AutoIndexingTaskInterval, metrics),
			background.NewIndexabilityUpdater(dbStoreShim, gitserverClient, config.MinimumSearchCount, float64(config.MinimumSearchRatio)/100, config.MinimumPreciseCount, config.AutoIndexingTaskInterval, metrics),
		)
	}

	return routines, nil
}
