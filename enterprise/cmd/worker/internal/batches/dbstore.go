package batches

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/memo"
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

	db, err := workerdb.InitDBWithLogger(logger)
	if err != nil {
		return nil, err
	}

	return store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey), nil
})

// InitReconcilerWorkerStore initializes and returns a dbworker.Store instance for the reconciler worker.
func InitReconcilerWorkerStore() (dbworkerstore.Store[*types.Changeset], error) {
	return initReconcilerWorkerStore.Init()
}

var initReconcilerWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.Changeset], error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.reconciler", "reconciler worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.InitDBWithLogger(observationContext.Logger)
	if err != nil {
		return nil, err
	}

	return store.NewReconcilerWorkerStore(db.Handle(), observationContext), nil
})

// InitBulkOperationWorkerStore initializes and returns a dbworker.Store instance for the bulk operation processor worker.
func InitBulkOperationWorkerStore() (dbworkerstore.Store[*types.ChangesetJob], error) {
	return initBulkOperationWorkerStore.Init()
}

var initBulkOperationWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.ChangesetJob], error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.bulk_ops", "bulk operation worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.InitDBWithLogger(observationContext.Logger)
	if err != nil {
		return nil, err
	}

	return store.NewBulkOperationWorkerStore(db.Handle(), observationContext), nil
})

// InitBatchSpecWorkspaceExecutionWorkerStore initializes and returns a dbworkerstore.Store instance for the batch spec workspace execution worker.
func InitBatchSpecWorkspaceExecutionWorkerStore() (dbworkerstore.Store[*types.BatchSpecWorkspaceExecutionJob], error) {
	return initBatchSpecWorkspaceExecutionWorkerStore.Init()
}

var initBatchSpecWorkspaceExecutionWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.BatchSpecWorkspaceExecutionJob], error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.execution", "the batch spec workspace execution worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.InitDBWithLogger(observationContext.Logger)
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecWorkspaceExecutionWorkerStore(db.Handle(), observationContext), nil
})

// InitBatchSpecResolutionWorkerStore initializes and returns a dbworker.Store instance for the batch spec workspace resolution worker.
func InitBatchSpecResolutionWorkerStore() (dbworkerstore.Store[*types.BatchSpecResolutionJob], error) {
	return initBatchSpecResolutionWorkerStore.Init()
}

var initBatchSpecResolutionWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.BatchSpecResolutionJob], error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.batch_spec_resolution", "the batch spec resolution worker store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.InitDBWithLogger(observationContext.Logger)
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecResolutionWorkerStore(db.Handle(), observationContext), nil
})
