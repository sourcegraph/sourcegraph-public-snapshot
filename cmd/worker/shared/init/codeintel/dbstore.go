package codeintel

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// InitDBStore initializes and returns a db store instance.
func InitDBStore() (*dbstore.Store, error) {
	return initDBStore.Init()
}

var initDBStore = memo.NewMemoizedConstructor(func() (*dbstore.Store, error) {
	logger := log.Scoped("store", "codeintel db store")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return dbstore.NewWithDB(database.NewDB(logger, db), observationContext), nil
})

// InitDependencySyncingStore initializes and returns a dependency index store.
func InitDependencySyncingStore() (dbworkerstore.Store, error) {
	return initDependencySyncStore.Init()
}

var initDependencySyncStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.dependency_sync", "dependency sync store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return dbstore.WorkerutilDependencySyncStore(dbStore, observationContext), nil
})

func InitDependencyIndexingStore() (dbworkerstore.Store, error) {
	return initDependenyIndexStore.Init()
}

var initDependenyIndexStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.dependency_index", "dependency index store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return dbstore.WorkerutilDependencyIndexStore(dbStore, observationContext), nil
})
