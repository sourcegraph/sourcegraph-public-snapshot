package batches

import (
	"database/sql"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// InitStore initializes and returns a *store.Store instance.
func InitStore() (*store.Store, error) {
	conn, err := initStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*store.Store), nil
}

var initStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.batches", "batches store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey), nil
})

// InitReconcilerWorkerStore initializes and returns a dbworker.Store instance for the reconciler worker.
func InitReconcilerWorkerStore() (dbworkerstore.Store, error) {
	conn, err := initReconcilerWorkerStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(dbworkerstore.Store), nil
}

var initReconcilerWorkerStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.reconciler", "reconciler worker store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
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
	conn, err := initBulkOperationWorkerStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(dbworkerstore.Store), nil
}

var initBulkOperationWorkerStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.bulk_ops", "bulk operation worker store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.NewBulkOperationWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext), nil
})

// InitBatchSpecWorkspaceExecutionWorkerStore initializes and returns a store.BatchSpecWorkspaceExecutionWorkerStore instance for the batch spec workspace execution worker.
func InitBatchSpecWorkspaceExecutionWorkerStore() (store.BatchSpecWorkspaceExecutionWorkerStore, error) {
	conn, err := initBatchSpecWorkspaceExecutionWorkerStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(store.BatchSpecWorkspaceExecutionWorkerStore), nil
}

var initBatchSpecWorkspaceExecutionWorkerStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.execution", "the batch spec workspace execution worker store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
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
	conn, err := initBatchSpecResolutionWorkerStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(dbworkerstore.Store), nil
}

var initBatchSpecResolutionWorkerStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.batch_spec_resolution", "the batch spec resolution worker store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecResolutionWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext), nil
})
