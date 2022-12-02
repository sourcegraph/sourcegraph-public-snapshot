package batches

import (
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// InitStore initializes and returns a *store.Store instance.
func InitStore() (*store.Store, error) {
	return initStore.Init()
}

var initStore = memo.NewMemoizedConstructor(func() (*store.Store, error) {
	observationContext := observation.NewContext(log.Scoped("store.batches", "batches store"))

	db, err := workerdb.InitDB(observationContext)
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
	observationContext := observation.NewContext(log.Scoped("store.reconciler", "reconciler worker store"))

	db, err := workerdb.InitDB(observationContext)
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
	observationContext := observation.NewContext(log.Scoped("store.bulk_ops", "bulk operation worker store"))

	db, err := workerdb.InitDB(observationContext)
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
	observationContext := observation.NewContext(log.Scoped("store.execution", "the batch spec workspace execution worker store"))

	db, err := workerdb.InitDB(observationContext)
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
	observationContext := observation.NewContext(log.Scoped("store.batch_spec_resolution", "the batch spec resolution worker store"))

	db, err := workerdb.InitDB(observationContext)
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecResolutionWorkerStore(db.Handle(), observationContext), nil
})
