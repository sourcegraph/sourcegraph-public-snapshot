package codeintel

import (
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// InitDBStore initializes and returns a db store instance.
func InitDBStore() (*dbstore.Store, error) {
	conn, err := initDBStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*dbstore.Store), nil
}

var initDBStore = shared.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := shared.InitDatabase()
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

var initDependencySyncStore = shared.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
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

var initDependenyIndexStore = shared.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return dbstore.WorkerutilDependencyIndexStore(dbStore, observationContext), nil
})
