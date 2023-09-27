// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge repos

import (
	"context"
	"sync"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	bbsestore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/repos) used for unit
// testing.
type MockStore struct {
	// CrebteExternblServiceRepoFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// CrebteExternblServiceRepo.
	CrebteExternblServiceRepoFunc *StoreCrebteExternblServiceRepoFunc
	// DeleteExternblServiceRepoFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// DeleteExternblServiceRepo.
	DeleteExternblServiceRepoFunc *StoreDeleteExternblServiceRepoFunc
	// DeleteExternblServiceReposNotInFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// DeleteExternblServiceReposNotIn.
	DeleteExternblServiceReposNotInFunc *StoreDeleteExternblServiceReposNotInFunc
	// DoneFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Done.
	DoneFunc *StoreDoneFunc
	// EnqueueSingleSyncJobFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method EnqueueSingleSyncJob.
	EnqueueSingleSyncJobFunc *StoreEnqueueSingleSyncJobFunc
	// EnqueueSyncJobsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method EnqueueSyncJobs.
	EnqueueSyncJobsFunc *StoreEnqueueSyncJobsFunc
	// ExternblServiceStoreFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ExternblServiceStore.
	ExternblServiceStoreFunc *StoreExternblServiceStoreFunc
	// GitserverReposStoreFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GitserverReposStore.
	GitserverReposStoreFunc *StoreGitserverReposStoreFunc
	// HbndleFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hbndle.
	HbndleFunc *StoreHbndleFunc
	// ListSyncJobsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ListSyncJobs.
	ListSyncJobsFunc *StoreListSyncJobsFunc
	// RepoStoreFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method RepoStore.
	RepoStoreFunc *StoreRepoStoreFunc
	// SetMetricsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SetMetrics.
	SetMetricsFunc *StoreSetMetricsFunc
	// TrbnsbctFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Trbnsbct.
	TrbnsbctFunc *StoreTrbnsbctFunc
	// UpdbteExternblServiceRepoFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// UpdbteExternblServiceRepo.
	UpdbteExternblServiceRepoFunc *StoreUpdbteExternblServiceRepoFunc
	// UpdbteRepoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method UpdbteRepo.
	UpdbteRepoFunc *StoreUpdbteRepoFunc
	// WithFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method With.
	WithFunc *StoreWithFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		CrebteExternblServiceRepoFunc: &StoreCrebteExternblServiceRepoFunc{
			defbultHook: func(context.Context, *types.ExternblService, *types.Repo) (r0 error) {
				return
			},
		},
		DeleteExternblServiceRepoFunc: &StoreDeleteExternblServiceRepoFunc{
			defbultHook: func(context.Context, *types.ExternblService, bpi.RepoID) (r0 error) {
				return
			},
		},
		DeleteExternblServiceReposNotInFunc: &StoreDeleteExternblServiceReposNotInFunc{
			defbultHook: func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) (r0 []bpi.RepoID, r1 error) {
				return
			},
		},
		DoneFunc: &StoreDoneFunc{
			defbultHook: func(error) (r0 error) {
				return
			},
		},
		EnqueueSingleSyncJobFunc: &StoreEnqueueSingleSyncJobFunc{
			defbultHook: func(context.Context, int64) (r0 error) {
				return
			},
		},
		EnqueueSyncJobsFunc: &StoreEnqueueSyncJobsFunc{
			defbultHook: func(context.Context, bool) (r0 error) {
				return
			},
		},
		ExternblServiceStoreFunc: &StoreExternblServiceStoreFunc{
			defbultHook: func() (r0 dbtbbbse.ExternblServiceStore) {
				return
			},
		},
		GitserverReposStoreFunc: &StoreGitserverReposStoreFunc{
			defbultHook: func() (r0 dbtbbbse.GitserverRepoStore) {
				return
			},
		},
		HbndleFunc: &StoreHbndleFunc{
			defbultHook: func() (r0 bbsestore.TrbnsbctbbleHbndle) {
				return
			},
		},
		ListSyncJobsFunc: &StoreListSyncJobsFunc{
			defbultHook: func(context.Context) (r0 []SyncJob, r1 error) {
				return
			},
		},
		RepoStoreFunc: &StoreRepoStoreFunc{
			defbultHook: func() (r0 dbtbbbse.RepoStore) {
				return
			},
		},
		SetMetricsFunc: &StoreSetMetricsFunc{
			defbultHook: func(StoreMetrics) {
				return
			},
		},
		TrbnsbctFunc: &StoreTrbnsbctFunc{
			defbultHook: func(context.Context) (r0 Store, r1 error) {
				return
			},
		},
		UpdbteExternblServiceRepoFunc: &StoreUpdbteExternblServiceRepoFunc{
			defbultHook: func(context.Context, *types.ExternblService, *types.Repo) (r0 error) {
				return
			},
		},
		UpdbteRepoFunc: &StoreUpdbteRepoFunc{
			defbultHook: func(context.Context, *types.Repo) (r0 *types.Repo, r1 error) {
				return
			},
		},
		WithFunc: &StoreWithFunc{
			defbultHook: func(bbsestore.ShbrebbleStore) (r0 Store) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		CrebteExternblServiceRepoFunc: &StoreCrebteExternblServiceRepoFunc{
			defbultHook: func(context.Context, *types.ExternblService, *types.Repo) error {
				pbnic("unexpected invocbtion of MockStore.CrebteExternblServiceRepo")
			},
		},
		DeleteExternblServiceRepoFunc: &StoreDeleteExternblServiceRepoFunc{
			defbultHook: func(context.Context, *types.ExternblService, bpi.RepoID) error {
				pbnic("unexpected invocbtion of MockStore.DeleteExternblServiceRepo")
			},
		},
		DeleteExternblServiceReposNotInFunc: &StoreDeleteExternblServiceReposNotInFunc{
			defbultHook: func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error) {
				pbnic("unexpected invocbtion of MockStore.DeleteExternblServiceReposNotIn")
			},
		},
		DoneFunc: &StoreDoneFunc{
			defbultHook: func(error) error {
				pbnic("unexpected invocbtion of MockStore.Done")
			},
		},
		EnqueueSingleSyncJobFunc: &StoreEnqueueSingleSyncJobFunc{
			defbultHook: func(context.Context, int64) error {
				pbnic("unexpected invocbtion of MockStore.EnqueueSingleSyncJob")
			},
		},
		EnqueueSyncJobsFunc: &StoreEnqueueSyncJobsFunc{
			defbultHook: func(context.Context, bool) error {
				pbnic("unexpected invocbtion of MockStore.EnqueueSyncJobs")
			},
		},
		ExternblServiceStoreFunc: &StoreExternblServiceStoreFunc{
			defbultHook: func() dbtbbbse.ExternblServiceStore {
				pbnic("unexpected invocbtion of MockStore.ExternblServiceStore")
			},
		},
		GitserverReposStoreFunc: &StoreGitserverReposStoreFunc{
			defbultHook: func() dbtbbbse.GitserverRepoStore {
				pbnic("unexpected invocbtion of MockStore.GitserverReposStore")
			},
		},
		HbndleFunc: &StoreHbndleFunc{
			defbultHook: func() bbsestore.TrbnsbctbbleHbndle {
				pbnic("unexpected invocbtion of MockStore.Hbndle")
			},
		},
		ListSyncJobsFunc: &StoreListSyncJobsFunc{
			defbultHook: func(context.Context) ([]SyncJob, error) {
				pbnic("unexpected invocbtion of MockStore.ListSyncJobs")
			},
		},
		RepoStoreFunc: &StoreRepoStoreFunc{
			defbultHook: func() dbtbbbse.RepoStore {
				pbnic("unexpected invocbtion of MockStore.RepoStore")
			},
		},
		SetMetricsFunc: &StoreSetMetricsFunc{
			defbultHook: func(StoreMetrics) {
				pbnic("unexpected invocbtion of MockStore.SetMetrics")
			},
		},
		TrbnsbctFunc: &StoreTrbnsbctFunc{
			defbultHook: func(context.Context) (Store, error) {
				pbnic("unexpected invocbtion of MockStore.Trbnsbct")
			},
		},
		UpdbteExternblServiceRepoFunc: &StoreUpdbteExternblServiceRepoFunc{
			defbultHook: func(context.Context, *types.ExternblService, *types.Repo) error {
				pbnic("unexpected invocbtion of MockStore.UpdbteExternblServiceRepo")
			},
		},
		UpdbteRepoFunc: &StoreUpdbteRepoFunc{
			defbultHook: func(context.Context, *types.Repo) (*types.Repo, error) {
				pbnic("unexpected invocbtion of MockStore.UpdbteRepo")
			},
		},
		WithFunc: &StoreWithFunc{
			defbultHook: func(bbsestore.ShbrebbleStore) Store {
				pbnic("unexpected invocbtion of MockStore.With")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		CrebteExternblServiceRepoFunc: &StoreCrebteExternblServiceRepoFunc{
			defbultHook: i.CrebteExternblServiceRepo,
		},
		DeleteExternblServiceRepoFunc: &StoreDeleteExternblServiceRepoFunc{
			defbultHook: i.DeleteExternblServiceRepo,
		},
		DeleteExternblServiceReposNotInFunc: &StoreDeleteExternblServiceReposNotInFunc{
			defbultHook: i.DeleteExternblServiceReposNotIn,
		},
		DoneFunc: &StoreDoneFunc{
			defbultHook: i.Done,
		},
		EnqueueSingleSyncJobFunc: &StoreEnqueueSingleSyncJobFunc{
			defbultHook: i.EnqueueSingleSyncJob,
		},
		EnqueueSyncJobsFunc: &StoreEnqueueSyncJobsFunc{
			defbultHook: i.EnqueueSyncJobs,
		},
		ExternblServiceStoreFunc: &StoreExternblServiceStoreFunc{
			defbultHook: i.ExternblServiceStore,
		},
		GitserverReposStoreFunc: &StoreGitserverReposStoreFunc{
			defbultHook: i.GitserverReposStore,
		},
		HbndleFunc: &StoreHbndleFunc{
			defbultHook: i.Hbndle,
		},
		ListSyncJobsFunc: &StoreListSyncJobsFunc{
			defbultHook: i.ListSyncJobs,
		},
		RepoStoreFunc: &StoreRepoStoreFunc{
			defbultHook: i.RepoStore,
		},
		SetMetricsFunc: &StoreSetMetricsFunc{
			defbultHook: i.SetMetrics,
		},
		TrbnsbctFunc: &StoreTrbnsbctFunc{
			defbultHook: i.Trbnsbct,
		},
		UpdbteExternblServiceRepoFunc: &StoreUpdbteExternblServiceRepoFunc{
			defbultHook: i.UpdbteExternblServiceRepo,
		},
		UpdbteRepoFunc: &StoreUpdbteRepoFunc{
			defbultHook: i.UpdbteRepo,
		},
		WithFunc: &StoreWithFunc{
			defbultHook: i.With,
		},
	}
}

// StoreCrebteExternblServiceRepoFunc describes the behbvior when the
// CrebteExternblServiceRepo method of the pbrent MockStore instbnce is
// invoked.
type StoreCrebteExternblServiceRepoFunc struct {
	defbultHook func(context.Context, *types.ExternblService, *types.Repo) error
	hooks       []func(context.Context, *types.ExternblService, *types.Repo) error
	history     []StoreCrebteExternblServiceRepoFuncCbll
	mutex       sync.Mutex
}

// CrebteExternblServiceRepo delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) CrebteExternblServiceRepo(v0 context.Context, v1 *types.ExternblService, v2 *types.Repo) error {
	r0 := m.CrebteExternblServiceRepoFunc.nextHook()(v0, v1, v2)
	m.CrebteExternblServiceRepoFunc.bppendCbll(StoreCrebteExternblServiceRepoFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// CrebteExternblServiceRepo method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreCrebteExternblServiceRepoFunc) SetDefbultHook(hook func(context.Context, *types.ExternblService, *types.Repo) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteExternblServiceRepo method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreCrebteExternblServiceRepoFunc) PushHook(hook func(context.Context, *types.ExternblService, *types.Repo) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreCrebteExternblServiceRepoFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *types.ExternblService, *types.Repo) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreCrebteExternblServiceRepoFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *types.ExternblService, *types.Repo) error {
		return r0
	})
}

func (f *StoreCrebteExternblServiceRepoFunc) nextHook() func(context.Context, *types.ExternblService, *types.Repo) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreCrebteExternblServiceRepoFunc) bppendCbll(r0 StoreCrebteExternblServiceRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreCrebteExternblServiceRepoFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreCrebteExternblServiceRepoFunc) History() []StoreCrebteExternblServiceRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreCrebteExternblServiceRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreCrebteExternblServiceRepoFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteExternblServiceRepo on bn instbnce of
// MockStore.
type StoreCrebteExternblServiceRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.ExternblService
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *types.Repo
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreCrebteExternblServiceRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreCrebteExternblServiceRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreDeleteExternblServiceRepoFunc describes the behbvior when the
// DeleteExternblServiceRepo method of the pbrent MockStore instbnce is
// invoked.
type StoreDeleteExternblServiceRepoFunc struct {
	defbultHook func(context.Context, *types.ExternblService, bpi.RepoID) error
	hooks       []func(context.Context, *types.ExternblService, bpi.RepoID) error
	history     []StoreDeleteExternblServiceRepoFuncCbll
	mutex       sync.Mutex
}

// DeleteExternblServiceRepo delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) DeleteExternblServiceRepo(v0 context.Context, v1 *types.ExternblService, v2 bpi.RepoID) error {
	r0 := m.DeleteExternblServiceRepoFunc.nextHook()(v0, v1, v2)
	m.DeleteExternblServiceRepoFunc.bppendCbll(StoreDeleteExternblServiceRepoFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// DeleteExternblServiceRepo method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreDeleteExternblServiceRepoFunc) SetDefbultHook(hook func(context.Context, *types.ExternblService, bpi.RepoID) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteExternblServiceRepo method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreDeleteExternblServiceRepoFunc) PushHook(hook func(context.Context, *types.ExternblService, bpi.RepoID) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDeleteExternblServiceRepoFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *types.ExternblService, bpi.RepoID) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDeleteExternblServiceRepoFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *types.ExternblService, bpi.RepoID) error {
		return r0
	})
}

func (f *StoreDeleteExternblServiceRepoFunc) nextHook() func(context.Context, *types.ExternblService, bpi.RepoID) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteExternblServiceRepoFunc) bppendCbll(r0 StoreDeleteExternblServiceRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDeleteExternblServiceRepoFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreDeleteExternblServiceRepoFunc) History() []StoreDeleteExternblServiceRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDeleteExternblServiceRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteExternblServiceRepoFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteExternblServiceRepo on bn instbnce of
// MockStore.
type StoreDeleteExternblServiceRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.ExternblService
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDeleteExternblServiceRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDeleteExternblServiceRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreDeleteExternblServiceReposNotInFunc describes the behbvior when the
// DeleteExternblServiceReposNotIn method of the pbrent MockStore instbnce
// is invoked.
type StoreDeleteExternblServiceReposNotInFunc struct {
	defbultHook func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error)
	hooks       []func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error)
	history     []StoreDeleteExternblServiceReposNotInFuncCbll
	mutex       sync.Mutex
}

// DeleteExternblServiceReposNotIn delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) DeleteExternblServiceReposNotIn(v0 context.Context, v1 *types.ExternblService, v2 mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error) {
	r0, r1 := m.DeleteExternblServiceReposNotInFunc.nextHook()(v0, v1, v2)
	m.DeleteExternblServiceReposNotInFunc.bppendCbll(StoreDeleteExternblServiceReposNotInFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// DeleteExternblServiceReposNotIn method of the pbrent MockStore instbnce
// is invoked bnd the hook queue is empty.
func (f *StoreDeleteExternblServiceReposNotInFunc) SetDefbultHook(hook func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteExternblServiceReposNotIn method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreDeleteExternblServiceReposNotInFunc) PushHook(hook func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDeleteExternblServiceReposNotInFunc) SetDefbultReturn(r0 []bpi.RepoID, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDeleteExternblServiceReposNotInFunc) PushReturn(r0 []bpi.RepoID, r1 error) {
	f.PushHook(func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error) {
		return r0, r1
	})
}

func (f *StoreDeleteExternblServiceReposNotInFunc) nextHook() func(context.Context, *types.ExternblService, mbp[bpi.RepoID]struct{}) ([]bpi.RepoID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteExternblServiceReposNotInFunc) bppendCbll(r0 StoreDeleteExternblServiceReposNotInFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// StoreDeleteExternblServiceReposNotInFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreDeleteExternblServiceReposNotInFunc) History() []StoreDeleteExternblServiceReposNotInFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDeleteExternblServiceReposNotInFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteExternblServiceReposNotInFuncCbll is bn object thbt describes
// bn invocbtion of method DeleteExternblServiceReposNotIn on bn instbnce of
// MockStore.
type StoreDeleteExternblServiceReposNotInFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.ExternblService
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 mbp[bpi.RepoID]struct{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bpi.RepoID
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDeleteExternblServiceReposNotInFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDeleteExternblServiceReposNotInFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreDoneFunc describes the behbvior when the Done method of the pbrent
// MockStore instbnce is invoked.
type StoreDoneFunc struct {
	defbultHook func(error) error
	hooks       []func(error) error
	history     []StoreDoneFuncCbll
	mutex       sync.Mutex
}

// Done delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.bppendCbll(StoreDoneFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Done method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDoneFunc) SetDefbultHook(hook func(error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Done method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDoneFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *StoreDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDoneFunc) bppendCbll(r0 StoreDoneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDoneFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreDoneFunc) History() []StoreDoneFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDoneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDoneFuncCbll is bn object thbt describes bn invocbtion of method
// Done on bn instbnce of MockStore.
type StoreDoneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDoneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDoneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreEnqueueSingleSyncJobFunc describes the behbvior when the
// EnqueueSingleSyncJob method of the pbrent MockStore instbnce is invoked.
type StoreEnqueueSingleSyncJobFunc struct {
	defbultHook func(context.Context, int64) error
	hooks       []func(context.Context, int64) error
	history     []StoreEnqueueSingleSyncJobFuncCbll
	mutex       sync.Mutex
}

// EnqueueSingleSyncJob delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) EnqueueSingleSyncJob(v0 context.Context, v1 int64) error {
	r0 := m.EnqueueSingleSyncJobFunc.nextHook()(v0, v1)
	m.EnqueueSingleSyncJobFunc.bppendCbll(StoreEnqueueSingleSyncJobFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the EnqueueSingleSyncJob
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreEnqueueSingleSyncJobFunc) SetDefbultHook(hook func(context.Context, int64) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// EnqueueSingleSyncJob method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreEnqueueSingleSyncJobFunc) PushHook(hook func(context.Context, int64) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreEnqueueSingleSyncJobFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int64) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreEnqueueSingleSyncJobFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int64) error {
		return r0
	})
}

func (f *StoreEnqueueSingleSyncJobFunc) nextHook() func(context.Context, int64) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreEnqueueSingleSyncJobFunc) bppendCbll(r0 StoreEnqueueSingleSyncJobFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreEnqueueSingleSyncJobFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreEnqueueSingleSyncJobFunc) History() []StoreEnqueueSingleSyncJobFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreEnqueueSingleSyncJobFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreEnqueueSingleSyncJobFuncCbll is bn object thbt describes bn
// invocbtion of method EnqueueSingleSyncJob on bn instbnce of MockStore.
type StoreEnqueueSingleSyncJobFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreEnqueueSingleSyncJobFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreEnqueueSingleSyncJobFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreEnqueueSyncJobsFunc describes the behbvior when the EnqueueSyncJobs
// method of the pbrent MockStore instbnce is invoked.
type StoreEnqueueSyncJobsFunc struct {
	defbultHook func(context.Context, bool) error
	hooks       []func(context.Context, bool) error
	history     []StoreEnqueueSyncJobsFuncCbll
	mutex       sync.Mutex
}

// EnqueueSyncJobs delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) EnqueueSyncJobs(v0 context.Context, v1 bool) error {
	r0 := m.EnqueueSyncJobsFunc.nextHook()(v0, v1)
	m.EnqueueSyncJobsFunc.bppendCbll(StoreEnqueueSyncJobsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the EnqueueSyncJobs
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreEnqueueSyncJobsFunc) SetDefbultHook(hook func(context.Context, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// EnqueueSyncJobs method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreEnqueueSyncJobsFunc) PushHook(hook func(context.Context, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreEnqueueSyncJobsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreEnqueueSyncJobsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bool) error {
		return r0
	})
}

func (f *StoreEnqueueSyncJobsFunc) nextHook() func(context.Context, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreEnqueueSyncJobsFunc) bppendCbll(r0 StoreEnqueueSyncJobsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreEnqueueSyncJobsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreEnqueueSyncJobsFunc) History() []StoreEnqueueSyncJobsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreEnqueueSyncJobsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreEnqueueSyncJobsFuncCbll is bn object thbt describes bn invocbtion of
// method EnqueueSyncJobs on bn instbnce of MockStore.
type StoreEnqueueSyncJobsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreEnqueueSyncJobsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreEnqueueSyncJobsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreExternblServiceStoreFunc describes the behbvior when the
// ExternblServiceStore method of the pbrent MockStore instbnce is invoked.
type StoreExternblServiceStoreFunc struct {
	defbultHook func() dbtbbbse.ExternblServiceStore
	hooks       []func() dbtbbbse.ExternblServiceStore
	history     []StoreExternblServiceStoreFuncCbll
	mutex       sync.Mutex
}

// ExternblServiceStore delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) ExternblServiceStore() dbtbbbse.ExternblServiceStore {
	r0 := m.ExternblServiceStoreFunc.nextHook()()
	m.ExternblServiceStoreFunc.bppendCbll(StoreExternblServiceStoreFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ExternblServiceStore
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreExternblServiceStoreFunc) SetDefbultHook(hook func() dbtbbbse.ExternblServiceStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExternblServiceStore method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreExternblServiceStoreFunc) PushHook(hook func() dbtbbbse.ExternblServiceStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreExternblServiceStoreFunc) SetDefbultReturn(r0 dbtbbbse.ExternblServiceStore) {
	f.SetDefbultHook(func() dbtbbbse.ExternblServiceStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreExternblServiceStoreFunc) PushReturn(r0 dbtbbbse.ExternblServiceStore) {
	f.PushHook(func() dbtbbbse.ExternblServiceStore {
		return r0
	})
}

func (f *StoreExternblServiceStoreFunc) nextHook() func() dbtbbbse.ExternblServiceStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreExternblServiceStoreFunc) bppendCbll(r0 StoreExternblServiceStoreFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreExternblServiceStoreFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreExternblServiceStoreFunc) History() []StoreExternblServiceStoreFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreExternblServiceStoreFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreExternblServiceStoreFuncCbll is bn object thbt describes bn
// invocbtion of method ExternblServiceStore on bn instbnce of MockStore.
type StoreExternblServiceStoreFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.ExternblServiceStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreExternblServiceStoreFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreExternblServiceStoreFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreGitserverReposStoreFunc describes the behbvior when the
// GitserverReposStore method of the pbrent MockStore instbnce is invoked.
type StoreGitserverReposStoreFunc struct {
	defbultHook func() dbtbbbse.GitserverRepoStore
	hooks       []func() dbtbbbse.GitserverRepoStore
	history     []StoreGitserverReposStoreFuncCbll
	mutex       sync.Mutex
}

// GitserverReposStore delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GitserverReposStore() dbtbbbse.GitserverRepoStore {
	r0 := m.GitserverReposStoreFunc.nextHook()()
	m.GitserverReposStoreFunc.bppendCbll(StoreGitserverReposStoreFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GitserverReposStore
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreGitserverReposStoreFunc) SetDefbultHook(hook func() dbtbbbse.GitserverRepoStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GitserverReposStore method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreGitserverReposStoreFunc) PushHook(hook func() dbtbbbse.GitserverRepoStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGitserverReposStoreFunc) SetDefbultReturn(r0 dbtbbbse.GitserverRepoStore) {
	f.SetDefbultHook(func() dbtbbbse.GitserverRepoStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGitserverReposStoreFunc) PushReturn(r0 dbtbbbse.GitserverRepoStore) {
	f.PushHook(func() dbtbbbse.GitserverRepoStore {
		return r0
	})
}

func (f *StoreGitserverReposStoreFunc) nextHook() func() dbtbbbse.GitserverRepoStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGitserverReposStoreFunc) bppendCbll(r0 StoreGitserverReposStoreFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGitserverReposStoreFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGitserverReposStoreFunc) History() []StoreGitserverReposStoreFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGitserverReposStoreFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGitserverReposStoreFuncCbll is bn object thbt describes bn
// invocbtion of method GitserverReposStore on bn instbnce of MockStore.
type StoreGitserverReposStoreFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.GitserverRepoStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGitserverReposStoreFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGitserverReposStoreFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreHbndleFunc describes the behbvior when the Hbndle method of the
// pbrent MockStore instbnce is invoked.
type StoreHbndleFunc struct {
	defbultHook func() bbsestore.TrbnsbctbbleHbndle
	hooks       []func() bbsestore.TrbnsbctbbleHbndle
	history     []StoreHbndleFuncCbll
	mutex       sync.Mutex
}

// Hbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Hbndle() bbsestore.TrbnsbctbbleHbndle {
	r0 := m.HbndleFunc.nextHook()()
	m.HbndleFunc.bppendCbll(StoreHbndleFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Hbndle method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreHbndleFunc) SetDefbultHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hbndle method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreHbndleFunc) PushHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreHbndleFunc) SetDefbultReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.SetDefbultHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreHbndleFunc) PushReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.PushHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

func (f *StoreHbndleFunc) nextHook() func() bbsestore.TrbnsbctbbleHbndle {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreHbndleFunc) bppendCbll(r0 StoreHbndleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreHbndleFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreHbndleFunc) History() []StoreHbndleFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreHbndleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreHbndleFuncCbll is bn object thbt describes bn invocbtion of method
// Hbndle on bn instbnce of MockStore.
type StoreHbndleFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bbsestore.TrbnsbctbbleHbndle
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreHbndleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreHbndleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreListSyncJobsFunc describes the behbvior when the ListSyncJobs method
// of the pbrent MockStore instbnce is invoked.
type StoreListSyncJobsFunc struct {
	defbultHook func(context.Context) ([]SyncJob, error)
	hooks       []func(context.Context) ([]SyncJob, error)
	history     []StoreListSyncJobsFuncCbll
	mutex       sync.Mutex
}

// ListSyncJobs delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) ListSyncJobs(v0 context.Context) ([]SyncJob, error) {
	r0, r1 := m.ListSyncJobsFunc.nextHook()(v0)
	m.ListSyncJobsFunc.bppendCbll(StoreListSyncJobsFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListSyncJobs method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreListSyncJobsFunc) SetDefbultHook(hook func(context.Context) ([]SyncJob, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListSyncJobs method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreListSyncJobsFunc) PushHook(hook func(context.Context) ([]SyncJob, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreListSyncJobsFunc) SetDefbultReturn(r0 []SyncJob, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]SyncJob, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreListSyncJobsFunc) PushReturn(r0 []SyncJob, r1 error) {
	f.PushHook(func(context.Context) ([]SyncJob, error) {
		return r0, r1
	})
}

func (f *StoreListSyncJobsFunc) nextHook() func(context.Context) ([]SyncJob, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreListSyncJobsFunc) bppendCbll(r0 StoreListSyncJobsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreListSyncJobsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreListSyncJobsFunc) History() []StoreListSyncJobsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreListSyncJobsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreListSyncJobsFuncCbll is bn object thbt describes bn invocbtion of
// method ListSyncJobs on bn instbnce of MockStore.
type StoreListSyncJobsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []SyncJob
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreListSyncJobsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreListSyncJobsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreRepoStoreFunc describes the behbvior when the RepoStore method of
// the pbrent MockStore instbnce is invoked.
type StoreRepoStoreFunc struct {
	defbultHook func() dbtbbbse.RepoStore
	hooks       []func() dbtbbbse.RepoStore
	history     []StoreRepoStoreFuncCbll
	mutex       sync.Mutex
}

// RepoStore delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) RepoStore() dbtbbbse.RepoStore {
	r0 := m.RepoStoreFunc.nextHook()()
	m.RepoStoreFunc.bppendCbll(StoreRepoStoreFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the RepoStore method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreRepoStoreFunc) SetDefbultHook(hook func() dbtbbbse.RepoStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoStore method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreRepoStoreFunc) PushHook(hook func() dbtbbbse.RepoStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreRepoStoreFunc) SetDefbultReturn(r0 dbtbbbse.RepoStore) {
	f.SetDefbultHook(func() dbtbbbse.RepoStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreRepoStoreFunc) PushReturn(r0 dbtbbbse.RepoStore) {
	f.PushHook(func() dbtbbbse.RepoStore {
		return r0
	})
}

func (f *StoreRepoStoreFunc) nextHook() func() dbtbbbse.RepoStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreRepoStoreFunc) bppendCbll(r0 StoreRepoStoreFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreRepoStoreFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreRepoStoreFunc) History() []StoreRepoStoreFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreRepoStoreFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreRepoStoreFuncCbll is bn object thbt describes bn invocbtion of
// method RepoStore on bn instbnce of MockStore.
type StoreRepoStoreFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.RepoStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreRepoStoreFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreRepoStoreFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreSetMetricsFunc describes the behbvior when the SetMetrics method of
// the pbrent MockStore instbnce is invoked.
type StoreSetMetricsFunc struct {
	defbultHook func(StoreMetrics)
	hooks       []func(StoreMetrics)
	history     []StoreSetMetricsFuncCbll
	mutex       sync.Mutex
}

// SetMetrics delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) SetMetrics(v0 StoreMetrics) {
	m.SetMetricsFunc.nextHook()(v0)
	m.SetMetricsFunc.bppendCbll(StoreSetMetricsFuncCbll{v0})
	return
}

// SetDefbultHook sets function thbt is cblled when the SetMetrics method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreSetMetricsFunc) SetDefbultHook(hook func(StoreMetrics)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetMetrics method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreSetMetricsFunc) PushHook(hook func(StoreMetrics)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSetMetricsFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(StoreMetrics) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSetMetricsFunc) PushReturn() {
	f.PushHook(func(StoreMetrics) {
		return
	})
}

func (f *StoreSetMetricsFunc) nextHook() func(StoreMetrics) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSetMetricsFunc) bppendCbll(r0 StoreSetMetricsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreSetMetricsFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreSetMetricsFunc) History() []StoreSetMetricsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSetMetricsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSetMetricsFuncCbll is bn object thbt describes bn invocbtion of
// method SetMetrics on bn instbnce of MockStore.
type StoreSetMetricsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 StoreMetrics
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSetMetricsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSetMetricsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// StoreTrbnsbctFunc describes the behbvior when the Trbnsbct method of the
// pbrent MockStore instbnce is invoked.
type StoreTrbnsbctFunc struct {
	defbultHook func(context.Context) (Store, error)
	hooks       []func(context.Context) (Store, error)
	history     []StoreTrbnsbctFuncCbll
	mutex       sync.Mutex
}

// Trbnsbct delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Trbnsbct(v0 context.Context) (Store, error) {
	r0, r1 := m.TrbnsbctFunc.nextHook()(v0)
	m.TrbnsbctFunc.bppendCbll(StoreTrbnsbctFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Trbnsbct method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreTrbnsbctFunc) SetDefbultHook(hook func(context.Context) (Store, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Trbnsbct method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreTrbnsbctFunc) PushHook(hook func(context.Context) (Store, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreTrbnsbctFunc) SetDefbultReturn(r0 Store, r1 error) {
	f.SetDefbultHook(func(context.Context) (Store, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreTrbnsbctFunc) PushReturn(r0 Store, r1 error) {
	f.PushHook(func(context.Context) (Store, error) {
		return r0, r1
	})
}

func (f *StoreTrbnsbctFunc) nextHook() func(context.Context) (Store, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTrbnsbctFunc) bppendCbll(r0 StoreTrbnsbctFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreTrbnsbctFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreTrbnsbctFunc) History() []StoreTrbnsbctFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreTrbnsbctFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTrbnsbctFuncCbll is bn object thbt describes bn invocbtion of method
// Trbnsbct on bn instbnce of MockStore.
type StoreTrbnsbctFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Store
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreTrbnsbctFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreTrbnsbctFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreUpdbteExternblServiceRepoFunc describes the behbvior when the
// UpdbteExternblServiceRepo method of the pbrent MockStore instbnce is
// invoked.
type StoreUpdbteExternblServiceRepoFunc struct {
	defbultHook func(context.Context, *types.ExternblService, *types.Repo) error
	hooks       []func(context.Context, *types.ExternblService, *types.Repo) error
	history     []StoreUpdbteExternblServiceRepoFuncCbll
	mutex       sync.Mutex
}

// UpdbteExternblServiceRepo delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) UpdbteExternblServiceRepo(v0 context.Context, v1 *types.ExternblService, v2 *types.Repo) error {
	r0 := m.UpdbteExternblServiceRepoFunc.nextHook()(v0, v1, v2)
	m.UpdbteExternblServiceRepoFunc.bppendCbll(StoreUpdbteExternblServiceRepoFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteExternblServiceRepo method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreUpdbteExternblServiceRepoFunc) SetDefbultHook(hook func(context.Context, *types.ExternblService, *types.Repo) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteExternblServiceRepo method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreUpdbteExternblServiceRepoFunc) PushHook(hook func(context.Context, *types.ExternblService, *types.Repo) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpdbteExternblServiceRepoFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *types.ExternblService, *types.Repo) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpdbteExternblServiceRepoFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *types.ExternblService, *types.Repo) error {
		return r0
	})
}

func (f *StoreUpdbteExternblServiceRepoFunc) nextHook() func(context.Context, *types.ExternblService, *types.Repo) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdbteExternblServiceRepoFunc) bppendCbll(r0 StoreUpdbteExternblServiceRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUpdbteExternblServiceRepoFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreUpdbteExternblServiceRepoFunc) History() []StoreUpdbteExternblServiceRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUpdbteExternblServiceRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdbteExternblServiceRepoFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteExternblServiceRepo on bn instbnce of
// MockStore.
type StoreUpdbteExternblServiceRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.ExternblService
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *types.Repo
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpdbteExternblServiceRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpdbteExternblServiceRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreUpdbteRepoFunc describes the behbvior when the UpdbteRepo method of
// the pbrent MockStore instbnce is invoked.
type StoreUpdbteRepoFunc struct {
	defbultHook func(context.Context, *types.Repo) (*types.Repo, error)
	hooks       []func(context.Context, *types.Repo) (*types.Repo, error)
	history     []StoreUpdbteRepoFuncCbll
	mutex       sync.Mutex
}

// UpdbteRepo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) UpdbteRepo(v0 context.Context, v1 *types.Repo) (*types.Repo, error) {
	r0, r1 := m.UpdbteRepoFunc.nextHook()(v0, v1)
	m.UpdbteRepoFunc.bppendCbll(StoreUpdbteRepoFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the UpdbteRepo method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreUpdbteRepoFunc) SetDefbultHook(hook func(context.Context, *types.Repo) (*types.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteRepo method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreUpdbteRepoFunc) PushHook(hook func(context.Context, *types.Repo) (*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpdbteRepoFunc) SetDefbultReturn(r0 *types.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.Repo) (*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpdbteRepoFunc) PushReturn(r0 *types.Repo, r1 error) {
	f.PushHook(func(context.Context, *types.Repo) (*types.Repo, error) {
		return r0, r1
	})
}

func (f *StoreUpdbteRepoFunc) nextHook() func(context.Context, *types.Repo) (*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdbteRepoFunc) bppendCbll(r0 StoreUpdbteRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUpdbteRepoFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreUpdbteRepoFunc) History() []StoreUpdbteRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUpdbteRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdbteRepoFuncCbll is bn object thbt describes bn invocbtion of
// method UpdbteRepo on bn instbnce of MockStore.
type StoreUpdbteRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.Repo
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpdbteRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpdbteRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreWithFunc describes the behbvior when the With method of the pbrent
// MockStore instbnce is invoked.
type StoreWithFunc struct {
	defbultHook func(bbsestore.ShbrebbleStore) Store
	hooks       []func(bbsestore.ShbrebbleStore) Store
	history     []StoreWithFuncCbll
	mutex       sync.Mutex
}

// With delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) With(v0 bbsestore.ShbrebbleStore) Store {
	r0 := m.WithFunc.nextHook()(v0)
	m.WithFunc.bppendCbll(StoreWithFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the With method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreWithFunc) SetDefbultHook(hook func(bbsestore.ShbrebbleStore) Store) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// With method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreWithFunc) PushHook(hook func(bbsestore.ShbrebbleStore) Store) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreWithFunc) SetDefbultReturn(r0 Store) {
	f.SetDefbultHook(func(bbsestore.ShbrebbleStore) Store {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreWithFunc) PushReturn(r0 Store) {
	f.PushHook(func(bbsestore.ShbrebbleStore) Store {
		return r0
	})
}

func (f *StoreWithFunc) nextHook() func(bbsestore.ShbrebbleStore) Store {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreWithFunc) bppendCbll(r0 StoreWithFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreWithFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreWithFunc) History() []StoreWithFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreWithFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreWithFuncCbll is bn object thbt describes bn invocbtion of method
// With on bn instbnce of MockStore.
type StoreWithFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 bbsestore.ShbrebbleStore
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Store
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreWithFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreWithFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
