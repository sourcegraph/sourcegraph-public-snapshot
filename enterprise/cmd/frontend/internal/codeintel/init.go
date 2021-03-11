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
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services) error {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	if err := initServices(ctx, db); err != nil {
		return err
	}

	if err := registerMigrations(ctx, db, outOfBandMigrationRunner); err != nil {
		return err
	}

	resolver, err := newResolver(ctx, db, observationContext)
	if err != nil {
		return err
	}

	uploadHandler, err := newUploadHandler(ctx, db)
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

func newResolver(ctx context.Context, db dbutil.DB, observationContext *observation.Context) (gql.CodeIntelResolver, error) {
	hunkCache, err := codeintelresolvers.NewHunkCache(config.HunkCacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hunk cache: %s", err)
	}

	innerResolver := codeintelresolvers.NewResolver(
		services.dbStore,
		services.lsifStore,
		services.gitserverClient,
		services.indexEnqueuer,
		hunkCache,
		observationContext,
	)
	resolver := codeintelgqlresolvers.NewResolver(db, innerResolver)

	return resolver, err
}

func newUploadHandler(ctx context.Context, db dbutil.DB) (func(internal bool) http.Handler, error) {
	internalHandler, err := NewCodeIntelUploadHandler(ctx, db, true)
	if err != nil {
		return nil, err
	}

	externalHandler, err := NewCodeIntelUploadHandler(ctx, db, false)
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
	return []goroutine.BackgroundRoutine{
		commitgraph.NewUpdater(
			services.dbStore,
			services.gitserverClient,
			config.CommitGraphUpdateTaskInterval,
			observationContext,
		),
	}
}

func newIndexingRoutines(observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		indexing.NewIndexScheduler(
			services.dbStore,
			services.indexEnqueuer,
			config.IndexBatchSize,
			config.MinimumTimeSinceLastEnqueue,
			config.MinimumSearchCount,
			float64(config.MinimumSearchRatio)/100,
			config.MinimumPreciseCount,
			config.AutoIndexingTaskInterval,
			observationContext,
		),
		indexing.NewIndexabilityUpdater(
			services.dbStore,
			services.gitserverClient,
			config.MinimumSearchCount,
			float64(config.MinimumSearchRatio)/100,
			config.MinimumPreciseCount,
			config.AutoIndexingSkipManualInterval,
			config.AutoIndexingTaskInterval,
			observationContext,
		),
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
