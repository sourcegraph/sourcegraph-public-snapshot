// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge dependencies

import (
	"context"
	"sync"
	"time"

	sqlf "github.com/keegbncsmith/sqlf"
	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	store "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	shbred2 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	dependencies "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	shbred1 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	bbsestore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	executor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	protocol "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
	workerutil "github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	store1 "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// MockDependenciesService is b mock implementbtion of the
// DependenciesService interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockDependenciesService struct {
	// InsertPbckbgeRepoRefsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertPbckbgeRepoRefs.
	InsertPbckbgeRepoRefsFunc *DependenciesServiceInsertPbckbgeRepoRefsFunc
	// ListPbckbgeRepoFiltersFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListPbckbgeRepoFilters.
	ListPbckbgeRepoFiltersFunc *DependenciesServiceListPbckbgeRepoFiltersFunc
}

// NewMockDependenciesService crebtes b new mock of the DependenciesService
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockDependenciesService() *MockDependenciesService {
	return &MockDependenciesService{
		InsertPbckbgeRepoRefsFunc: &DependenciesServiceInsertPbckbgeRepoRefsFunc{
			defbultHook: func(context.Context, []shbred.MinimblPbckbgeRepoRef) (r0 []shbred.PbckbgeRepoReference, r1 []shbred.PbckbgeRepoRefVersion, r2 error) {
				return
			},
		},
		ListPbckbgeRepoFiltersFunc: &DependenciesServiceListPbckbgeRepoFiltersFunc{
			defbultHook: func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) (r0 []shbred.PbckbgeRepoFilter, r1 bool, r2 error) {
				return
			},
		},
	}
}

// NewStrictMockDependenciesService crebtes b new mock of the
// DependenciesService interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockDependenciesService() *MockDependenciesService {
	return &MockDependenciesService{
		InsertPbckbgeRepoRefsFunc: &DependenciesServiceInsertPbckbgeRepoRefsFunc{
			defbultHook: func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error) {
				pbnic("unexpected invocbtion of MockDependenciesService.InsertPbckbgeRepoRefs")
			},
		},
		ListPbckbgeRepoFiltersFunc: &DependenciesServiceListPbckbgeRepoFiltersFunc{
			defbultHook: func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error) {
				pbnic("unexpected invocbtion of MockDependenciesService.ListPbckbgeRepoFilters")
			},
		},
	}
}

// NewMockDependenciesServiceFrom crebtes b new mock of the
// MockDependenciesService interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockDependenciesServiceFrom(i DependenciesService) *MockDependenciesService {
	return &MockDependenciesService{
		InsertPbckbgeRepoRefsFunc: &DependenciesServiceInsertPbckbgeRepoRefsFunc{
			defbultHook: i.InsertPbckbgeRepoRefs,
		},
		ListPbckbgeRepoFiltersFunc: &DependenciesServiceListPbckbgeRepoFiltersFunc{
			defbultHook: i.ListPbckbgeRepoFilters,
		},
	}
}

// DependenciesServiceInsertPbckbgeRepoRefsFunc describes the behbvior when
// the InsertPbckbgeRepoRefs method of the pbrent MockDependenciesService
// instbnce is invoked.
type DependenciesServiceInsertPbckbgeRepoRefsFunc struct {
	defbultHook func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error)
	hooks       []func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error)
	history     []DependenciesServiceInsertPbckbgeRepoRefsFuncCbll
	mutex       sync.Mutex
}

// InsertPbckbgeRepoRefs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDependenciesService) InsertPbckbgeRepoRefs(v0 context.Context, v1 []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error) {
	r0, r1, r2 := m.InsertPbckbgeRepoRefsFunc.nextHook()(v0, v1)
	m.InsertPbckbgeRepoRefsFunc.bppendCbll(DependenciesServiceInsertPbckbgeRepoRefsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// InsertPbckbgeRepoRefs method of the pbrent MockDependenciesService
// instbnce is invoked bnd the hook queue is empty.
func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) SetDefbultHook(hook func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertPbckbgeRepoRefs method of the pbrent MockDependenciesService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) PushHook(hook func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) SetDefbultReturn(r0 []shbred.PbckbgeRepoReference, r1 []shbred.PbckbgeRepoRefVersion, r2 error) {
	f.SetDefbultHook(func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) PushReturn(r0 []shbred.PbckbgeRepoReference, r1 []shbred.PbckbgeRepoRefVersion, r2 error) {
	f.PushHook(func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error) {
		return r0, r1, r2
	})
}

func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) nextHook() func(context.Context, []shbred.MinimblPbckbgeRepoRef) ([]shbred.PbckbgeRepoReference, []shbred.PbckbgeRepoRefVersion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) bppendCbll(r0 DependenciesServiceInsertPbckbgeRepoRefsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// DependenciesServiceInsertPbckbgeRepoRefsFuncCbll objects describing the
// invocbtions of this function.
func (f *DependenciesServiceInsertPbckbgeRepoRefsFunc) History() []DependenciesServiceInsertPbckbgeRepoRefsFuncCbll {
	f.mutex.Lock()
	history := mbke([]DependenciesServiceInsertPbckbgeRepoRefsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DependenciesServiceInsertPbckbgeRepoRefsFuncCbll is bn object thbt
// describes bn invocbtion of method InsertPbckbgeRepoRefs on bn instbnce of
// MockDependenciesService.
type DependenciesServiceInsertPbckbgeRepoRefsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []shbred.MinimblPbckbgeRepoRef
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.PbckbgeRepoReference
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []shbred.PbckbgeRepoRefVersion
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DependenciesServiceInsertPbckbgeRepoRefsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DependenciesServiceInsertPbckbgeRepoRefsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// DependenciesServiceListPbckbgeRepoFiltersFunc describes the behbvior when
// the ListPbckbgeRepoFilters method of the pbrent MockDependenciesService
// instbnce is invoked.
type DependenciesServiceListPbckbgeRepoFiltersFunc struct {
	defbultHook func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error)
	hooks       []func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error)
	history     []DependenciesServiceListPbckbgeRepoFiltersFuncCbll
	mutex       sync.Mutex
}

// ListPbckbgeRepoFilters delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDependenciesService) ListPbckbgeRepoFilters(v0 context.Context, v1 dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error) {
	r0, r1, r2 := m.ListPbckbgeRepoFiltersFunc.nextHook()(v0, v1)
	m.ListPbckbgeRepoFiltersFunc.bppendCbll(DependenciesServiceListPbckbgeRepoFiltersFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ListPbckbgeRepoFilters method of the pbrent MockDependenciesService
// instbnce is invoked bnd the hook queue is empty.
func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) SetDefbultHook(hook func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListPbckbgeRepoFilters method of the pbrent MockDependenciesService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) PushHook(hook func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) SetDefbultReturn(r0 []shbred.PbckbgeRepoFilter, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) PushReturn(r0 []shbred.PbckbgeRepoFilter, r1 bool, r2 error) {
	f.PushHook(func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error) {
		return r0, r1, r2
	})
}

func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) nextHook() func(context.Context, dependencies.ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) bppendCbll(r0 DependenciesServiceListPbckbgeRepoFiltersFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// DependenciesServiceListPbckbgeRepoFiltersFuncCbll objects describing the
// invocbtions of this function.
func (f *DependenciesServiceListPbckbgeRepoFiltersFunc) History() []DependenciesServiceListPbckbgeRepoFiltersFuncCbll {
	f.mutex.Lock()
	history := mbke([]DependenciesServiceListPbckbgeRepoFiltersFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DependenciesServiceListPbckbgeRepoFiltersFuncCbll is bn object thbt
// describes bn invocbtion of method ListPbckbgeRepoFilters on bn instbnce
// of MockDependenciesService.
type DependenciesServiceListPbckbgeRepoFiltersFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 dependencies.ListPbckbgeRepoRefFiltersOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.PbckbgeRepoFilter
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DependenciesServiceListPbckbgeRepoFiltersFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DependenciesServiceListPbckbgeRepoFiltersFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// MockExternblServiceStore is b mock implementbtion of the
// ExternblServiceStore interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockExternblServiceStore struct {
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *ExternblServiceStoreListFunc
	// UpsertFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Upsert.
	UpsertFunc *ExternblServiceStoreUpsertFunc
}

// NewMockExternblServiceStore crebtes b new mock of the
// ExternblServiceStore interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockExternblServiceStore() *MockExternblServiceStore {
	return &MockExternblServiceStore{
		ListFunc: &ExternblServiceStoreListFunc{
			defbultHook: func(context.Context, dbtbbbse.ExternblServicesListOptions) (r0 []*types.ExternblService, r1 error) {
				return
			},
		},
		UpsertFunc: &ExternblServiceStoreUpsertFunc{
			defbultHook: func(context.Context, ...*types.ExternblService) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockExternblServiceStore crebtes b new mock of the
// ExternblServiceStore interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockExternblServiceStore() *MockExternblServiceStore {
	return &MockExternblServiceStore{
		ListFunc: &ExternblServiceStoreListFunc{
			defbultHook: func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				pbnic("unexpected invocbtion of MockExternblServiceStore.List")
			},
		},
		UpsertFunc: &ExternblServiceStoreUpsertFunc{
			defbultHook: func(context.Context, ...*types.ExternblService) error {
				pbnic("unexpected invocbtion of MockExternblServiceStore.Upsert")
			},
		},
	}
}

// NewMockExternblServiceStoreFrom crebtes b new mock of the
// MockExternblServiceStore interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockExternblServiceStoreFrom(i ExternblServiceStore) *MockExternblServiceStore {
	return &MockExternblServiceStore{
		ListFunc: &ExternblServiceStoreListFunc{
			defbultHook: i.List,
		},
		UpsertFunc: &ExternblServiceStoreUpsertFunc{
			defbultHook: i.Upsert,
		},
	}
}

// ExternblServiceStoreListFunc describes the behbvior when the List method
// of the pbrent MockExternblServiceStore instbnce is invoked.
type ExternblServiceStoreListFunc struct {
	defbultHook func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
	hooks       []func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
	history     []ExternblServiceStoreListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockExternblServiceStore) List(v0 context.Context, v1 dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.bppendCbll(ExternblServiceStoreListFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockExternblServiceStore instbnce is invoked bnd the hook queue is
// empty.
func (f *ExternblServiceStoreListFunc) SetDefbultHook(hook func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockExternblServiceStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ExternblServiceStoreListFunc) PushHook(hook func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ExternblServiceStoreListFunc) SetDefbultReturn(r0 []*types.ExternblService, r1 error) {
	f.SetDefbultHook(func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ExternblServiceStoreListFunc) PushReturn(r0 []*types.ExternblService, r1 error) {
	f.PushHook(func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		return r0, r1
	})
}

func (f *ExternblServiceStoreListFunc) nextHook() func(context.Context, dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExternblServiceStoreListFunc) bppendCbll(r0 ExternblServiceStoreListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ExternblServiceStoreListFuncCbll objects
// describing the invocbtions of this function.
func (f *ExternblServiceStoreListFunc) History() []ExternblServiceStoreListFuncCbll {
	f.mutex.Lock()
	history := mbke([]ExternblServiceStoreListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExternblServiceStoreListFuncCbll is bn object thbt describes bn
// invocbtion of method List on bn instbnce of MockExternblServiceStore.
type ExternblServiceStoreListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 dbtbbbse.ExternblServicesListOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.ExternblService
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ExternblServiceStoreListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ExternblServiceStoreListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ExternblServiceStoreUpsertFunc describes the behbvior when the Upsert
// method of the pbrent MockExternblServiceStore instbnce is invoked.
type ExternblServiceStoreUpsertFunc struct {
	defbultHook func(context.Context, ...*types.ExternblService) error
	hooks       []func(context.Context, ...*types.ExternblService) error
	history     []ExternblServiceStoreUpsertFuncCbll
	mutex       sync.Mutex
}

// Upsert delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockExternblServiceStore) Upsert(v0 context.Context, v1 ...*types.ExternblService) error {
	r0 := m.UpsertFunc.nextHook()(v0, v1...)
	m.UpsertFunc.bppendCbll(ExternblServiceStoreUpsertFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Upsert method of the
// pbrent MockExternblServiceStore instbnce is invoked bnd the hook queue is
// empty.
func (f *ExternblServiceStoreUpsertFunc) SetDefbultHook(hook func(context.Context, ...*types.ExternblService) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Upsert method of the pbrent MockExternblServiceStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ExternblServiceStoreUpsertFunc) PushHook(hook func(context.Context, ...*types.ExternblService) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ExternblServiceStoreUpsertFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, ...*types.ExternblService) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ExternblServiceStoreUpsertFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, ...*types.ExternblService) error {
		return r0
	})
}

func (f *ExternblServiceStoreUpsertFunc) nextHook() func(context.Context, ...*types.ExternblService) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExternblServiceStoreUpsertFunc) bppendCbll(r0 ExternblServiceStoreUpsertFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ExternblServiceStoreUpsertFuncCbll objects
// describing the invocbtions of this function.
func (f *ExternblServiceStoreUpsertFunc) History() []ExternblServiceStoreUpsertFuncCbll {
	f.mutex.Lock()
	history := mbke([]ExternblServiceStoreUpsertFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExternblServiceStoreUpsertFuncCbll is bn object thbt describes bn
// invocbtion of method Upsert on bn instbnce of MockExternblServiceStore.
type ExternblServiceStoreUpsertFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []*types.ExternblService
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ExternblServiceStoreUpsertFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ExternblServiceStoreUpsertFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockGitserverRepoStore is b mock implementbtion of the GitserverRepoStore
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockGitserverRepoStore struct {
	// GetByNbmesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetByNbmes.
	GetByNbmesFunc *GitserverRepoStoreGetByNbmesFunc
}

// NewMockGitserverRepoStore crebtes b new mock of the GitserverRepoStore
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGitserverRepoStore() *MockGitserverRepoStore {
	return &MockGitserverRepoStore{
		GetByNbmesFunc: &GitserverRepoStoreGetByNbmesFunc{
			defbultHook: func(context.Context, ...bpi.RepoNbme) (r0 mbp[bpi.RepoNbme]*types.GitserverRepo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockGitserverRepoStore crebtes b new mock of the
// GitserverRepoStore interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockGitserverRepoStore() *MockGitserverRepoStore {
	return &MockGitserverRepoStore{
		GetByNbmesFunc: &GitserverRepoStoreGetByNbmesFunc{
			defbultHook: func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error) {
				pbnic("unexpected invocbtion of MockGitserverRepoStore.GetByNbmes")
			},
		},
	}
}

// NewMockGitserverRepoStoreFrom crebtes b new mock of the
// MockGitserverRepoStore interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockGitserverRepoStoreFrom(i GitserverRepoStore) *MockGitserverRepoStore {
	return &MockGitserverRepoStore{
		GetByNbmesFunc: &GitserverRepoStoreGetByNbmesFunc{
			defbultHook: i.GetByNbmes,
		},
	}
}

// GitserverRepoStoreGetByNbmesFunc describes the behbvior when the
// GetByNbmes method of the pbrent MockGitserverRepoStore instbnce is
// invoked.
type GitserverRepoStoreGetByNbmesFunc struct {
	defbultHook func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error)
	hooks       []func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error)
	history     []GitserverRepoStoreGetByNbmesFuncCbll
	mutex       sync.Mutex
}

// GetByNbmes delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverRepoStore) GetByNbmes(v0 context.Context, v1 ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error) {
	r0, r1 := m.GetByNbmesFunc.nextHook()(v0, v1...)
	m.GetByNbmesFunc.bppendCbll(GitserverRepoStoreGetByNbmesFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByNbmes method of
// the pbrent MockGitserverRepoStore instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverRepoStoreGetByNbmesFunc) SetDefbultHook(hook func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByNbmes method of the pbrent MockGitserverRepoStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverRepoStoreGetByNbmesFunc) PushHook(hook func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverRepoStoreGetByNbmesFunc) SetDefbultReturn(r0 mbp[bpi.RepoNbme]*types.GitserverRepo, r1 error) {
	f.SetDefbultHook(func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverRepoStoreGetByNbmesFunc) PushReturn(r0 mbp[bpi.RepoNbme]*types.GitserverRepo, r1 error) {
	f.PushHook(func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error) {
		return r0, r1
	})
}

func (f *GitserverRepoStoreGetByNbmesFunc) nextHook() func(context.Context, ...bpi.RepoNbme) (mbp[bpi.RepoNbme]*types.GitserverRepo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverRepoStoreGetByNbmesFunc) bppendCbll(r0 GitserverRepoStoreGetByNbmesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverRepoStoreGetByNbmesFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverRepoStoreGetByNbmesFunc) History() []GitserverRepoStoreGetByNbmesFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverRepoStoreGetByNbmesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverRepoStoreGetByNbmesFuncCbll is bn object thbt describes bn
// invocbtion of method GetByNbmes on bn instbnce of MockGitserverRepoStore.
type GitserverRepoStoreGetByNbmesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[bpi.RepoNbme]*types.GitserverRepo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GitserverRepoStoreGetByNbmesFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverRepoStoreGetByNbmesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockIndexEnqueuer is b mock implementbtion of the IndexEnqueuer interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockIndexEnqueuer struct {
	// QueueIndexesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method QueueIndexes.
	QueueIndexesFunc *IndexEnqueuerQueueIndexesFunc
	// QueueIndexesForPbckbgeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method QueueIndexesForPbckbge.
	QueueIndexesForPbckbgeFunc *IndexEnqueuerQueueIndexesForPbckbgeFunc
}

// NewMockIndexEnqueuer crebtes b new mock of the IndexEnqueuer interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockIndexEnqueuer() *MockIndexEnqueuer {
	return &MockIndexEnqueuer{
		QueueIndexesFunc: &IndexEnqueuerQueueIndexesFunc{
			defbultHook: func(context.Context, int, string, string, bool, bool) (r0 []shbred1.Index, r1 error) {
				return
			},
		},
		QueueIndexesForPbckbgeFunc: &IndexEnqueuerQueueIndexesForPbckbgeFunc{
			defbultHook: func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockIndexEnqueuer crebtes b new mock of the IndexEnqueuer
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockIndexEnqueuer() *MockIndexEnqueuer {
	return &MockIndexEnqueuer{
		QueueIndexesFunc: &IndexEnqueuerQueueIndexesFunc{
			defbultHook: func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error) {
				pbnic("unexpected invocbtion of MockIndexEnqueuer.QueueIndexes")
			},
		},
		QueueIndexesForPbckbgeFunc: &IndexEnqueuerQueueIndexesForPbckbgeFunc{
			defbultHook: func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
				pbnic("unexpected invocbtion of MockIndexEnqueuer.QueueIndexesForPbckbge")
			},
		},
	}
}

// NewMockIndexEnqueuerFrom crebtes b new mock of the MockIndexEnqueuer
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockIndexEnqueuerFrom(i IndexEnqueuer) *MockIndexEnqueuer {
	return &MockIndexEnqueuer{
		QueueIndexesFunc: &IndexEnqueuerQueueIndexesFunc{
			defbultHook: i.QueueIndexes,
		},
		QueueIndexesForPbckbgeFunc: &IndexEnqueuerQueueIndexesForPbckbgeFunc{
			defbultHook: i.QueueIndexesForPbckbge,
		},
	}
}

// IndexEnqueuerQueueIndexesFunc describes the behbvior when the
// QueueIndexes method of the pbrent MockIndexEnqueuer instbnce is invoked.
type IndexEnqueuerQueueIndexesFunc struct {
	defbultHook func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error)
	hooks       []func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error)
	history     []IndexEnqueuerQueueIndexesFuncCbll
	mutex       sync.Mutex
}

// QueueIndexes delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockIndexEnqueuer) QueueIndexes(v0 context.Context, v1 int, v2 string, v3 string, v4 bool, v5 bool) ([]shbred1.Index, error) {
	r0, r1 := m.QueueIndexesFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.QueueIndexesFunc.bppendCbll(IndexEnqueuerQueueIndexesFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the QueueIndexes method
// of the pbrent MockIndexEnqueuer instbnce is invoked bnd the hook queue is
// empty.
func (f *IndexEnqueuerQueueIndexesFunc) SetDefbultHook(hook func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueueIndexes method of the pbrent MockIndexEnqueuer instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *IndexEnqueuerQueueIndexesFunc) PushHook(hook func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *IndexEnqueuerQueueIndexesFunc) SetDefbultReturn(r0 []shbred1.Index, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *IndexEnqueuerQueueIndexesFunc) PushReturn(r0 []shbred1.Index, r1 error) {
	f.PushHook(func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error) {
		return r0, r1
	})
}

func (f *IndexEnqueuerQueueIndexesFunc) nextHook() func(context.Context, int, string, string, bool, bool) ([]shbred1.Index, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *IndexEnqueuerQueueIndexesFunc) bppendCbll(r0 IndexEnqueuerQueueIndexesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of IndexEnqueuerQueueIndexesFuncCbll objects
// describing the invocbtions of this function.
func (f *IndexEnqueuerQueueIndexesFunc) History() []IndexEnqueuerQueueIndexesFuncCbll {
	f.mutex.Lock()
	history := mbke([]IndexEnqueuerQueueIndexesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// IndexEnqueuerQueueIndexesFuncCbll is bn object thbt describes bn
// invocbtion of method QueueIndexes on bn instbnce of MockIndexEnqueuer.
type IndexEnqueuerQueueIndexesFuncCbll struct {
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
	Arg4 bool
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.Index
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c IndexEnqueuerQueueIndexesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c IndexEnqueuerQueueIndexesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// IndexEnqueuerQueueIndexesForPbckbgeFunc describes the behbvior when the
// QueueIndexesForPbckbge method of the pbrent MockIndexEnqueuer instbnce is
// invoked.
type IndexEnqueuerQueueIndexesForPbckbgeFunc struct {
	defbultHook func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error
	hooks       []func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error
	history     []IndexEnqueuerQueueIndexesForPbckbgeFuncCbll
	mutex       sync.Mutex
}

// QueueIndexesForPbckbge delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockIndexEnqueuer) QueueIndexesForPbckbge(v0 context.Context, v1 shbred.MinimiblVersionedPbckbgeRepo, v2 bool) error {
	r0 := m.QueueIndexesForPbckbgeFunc.nextHook()(v0, v1, v2)
	m.QueueIndexesForPbckbgeFunc.bppendCbll(IndexEnqueuerQueueIndexesForPbckbgeFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// QueueIndexesForPbckbge method of the pbrent MockIndexEnqueuer instbnce is
// invoked bnd the hook queue is empty.
func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) SetDefbultHook(hook func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueueIndexesForPbckbge method of the pbrent MockIndexEnqueuer instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) PushHook(hook func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
		return r0
	})
}

func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) nextHook() func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) bppendCbll(r0 IndexEnqueuerQueueIndexesForPbckbgeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of IndexEnqueuerQueueIndexesForPbckbgeFuncCbll
// objects describing the invocbtions of this function.
func (f *IndexEnqueuerQueueIndexesForPbckbgeFunc) History() []IndexEnqueuerQueueIndexesForPbckbgeFuncCbll {
	f.mutex.Lock()
	history := mbke([]IndexEnqueuerQueueIndexesForPbckbgeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// IndexEnqueuerQueueIndexesForPbckbgeFuncCbll is bn object thbt describes
// bn invocbtion of method QueueIndexesForPbckbge on bn instbnce of
// MockIndexEnqueuer.
type IndexEnqueuerQueueIndexesForPbckbgeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.MinimiblVersionedPbckbgeRepo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c IndexEnqueuerQueueIndexesForPbckbgeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c IndexEnqueuerQueueIndexesForPbckbgeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockRepoUpdbterClient is b mock implementbtion of the RepoUpdbterClient
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockRepoUpdbterClient struct {
	// RepoLookupFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method RepoLookup.
	RepoLookupFunc *RepoUpdbterClientRepoLookupFunc
}

// NewMockRepoUpdbterClient crebtes b new mock of the RepoUpdbterClient
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockRepoUpdbterClient() *MockRepoUpdbterClient {
	return &MockRepoUpdbterClient{
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
		RepoLookupFunc: &RepoUpdbterClientRepoLookupFunc{
			defbultHook: i.RepoLookup,
		},
	}
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

// MockReposStore is b mock implementbtion of the ReposStore interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockReposStore struct {
	// ListMinimblReposFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListMinimblRepos.
	ListMinimblReposFunc *ReposStoreListMinimblReposFunc
}

// NewMockReposStore crebtes b new mock of the ReposStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockReposStore() *MockReposStore {
	return &MockReposStore{
		ListMinimblReposFunc: &ReposStoreListMinimblReposFunc{
			defbultHook: func(context.Context, dbtbbbse.ReposListOptions) (r0 []types.MinimblRepo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockReposStore crebtes b new mock of the ReposStore interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockReposStore() *MockReposStore {
	return &MockReposStore{
		ListMinimblReposFunc: &ReposStoreListMinimblReposFunc{
			defbultHook: func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
				pbnic("unexpected invocbtion of MockReposStore.ListMinimblRepos")
			},
		},
	}
}

// NewMockReposStoreFrom crebtes b new mock of the MockReposStore interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockReposStoreFrom(i ReposStore) *MockReposStore {
	return &MockReposStore{
		ListMinimblReposFunc: &ReposStoreListMinimblReposFunc{
			defbultHook: i.ListMinimblRepos,
		},
	}
}

// ReposStoreListMinimblReposFunc describes the behbvior when the
// ListMinimblRepos method of the pbrent MockReposStore instbnce is invoked.
type ReposStoreListMinimblReposFunc struct {
	defbultHook func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error)
	hooks       []func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error)
	history     []ReposStoreListMinimblReposFuncCbll
	mutex       sync.Mutex
}

// ListMinimblRepos delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposStore) ListMinimblRepos(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
	r0, r1 := m.ListMinimblReposFunc.nextHook()(v0, v1)
	m.ListMinimblReposFunc.bppendCbll(ReposStoreListMinimblReposFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListMinimblRepos
// method of the pbrent MockReposStore instbnce is invoked bnd the hook
// queue is empty.
func (f *ReposStoreListMinimblReposFunc) SetDefbultHook(hook func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListMinimblRepos method of the pbrent MockReposStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ReposStoreListMinimblReposFunc) PushHook(hook func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposStoreListMinimblReposFunc) SetDefbultReturn(r0 []types.MinimblRepo, r1 error) {
	f.SetDefbultHook(func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposStoreListMinimblReposFunc) PushReturn(r0 []types.MinimblRepo, r1 error) {
	f.PushHook(func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		return r0, r1
	})
}

func (f *ReposStoreListMinimblReposFunc) nextHook() func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposStoreListMinimblReposFunc) bppendCbll(r0 ReposStoreListMinimblReposFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposStoreListMinimblReposFuncCbll objects
// describing the invocbtions of this function.
func (f *ReposStoreListMinimblReposFunc) History() []ReposStoreListMinimblReposFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposStoreListMinimblReposFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposStoreListMinimblReposFuncCbll is bn object thbt describes bn
// invocbtion of method ListMinimblRepos on bn instbnce of MockReposStore.
type ReposStoreListMinimblReposFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 dbtbbbse.ReposListOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []types.MinimblRepo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposStoreListMinimblReposFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposStoreListMinimblReposFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockUplobdService is b mock implementbtion of the UplobdService interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/bbckground/dependencies)
// used for unit testing.
type MockUplobdService struct {
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
		GetUplobdByIDFunc: &UplobdServiceGetUplobdByIDFunc{
			defbultHook: i.GetUplobdByID,
		},
		ReferencesForUplobdFunc: &UplobdServiceReferencesForUplobdFunc{
			defbultHook: i.ReferencesForUplobd,
		},
	}
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

// MockPbckbgeReferenceScbnner is b mock implementbtion of the
// PbckbgeReferenceScbnner interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred)
// used for unit testing.
type MockPbckbgeReferenceScbnner struct {
	// CloseFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Close.
	CloseFunc *PbckbgeReferenceScbnnerCloseFunc
	// NextFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Next.
	NextFunc *PbckbgeReferenceScbnnerNextFunc
}

// NewMockPbckbgeReferenceScbnner crebtes b new mock of the
// PbckbgeReferenceScbnner interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockPbckbgeReferenceScbnner() *MockPbckbgeReferenceScbnner {
	return &MockPbckbgeReferenceScbnner{
		CloseFunc: &PbckbgeReferenceScbnnerCloseFunc{
			defbultHook: func() (r0 error) {
				return
			},
		},
		NextFunc: &PbckbgeReferenceScbnnerNextFunc{
			defbultHook: func() (r0 shbred1.PbckbgeReference, r1 bool, r2 error) {
				return
			},
		},
	}
}

// NewStrictMockPbckbgeReferenceScbnner crebtes b new mock of the
// PbckbgeReferenceScbnner interfbce. All methods pbnic on invocbtion,
// unless overwritten.
func NewStrictMockPbckbgeReferenceScbnner() *MockPbckbgeReferenceScbnner {
	return &MockPbckbgeReferenceScbnner{
		CloseFunc: &PbckbgeReferenceScbnnerCloseFunc{
			defbultHook: func() error {
				pbnic("unexpected invocbtion of MockPbckbgeReferenceScbnner.Close")
			},
		},
		NextFunc: &PbckbgeReferenceScbnnerNextFunc{
			defbultHook: func() (shbred1.PbckbgeReference, bool, error) {
				pbnic("unexpected invocbtion of MockPbckbgeReferenceScbnner.Next")
			},
		},
	}
}

// NewMockPbckbgeReferenceScbnnerFrom crebtes b new mock of the
// MockPbckbgeReferenceScbnner interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockPbckbgeReferenceScbnnerFrom(i shbred1.PbckbgeReferenceScbnner) *MockPbckbgeReferenceScbnner {
	return &MockPbckbgeReferenceScbnner{
		CloseFunc: &PbckbgeReferenceScbnnerCloseFunc{
			defbultHook: i.Close,
		},
		NextFunc: &PbckbgeReferenceScbnnerNextFunc{
			defbultHook: i.Next,
		},
	}
}

// PbckbgeReferenceScbnnerCloseFunc describes the behbvior when the Close
// method of the pbrent MockPbckbgeReferenceScbnner instbnce is invoked.
type PbckbgeReferenceScbnnerCloseFunc struct {
	defbultHook func() error
	hooks       []func() error
	history     []PbckbgeReferenceScbnnerCloseFuncCbll
	mutex       sync.Mutex
}

// Close delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockPbckbgeReferenceScbnner) Close() error {
	r0 := m.CloseFunc.nextHook()()
	m.CloseFunc.bppendCbll(PbckbgeReferenceScbnnerCloseFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Close method of the
// pbrent MockPbckbgeReferenceScbnner instbnce is invoked bnd the hook queue
// is empty.
func (f *PbckbgeReferenceScbnnerCloseFunc) SetDefbultHook(hook func() error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Close method of the pbrent MockPbckbgeReferenceScbnner instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *PbckbgeReferenceScbnnerCloseFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *PbckbgeReferenceScbnnerCloseFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func() error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *PbckbgeReferenceScbnnerCloseFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *PbckbgeReferenceScbnnerCloseFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *PbckbgeReferenceScbnnerCloseFunc) bppendCbll(r0 PbckbgeReferenceScbnnerCloseFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of PbckbgeReferenceScbnnerCloseFuncCbll
// objects describing the invocbtions of this function.
func (f *PbckbgeReferenceScbnnerCloseFunc) History() []PbckbgeReferenceScbnnerCloseFuncCbll {
	f.mutex.Lock()
	history := mbke([]PbckbgeReferenceScbnnerCloseFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// PbckbgeReferenceScbnnerCloseFuncCbll is bn object thbt describes bn
// invocbtion of method Close on bn instbnce of MockPbckbgeReferenceScbnner.
type PbckbgeReferenceScbnnerCloseFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c PbckbgeReferenceScbnnerCloseFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c PbckbgeReferenceScbnnerCloseFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// PbckbgeReferenceScbnnerNextFunc describes the behbvior when the Next
// method of the pbrent MockPbckbgeReferenceScbnner instbnce is invoked.
type PbckbgeReferenceScbnnerNextFunc struct {
	defbultHook func() (shbred1.PbckbgeReference, bool, error)
	hooks       []func() (shbred1.PbckbgeReference, bool, error)
	history     []PbckbgeReferenceScbnnerNextFuncCbll
	mutex       sync.Mutex
}

// Next delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockPbckbgeReferenceScbnner) Next() (shbred1.PbckbgeReference, bool, error) {
	r0, r1, r2 := m.NextFunc.nextHook()()
	m.NextFunc.bppendCbll(PbckbgeReferenceScbnnerNextFuncCbll{r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Next method of the
// pbrent MockPbckbgeReferenceScbnner instbnce is invoked bnd the hook queue
// is empty.
func (f *PbckbgeReferenceScbnnerNextFunc) SetDefbultHook(hook func() (shbred1.PbckbgeReference, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Next method of the pbrent MockPbckbgeReferenceScbnner instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *PbckbgeReferenceScbnnerNextFunc) PushHook(hook func() (shbred1.PbckbgeReference, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *PbckbgeReferenceScbnnerNextFunc) SetDefbultReturn(r0 shbred1.PbckbgeReference, r1 bool, r2 error) {
	f.SetDefbultHook(func() (shbred1.PbckbgeReference, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *PbckbgeReferenceScbnnerNextFunc) PushReturn(r0 shbred1.PbckbgeReference, r1 bool, r2 error) {
	f.PushHook(func() (shbred1.PbckbgeReference, bool, error) {
		return r0, r1, r2
	})
}

func (f *PbckbgeReferenceScbnnerNextFunc) nextHook() func() (shbred1.PbckbgeReference, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *PbckbgeReferenceScbnnerNextFunc) bppendCbll(r0 PbckbgeReferenceScbnnerNextFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of PbckbgeReferenceScbnnerNextFuncCbll objects
// describing the invocbtions of this function.
func (f *PbckbgeReferenceScbnnerNextFunc) History() []PbckbgeReferenceScbnnerNextFuncCbll {
	f.mutex.Lock()
	history := mbke([]PbckbgeReferenceScbnnerNextFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// PbckbgeReferenceScbnnerNextFuncCbll is bn object thbt describes bn
// invocbtion of method Next on bn instbnce of MockPbckbgeReferenceScbnner.
type PbckbgeReferenceScbnnerNextFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred1.PbckbgeReference
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c PbckbgeReferenceScbnnerNextFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c PbckbgeReferenceScbnnerNextFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

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
			defbultHook: func(context.Context, int) (r0 shbred2.IndexConfigurbtion, r1 bool, r2 error) {
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
			defbultHook: func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error) {
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
	defbultHook func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error)
	hooks       []func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error)
	history     []StoreGetIndexConfigurbtionByRepositoryIDFuncCbll
	mutex       sync.Mutex
}

// GetIndexConfigurbtionByRepositoryID delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockStore) GetIndexConfigurbtionByRepositoryID(v0 context.Context, v1 int) (shbred2.IndexConfigurbtion, bool, error) {
	r0, r1, r2 := m.GetIndexConfigurbtionByRepositoryIDFunc.nextHook()(v0, v1)
	m.GetIndexConfigurbtionByRepositoryIDFunc.bppendCbll(StoreGetIndexConfigurbtionByRepositoryIDFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce is invoked bnd the hook queue is empty.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) SetDefbultHook(hook func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetIndexConfigurbtionByRepositoryID method of the pbrent MockStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) PushHook(hook func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) SetDefbultReturn(r0 shbred2.IndexConfigurbtion, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) PushReturn(r0 shbred2.IndexConfigurbtion, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreGetIndexConfigurbtionByRepositoryIDFunc) nextHook() func(context.Context, int) (shbred2.IndexConfigurbtion, bool, error) {
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
	Result0 shbred2.IndexConfigurbtion
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

// MockWorkerStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store)
// used for unit testing.
type MockWorkerStore[T workerutil.Record] struct {
	// AddExecutionLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AddExecutionLogEntry.
	AddExecutionLogEntryFunc *WorkerStoreAddExecutionLogEntryFunc[T]
	// DequeueFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Dequeue.
	DequeueFunc *WorkerStoreDequeueFunc[T]
	// HbndleFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hbndle.
	HbndleFunc *WorkerStoreHbndleFunc[T]
	// HebrtbebtFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method Hebrtbebt.
	HebrtbebtFunc *WorkerStoreHebrtbebtFunc[T]
	// MbrkCompleteFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkComplete.
	MbrkCompleteFunc *WorkerStoreMbrkCompleteFunc[T]
	// MbrkErroredFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkErrored.
	MbrkErroredFunc *WorkerStoreMbrkErroredFunc[T]
	// MbrkFbiledFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkFbiled.
	MbrkFbiledFunc *WorkerStoreMbrkFbiledFunc[T]
	// MbxDurbtionInQueueFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method MbxDurbtionInQueue.
	MbxDurbtionInQueueFunc *WorkerStoreMbxDurbtionInQueueFunc[T]
	// QueuedCountFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method QueuedCount.
	QueuedCountFunc *WorkerStoreQueuedCountFunc[T]
	// RequeueFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Requeue.
	RequeueFunc *WorkerStoreRequeueFunc[T]
	// ResetStblledFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ResetStblled.
	ResetStblledFunc *WorkerStoreResetStblledFunc[T]
	// UpdbteExecutionLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteExecutionLogEntry.
	UpdbteExecutionLogEntryFunc *WorkerStoreUpdbteExecutionLogEntryFunc[T]
	// WithFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method With.
	WithFunc *WorkerStoreWithFunc[T]
}

// NewMockWorkerStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockWorkerStore[T workerutil.Record]() *MockWorkerStore[T] {
	return &MockWorkerStore[T]{
		AddExecutionLogEntryFunc: &WorkerStoreAddExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (r0 int, r1 error) {
				return
			},
		},
		DequeueFunc: &WorkerStoreDequeueFunc[T]{
			defbultHook: func(context.Context, string, []*sqlf.Query) (r0 T, r1 bool, r2 error) {
				return
			},
		},
		HbndleFunc: &WorkerStoreHbndleFunc[T]{
			defbultHook: func() (r0 bbsestore.TrbnsbctbbleHbndle) {
				return
			},
		},
		HebrtbebtFunc: &WorkerStoreHebrtbebtFunc[T]{
			defbultHook: func(context.Context, []string, store1.HebrtbebtOptions) (r0 []string, r1 []string, r2 error) {
				return
			},
		},
		MbrkCompleteFunc: &WorkerStoreMbrkCompleteFunc[T]{
			defbultHook: func(context.Context, int, store1.MbrkFinblOptions) (r0 bool, r1 error) {
				return
			},
		},
		MbrkErroredFunc: &WorkerStoreMbrkErroredFunc[T]{
			defbultHook: func(context.Context, int, string, store1.MbrkFinblOptions) (r0 bool, r1 error) {
				return
			},
		},
		MbrkFbiledFunc: &WorkerStoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, int, string, store1.MbrkFinblOptions) (r0 bool, r1 error) {
				return
			},
		},
		MbxDurbtionInQueueFunc: &WorkerStoreMbxDurbtionInQueueFunc[T]{
			defbultHook: func(context.Context) (r0 time.Durbtion, r1 error) {
				return
			},
		},
		QueuedCountFunc: &WorkerStoreQueuedCountFunc[T]{
			defbultHook: func(context.Context, bool) (r0 int, r1 error) {
				return
			},
		},
		RequeueFunc: &WorkerStoreRequeueFunc[T]{
			defbultHook: func(context.Context, int, time.Time) (r0 error) {
				return
			},
		},
		ResetStblledFunc: &WorkerStoreResetStblledFunc[T]{
			defbultHook: func(context.Context) (r0 mbp[int]time.Durbtion, r1 mbp[int]time.Durbtion, r2 error) {
				return
			},
		},
		UpdbteExecutionLogEntryFunc: &WorkerStoreUpdbteExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (r0 error) {
				return
			},
		},
		WithFunc: &WorkerStoreWithFunc[T]{
			defbultHook: func(bbsestore.ShbrebbleStore) (r0 store1.Store[T]) {
				return
			},
		},
	}
}

// NewStrictMockWorkerStore crebtes b new mock of the Store interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockWorkerStore[T workerutil.Record]() *MockWorkerStore[T] {
	return &MockWorkerStore[T]{
		AddExecutionLogEntryFunc: &WorkerStoreAddExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.AddExecutionLogEntry")
			},
		},
		DequeueFunc: &WorkerStoreDequeueFunc[T]{
			defbultHook: func(context.Context, string, []*sqlf.Query) (T, bool, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.Dequeue")
			},
		},
		HbndleFunc: &WorkerStoreHbndleFunc[T]{
			defbultHook: func() bbsestore.TrbnsbctbbleHbndle {
				pbnic("unexpected invocbtion of MockWorkerStore.Hbndle")
			},
		},
		HebrtbebtFunc: &WorkerStoreHebrtbebtFunc[T]{
			defbultHook: func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.Hebrtbebt")
			},
		},
		MbrkCompleteFunc: &WorkerStoreMbrkCompleteFunc[T]{
			defbultHook: func(context.Context, int, store1.MbrkFinblOptions) (bool, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.MbrkComplete")
			},
		},
		MbrkErroredFunc: &WorkerStoreMbrkErroredFunc[T]{
			defbultHook: func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.MbrkErrored")
			},
		},
		MbrkFbiledFunc: &WorkerStoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.MbrkFbiled")
			},
		},
		MbxDurbtionInQueueFunc: &WorkerStoreMbxDurbtionInQueueFunc[T]{
			defbultHook: func(context.Context) (time.Durbtion, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.MbxDurbtionInQueue")
			},
		},
		QueuedCountFunc: &WorkerStoreQueuedCountFunc[T]{
			defbultHook: func(context.Context, bool) (int, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.QueuedCount")
			},
		},
		RequeueFunc: &WorkerStoreRequeueFunc[T]{
			defbultHook: func(context.Context, int, time.Time) error {
				pbnic("unexpected invocbtion of MockWorkerStore.Requeue")
			},
		},
		ResetStblledFunc: &WorkerStoreResetStblledFunc[T]{
			defbultHook: func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
				pbnic("unexpected invocbtion of MockWorkerStore.ResetStblled")
			},
		},
		UpdbteExecutionLogEntryFunc: &WorkerStoreUpdbteExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error {
				pbnic("unexpected invocbtion of MockWorkerStore.UpdbteExecutionLogEntry")
			},
		},
		WithFunc: &WorkerStoreWithFunc[T]{
			defbultHook: func(bbsestore.ShbrebbleStore) store1.Store[T] {
				pbnic("unexpected invocbtion of MockWorkerStore.With")
			},
		},
	}
}

// NewMockWorkerStoreFrom crebtes b new mock of the MockWorkerStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockWorkerStoreFrom[T workerutil.Record](i store1.Store[T]) *MockWorkerStore[T] {
	return &MockWorkerStore[T]{
		AddExecutionLogEntryFunc: &WorkerStoreAddExecutionLogEntryFunc[T]{
			defbultHook: i.AddExecutionLogEntry,
		},
		DequeueFunc: &WorkerStoreDequeueFunc[T]{
			defbultHook: i.Dequeue,
		},
		HbndleFunc: &WorkerStoreHbndleFunc[T]{
			defbultHook: i.Hbndle,
		},
		HebrtbebtFunc: &WorkerStoreHebrtbebtFunc[T]{
			defbultHook: i.Hebrtbebt,
		},
		MbrkCompleteFunc: &WorkerStoreMbrkCompleteFunc[T]{
			defbultHook: i.MbrkComplete,
		},
		MbrkErroredFunc: &WorkerStoreMbrkErroredFunc[T]{
			defbultHook: i.MbrkErrored,
		},
		MbrkFbiledFunc: &WorkerStoreMbrkFbiledFunc[T]{
			defbultHook: i.MbrkFbiled,
		},
		MbxDurbtionInQueueFunc: &WorkerStoreMbxDurbtionInQueueFunc[T]{
			defbultHook: i.MbxDurbtionInQueue,
		},
		QueuedCountFunc: &WorkerStoreQueuedCountFunc[T]{
			defbultHook: i.QueuedCount,
		},
		RequeueFunc: &WorkerStoreRequeueFunc[T]{
			defbultHook: i.Requeue,
		},
		ResetStblledFunc: &WorkerStoreResetStblledFunc[T]{
			defbultHook: i.ResetStblled,
		},
		UpdbteExecutionLogEntryFunc: &WorkerStoreUpdbteExecutionLogEntryFunc[T]{
			defbultHook: i.UpdbteExecutionLogEntry,
		},
		WithFunc: &WorkerStoreWithFunc[T]{
			defbultHook: i.With,
		},
	}
}

// WorkerStoreAddExecutionLogEntryFunc describes the behbvior when the
// AddExecutionLogEntry method of the pbrent MockWorkerStore instbnce is
// invoked.
type WorkerStoreAddExecutionLogEntryFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error)
	hooks       []func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error)
	history     []WorkerStoreAddExecutionLogEntryFuncCbll[T]
	mutex       sync.Mutex
}

// AddExecutionLogEntry delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) AddExecutionLogEntry(v0 context.Context, v1 int, v2 executor.ExecutionLogEntry, v3 store1.ExecutionLogEntryOptions) (int, error) {
	r0, r1 := m.AddExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3)
	m.AddExecutionLogEntryFunc.bppendCbll(WorkerStoreAddExecutionLogEntryFuncCbll[T]{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AddExecutionLogEntry
// method of the pbrent MockWorkerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *WorkerStoreAddExecutionLogEntryFunc[T]) SetDefbultHook(hook func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddExecutionLogEntry method of the pbrent MockWorkerStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *WorkerStoreAddExecutionLogEntryFunc[T]) PushHook(hook func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreAddExecutionLogEntryFunc[T]) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreAddExecutionLogEntryFunc[T]) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error) {
		return r0, r1
	})
}

func (f *WorkerStoreAddExecutionLogEntryFunc[T]) nextHook() func(context.Context, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreAddExecutionLogEntryFunc[T]) bppendCbll(r0 WorkerStoreAddExecutionLogEntryFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreAddExecutionLogEntryFuncCbll
// objects describing the invocbtions of this function.
func (f *WorkerStoreAddExecutionLogEntryFunc[T]) History() []WorkerStoreAddExecutionLogEntryFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreAddExecutionLogEntryFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreAddExecutionLogEntryFuncCbll is bn object thbt describes bn
// invocbtion of method AddExecutionLogEntry on bn instbnce of
// MockWorkerStore.
type WorkerStoreAddExecutionLogEntryFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 executor.ExecutionLogEntry
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 store1.ExecutionLogEntryOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreAddExecutionLogEntryFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreAddExecutionLogEntryFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// WorkerStoreDequeueFunc describes the behbvior when the Dequeue method of
// the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreDequeueFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, string, []*sqlf.Query) (T, bool, error)
	hooks       []func(context.Context, string, []*sqlf.Query) (T, bool, error)
	history     []WorkerStoreDequeueFuncCbll[T]
	mutex       sync.Mutex
}

// Dequeue delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) Dequeue(v0 context.Context, v1 string, v2 []*sqlf.Query) (T, bool, error) {
	r0, r1, r2 := m.DequeueFunc.nextHook()(v0, v1, v2)
	m.DequeueFunc.bppendCbll(WorkerStoreDequeueFuncCbll[T]{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Dequeue method of
// the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreDequeueFunc[T]) SetDefbultHook(hook func(context.Context, string, []*sqlf.Query) (T, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Dequeue method of the pbrent MockWorkerStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkerStoreDequeueFunc[T]) PushHook(hook func(context.Context, string, []*sqlf.Query) (T, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreDequeueFunc[T]) SetDefbultReturn(r0 T, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, []*sqlf.Query) (T, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreDequeueFunc[T]) PushReturn(r0 T, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, []*sqlf.Query) (T, bool, error) {
		return r0, r1, r2
	})
}

func (f *WorkerStoreDequeueFunc[T]) nextHook() func(context.Context, string, []*sqlf.Query) (T, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreDequeueFunc[T]) bppendCbll(r0 WorkerStoreDequeueFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreDequeueFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreDequeueFunc[T]) History() []WorkerStoreDequeueFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreDequeueFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreDequeueFuncCbll is bn object thbt describes bn invocbtion of
// method Dequeue on bn instbnce of MockWorkerStore.
type WorkerStoreDequeueFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []*sqlf.Query
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 T
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreDequeueFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreDequeueFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// WorkerStoreHbndleFunc describes the behbvior when the Hbndle method of
// the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreHbndleFunc[T workerutil.Record] struct {
	defbultHook func() bbsestore.TrbnsbctbbleHbndle
	hooks       []func() bbsestore.TrbnsbctbbleHbndle
	history     []WorkerStoreHbndleFuncCbll[T]
	mutex       sync.Mutex
}

// Hbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) Hbndle() bbsestore.TrbnsbctbbleHbndle {
	r0 := m.HbndleFunc.nextHook()()
	m.HbndleFunc.bppendCbll(WorkerStoreHbndleFuncCbll[T]{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Hbndle method of the
// pbrent MockWorkerStore instbnce is invoked bnd the hook queue is empty.
func (f *WorkerStoreHbndleFunc[T]) SetDefbultHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hbndle method of the pbrent MockWorkerStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkerStoreHbndleFunc[T]) PushHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreHbndleFunc[T]) SetDefbultReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.SetDefbultHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreHbndleFunc[T]) PushReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.PushHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

func (f *WorkerStoreHbndleFunc[T]) nextHook() func() bbsestore.TrbnsbctbbleHbndle {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreHbndleFunc[T]) bppendCbll(r0 WorkerStoreHbndleFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreHbndleFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreHbndleFunc[T]) History() []WorkerStoreHbndleFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreHbndleFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreHbndleFuncCbll is bn object thbt describes bn invocbtion of
// method Hbndle on bn instbnce of MockWorkerStore.
type WorkerStoreHbndleFuncCbll[T workerutil.Record] struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bbsestore.TrbnsbctbbleHbndle
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreHbndleFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreHbndleFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// WorkerStoreHebrtbebtFunc describes the behbvior when the Hebrtbebt method
// of the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreHebrtbebtFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error)
	hooks       []func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error)
	history     []WorkerStoreHebrtbebtFuncCbll[T]
	mutex       sync.Mutex
}

// Hebrtbebt delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) Hebrtbebt(v0 context.Context, v1 []string, v2 store1.HebrtbebtOptions) ([]string, []string, error) {
	r0, r1, r2 := m.HebrtbebtFunc.nextHook()(v0, v1, v2)
	m.HebrtbebtFunc.bppendCbll(WorkerStoreHebrtbebtFuncCbll[T]{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Hebrtbebt method of
// the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreHebrtbebtFunc[T]) SetDefbultHook(hook func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hebrtbebt method of the pbrent MockWorkerStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkerStoreHebrtbebtFunc[T]) PushHook(hook func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreHebrtbebtFunc[T]) SetDefbultReturn(r0 []string, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreHebrtbebtFunc[T]) PushReturn(r0 []string, r1 []string, r2 error) {
	f.PushHook(func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error) {
		return r0, r1, r2
	})
}

func (f *WorkerStoreHebrtbebtFunc[T]) nextHook() func(context.Context, []string, store1.HebrtbebtOptions) ([]string, []string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreHebrtbebtFunc[T]) bppendCbll(r0 WorkerStoreHebrtbebtFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreHebrtbebtFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreHebrtbebtFunc[T]) History() []WorkerStoreHebrtbebtFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreHebrtbebtFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreHebrtbebtFuncCbll is bn object thbt describes bn invocbtion of
// method Hebrtbebt on bn instbnce of MockWorkerStore.
type WorkerStoreHebrtbebtFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 store1.HebrtbebtOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreHebrtbebtFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreHebrtbebtFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// WorkerStoreMbrkCompleteFunc describes the behbvior when the MbrkComplete
// method of the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreMbrkCompleteFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, store1.MbrkFinblOptions) (bool, error)
	hooks       []func(context.Context, int, store1.MbrkFinblOptions) (bool, error)
	history     []WorkerStoreMbrkCompleteFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkComplete delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) MbrkComplete(v0 context.Context, v1 int, v2 store1.MbrkFinblOptions) (bool, error) {
	r0, r1 := m.MbrkCompleteFunc.nextHook()(v0, v1, v2)
	m.MbrkCompleteFunc.bppendCbll(WorkerStoreMbrkCompleteFuncCbll[T]{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkComplete method
// of the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreMbrkCompleteFunc[T]) SetDefbultHook(hook func(context.Context, int, store1.MbrkFinblOptions) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkComplete method of the pbrent MockWorkerStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WorkerStoreMbrkCompleteFunc[T]) PushHook(hook func(context.Context, int, store1.MbrkFinblOptions) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreMbrkCompleteFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, store1.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreMbrkCompleteFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, store1.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

func (f *WorkerStoreMbrkCompleteFunc[T]) nextHook() func(context.Context, int, store1.MbrkFinblOptions) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreMbrkCompleteFunc[T]) bppendCbll(r0 WorkerStoreMbrkCompleteFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreMbrkCompleteFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreMbrkCompleteFunc[T]) History() []WorkerStoreMbrkCompleteFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreMbrkCompleteFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreMbrkCompleteFuncCbll is bn object thbt describes bn invocbtion
// of method MbrkComplete on bn instbnce of MockWorkerStore.
type WorkerStoreMbrkCompleteFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 store1.MbrkFinblOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreMbrkCompleteFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreMbrkCompleteFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// WorkerStoreMbrkErroredFunc describes the behbvior when the MbrkErrored
// method of the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreMbrkErroredFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)
	hooks       []func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)
	history     []WorkerStoreMbrkErroredFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkErrored delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) MbrkErrored(v0 context.Context, v1 int, v2 string, v3 store1.MbrkFinblOptions) (bool, error) {
	r0, r1 := m.MbrkErroredFunc.nextHook()(v0, v1, v2, v3)
	m.MbrkErroredFunc.bppendCbll(WorkerStoreMbrkErroredFuncCbll[T]{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkErrored method
// of the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreMbrkErroredFunc[T]) SetDefbultHook(hook func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkErrored method of the pbrent MockWorkerStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WorkerStoreMbrkErroredFunc[T]) PushHook(hook func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreMbrkErroredFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreMbrkErroredFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

func (f *WorkerStoreMbrkErroredFunc[T]) nextHook() func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreMbrkErroredFunc[T]) bppendCbll(r0 WorkerStoreMbrkErroredFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreMbrkErroredFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreMbrkErroredFunc[T]) History() []WorkerStoreMbrkErroredFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreMbrkErroredFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreMbrkErroredFuncCbll is bn object thbt describes bn invocbtion
// of method MbrkErrored on bn instbnce of MockWorkerStore.
type WorkerStoreMbrkErroredFuncCbll[T workerutil.Record] struct {
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
	Arg3 store1.MbrkFinblOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreMbrkErroredFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreMbrkErroredFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// WorkerStoreMbrkFbiledFunc describes the behbvior when the MbrkFbiled
// method of the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreMbrkFbiledFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)
	hooks       []func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)
	history     []WorkerStoreMbrkFbiledFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkFbiled delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) MbrkFbiled(v0 context.Context, v1 int, v2 string, v3 store1.MbrkFinblOptions) (bool, error) {
	r0, r1 := m.MbrkFbiledFunc.nextHook()(v0, v1, v2, v3)
	m.MbrkFbiledFunc.bppendCbll(WorkerStoreMbrkFbiledFuncCbll[T]{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkFbiled method of
// the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreMbrkFbiledFunc[T]) SetDefbultHook(hook func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkFbiled method of the pbrent MockWorkerStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkerStoreMbrkFbiledFunc[T]) PushHook(hook func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreMbrkFbiledFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreMbrkFbiledFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

func (f *WorkerStoreMbrkFbiledFunc[T]) nextHook() func(context.Context, int, string, store1.MbrkFinblOptions) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreMbrkFbiledFunc[T]) bppendCbll(r0 WorkerStoreMbrkFbiledFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreMbrkFbiledFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreMbrkFbiledFunc[T]) History() []WorkerStoreMbrkFbiledFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreMbrkFbiledFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreMbrkFbiledFuncCbll is bn object thbt describes bn invocbtion
// of method MbrkFbiled on bn instbnce of MockWorkerStore.
type WorkerStoreMbrkFbiledFuncCbll[T workerutil.Record] struct {
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
	Arg3 store1.MbrkFinblOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreMbrkFbiledFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreMbrkFbiledFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// WorkerStoreMbxDurbtionInQueueFunc describes the behbvior when the
// MbxDurbtionInQueue method of the pbrent MockWorkerStore instbnce is
// invoked.
type WorkerStoreMbxDurbtionInQueueFunc[T workerutil.Record] struct {
	defbultHook func(context.Context) (time.Durbtion, error)
	hooks       []func(context.Context) (time.Durbtion, error)
	history     []WorkerStoreMbxDurbtionInQueueFuncCbll[T]
	mutex       sync.Mutex
}

// MbxDurbtionInQueue delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) MbxDurbtionInQueue(v0 context.Context) (time.Durbtion, error) {
	r0, r1 := m.MbxDurbtionInQueueFunc.nextHook()(v0)
	m.MbxDurbtionInQueueFunc.bppendCbll(WorkerStoreMbxDurbtionInQueueFuncCbll[T]{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbxDurbtionInQueue
// method of the pbrent MockWorkerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) SetDefbultHook(hook func(context.Context) (time.Durbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbxDurbtionInQueue method of the pbrent MockWorkerStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) PushHook(hook func(context.Context) (time.Durbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) SetDefbultReturn(r0 time.Durbtion, r1 error) {
	f.SetDefbultHook(func(context.Context) (time.Durbtion, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) PushReturn(r0 time.Durbtion, r1 error) {
	f.PushHook(func(context.Context) (time.Durbtion, error) {
		return r0, r1
	})
}

func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) nextHook() func(context.Context) (time.Durbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) bppendCbll(r0 WorkerStoreMbxDurbtionInQueueFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreMbxDurbtionInQueueFuncCbll
// objects describing the invocbtions of this function.
func (f *WorkerStoreMbxDurbtionInQueueFunc[T]) History() []WorkerStoreMbxDurbtionInQueueFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreMbxDurbtionInQueueFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreMbxDurbtionInQueueFuncCbll is bn object thbt describes bn
// invocbtion of method MbxDurbtionInQueue on bn instbnce of
// MockWorkerStore.
type WorkerStoreMbxDurbtionInQueueFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 time.Durbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreMbxDurbtionInQueueFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreMbxDurbtionInQueueFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// WorkerStoreQueuedCountFunc describes the behbvior when the QueuedCount
// method of the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreQueuedCountFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, bool) (int, error)
	hooks       []func(context.Context, bool) (int, error)
	history     []WorkerStoreQueuedCountFuncCbll[T]
	mutex       sync.Mutex
}

// QueuedCount delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) QueuedCount(v0 context.Context, v1 bool) (int, error) {
	r0, r1 := m.QueuedCountFunc.nextHook()(v0, v1)
	m.QueuedCountFunc.bppendCbll(WorkerStoreQueuedCountFuncCbll[T]{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the QueuedCount method
// of the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreQueuedCountFunc[T]) SetDefbultHook(hook func(context.Context, bool) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueuedCount method of the pbrent MockWorkerStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WorkerStoreQueuedCountFunc[T]) PushHook(hook func(context.Context, bool) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreQueuedCountFunc[T]) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, bool) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreQueuedCountFunc[T]) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, bool) (int, error) {
		return r0, r1
	})
}

func (f *WorkerStoreQueuedCountFunc[T]) nextHook() func(context.Context, bool) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreQueuedCountFunc[T]) bppendCbll(r0 WorkerStoreQueuedCountFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreQueuedCountFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreQueuedCountFunc[T]) History() []WorkerStoreQueuedCountFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreQueuedCountFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreQueuedCountFuncCbll is bn object thbt describes bn invocbtion
// of method QueuedCount on bn instbnce of MockWorkerStore.
type WorkerStoreQueuedCountFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreQueuedCountFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreQueuedCountFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// WorkerStoreRequeueFunc describes the behbvior when the Requeue method of
// the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreRequeueFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, time.Time) error
	hooks       []func(context.Context, int, time.Time) error
	history     []WorkerStoreRequeueFuncCbll[T]
	mutex       sync.Mutex
}

// Requeue delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) Requeue(v0 context.Context, v1 int, v2 time.Time) error {
	r0 := m.RequeueFunc.nextHook()(v0, v1, v2)
	m.RequeueFunc.bppendCbll(WorkerStoreRequeueFuncCbll[T]{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Requeue method of
// the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreRequeueFunc[T]) SetDefbultHook(hook func(context.Context, int, time.Time) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Requeue method of the pbrent MockWorkerStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkerStoreRequeueFunc[T]) PushHook(hook func(context.Context, int, time.Time) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreRequeueFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, time.Time) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreRequeueFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, time.Time) error {
		return r0
	})
}

func (f *WorkerStoreRequeueFunc[T]) nextHook() func(context.Context, int, time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreRequeueFunc[T]) bppendCbll(r0 WorkerStoreRequeueFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreRequeueFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreRequeueFunc[T]) History() []WorkerStoreRequeueFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreRequeueFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreRequeueFuncCbll is bn object thbt describes bn invocbtion of
// method Requeue on bn instbnce of MockWorkerStore.
type WorkerStoreRequeueFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreRequeueFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreRequeueFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// WorkerStoreResetStblledFunc describes the behbvior when the ResetStblled
// method of the pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreResetStblledFunc[T workerutil.Record] struct {
	defbultHook func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)
	hooks       []func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)
	history     []WorkerStoreResetStblledFuncCbll[T]
	mutex       sync.Mutex
}

// ResetStblled delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) ResetStblled(v0 context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
	r0, r1, r2 := m.ResetStblledFunc.nextHook()(v0)
	m.ResetStblledFunc.bppendCbll(WorkerStoreResetStblledFuncCbll[T]{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ResetStblled method
// of the pbrent MockWorkerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *WorkerStoreResetStblledFunc[T]) SetDefbultHook(hook func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResetStblled method of the pbrent MockWorkerStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WorkerStoreResetStblledFunc[T]) PushHook(hook func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreResetStblledFunc[T]) SetDefbultReturn(r0 mbp[int]time.Durbtion, r1 mbp[int]time.Durbtion, r2 error) {
	f.SetDefbultHook(func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreResetStblledFunc[T]) PushReturn(r0 mbp[int]time.Durbtion, r1 mbp[int]time.Durbtion, r2 error) {
	f.PushHook(func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
		return r0, r1, r2
	})
}

func (f *WorkerStoreResetStblledFunc[T]) nextHook() func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreResetStblledFunc[T]) bppendCbll(r0 WorkerStoreResetStblledFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreResetStblledFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkerStoreResetStblledFunc[T]) History() []WorkerStoreResetStblledFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreResetStblledFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreResetStblledFuncCbll is bn object thbt describes bn invocbtion
// of method ResetStblled on bn instbnce of MockWorkerStore.
type WorkerStoreResetStblledFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[int]time.Durbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 mbp[int]time.Durbtion
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreResetStblledFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreResetStblledFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// WorkerStoreUpdbteExecutionLogEntryFunc describes the behbvior when the
// UpdbteExecutionLogEntry method of the pbrent MockWorkerStore instbnce is
// invoked.
type WorkerStoreUpdbteExecutionLogEntryFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error
	hooks       []func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error
	history     []WorkerStoreUpdbteExecutionLogEntryFuncCbll[T]
	mutex       sync.Mutex
}

// UpdbteExecutionLogEntry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) UpdbteExecutionLogEntry(v0 context.Context, v1 int, v2 int, v3 executor.ExecutionLogEntry, v4 store1.ExecutionLogEntryOptions) error {
	r0 := m.UpdbteExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3, v4)
	m.UpdbteExecutionLogEntryFunc.bppendCbll(WorkerStoreUpdbteExecutionLogEntryFuncCbll[T]{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteExecutionLogEntry method of the pbrent MockWorkerStore instbnce is
// invoked bnd the hook queue is empty.
func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) SetDefbultHook(hook func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteExecutionLogEntry method of the pbrent MockWorkerStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) PushHook(hook func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error {
		return r0
	})
}

func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) nextHook() func(context.Context, int, int, executor.ExecutionLogEntry, store1.ExecutionLogEntryOptions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) bppendCbll(r0 WorkerStoreUpdbteExecutionLogEntryFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreUpdbteExecutionLogEntryFuncCbll
// objects describing the invocbtions of this function.
func (f *WorkerStoreUpdbteExecutionLogEntryFunc[T]) History() []WorkerStoreUpdbteExecutionLogEntryFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreUpdbteExecutionLogEntryFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreUpdbteExecutionLogEntryFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteExecutionLogEntry on bn instbnce of
// MockWorkerStore.
type WorkerStoreUpdbteExecutionLogEntryFuncCbll[T workerutil.Record] struct {
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
	Arg3 executor.ExecutionLogEntry
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 store1.ExecutionLogEntryOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreUpdbteExecutionLogEntryFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreUpdbteExecutionLogEntryFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// WorkerStoreWithFunc describes the behbvior when the With method of the
// pbrent MockWorkerStore instbnce is invoked.
type WorkerStoreWithFunc[T workerutil.Record] struct {
	defbultHook func(bbsestore.ShbrebbleStore) store1.Store[T]
	hooks       []func(bbsestore.ShbrebbleStore) store1.Store[T]
	history     []WorkerStoreWithFuncCbll[T]
	mutex       sync.Mutex
}

// With delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkerStore[T]) With(v0 bbsestore.ShbrebbleStore) store1.Store[T] {
	r0 := m.WithFunc.nextHook()(v0)
	m.WithFunc.bppendCbll(WorkerStoreWithFuncCbll[T]{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the With method of the
// pbrent MockWorkerStore instbnce is invoked bnd the hook queue is empty.
func (f *WorkerStoreWithFunc[T]) SetDefbultHook(hook func(bbsestore.ShbrebbleStore) store1.Store[T]) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// With method of the pbrent MockWorkerStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkerStoreWithFunc[T]) PushHook(hook func(bbsestore.ShbrebbleStore) store1.Store[T]) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkerStoreWithFunc[T]) SetDefbultReturn(r0 store1.Store[T]) {
	f.SetDefbultHook(func(bbsestore.ShbrebbleStore) store1.Store[T] {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkerStoreWithFunc[T]) PushReturn(r0 store1.Store[T]) {
	f.PushHook(func(bbsestore.ShbrebbleStore) store1.Store[T] {
		return r0
	})
}

func (f *WorkerStoreWithFunc[T]) nextHook() func(bbsestore.ShbrebbleStore) store1.Store[T] {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkerStoreWithFunc[T]) bppendCbll(r0 WorkerStoreWithFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkerStoreWithFuncCbll objects describing
// the invocbtions of this function.
func (f *WorkerStoreWithFunc[T]) History() []WorkerStoreWithFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WorkerStoreWithFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkerStoreWithFuncCbll is bn object thbt describes bn invocbtion of
// method With on bn instbnce of MockWorkerStore.
type WorkerStoreWithFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 bbsestore.ShbrebbleStore
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 store1.Store[T]
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkerStoreWithFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkerStoreWithFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
