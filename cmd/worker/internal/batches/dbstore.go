package batches

import (
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches/types"
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
	observationCtx := observation.NewContext(log.Scoped("store.batches"))

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return store.New(db, observationCtx, keyring.Default().BatchChangesCredentialKey), nil
})

// InitReconcilerWorkerStore initializes and returns a dbworker.Store instance for the reconciler worker.
func InitReconcilerWorkerStore() (dbworkerstore.Store[*types.Changeset], error) {
	return initReconcilerWorkerStore.Init()
}

var initReconcilerWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.Changeset], error) {
	observationCtx := observation.NewContext(log.Scoped("store.reconciler"))

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return store.NewReconcilerWorkerStore(observationCtx, db.Handle()), nil
})

// InitBulkOperationWorkerStore initializes and returns a dbworker.Store instance for the bulk operation processor worker.
func InitBulkOperationWorkerStore() (dbworkerstore.Store[*types.ChangesetJob], error) {
	return initBulkOperationWorkerStore.Init()
}

var initBulkOperationWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.ChangesetJob], error) {
	observationCtx := observation.NewContext(log.Scoped("store.bulk_ops"))

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return store.NewBulkOperationWorkerStore(observationCtx, db.Handle()), nil
})

// InitBatchSpecWorkspaceExecutionWorkerStore initializes and returns a dbworkerstore.Store instance for the batch spec workspace execution worker.
func InitBatchSpecWorkspaceExecutionWorkerStore() (dbworkerstore.Store[*types.BatchSpecWorkspaceExecutionJob], error) {
	return initBatchSpecWorkspaceExecutionWorkerStore.Init()
}

var initBatchSpecWorkspaceExecutionWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.BatchSpecWorkspaceExecutionJob], error) {
	observationCtx := observation.NewContext(log.Scoped("store.execution"))

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecWorkspaceExecutionWorkerStore(observationCtx, db.Handle()), nil
})

// InitBatchSpecResolutionWorkerStore initializes and returns a dbworker.Store instance for the batch spec workspace resolution worker.
func InitBatchSpecResolutionWorkerStore() (dbworkerstore.Store[*types.BatchSpecResolutionJob], error) {
	return initBatchSpecResolutionWorkerStore.Init()
}

var initBatchSpecResolutionWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*types.BatchSpecResolutionJob], error) {
	observationCtx := observation.NewContext(log.Scoped("store.batch_spec_resolution"))

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return store.NewBatchSpecResolutionWorkerStore(observationCtx, db.Handle()), nil
})
