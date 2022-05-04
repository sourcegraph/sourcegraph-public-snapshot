package codeintel

import (
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// InitDBStore initializes and returns a db store instance.
func InitDBStore() (*dbstore.Store, error) {
	conn, err := initDBStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*dbstore.Store), nil
}

var initDBStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store", "codeintel db store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return dbstore.NewWithDB(db, observationContext), nil
})

// InitDependencySyncingStore initializes and returns a dependency index store.
func InitDependencySyncingStore() (dbworkerstore.Store, error) {
	store, err := initDependencySyncStore.Init()
	if err != nil {
		return nil, err
	}

	return store.(dbworkerstore.Store), nil
}

var initDependencySyncStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.dependency_sync", "dependency sync store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return dbstore.WorkerutilDependencySyncStore(dbStore, observationContext), nil
})

func InitDependencyIndexingStore() (dbworkerstore.Store, error) {
	store, err := initDependenyIndexStore.Init()
	if err != nil {
		return nil, err
	}

	return store.(dbworkerstore.Store), nil
}

var initDependenyIndexStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.dependency_index", "dependency index store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return dbstore.WorkerutilDependencyIndexStore(dbStore, observationContext), nil
})
