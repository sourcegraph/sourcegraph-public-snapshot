// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge butoindexing

import (
	"context"
	"sync"
	"time"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	store "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	shbred1 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	protocol "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store)
// used for unit testing.
type MockStore struct {
	// GetIndexConfigurbtionByRepositoryIDFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetIndexConfigurbtionByRepositoryID.
	GetIndexConfigurbtionByRepositoryIDFunc *StoreGetIndexConfigurbtionByRepositoryIDFunc
	// GetInferenceScriptFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetInferenceScript.
	GetInferenceScriptFunc *StoreGetInferenceScriptFunc
	// GetLbstIndexScbnForRepositoryFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetLbstIndexScbnForRepository.
	GetLbstIndexScbnForRepositoryFunc *StoreGetLbstIndexScbnForRepositoryFunc
	// GetQueuedRepoRevFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetQueuedRepoRev.
	GetQueuedRepoRevFunc *StoreGetQueuedRepoRevFunc
	// GetRepositoriesForIndexScbnFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetRepositoriesForIndexScbn.
	GetRepositoriesForIndexScbnFunc *StoreGetRepositoriesForIndexScbnFunc
	// InsertDependencyIndexingJobFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// InsertDependencyIndexingJob.
	InsertDependencyIndexingJobFunc *StoreInsertDependencyIndexingJobFunc
	// InsertIndexesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertIndexes.
	InsertIndexesFunc *StoreInsertIndexesFunc
	// IsQueuedFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method IsQueued.
	IsQueuedFunc *StoreIsQueuedFunc
	// IsQueuedRootIndexerFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method IsQueuedRootIndexer.
	IsQueuedRootIndexerFunc *StoreIsQueuedRootIndexerFunc
	// MbrkRepoRevsAsProcessedFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method MbrkRepoRevsAsProcessed.
	MbrkRepoRevsAsProcessedFunc *StoreMbrkRepoRevsAsProcessedFunc
	// QueueRepoRevFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method QueueRepoRev.
	QueueRepoRevFunc *StoreQueueRepoRevFunc
	// RepositoryExceptionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RepositoryExceptions.
	RepositoryExceptionsFunc *StoreRepositoryExceptionsFunc
	// RepositoryIDsWithConfigurbtionFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// RepositoryIDsWithConfigurbtion.
	RepositoryIDsWithConfigurbtionFunc *StoreRepositoryIDsWithConfigurbtionFunc
	// SetConfigurbtionSummbryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetConfigurbtionSummbry.
	SetConfigurbtionSummbryFunc *StoreSetConfigurbtionSummbryFunc
	// SetInferenceScriptFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetInferenceScript.
	SetInferenceScriptFunc *StoreSetInferenceScriptFunc
	// SetRepositoryExceptionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetRepositoryExceptions.
	SetRepositoryExceptionsFunc *StoreSetRepositoryExceptionsFunc
	// TopRepositoriesToConfigureFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// TopRepositoriesToConfigure.
	TopRepositoriesToConfigureFunc *StoreTopRepositoriesToConfigureFunc
	// TruncbteConfigurbtionSummbryFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// TruncbteConfigurbtionSummbry.
	TruncbteConfigurbtionSummbryFunc *StoreTruncbteConfigurbtionSummbryFunc
	// UpdbteIndexConfigurbtionByRepositoryIDFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// UpdbteIndexConfigurbtionByRepositoryID.
	UpdbteIndexConfigurbtionByRepositoryIDFunc *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc
	// WithTrbnsbctionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithTrbnsbction.
	WithTrbnsbctionFunc *StoreWithTrbnsbctionFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		GetIndexConfigurbtionByRepositoryIDFunc: &StoreGetIndexConfigurbtionByRepositoryIDFunc{
			defbultHook: func(context.Context, int) (r0 shbred.IndexConfigurbtion, r1 bool, r2 error) {
				return
			},
		},
		GetInferenceScriptFunc: &StoreGetInferenceScriptFunc{
			defbultHook: func(context.Context) (r0 string, r1 error) {
				return
			},
		},
		GetLbstIndexScbnForRepositoryFunc: &StoreGetLbstIndexScbnForRepositoryFunc{
			defbultHook: func(context.Context, int) (r0 *time.Time, r1 error) {
				return
			},
		},
		GetQueuedRepoRevFunc: &StoreGetQueuedRepoRevFunc{
			defbultHook: func(context.Context, int) (r0 []store.RepoRev, r1 error) {
				return
			},
		},
		GetRepositoriesForIndexScbnFunc: &StoreGetRepositoriesForIndexScbnFunc{
			defbultHook: func(context.Context, time.Durbtion, bool, *int, int, time.Time) (r0 []int, r1 error) {
				return
			},
		},
		InsertDependencyIndexingJobFunc: &StoreInsertDependencyIndexingJobFunc{
			defbultHook: func(context.Context, int, string, time.Time) (r0 int, r1 error) {
				return
			},
		},
		InsertIndexesFunc: &StoreInsertIndexesFunc{
			defbultHook: func(context.Context, []shbred1.Index) (r0 []shbred1.Index, r1 error) {
				return
			},
		},
		IsQueuedFunc: &StoreIsQueuedFunc{
			defbultHook: func(context.Context, int, string) (r0 bool, r1 error) {
				return
			},
		},
		IsQueuedRootIndexerFunc: &StoreIsQueuedRootIndexerFunc{
			defbultHook: func(context.Context, int, string, string, string) (r0 bool, r1 error) {
				return
			},
		},
		MbrkRepoRevsAsProcessedFunc: &StoreMbrkRepoRevsAsProcessedFunc{
			defbultHook: func(context.Context, []int) (r0 error) {
				return
			},
		},
		QueueRepoRevFunc: &StoreQueueRepoRevFunc{
			defbultHook: func(context.Context, int, string) (r0 error) {
				return
			},
		},
		RepositoryExceptionsFunc: &StoreRepositoryExceptionsFunc{
			defbultHook: func(context.Context, int) (r0 bool, r1 bool, r2 error) {
				return
			},
		},
		RepositoryIDsWithConfigurbtionFunc: &StoreRepositoryIDsWithConfigurbtionFunc{
			defbultHook: func(context.Context, int, int) (r0 []shbred1.RepositoryWithAvbilbbleIndexers, r1 int, r2 error) {
				return
			},
		},
		SetConfigurbtionSummbryFunc: &StoreSetConfigurbtionSummbryFunc{
			defbultHook: func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) (r0 error) {
				return
			},
		},
		SetInferenceScriptFunc: &StoreSetInferenceScriptFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		SetRepositoryExceptionsFunc: &StoreSetRepositoryExceptionsFunc{
			defbultHook: func(context.Context, int, bool, bool) (r0 error) {
				return
			},
		},
		TopRepositoriesToConfigureFunc: &StoreTopRepositoriesToConfigureFunc{
			defbultHook: func(context.Context, int) (r0 []shbred1.RepositoryWithCount, r1 error) {
				return
			},
		},
		TruncbteConfigurbtionSummbryFunc: &StoreTruncbteConfigurbtionSummbryFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
		UpdbteIndexConfigurbtionByRepositoryIDFunc: &StoreUpdbteIndexConfigurbtionByRepositoryIDFunc{
			defbultHook: func(context.Context, int, []byte) (r0 error) {
				return
			},
		},
		WithTrbnsbctionFunc: &StoreWithTrbnsbctionFunc{
			defbultHook: func(context.Context, func(tx store.Store) error) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		GetIndexConfigurbtionByRepositoryIDFunc: &StoreGetIndexConfigurbtionByRepositoryIDFunc{
			defbultHook: func(context.Context, int) (shbred.IndexConfigurbtion, bool, error) {
				pbnic("unexpected invocbtion of MockStore.GetIndexConfigurbtionByRepositoryID")
			},
		},
		GetInferenceScriptFunc: &StoreGetInferenceScriptFunc{
			defbultHook: func(context.Context) (string, error) {
				pbnic("unexpected invocbtion of MockStore.GetInferenceScript")
			},
		},
		GetLbstIndexScbnForRepositoryFunc: &StoreGetLbstIndexScbnForRepositoryFunc{
			defbultHook: func(context.Context, int) (*time.Time, error) {
				pbnic("unexpected invocbtion of MockStore.GetLbstIndexScbnForRepository")
			},
		},
		GetQueuedRepoRevFunc: &StoreGetQueuedRepoRevFunc{
			defbultHook: func(context.Context, int) ([]store.RepoRev, error) {
				pbnic("unexpected invocbtion of MockStore.GetQueuedRepoRev")
			},
		},
		GetRepositoriesForIndexScbnFunc: &StoreGetRepositoriesForIndexScbnFunc{
			defbultHook: func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error) {
				pbnic("unexpected invocbtion of MockStore.GetRepositoriesForIndexScbn")
			},
		},
		InsertDependencyIndexingJobFunc: &StoreInsertDependencyIndexingJobFunc{
			defbultHook: func(context.Context, int, string, time.Time) (int, error) {
				pbnic("unexpected invocbtion of MockStore.InsertDependencyIndexingJob")
			},
		},
		InsertIndexesFunc: &StoreInsertIndexesFunc{
			defbultHook: func(context.Context, []shbred1.Index) ([]shbred1.Index, error) {
				pbnic("unexpected invocbtion of MockStore.InsertIndexes")
			},
		},
		IsQueuedFunc: &StoreIsQueuedFunc{
			defbultHook: func(context.Context, int, string) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.IsQueued")
			},
		},
		IsQueuedRootIndexerFunc: &StoreIsQueuedRootIndexerFunc{
			defbultHook: func(context.Context, int, string, string, string) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.IsQueuedRootIndexer")
			},
		},
		MbrkRepoRevsAsProcessedFunc: &StoreMbrkRepoRevsAsProcessedFunc{
			defbultHook: func(context.Context, []int) error {
				pbnic("unexpected invocbtion of MockStore.MbrkRepoRevsAsProcessed")
			},
		},
		QueueRepoRevFunc: &StoreQueueRepoRevFunc{
			defbultHook: func(context.Context, int, string) error {
				pbnic("unexpected invocbtion of MockStore.QueueRepoRev")
			},
		},
		RepositoryExceptionsFunc: &StoreRepositoryExceptionsFunc{
			defbultHook: func(context.Context, int) (bool, bool, error) {
				pbnic("unexpected invocbtion of MockStore.RepositoryExceptions")
			},
		},
		RepositoryIDsWithConfigurbtionFunc: &StoreRepositoryIDsWithConfigurbtionFunc{
			defbultHook: func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error) {
				pbnic("unexpected invocbtion of MockStore.RepositoryIDsWithConfigurbtion")
			},
		},
		SetConfigurbtionSummbryFunc: &StoreSetConfigurbtionSummbryFunc{
			defbultHook: func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error {
				pbnic("unexpected invocbtion of MockStore.SetConfigurbtionSummbry")
			},
		},
		SetInferenceScriptFunc: &StoreSetInferenceScriptFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockStore.SetInferenceScript")
			},
		},
		SetRepositoryExceptionsFunc: &StoreSetRepositoryExceptionsFunc{
			defbultHook: func(context.Context, int, bool, bool) error {
				pbnic("unexpected invocbtion of MockStore.SetRepositoryExceptions")
			},
		},
		TopRepositoriesToConfigureFunc: &StoreTopRepositoriesToConfigureFunc{
			defbultHook: func(context.Context, int) ([]shbred1.RepositoryWithCount, error) {
				pbnic("unexpected invocbtion of MockStore.TopRepositoriesToConfigure")
			},
		},
		TruncbteConfigurbtionSummbryFunc: &StoreTruncbteConfigurbtionSummbryFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockStore.TruncbteConfigurbtionSummbry")
			},
		},
		UpdbteIndexConfigurbtionByRepositoryIDFunc: &StoreUpdbteIndexConfigurbtionByRepositoryIDFunc{
			defbultHook: func(context.Context, int, []byte) error {
				pbnic("unexpected invocbtion of MockStore.UpdbteIndexConfigurbtionByRepositoryID")
			},
		},
		WithTrbnsbctionFunc: &StoreWithTrbnsbctionFunc{
			defbultHook: func(context.Context, func(tx store.Store) error) error {
				pbnic("unexpected invocbtion of MockStore.WithTrbnsbction")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i store.Store) *MockStore {
	return &MockStore{
		GetIndexConfigurbtionByRepositoryIDFunc: &StoreGetIndexConfigurbtionByRepositoryIDFunc{
			defbultHook: i.GetIndexConfigurbtionByRepositoryID,
		},
		GetInferenceScriptFunc: &StoreGetInferenceScriptFunc{
			defbultHook: i.GetInferenceScript,
		},
		GetLbstIndexScbnForRepositoryFunc: &StoreGetLbstIndexScbnForRepositoryFunc{
			defbultHook: i.GetLbstIndexScbnForRepository,
		},
		GetQueuedRepoRevFunc: &StoreGetQueuedRepoRevFunc{
			defbultHook: i.GetQueuedRepoRev,
		},
		GetRepositoriesForIndexScbnFunc: &StoreGetRepositoriesForIndexScbnFunc{
			defbultHook: i.GetRepositoriesForIndexScbn,
		},
		InsertDependencyIndexingJobFunc: &StoreInsertDependencyIndexingJobFunc{
			defbultHook: i.InsertDependencyIndexingJob,
		},
		InsertIndexesFunc: &StoreInsertIndexesFunc{
			defbultHook: i.InsertIndexes,
		},
		IsQueuedFunc: &StoreIsQueuedFunc{
			defbultHook: i.IsQueued,
		},
		IsQueuedRootIndexerFunc: &StoreIsQueuedRootIndexerFunc{
			defbultHook: i.IsQueuedRootIndexer,
		},
		MbrkRepoRevsAsProcessedFunc: &StoreMbrkRepoRevsAsProcessedFunc{
			defbultHook: i.MbrkRepoRevsAsProcessed,
		},
		QueueRepoRevFunc: &StoreQueueRepoRevFunc{
			defbultHook: i.QueueRepoRev,
		},
		RepositoryExceptionsFunc: &StoreRepositoryExceptionsFunc{
			defbultHook: i.RepositoryExceptions,
		},
		RepositoryIDsWithConfigurbtionFunc: &StoreRepositoryIDsWithConfigurbtionFunc{
			defbultHook: i.RepositoryIDsWithConfigurbtion,
		},
		SetConfigurbtionSummbryFunc: &StoreSetConfigurbtionSummbryFunc{
			defbultHook: i.SetConfigurbtionSummbry,
		},
		SetInferenceScriptFunc: &StoreSetInferenceScriptFunc{
			defbultHook: i.SetInferenceScript,
		},
		SetRepositoryExceptionsFunc: &StoreSetRepositoryExceptionsFunc{
			defbultHook: i.SetRepositoryExceptions,
		},
		TopRepositoriesToConfigureFunc: &StoreTopRepositoriesToConfigureFunc{
			defbultHook: i.TopRepositoriesToConfigure,
		},
		TruncbteConfigurbtionSummbryFunc: &StoreTruncbteConfigurbtionSummbryFunc{
			defbultHook: i.TruncbteConfigurbtionSummbry,
		},
		UpdbteIndexConfigurbtionByRepositoryIDFunc: &StoreUpdbteIndexConfigurbtionByRepositoryIDFunc{
			defbultHook: i.UpdbteIndexConfigurbtionByRepositoryID,
		},
		WithTrbnsbctionFunc: &StoreWithTrbnsbctionFunc{
			defbultHook: i.WithTrbnsbction,
		},
	}
}

// StoreGetIndexConfigurbtionByRepositoryIDFunc describes the behbvior when
// the GetIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce is invoked.
type StoreGetIndexConfigurbtionByRepositoryIDFunc struct {
	defbultHook func(context.Context, int) (shbred.IndexConfigurbtion, bool, error)
	hooks       []func(context.Context, int) (shbred.IndexConfigurbtion, bool, error)
	history     []StoreGetIndexConfigurbtionByRepositoryIDFuncCbll
	mutex       sync.Mutex
}

// GetIndexConfigurbtionByRepositoryID delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockStore) GetIndexConfigurbtionByRepositoryID(v0 context.Context, v1 int) (shbred.IndexConfigurbtion, bool, error) {
	r0, r1, r2 := m.GetIndexConfigurbtionByRepositoryIDFunc.nextHook()(v0, v1)
	m.GetIndexConfigurbtionByRepositoryIDFunc.bppendCbll(StoreGetIndexConfigurbtionByRepositoryIDFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce is invoked bnd the hook queue is empty.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) SetDefbultHook(hook func(context.Context, int) (shbred.IndexConfigurbtion, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) PushHook(hook func(context.Context, int) (shbred.IndexConfigurbtion, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) SetDefbultReturn(r0 shbred.IndexConfigurbtion, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred.IndexConfigurbtion, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) PushReturn(r0 shbred.IndexConfigurbtion, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (shbred.IndexConfigurbtion, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) nextHook() func(context.Context, int) (shbred.IndexConfigurbtion, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) bppendCbll(r0 StoreGetIndexConfigurbtionByRepositoryIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// StoreGetIndexConfigurbtionByRepositoryIDFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) History() []StoreGetIndexConfigurbtionByRepositoryIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetIndexConfigurbtionByRepositoryIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetIndexConfigurbtionByRepositoryIDFuncCbll is bn object thbt
// describes bn invocbtion of method GetIndexConfigurbtionByRepositoryID on
// bn instbnce of MockStore.
type StoreGetIndexConfigurbtionByRepositoryIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.IndexConfigurbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetIndexConfigurbtionByRepositoryIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetIndexConfigurbtionByRepositoryIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreGetInferenceScriptFunc describes the behbvior when the
// GetInferenceScript method of the pbrent MockStore instbnce is invoked.
type StoreGetInferenceScriptFunc struct {
	defbultHook func(context.Context) (string, error)
	hooks       []func(context.Context) (string, error)
	history     []StoreGetInferenceScriptFuncCbll
	mutex       sync.Mutex
}

// GetInferenceScript delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetInferenceScript(v0 context.Context) (string, error) {
	r0, r1 := m.GetInferenceScriptFunc.nextHook()(v0)
	m.GetInferenceScriptFunc.bppendCbll(StoreGetInferenceScriptFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetInferenceScript
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreGetInferenceScriptFunc) SetDefbultHook(hook func(context.Context) (string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetInferenceScript method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreGetInferenceScriptFunc) PushHook(hook func(context.Context) (string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetInferenceScriptFunc) SetDefbultReturn(r0 string, r1 error) {
	f.SetDefbultHook(func(context.Context) (string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetInferenceScriptFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(context.Context) (string, error) {
		return r0, r1
	})
}

func (f *StoreGetInferenceScriptFunc) nextHook() func(context.Context) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetInferenceScriptFunc) bppendCbll(r0 StoreGetInferenceScriptFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetInferenceScriptFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetInferenceScriptFunc) History() []StoreGetInferenceScriptFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetInferenceScriptFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetInferenceScriptFuncCbll is bn object thbt describes bn invocbtion
// of method GetInferenceScript on bn instbnce of MockStore.
type StoreGetInferenceScriptFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetInferenceScriptFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetInferenceScriptFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetLbstIndexScbnForRepositoryFunc describes the behbvior when the
// GetLbstIndexScbnForRepository method of the pbrent MockStore instbnce is
// invoked.
type StoreGetLbstIndexScbnForRepositoryFunc struct {
	defbultHook func(context.Context, int) (*time.Time, error)
	hooks       []func(context.Context, int) (*time.Time, error)
	history     []StoreGetLbstIndexScbnForRepositoryFuncCbll
	mutex       sync.Mutex
}

// GetLbstIndexScbnForRepository delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetLbstIndexScbnForRepository(v0 context.Context, v1 int) (*time.Time, error) {
	r0, r1 := m.GetLbstIndexScbnForRepositoryFunc.nextHook()(v0, v1)
	m.GetLbstIndexScbnForRepositoryFunc.bppendCbll(StoreGetLbstIndexScbnForRepositoryFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetLbstIndexScbnForRepository method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreGetLbstIndexScbnForRepositoryFunc) SetDefbultHook(hook func(context.Context, int) (*time.Time, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetLbstIndexScbnForRepository method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreGetLbstIndexScbnForRepositoryFunc) PushHook(hook func(context.Context, int) (*time.Time, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetLbstIndexScbnForRepositoryFunc) SetDefbultReturn(r0 *time.Time, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (*time.Time, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetLbstIndexScbnForRepositoryFunc) PushReturn(r0 *time.Time, r1 error) {
	f.PushHook(func(context.Context, int) (*time.Time, error) {
		return r0, r1
	})
}

func (f *StoreGetLbstIndexScbnForRepositoryFunc) nextHook() func(context.Context, int) (*time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetLbstIndexScbnForRepositoryFunc) bppendCbll(r0 StoreGetLbstIndexScbnForRepositoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetLbstIndexScbnForRepositoryFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreGetLbstIndexScbnForRepositoryFunc) History() []StoreGetLbstIndexScbnForRepositoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetLbstIndexScbnForRepositoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetLbstIndexScbnForRepositoryFuncCbll is bn object thbt describes bn
// invocbtion of method GetLbstIndexScbnForRepository on bn instbnce of
// MockStore.
type StoreGetLbstIndexScbnForRepositoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *time.Time
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetLbstIndexScbnForRepositoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetLbstIndexScbnForRepositoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetQueuedRepoRevFunc describes the behbvior when the
// GetQueuedRepoRev method of the pbrent MockStore instbnce is invoked.
type StoreGetQueuedRepoRevFunc struct {
	defbultHook func(context.Context, int) ([]store.RepoRev, error)
	hooks       []func(context.Context, int) ([]store.RepoRev, error)
	history     []StoreGetQueuedRepoRevFuncCbll
	mutex       sync.Mutex
}

// GetQueuedRepoRev delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetQueuedRepoRev(v0 context.Context, v1 int) ([]store.RepoRev, error) {
	r0, r1 := m.GetQueuedRepoRevFunc.nextHook()(v0, v1)
	m.GetQueuedRepoRevFunc.bppendCbll(StoreGetQueuedRepoRevFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetQueuedRepoRev
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreGetQueuedRepoRevFunc) SetDefbultHook(hook func(context.Context, int) ([]store.RepoRev, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetQueuedRepoRev method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreGetQueuedRepoRevFunc) PushHook(hook func(context.Context, int) ([]store.RepoRev, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetQueuedRepoRevFunc) SetDefbultReturn(r0 []store.RepoRev, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]store.RepoRev, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetQueuedRepoRevFunc) PushReturn(r0 []store.RepoRev, r1 error) {
	f.PushHook(func(context.Context, int) ([]store.RepoRev, error) {
		return r0, r1
	})
}

func (f *StoreGetQueuedRepoRevFunc) nextHook() func(context.Context, int) ([]store.RepoRev, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetQueuedRepoRevFunc) bppendCbll(r0 StoreGetQueuedRepoRevFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetQueuedRepoRevFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetQueuedRepoRevFunc) History() []StoreGetQueuedRepoRevFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetQueuedRepoRevFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetQueuedRepoRevFuncCbll is bn object thbt describes bn invocbtion
// of method GetQueuedRepoRev on bn instbnce of MockStore.
type StoreGetQueuedRepoRevFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []store.RepoRev
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetQueuedRepoRevFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetQueuedRepoRevFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetRepositoriesForIndexScbnFunc describes the behbvior when the
// GetRepositoriesForIndexScbn method of the pbrent MockStore instbnce is
// invoked.
type StoreGetRepositoriesForIndexScbnFunc struct {
	defbultHook func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error)
	hooks       []func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error)
	history     []StoreGetRepositoriesForIndexScbnFuncCbll
	mutex       sync.Mutex
}

// GetRepositoriesForIndexScbn delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetRepositoriesForIndexScbn(v0 context.Context, v1 time.Durbtion, v2 bool, v3 *int, v4 int, v5 time.Time) ([]int, error) {
	r0, r1 := m.GetRepositoriesForIndexScbnFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.GetRepositoriesForIndexScbnFunc.bppendCbll(StoreGetRepositoriesForIndexScbnFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetRepositoriesForIndexScbn method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreGetRepositoriesForIndexScbnFunc) SetDefbultHook(hook func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepositoriesForIndexScbn method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreGetRepositoriesForIndexScbnFunc) PushHook(hook func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetRepositoriesForIndexScbnFunc) SetDefbultReturn(r0 []int, r1 error) {
	f.SetDefbultHook(func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetRepositoriesForIndexScbnFunc) PushReturn(r0 []int, r1 error) {
	f.PushHook(func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error) {
		return r0, r1
	})
}

func (f *StoreGetRepositoriesForIndexScbnFunc) nextHook() func(context.Context, time.Durbtion, bool, *int, int, time.Time) ([]int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetRepositoriesForIndexScbnFunc) bppendCbll(r0 StoreGetRepositoriesForIndexScbnFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetRepositoriesForIndexScbnFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreGetRepositoriesForIndexScbnFunc) History() []StoreGetRepositoriesForIndexScbnFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetRepositoriesForIndexScbnFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetRepositoriesForIndexScbnFuncCbll is bn object thbt describes bn
// invocbtion of method GetRepositoriesForIndexScbn on bn instbnce of
// MockStore.
type StoreGetRepositoriesForIndexScbnFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 time.Durbtion
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetRepositoriesForIndexScbnFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetRepositoriesForIndexScbnFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreInsertDependencyIndexingJobFunc describes the behbvior when the
// InsertDependencyIndexingJob method of the pbrent MockStore instbnce is
// invoked.
type StoreInsertDependencyIndexingJobFunc struct {
	defbultHook func(context.Context, int, string, time.Time) (int, error)
	hooks       []func(context.Context, int, string, time.Time) (int, error)
	history     []StoreInsertDependencyIndexingJobFuncCbll
	mutex       sync.Mutex
}

// InsertDependencyIndexingJob delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertDependencyIndexingJob(v0 context.Context, v1 int, v2 string, v3 time.Time) (int, error) {
	r0, r1 := m.InsertDependencyIndexingJobFunc.nextHook()(v0, v1, v2, v3)
	m.InsertDependencyIndexingJobFunc.bppendCbll(StoreInsertDependencyIndexingJobFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// InsertDependencyIndexingJob method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreInsertDependencyIndexingJobFunc) SetDefbultHook(hook func(context.Context, int, string, time.Time) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertDependencyIndexingJob method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreInsertDependencyIndexingJobFunc) PushHook(hook func(context.Context, int, string, time.Time) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertDependencyIndexingJobFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, time.Time) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertDependencyIndexingJobFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, int, string, time.Time) (int, error) {
		return r0, r1
	})
}

func (f *StoreInsertDependencyIndexingJobFunc) nextHook() func(context.Context, int, string, time.Time) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertDependencyIndexingJobFunc) bppendCbll(r0 StoreInsertDependencyIndexingJobFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertDependencyIndexingJobFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreInsertDependencyIndexingJobFunc) History() []StoreInsertDependencyIndexingJobFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertDependencyIndexingJobFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertDependencyIndexingJobFuncCbll is bn object thbt describes bn
// invocbtion of method InsertDependencyIndexingJob on bn instbnce of
// MockStore.
type StoreInsertDependencyIndexingJobFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertDependencyIndexingJobFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertDependencyIndexingJobFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreInsertIndexesFunc describes the behbvior when the InsertIndexes
// method of the pbrent MockStore instbnce is invoked.
type StoreInsertIndexesFunc struct {
	defbultHook func(context.Context, []shbred1.Index) ([]shbred1.Index, error)
	hooks       []func(context.Context, []shbred1.Index) ([]shbred1.Index, error)
	history     []StoreInsertIndexesFuncCbll
	mutex       sync.Mutex
}

// InsertIndexes delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertIndexes(v0 context.Context, v1 []shbred1.Index) ([]shbred1.Index, error) {
	r0, r1 := m.InsertIndexesFunc.nextHook()(v0, v1)
	m.InsertIndexesFunc.bppendCbll(StoreInsertIndexesFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the InsertIndexes method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreInsertIndexesFunc) SetDefbultHook(hook func(context.Context, []shbred1.Index) ([]shbred1.Index, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertIndexes method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreInsertIndexesFunc) PushHook(hook func(context.Context, []shbred1.Index) ([]shbred1.Index, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertIndexesFunc) SetDefbultReturn(r0 []shbred1.Index, r1 error) {
	f.SetDefbultHook(func(context.Context, []shbred1.Index) ([]shbred1.Index, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertIndexesFunc) PushReturn(r0 []shbred1.Index, r1 error) {
	f.PushHook(func(context.Context, []shbred1.Index) ([]shbred1.Index, error) {
		return r0, r1
	})
}

func (f *StoreInsertIndexesFunc) nextHook() func(context.Context, []shbred1.Index) ([]shbred1.Index, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertIndexesFunc) bppendCbll(r0 StoreInsertIndexesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertIndexesFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreInsertIndexesFunc) History() []StoreInsertIndexesFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertIndexesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertIndexesFuncCbll is bn object thbt describes bn invocbtion of
// method InsertIndexes on bn instbnce of MockStore.
type StoreInsertIndexesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []shbred1.Index
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.Index
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertIndexesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertIndexesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreIsQueuedFunc describes the behbvior when the IsQueued method of the
// pbrent MockStore instbnce is invoked.
type StoreIsQueuedFunc struct {
	defbultHook func(context.Context, int, string) (bool, error)
	hooks       []func(context.Context, int, string) (bool, error)
	history     []StoreIsQueuedFuncCbll
	mutex       sync.Mutex
}

// IsQueued delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) IsQueued(v0 context.Context, v1 int, v2 string) (bool, error) {
	r0, r1 := m.IsQueuedFunc.nextHook()(v0, v1, v2)
	m.IsQueuedFunc.bppendCbll(StoreIsQueuedFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the IsQueued method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreIsQueuedFunc) SetDefbultHook(hook func(context.Context, int, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IsQueued method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreIsQueuedFunc) PushHook(hook func(context.Context, int, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIsQueuedFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIsQueuedFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreIsQueuedFunc) nextHook() func(context.Context, int, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIsQueuedFunc) bppendCbll(r0 StoreIsQueuedFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIsQueuedFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreIsQueuedFunc) History() []StoreIsQueuedFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIsQueuedFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIsQueuedFuncCbll is bn object thbt describes bn invocbtion of method
// IsQueued on bn instbnce of MockStore.
type StoreIsQueuedFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIsQueuedFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIsQueuedFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreIsQueuedRootIndexerFunc describes the behbvior when the
// IsQueuedRootIndexer method of the pbrent MockStore instbnce is invoked.
type StoreIsQueuedRootIndexerFunc struct {
	defbultHook func(context.Context, int, string, string, string) (bool, error)
	hooks       []func(context.Context, int, string, string, string) (bool, error)
	history     []StoreIsQueuedRootIndexerFuncCbll
	mutex       sync.Mutex
}

// IsQueuedRootIndexer delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) IsQueuedRootIndexer(v0 context.Context, v1 int, v2 string, v3 string, v4 string) (bool, error) {
	r0, r1 := m.IsQueuedRootIndexerFunc.nextHook()(v0, v1, v2, v3, v4)
	m.IsQueuedRootIndexerFunc.bppendCbll(StoreIsQueuedRootIndexerFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the IsQueuedRootIndexer
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreIsQueuedRootIndexerFunc) SetDefbultHook(hook func(context.Context, int, string, string, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IsQueuedRootIndexer method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreIsQueuedRootIndexerFunc) PushHook(hook func(context.Context, int, string, string, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIsQueuedRootIndexerFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, string, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIsQueuedRootIndexerFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string, string, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreIsQueuedRootIndexerFunc) nextHook() func(context.Context, int, string, string, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIsQueuedRootIndexerFunc) bppendCbll(r0 StoreIsQueuedRootIndexerFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIsQueuedRootIndexerFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreIsQueuedRootIndexerFunc) History() []StoreIsQueuedRootIndexerFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIsQueuedRootIndexerFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIsQueuedRootIndexerFuncCbll is bn object thbt describes bn
// invocbtion of method IsQueuedRootIndexer on bn instbnce of MockStore.
type StoreIsQueuedRootIndexerFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIsQueuedRootIndexerFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIsQueuedRootIndexerFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreMbrkRepoRevsAsProcessedFunc describes the behbvior when the
// MbrkRepoRevsAsProcessed method of the pbrent MockStore instbnce is
// invoked.
type StoreMbrkRepoRevsAsProcessedFunc struct {
	defbultHook func(context.Context, []int) error
	hooks       []func(context.Context, []int) error
	history     []StoreMbrkRepoRevsAsProcessedFuncCbll
	mutex       sync.Mutex
}

// MbrkRepoRevsAsProcessed delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) MbrkRepoRevsAsProcessed(v0 context.Context, v1 []int) error {
	r0 := m.MbrkRepoRevsAsProcessedFunc.nextHook()(v0, v1)
	m.MbrkRepoRevsAsProcessedFunc.bppendCbll(StoreMbrkRepoRevsAsProcessedFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// MbrkRepoRevsAsProcessed method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreMbrkRepoRevsAsProcessedFunc) SetDefbultHook(hook func(context.Context, []int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkRepoRevsAsProcessed method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreMbrkRepoRevsAsProcessedFunc) PushHook(hook func(context.Context, []int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkRepoRevsAsProcessedFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkRepoRevsAsProcessedFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []int) error {
		return r0
	})
}

func (f *StoreMbrkRepoRevsAsProcessedFunc) nextHook() func(context.Context, []int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMbrkRepoRevsAsProcessedFunc) bppendCbll(r0 StoreMbrkRepoRevsAsProcessedFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreMbrkRepoRevsAsProcessedFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreMbrkRepoRevsAsProcessedFunc) History() []StoreMbrkRepoRevsAsProcessedFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreMbrkRepoRevsAsProcessedFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMbrkRepoRevsAsProcessedFuncCbll is bn object thbt describes bn
// invocbtion of method MbrkRepoRevsAsProcessed on bn instbnce of MockStore.
type StoreMbrkRepoRevsAsProcessedFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreMbrkRepoRevsAsProcessedFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkRepoRevsAsProcessedFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreQueueRepoRevFunc describes the behbvior when the QueueRepoRev method
// of the pbrent MockStore instbnce is invoked.
type StoreQueueRepoRevFunc struct {
	defbultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []StoreQueueRepoRevFuncCbll
	mutex       sync.Mutex
}

// QueueRepoRev delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) QueueRepoRev(v0 context.Context, v1 int, v2 string) error {
	r0 := m.QueueRepoRevFunc.nextHook()(v0, v1, v2)
	m.QueueRepoRevFunc.bppendCbll(StoreQueueRepoRevFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the QueueRepoRev method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreQueueRepoRevFunc) SetDefbultHook(hook func(context.Context, int, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueueRepoRev method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreQueueRepoRevFunc) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreQueueRepoRevFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreQueueRepoRevFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *StoreQueueRepoRevFunc) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreQueueRepoRevFunc) bppendCbll(r0 StoreQueueRepoRevFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreQueueRepoRevFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreQueueRepoRevFunc) History() []StoreQueueRepoRevFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreQueueRepoRevFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreQueueRepoRevFuncCbll is bn object thbt describes bn invocbtion of
// method QueueRepoRev on bn instbnce of MockStore.
type StoreQueueRepoRevFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreQueueRepoRevFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreQueueRepoRevFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreRepositoryExceptionsFunc describes the behbvior when the
// RepositoryExceptions method of the pbrent MockStore instbnce is invoked.
type StoreRepositoryExceptionsFunc struct {
	defbultHook func(context.Context, int) (bool, bool, error)
	hooks       []func(context.Context, int) (bool, bool, error)
	history     []StoreRepositoryExceptionsFuncCbll
	mutex       sync.Mutex
}

// RepositoryExceptions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) RepositoryExceptions(v0 context.Context, v1 int) (bool, bool, error) {
	r0, r1, r2 := m.RepositoryExceptionsFunc.nextHook()(v0, v1)
	m.RepositoryExceptionsFunc.bppendCbll(StoreRepositoryExceptionsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the RepositoryExceptions
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreRepositoryExceptionsFunc) SetDefbultHook(hook func(context.Context, int) (bool, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepositoryExceptions method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreRepositoryExceptionsFunc) PushHook(hook func(context.Context, int) (bool, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreRepositoryExceptionsFunc) SetDefbultReturn(r0 bool, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (bool, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreRepositoryExceptionsFunc) PushReturn(r0 bool, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (bool, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreRepositoryExceptionsFunc) nextHook() func(context.Context, int) (bool, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreRepositoryExceptionsFunc) bppendCbll(r0 StoreRepositoryExceptionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreRepositoryExceptionsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreRepositoryExceptionsFunc) History() []StoreRepositoryExceptionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreRepositoryExceptionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreRepositoryExceptionsFuncCbll is bn object thbt describes bn
// invocbtion of method RepositoryExceptions on bn instbnce of MockStore.
type StoreRepositoryExceptionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreRepositoryExceptionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreRepositoryExceptionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreRepositoryIDsWithConfigurbtionFunc describes the behbvior when the
// RepositoryIDsWithConfigurbtion method of the pbrent MockStore instbnce is
// invoked.
type StoreRepositoryIDsWithConfigurbtionFunc struct {
	defbultHook func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error)
	hooks       []func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error)
	history     []StoreRepositoryIDsWithConfigurbtionFuncCbll
	mutex       sync.Mutex
}

// RepositoryIDsWithConfigurbtion delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) RepositoryIDsWithConfigurbtion(v0 context.Context, v1 int, v2 int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error) {
	r0, r1, r2 := m.RepositoryIDsWithConfigurbtionFunc.nextHook()(v0, v1, v2)
	m.RepositoryIDsWithConfigurbtionFunc.bppendCbll(StoreRepositoryIDsWithConfigurbtionFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// RepositoryIDsWithConfigurbtion method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreRepositoryIDsWithConfigurbtionFunc) SetDefbultHook(hook func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepositoryIDsWithConfigurbtion method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreRepositoryIDsWithConfigurbtionFunc) PushHook(hook func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreRepositoryIDsWithConfigurbtionFunc) SetDefbultReturn(r0 []shbred1.RepositoryWithAvbilbbleIndexers, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreRepositoryIDsWithConfigurbtionFunc) PushReturn(r0 []shbred1.RepositoryWithAvbilbbleIndexers, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreRepositoryIDsWithConfigurbtionFunc) nextHook() func(context.Context, int, int) ([]shbred1.RepositoryWithAvbilbbleIndexers, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreRepositoryIDsWithConfigurbtionFunc) bppendCbll(r0 StoreRepositoryIDsWithConfigurbtionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreRepositoryIDsWithConfigurbtionFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreRepositoryIDsWithConfigurbtionFunc) History() []StoreRepositoryIDsWithConfigurbtionFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreRepositoryIDsWithConfigurbtionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreRepositoryIDsWithConfigurbtionFuncCbll is bn object thbt describes
// bn invocbtion of method RepositoryIDsWithConfigurbtion on bn instbnce of
// MockStore.
type StoreRepositoryIDsWithConfigurbtionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.RepositoryWithAvbilbbleIndexers
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreRepositoryIDsWithConfigurbtionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreRepositoryIDsWithConfigurbtionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreSetConfigurbtionSummbryFunc describes the behbvior when the
// SetConfigurbtionSummbry method of the pbrent MockStore instbnce is
// invoked.
type StoreSetConfigurbtionSummbryFunc struct {
	defbultHook func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error
	hooks       []func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error
	history     []StoreSetConfigurbtionSummbryFuncCbll
	mutex       sync.Mutex
}

// SetConfigurbtionSummbry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) SetConfigurbtionSummbry(v0 context.Context, v1 int, v2 int, v3 mbp[string]shbred1.AvbilbbleIndexer) error {
	r0 := m.SetConfigurbtionSummbryFunc.nextHook()(v0, v1, v2, v3)
	m.SetConfigurbtionSummbryFunc.bppendCbll(StoreSetConfigurbtionSummbryFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// SetConfigurbtionSummbry method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreSetConfigurbtionSummbryFunc) SetDefbultHook(hook func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetConfigurbtionSummbry method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreSetConfigurbtionSummbryFunc) PushHook(hook func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSetConfigurbtionSummbryFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSetConfigurbtionSummbryFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error {
		return r0
	})
}

func (f *StoreSetConfigurbtionSummbryFunc) nextHook() func(context.Context, int, int, mbp[string]shbred1.AvbilbbleIndexer) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSetConfigurbtionSummbryFunc) bppendCbll(r0 StoreSetConfigurbtionSummbryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreSetConfigurbtionSummbryFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreSetConfigurbtionSummbryFunc) History() []StoreSetConfigurbtionSummbryFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSetConfigurbtionSummbryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSetConfigurbtionSummbryFuncCbll is bn object thbt describes bn
// invocbtion of method SetConfigurbtionSummbry on bn instbnce of MockStore.
type StoreSetConfigurbtionSummbryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 mbp[string]shbred1.AvbilbbleIndexer
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSetConfigurbtionSummbryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSetConfigurbtionSummbryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreSetInferenceScriptFunc describes the behbvior when the
// SetInferenceScript method of the pbrent MockStore instbnce is invoked.
type StoreSetInferenceScriptFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []StoreSetInferenceScriptFuncCbll
	mutex       sync.Mutex
}

// SetInferenceScript delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) SetInferenceScript(v0 context.Context, v1 string) error {
	r0 := m.SetInferenceScriptFunc.nextHook()(v0, v1)
	m.SetInferenceScriptFunc.bppendCbll(StoreSetInferenceScriptFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetInferenceScript
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreSetInferenceScriptFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetInferenceScript method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreSetInferenceScriptFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSetInferenceScriptFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSetInferenceScriptFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *StoreSetInferenceScriptFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSetInferenceScriptFunc) bppendCbll(r0 StoreSetInferenceScriptFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreSetInferenceScriptFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreSetInferenceScriptFunc) History() []StoreSetInferenceScriptFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSetInferenceScriptFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSetInferenceScriptFuncCbll is bn object thbt describes bn invocbtion
// of method SetInferenceScript on bn instbnce of MockStore.
type StoreSetInferenceScriptFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSetInferenceScriptFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSetInferenceScriptFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreSetRepositoryExceptionsFunc describes the behbvior when the
// SetRepositoryExceptions method of the pbrent MockStore instbnce is
// invoked.
type StoreSetRepositoryExceptionsFunc struct {
	defbultHook func(context.Context, int, bool, bool) error
	hooks       []func(context.Context, int, bool, bool) error
	history     []StoreSetRepositoryExceptionsFuncCbll
	mutex       sync.Mutex
}

// SetRepositoryExceptions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) SetRepositoryExceptions(v0 context.Context, v1 int, v2 bool, v3 bool) error {
	r0 := m.SetRepositoryExceptionsFunc.nextHook()(v0, v1, v2, v3)
	m.SetRepositoryExceptionsFunc.bppendCbll(StoreSetRepositoryExceptionsFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// SetRepositoryExceptions method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreSetRepositoryExceptionsFunc) SetDefbultHook(hook func(context.Context, int, bool, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetRepositoryExceptions method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreSetRepositoryExceptionsFunc) PushHook(hook func(context.Context, int, bool, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSetRepositoryExceptionsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, bool, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSetRepositoryExceptionsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, bool, bool) error {
		return r0
	})
}

func (f *StoreSetRepositoryExceptionsFunc) nextHook() func(context.Context, int, bool, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSetRepositoryExceptionsFunc) bppendCbll(r0 StoreSetRepositoryExceptionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreSetRepositoryExceptionsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreSetRepositoryExceptionsFunc) History() []StoreSetRepositoryExceptionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSetRepositoryExceptionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSetRepositoryExceptionsFuncCbll is bn object thbt describes bn
// invocbtion of method SetRepositoryExceptions on bn instbnce of MockStore.
type StoreSetRepositoryExceptionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSetRepositoryExceptionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSetRepositoryExceptionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreTopRepositoriesToConfigureFunc describes the behbvior when the
// TopRepositoriesToConfigure method of the pbrent MockStore instbnce is
// invoked.
type StoreTopRepositoriesToConfigureFunc struct {
	defbultHook func(context.Context, int) ([]shbred1.RepositoryWithCount, error)
	hooks       []func(context.Context, int) ([]shbred1.RepositoryWithCount, error)
	history     []StoreTopRepositoriesToConfigureFuncCbll
	mutex       sync.Mutex
}

// TopRepositoriesToConfigure delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) TopRepositoriesToConfigure(v0 context.Context, v1 int) ([]shbred1.RepositoryWithCount, error) {
	r0, r1 := m.TopRepositoriesToConfigureFunc.nextHook()(v0, v1)
	m.TopRepositoriesToConfigureFunc.bppendCbll(StoreTopRepositoriesToConfigureFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// TopRepositoriesToConfigure method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreTopRepositoriesToConfigureFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred1.RepositoryWithCount, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// TopRepositoriesToConfigure method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreTopRepositoriesToConfigureFunc) PushHook(hook func(context.Context, int) ([]shbred1.RepositoryWithCount, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreTopRepositoriesToConfigureFunc) SetDefbultReturn(r0 []shbred1.RepositoryWithCount, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred1.RepositoryWithCount, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreTopRepositoriesToConfigureFunc) PushReturn(r0 []shbred1.RepositoryWithCount, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred1.RepositoryWithCount, error) {
		return r0, r1
	})
}

func (f *StoreTopRepositoriesToConfigureFunc) nextHook() func(context.Context, int) ([]shbred1.RepositoryWithCount, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTopRepositoriesToConfigureFunc) bppendCbll(r0 StoreTopRepositoriesToConfigureFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreTopRepositoriesToConfigureFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreTopRepositoriesToConfigureFunc) History() []StoreTopRepositoriesToConfigureFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreTopRepositoriesToConfigureFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTopRepositoriesToConfigureFuncCbll is bn object thbt describes bn
// invocbtion of method TopRepositoriesToConfigure on bn instbnce of
// MockStore.
type StoreTopRepositoriesToConfigureFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.RepositoryWithCount
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreTopRepositoriesToConfigureFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreTopRepositoriesToConfigureFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreTruncbteConfigurbtionSummbryFunc describes the behbvior when the
// TruncbteConfigurbtionSummbry method of the pbrent MockStore instbnce is
// invoked.
type StoreTruncbteConfigurbtionSummbryFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []StoreTruncbteConfigurbtionSummbryFuncCbll
	mutex       sync.Mutex
}

// TruncbteConfigurbtionSummbry delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) TruncbteConfigurbtionSummbry(v0 context.Context, v1 int) error {
	r0 := m.TruncbteConfigurbtionSummbryFunc.nextHook()(v0, v1)
	m.TruncbteConfigurbtionSummbryFunc.bppendCbll(StoreTruncbteConfigurbtionSummbryFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// TruncbteConfigurbtionSummbry method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreTruncbteConfigurbtionSummbryFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// TruncbteConfigurbtionSummbry method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreTruncbteConfigurbtionSummbryFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreTruncbteConfigurbtionSummbryFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreTruncbteConfigurbtionSummbryFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *StoreTruncbteConfigurbtionSummbryFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTruncbteConfigurbtionSummbryFunc) bppendCbll(r0 StoreTruncbteConfigurbtionSummbryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreTruncbteConfigurbtionSummbryFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreTruncbteConfigurbtionSummbryFunc) History() []StoreTruncbteConfigurbtionSummbryFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreTruncbteConfigurbtionSummbryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTruncbteConfigurbtionSummbryFuncCbll is bn object thbt describes bn
// invocbtion of method TruncbteConfigurbtionSummbry on bn instbnce of
// MockStore.
type StoreTruncbteConfigurbtionSummbryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreTruncbteConfigurbtionSummbryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreTruncbteConfigurbtionSummbryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreUpdbteIndexConfigurbtionByRepositoryIDFunc describes the behbvior
// when the UpdbteIndexConfigurbtionByRepositoryID method of the pbrent
// MockStore instbnce is invoked.
type StoreUpdbteIndexConfigurbtionByRepositoryIDFunc struct {
	defbultHook func(context.Context, int, []byte) error
	hooks       []func(context.Context, int, []byte) error
	history     []StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll
	mutex       sync.Mutex
}

// UpdbteIndexConfigurbtionByRepositoryID delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockStore) UpdbteIndexConfigurbtionByRepositoryID(v0 context.Context, v1 int, v2 []byte) error {
	r0 := m.UpdbteIndexConfigurbtionByRepositoryIDFunc.nextHook()(v0, v1, v2)
	m.UpdbteIndexConfigurbtionByRepositoryIDFunc.bppendCbll(StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce is invoked bnd the hook queue is empty.
func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) SetDefbultHook(hook func(context.Context, int, []byte) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) PushHook(hook func(context.Context, int, []byte) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, []byte) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, []byte) error {
		return r0
	})
}

func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) nextHook() func(context.Context, int, []byte) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) bppendCbll(r0 StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreUpdbteIndexConfigurbtionByRepositoryIDFunc) History() []StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll is bn object thbt
// describes bn invocbtion of method UpdbteIndexConfigurbtionByRepositoryID
// on bn instbnce of MockStore.
type StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []byte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpdbteIndexConfigurbtionByRepositoryIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreWithTrbnsbctionFunc describes the behbvior when the WithTrbnsbction
// method of the pbrent MockStore instbnce is invoked.
type StoreWithTrbnsbctionFunc struct {
	defbultHook func(context.Context, func(tx store.Store) error) error
	hooks       []func(context.Context, func(tx store.Store) error) error
	history     []StoreWithTrbnsbctionFuncCbll
	mutex       sync.Mutex
}

// WithTrbnsbction delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) WithTrbnsbction(v0 context.Context, v1 func(tx store.Store) error) error {
	r0 := m.WithTrbnsbctionFunc.nextHook()(v0, v1)
	m.WithTrbnsbctionFunc.bppendCbll(StoreWithTrbnsbctionFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithTrbnsbction
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreWithTrbnsbctionFunc) SetDefbultHook(hook func(context.Context, func(tx store.Store) error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithTrbnsbction method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreWithTrbnsbctionFunc) PushHook(hook func(context.Context, func(tx store.Store) error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreWithTrbnsbctionFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, func(tx store.Store) error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreWithTrbnsbctionFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, func(tx store.Store) error) error {
		return r0
	})
}

func (f *StoreWithTrbnsbctionFunc) nextHook() func(context.Context, func(tx store.Store) error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreWithTrbnsbctionFunc) bppendCbll(r0 StoreWithTrbnsbctionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreWithTrbnsbctionFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreWithTrbnsbctionFunc) History() []StoreWithTrbnsbctionFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreWithTrbnsbctionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreWithTrbnsbctionFuncCbll is bn object thbt describes bn invocbtion of
// method WithTrbnsbction on bn instbnce of MockStore.
type StoreWithTrbnsbctionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 func(tx store.Store) error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreWithTrbnsbctionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreWithTrbnsbctionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockInferenceService is b mock implementbtion of the InferenceService
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing) used
// for unit testing.
type MockInferenceService struct {
	// InferIndexJobsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InferIndexJobs.
	InferIndexJobsFunc *InferenceServiceInferIndexJobsFunc
}

// NewMockInferenceService crebtes b new mock of the InferenceService
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockInferenceService() *MockInferenceService {
	return &MockInferenceService{
		InferIndexJobsFunc: &InferenceServiceInferIndexJobsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, string) (r0 *shbred.InferenceResult, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockInferenceService crebtes b new mock of the InferenceService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockInferenceService() *MockInferenceService {
	return &MockInferenceService{
		InferIndexJobsFunc: &InferenceServiceInferIndexJobsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error) {
				pbnic("unexpected invocbtion of MockInferenceService.InferIndexJobs")
			},
		},
	}
}

// NewMockInferenceServiceFrom crebtes b new mock of the
// MockInferenceService interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockInferenceServiceFrom(i InferenceService) *MockInferenceService {
	return &MockInferenceService{
		InferIndexJobsFunc: &InferenceServiceInferIndexJobsFunc{
			defbultHook: i.InferIndexJobs,
		},
	}
}

// InferenceServiceInferIndexJobsFunc describes the behbvior when the
// InferIndexJobs method of the pbrent MockInferenceService instbnce is
// invoked.
type InferenceServiceInferIndexJobsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error)
	hooks       []func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error)
	history     []InferenceServiceInferIndexJobsFuncCbll
	mutex       sync.Mutex
}

// InferIndexJobs delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInferenceService) InferIndexJobs(v0 context.Context, v1 bpi.RepoNbme, v2 string, v3 string) (*shbred.InferenceResult, error) {
	r0, r1 := m.InferIndexJobsFunc.nextHook()(v0, v1, v2, v3)
	m.InferIndexJobsFunc.bppendCbll(InferenceServiceInferIndexJobsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the InferIndexJobs
// method of the pbrent MockInferenceService instbnce is invoked bnd the
// hook queue is empty.
func (f *InferenceServiceInferIndexJobsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InferIndexJobs method of the pbrent MockInferenceService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *InferenceServiceInferIndexJobsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InferenceServiceInferIndexJobsFunc) SetDefbultReturn(r0 *shbred.InferenceResult, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InferenceServiceInferIndexJobsFunc) PushReturn(r0 *shbred.InferenceResult, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error) {
		return r0, r1
	})
}

func (f *InferenceServiceInferIndexJobsFunc) nextHook() func(context.Context, bpi.RepoNbme, string, string) (*shbred.InferenceResult, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InferenceServiceInferIndexJobsFunc) bppendCbll(r0 InferenceServiceInferIndexJobsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InferenceServiceInferIndexJobsFuncCbll
// objects describing the invocbtions of this function.
func (f *InferenceServiceInferIndexJobsFunc) History() []InferenceServiceInferIndexJobsFuncCbll {
	f.mutex.Lock()
	history := mbke([]InferenceServiceInferIndexJobsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InferenceServiceInferIndexJobsFuncCbll is bn object thbt describes bn
// invocbtion of method InferIndexJobs on bn instbnce of
// MockInferenceService.
type InferenceServiceInferIndexJobsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *shbred.InferenceResult
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InferenceServiceInferIndexJobsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InferenceServiceInferIndexJobsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockRepoUpdbterClient is b mock implementbtion of the RepoUpdbterClient
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing) used
// for unit testing.
type MockRepoUpdbterClient struct {
	// EnqueueRepoUpdbteFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method EnqueueRepoUpdbte.
	EnqueueRepoUpdbteFunc *RepoUpdbterClientEnqueueRepoUpdbteFunc
	// RepoLookupFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method RepoLookup.
	RepoLookupFunc *RepoUpdbterClientRepoLookupFunc
}

// NewMockRepoUpdbterClient crebtes b new mock of the RepoUpdbterClient
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockRepoUpdbterClient() *MockRepoUpdbterClient {
	return &MockRepoUpdbterClient{
		EnqueueRepoUpdbteFunc: &RepoUpdbterClientEnqueueRepoUpdbteFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 *protocol.RepoUpdbteResponse, r1 error) {
				return
			},
		},
		RepoLookupFunc: &RepoUpdbterClientRepoLookupFunc{
			defbultHook: func(context.Context, protocol.RepoLookupArgs) (r0 *protocol.RepoLookupResult, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockRepoUpdbterClient crebtes b new mock of the
// RepoUpdbterClient interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockRepoUpdbterClient() *MockRepoUpdbterClient {
	return &MockRepoUpdbterClient{
		EnqueueRepoUpdbteFunc: &RepoUpdbterClientEnqueueRepoUpdbteFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
				pbnic("unexpected invocbtion of MockRepoUpdbterClient.EnqueueRepoUpdbte")
			},
		},
		RepoLookupFunc: &RepoUpdbterClientRepoLookupFunc{
			defbultHook: func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
				pbnic("unexpected invocbtion of MockRepoUpdbterClient.RepoLookup")
			},
		},
	}
}

// NewMockRepoUpdbterClientFrom crebtes b new mock of the
// MockRepoUpdbterClient interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockRepoUpdbterClientFrom(i RepoUpdbterClient) *MockRepoUpdbterClient {
	return &MockRepoUpdbterClient{
		EnqueueRepoUpdbteFunc: &RepoUpdbterClientEnqueueRepoUpdbteFunc{
			defbultHook: i.EnqueueRepoUpdbte,
		},
		RepoLookupFunc: &RepoUpdbterClientRepoLookupFunc{
			defbultHook: i.RepoLookup,
		},
	}
}

// RepoUpdbterClientEnqueueRepoUpdbteFunc describes the behbvior when the
// EnqueueRepoUpdbte method of the pbrent MockRepoUpdbterClient instbnce is
// invoked.
type RepoUpdbterClientEnqueueRepoUpdbteFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error)
	hooks       []func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error)
	history     []RepoUpdbterClientEnqueueRepoUpdbteFuncCbll
	mutex       sync.Mutex
}

// EnqueueRepoUpdbte delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoUpdbterClient) EnqueueRepoUpdbte(v0 context.Context, v1 bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
	r0, r1 := m.EnqueueRepoUpdbteFunc.nextHook()(v0, v1)
	m.EnqueueRepoUpdbteFunc.bppendCbll(RepoUpdbterClientEnqueueRepoUpdbteFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the EnqueueRepoUpdbte
// method of the pbrent MockRepoUpdbterClient instbnce is invoked bnd the
// hook queue is empty.
func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// EnqueueRepoUpdbte method of the pbrent MockRepoUpdbterClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) SetDefbultReturn(r0 *protocol.RepoUpdbteResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) PushReturn(r0 *protocol.RepoUpdbteResponse, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		return r0, r1
	})
}

func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) nextHook() func(context.Context, bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) bppendCbll(r0 RepoUpdbterClientEnqueueRepoUpdbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoUpdbterClientEnqueueRepoUpdbteFuncCbll
// objects describing the invocbtions of this function.
func (f *RepoUpdbterClientEnqueueRepoUpdbteFunc) History() []RepoUpdbterClientEnqueueRepoUpdbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoUpdbterClientEnqueueRepoUpdbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoUpdbterClientEnqueueRepoUpdbteFuncCbll is bn object thbt describes bn
// invocbtion of method EnqueueRepoUpdbte on bn instbnce of
// MockRepoUpdbterClient.
type RepoUpdbterClientEnqueueRepoUpdbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoUpdbteResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoUpdbterClientEnqueueRepoUpdbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoUpdbterClientEnqueueRepoUpdbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoUpdbterClientRepoLookupFunc describes the behbvior when the
// RepoLookup method of the pbrent MockRepoUpdbterClient instbnce is
// invoked.
type RepoUpdbterClientRepoLookupFunc struct {
	defbultHook func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
	hooks       []func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
	history     []RepoUpdbterClientRepoLookupFuncCbll
	mutex       sync.Mutex
}

// RepoLookup delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoUpdbterClient) RepoLookup(v0 context.Context, v1 protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
	r0, r1 := m.RepoLookupFunc.nextHook()(v0, v1)
	m.RepoLookupFunc.bppendCbll(RepoUpdbterClientRepoLookupFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RepoLookup method of
// the pbrent MockRepoUpdbterClient instbnce is invoked bnd the hook queue
// is empty.
func (f *RepoUpdbterClientRepoLookupFunc) SetDefbultHook(hook func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoLookup method of the pbrent MockRepoUpdbterClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *RepoUpdbterClientRepoLookupFunc) PushHook(hook func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoUpdbterClientRepoLookupFunc) SetDefbultReturn(r0 *protocol.RepoLookupResult, r1 error) {
	f.SetDefbultHook(func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoUpdbterClientRepoLookupFunc) PushReturn(r0 *protocol.RepoLookupResult, r1 error) {
	f.PushHook(func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		return r0, r1
	})
}

func (f *RepoUpdbterClientRepoLookupFunc) nextHook() func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoUpdbterClientRepoLookupFunc) bppendCbll(r0 RepoUpdbterClientRepoLookupFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoUpdbterClientRepoLookupFuncCbll objects
// describing the invocbtions of this function.
func (f *RepoUpdbterClientRepoLookupFunc) History() []RepoUpdbterClientRepoLookupFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoUpdbterClientRepoLookupFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoUpdbterClientRepoLookupFuncCbll is bn object thbt describes bn
// invocbtion of method RepoLookup on bn instbnce of MockRepoUpdbterClient.
type RepoUpdbterClientRepoLookupFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 protocol.RepoLookupArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoLookupResult
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoUpdbterClientRepoLookupFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoUpdbterClientRepoLookupFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockUplobdService is b mock implementbtion of the UplobdService interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing) used
// for unit testing.
type MockUplobdService struct {
	// GetRecentIndexesSummbryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRecentIndexesSummbry.
	GetRecentIndexesSummbryFunc *UplobdServiceGetRecentIndexesSummbryFunc
	// GetRecentUplobdsSummbryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRecentUplobdsSummbry.
	GetRecentUplobdsSummbryFunc *UplobdServiceGetRecentUplobdsSummbryFunc
	// GetUplobdByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetUplobdByID.
	GetUplobdByIDFunc *UplobdServiceGetUplobdByIDFunc
	// ReferencesForUplobdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReferencesForUplobd.
	ReferencesForUplobdFunc *UplobdServiceReferencesForUplobdFunc
}

// NewMockUplobdService crebtes b new mock of the UplobdService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockUplobdService() *MockUplobdService {
	return &MockUplobdService{
		GetRecentIndexesSummbryFunc: &UplobdServiceGetRecentIndexesSummbryFunc{
			defbultHook: func(context.Context, int) (r0 []shbred1.IndexesWithRepositoryNbmespbce, r1 error) {
				return
			},
		},
		GetRecentUplobdsSummbryFunc: &UplobdServiceGetRecentUplobdsSummbryFunc{
			defbultHook: func(context.Context, int) (r0 []shbred1.UplobdsWithRepositoryNbmespbce, r1 error) {
				return
			},
		},
		GetUplobdByIDFunc: &UplobdServiceGetUplobdByIDFunc{
			defbultHook: func(context.Context, int) (r0 shbred1.Uplobd, r1 bool, r2 error) {
				return
			},
		},
		ReferencesForUplobdFunc: &UplobdServiceReferencesForUplobdFunc{
			defbultHook: func(context.Context, int) (r0 shbred1.PbckbgeReferenceScbnner, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockUplobdService crebtes b new mock of the UplobdService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockUplobdService() *MockUplobdService {
	return &MockUplobdService{
		GetRecentIndexesSummbryFunc: &UplobdServiceGetRecentIndexesSummbryFunc{
			defbultHook: func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetRecentIndexesSummbry")
			},
		},
		GetRecentUplobdsSummbryFunc: &UplobdServiceGetRecentUplobdsSummbryFunc{
			defbultHook: func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetRecentUplobdsSummbry")
			},
		},
		GetUplobdByIDFunc: &UplobdServiceGetUplobdByIDFunc{
			defbultHook: func(context.Context, int) (shbred1.Uplobd, bool, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetUplobdByID")
			},
		},
		ReferencesForUplobdFunc: &UplobdServiceReferencesForUplobdFunc{
			defbultHook: func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error) {
				pbnic("unexpected invocbtion of MockUplobdService.ReferencesForUplobd")
			},
		},
	}
}

// NewMockUplobdServiceFrom crebtes b new mock of the MockUplobdService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockUplobdServiceFrom(i UplobdService) *MockUplobdService {
	return &MockUplobdService{
		GetRecentIndexesSummbryFunc: &UplobdServiceGetRecentIndexesSummbryFunc{
			defbultHook: i.GetRecentIndexesSummbry,
		},
		GetRecentUplobdsSummbryFunc: &UplobdServiceGetRecentUplobdsSummbryFunc{
			defbultHook: i.GetRecentUplobdsSummbry,
		},
		GetUplobdByIDFunc: &UplobdServiceGetUplobdByIDFunc{
			defbultHook: i.GetUplobdByID,
		},
		ReferencesForUplobdFunc: &UplobdServiceReferencesForUplobdFunc{
			defbultHook: i.ReferencesForUplobd,
		},
	}
}

// UplobdServiceGetRecentIndexesSummbryFunc describes the behbvior when the
// GetRecentIndexesSummbry method of the pbrent MockUplobdService instbnce
// is invoked.
type UplobdServiceGetRecentIndexesSummbryFunc struct {
	defbultHook func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error)
	hooks       []func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error)
	history     []UplobdServiceGetRecentIndexesSummbryFuncCbll
	mutex       sync.Mutex
}

// GetRecentIndexesSummbry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetRecentIndexesSummbry(v0 context.Context, v1 int) ([]shbred1.IndexesWithRepositoryNbmespbce, error) {
	r0, r1 := m.GetRecentIndexesSummbryFunc.nextHook()(v0, v1)
	m.GetRecentIndexesSummbryFunc.bppendCbll(UplobdServiceGetRecentIndexesSummbryFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetRecentIndexesSummbry method of the pbrent MockUplobdService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdServiceGetRecentIndexesSummbryFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRecentIndexesSummbry method of the pbrent MockUplobdService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdServiceGetRecentIndexesSummbryFunc) PushHook(hook func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetRecentIndexesSummbryFunc) SetDefbultReturn(r0 []shbred1.IndexesWithRepositoryNbmespbce, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetRecentIndexesSummbryFunc) PushReturn(r0 []shbred1.IndexesWithRepositoryNbmespbce, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

func (f *UplobdServiceGetRecentIndexesSummbryFunc) nextHook() func(context.Context, int) ([]shbred1.IndexesWithRepositoryNbmespbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetRecentIndexesSummbryFunc) bppendCbll(r0 UplobdServiceGetRecentIndexesSummbryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdServiceGetRecentIndexesSummbryFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdServiceGetRecentIndexesSummbryFunc) History() []UplobdServiceGetRecentIndexesSummbryFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetRecentIndexesSummbryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetRecentIndexesSummbryFuncCbll is bn object thbt describes
// bn invocbtion of method GetRecentIndexesSummbry on bn instbnce of
// MockUplobdService.
type UplobdServiceGetRecentIndexesSummbryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.IndexesWithRepositoryNbmespbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetRecentIndexesSummbryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetRecentIndexesSummbryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdServiceGetRecentUplobdsSummbryFunc describes the behbvior when the
// GetRecentUplobdsSummbry method of the pbrent MockUplobdService instbnce
// is invoked.
type UplobdServiceGetRecentUplobdsSummbryFunc struct {
	defbultHook func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error)
	hooks       []func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error)
	history     []UplobdServiceGetRecentUplobdsSummbryFuncCbll
	mutex       sync.Mutex
}

// GetRecentUplobdsSummbry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetRecentUplobdsSummbry(v0 context.Context, v1 int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error) {
	r0, r1 := m.GetRecentUplobdsSummbryFunc.nextHook()(v0, v1)
	m.GetRecentUplobdsSummbryFunc.bppendCbll(UplobdServiceGetRecentUplobdsSummbryFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetRecentUplobdsSummbry method of the pbrent MockUplobdService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdServiceGetRecentUplobdsSummbryFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRecentUplobdsSummbry method of the pbrent MockUplobdService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdServiceGetRecentUplobdsSummbryFunc) PushHook(hook func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetRecentUplobdsSummbryFunc) SetDefbultReturn(r0 []shbred1.UplobdsWithRepositoryNbmespbce, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetRecentUplobdsSummbryFunc) PushReturn(r0 []shbred1.UplobdsWithRepositoryNbmespbce, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

func (f *UplobdServiceGetRecentUplobdsSummbryFunc) nextHook() func(context.Context, int) ([]shbred1.UplobdsWithRepositoryNbmespbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetRecentUplobdsSummbryFunc) bppendCbll(r0 UplobdServiceGetRecentUplobdsSummbryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdServiceGetRecentUplobdsSummbryFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdServiceGetRecentUplobdsSummbryFunc) History() []UplobdServiceGetRecentUplobdsSummbryFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetRecentUplobdsSummbryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetRecentUplobdsSummbryFuncCbll is bn object thbt describes
// bn invocbtion of method GetRecentUplobdsSummbry on bn instbnce of
// MockUplobdService.
type UplobdServiceGetRecentUplobdsSummbryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.UplobdsWithRepositoryNbmespbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetRecentUplobdsSummbryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetRecentUplobdsSummbryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdServiceGetUplobdByIDFunc describes the behbvior when the
// GetUplobdByID method of the pbrent MockUplobdService instbnce is invoked.
type UplobdServiceGetUplobdByIDFunc struct {
	defbultHook func(context.Context, int) (shbred1.Uplobd, bool, error)
	hooks       []func(context.Context, int) (shbred1.Uplobd, bool, error)
	history     []UplobdServiceGetUplobdByIDFuncCbll
	mutex       sync.Mutex
}

// GetUplobdByID delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetUplobdByID(v0 context.Context, v1 int) (shbred1.Uplobd, bool, error) {
	r0, r1, r2 := m.GetUplobdByIDFunc.nextHook()(v0, v1)
	m.GetUplobdByIDFunc.bppendCbll(UplobdServiceGetUplobdByIDFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetUplobdByID method
// of the pbrent MockUplobdService instbnce is invoked bnd the hook queue is
// empty.
func (f *UplobdServiceGetUplobdByIDFunc) SetDefbultHook(hook func(context.Context, int) (shbred1.Uplobd, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobdByID method of the pbrent MockUplobdService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UplobdServiceGetUplobdByIDFunc) PushHook(hook func(context.Context, int) (shbred1.Uplobd, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetUplobdByIDFunc) SetDefbultReturn(r0 shbred1.Uplobd, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred1.Uplobd, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetUplobdByIDFunc) PushReturn(r0 shbred1.Uplobd, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (shbred1.Uplobd, bool, error) {
		return r0, r1, r2
	})
}

func (f *UplobdServiceGetUplobdByIDFunc) nextHook() func(context.Context, int) (shbred1.Uplobd, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetUplobdByIDFunc) bppendCbll(r0 UplobdServiceGetUplobdByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdServiceGetUplobdByIDFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdServiceGetUplobdByIDFunc) History() []UplobdServiceGetUplobdByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetUplobdByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetUplobdByIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetUplobdByID on bn instbnce of MockUplobdService.
type UplobdServiceGetUplobdByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred1.Uplobd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetUplobdByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetUplobdByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// UplobdServiceReferencesForUplobdFunc describes the behbvior when the
// ReferencesForUplobd method of the pbrent MockUplobdService instbnce is
// invoked.
type UplobdServiceReferencesForUplobdFunc struct {
	defbultHook func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error)
	hooks       []func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error)
	history     []UplobdServiceReferencesForUplobdFuncCbll
	mutex       sync.Mutex
}

// ReferencesForUplobd delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) ReferencesForUplobd(v0 context.Context, v1 int) (shbred1.PbckbgeReferenceScbnner, error) {
	r0, r1 := m.ReferencesForUplobdFunc.nextHook()(v0, v1)
	m.ReferencesForUplobdFunc.bppendCbll(UplobdServiceReferencesForUplobdFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ReferencesForUplobd
// method of the pbrent MockUplobdService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdServiceReferencesForUplobdFunc) SetDefbultHook(hook func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReferencesForUplobd method of the pbrent MockUplobdService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdServiceReferencesForUplobdFunc) PushHook(hook func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceReferencesForUplobdFunc) SetDefbultReturn(r0 shbred1.PbckbgeReferenceScbnner, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceReferencesForUplobdFunc) PushReturn(r0 shbred1.PbckbgeReferenceScbnner, r1 error) {
	f.PushHook(func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error) {
		return r0, r1
	})
}

func (f *UplobdServiceReferencesForUplobdFunc) nextHook() func(context.Context, int) (shbred1.PbckbgeReferenceScbnner, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceReferencesForUplobdFunc) bppendCbll(r0 UplobdServiceReferencesForUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdServiceReferencesForUplobdFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdServiceReferencesForUplobdFunc) History() []UplobdServiceReferencesForUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceReferencesForUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceReferencesForUplobdFuncCbll is bn object thbt describes bn
// invocbtion of method ReferencesForUplobd on bn instbnce of
// MockUplobdService.
type UplobdServiceReferencesForUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred1.PbckbgeReferenceScbnner
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceReferencesForUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceReferencesForUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
