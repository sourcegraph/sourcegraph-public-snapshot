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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/background/indexing"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/background/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
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

	routines := newBackgroundRoutines(observationContext)

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

func newBackgroundRoutines(observationContext *observation.Context) (routines []goroutine.BackgroundRoutine) {
	routines = append(routines, newCommitGraphRoutines(observationContext)...)
	routines = append(routines, newIndexingRoutines(observationContext)...)
	routines = append(routines, newJanitorRoutines(observationContext)...)
	return routines
}

func newCommitGraphRoutines(observationContext *observation.Context) []goroutine.BackgroundRoutine {
	dbStore := services.dbStore
	gitserverClient := services.gitserverClient
	operations := commitgraph.NewOperations(dbStore, observationContext)

	return []goroutine.BackgroundRoutine{
		commitgraph.NewUpdater(dbStore, gitserverClient, config.CommitGraphUpdateTaskInterval, operations),
	}
}

func newIndexingRoutines(observationContext *observation.Context) []goroutine.BackgroundRoutine {
	if !config.EnableAutoIndexing {
		return nil
	}

	dbStore := &indexing.DBStoreShim{services.dbStore}
	gitserverClient := services.gitserverClient
	operations := indexing.NewOperations(observationContext)

	return []goroutine.BackgroundRoutine{
		indexing.NewIndexScheduler(dbStore, gitserverClient, config.IndexBatchSize, config.MinimumTimeSinceLastEnqueue, config.MinimumSearchCount, float64(config.MinimumSearchRatio)/100, config.MinimumPreciseCount, config.AutoIndexingTaskInterval, operations),
		indexing.NewIndexabilityUpdater(dbStore, gitserverClient, config.MinimumSearchCount, float64(config.MinimumSearchRatio)/100, config.MinimumPreciseCount, config.AutoIndexingTaskInterval, operations),
	}
}

func newJanitorRoutines(observationContext *observation.Context) []goroutine.BackgroundRoutine {
	dbStore := &janitor.DBStoreShim{services.dbStore}
	uploadWorkerStore := dbstore.WorkerutilUploadStore(services.dbStore, observationContext)
	indexWorkerStore := dbstore.WorkerutilIndexStore(services.dbStore, observationContext)
	lsifStore := services.lsifStore
	metrics := janitor.NewMetrics(observationContext)

	return []goroutine.BackgroundRoutine{
		janitor.NewAbandonedUploadJanitor(dbStore, config.UploadTimeout, config.CleanupTaskInterval, metrics),
		janitor.NewDeletedRepositoryJanitor(dbStore, config.CleanupTaskInterval, metrics),
		janitor.NewHardDeleter(dbStore, lsifStore, config.CleanupTaskInterval, metrics),
		janitor.NewRecordExpirer(dbStore, config.DataTTL, config.CleanupTaskInterval, metrics),
		janitor.NewUploadResetter(uploadWorkerStore, config.CleanupTaskInterval, metrics, observationContext),
		janitor.NewIndexResetter(indexWorkerStore, config.CleanupTaskInterval, metrics, observationContext),
	}
}
