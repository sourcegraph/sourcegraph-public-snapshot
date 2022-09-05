package batches

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// InitStore initializes and returns a *store.Store instance.
func InitStore() (*store.Store, error) {
	return initStore.Init()
}

var initStore = memo.NewMemoizedConstructor(func() (*store.Store, error) {
	logger := log.Scoped("store.batches", "batches store")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.New(database.NewDB(logger, db), observationContext, keyring.Default().BatchChangesCredentialKey), nil
})

// InitReconcilerWorkerStore initializes and returns a dbworker.Store instance for the reconciler worker.
func InitReconcilerWorkerStore() (dbworkerstore.Store, error) {
	return initReconcilerWorkerStore.Init()
}

var initReconcilerWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.reconciler", "reconciler worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.NewReconcilerWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext), nil
})

// InitBulkOperationWorkerStore initializes and returns a dbworker.Store instance for the bulk operation processor worker.
func InitBulkOperationWorkerStore() (dbworkerstore.Store, error) {
	return initBulkOperationWorkerStore.Init()
}

var initBulkOperationWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.bulk_ops", "bulk operation worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.NewBulkOperationWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext), nil
})

// InitBatchSpecWorkspaceExecutionWorkerStore initializes and returns a dbworkerstore.Store instance for the batch spec workspace execution worker.
func InitBatchSpecWorkspaceExecutionWorkerStore() (dbworkerstore.Store, error) {
	return initBatchSpecWorkspaceExecutionWorkerStore.Init()
}

var initBatchSpecWorkspaceExecutionWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.execution", "the batch spec workspace execution worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecWorkspaceExecutionWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext), nil
})

// InitBatchSpecResolutionWorkerStore initializes and returns a dbworker.Store instance for the batch spec workspace resolution worker.
func InitBatchSpecResolutionWorkerStore() (dbworkerstore.Store, error) {
	return initBatchSpecResolutionWorkerStore.Init()
}

var initBatchSpecResolutionWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.batch_spec_resolution", "the batch spec resolution worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecResolutionWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext), nil
})
