// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge bbckground

import (
	"context"
	"sync"

	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// MockAutoIndexingService is b mock implementbtion of the
// AutoIndexingService interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/internbl/bbckground)
// used for unit testing.
type MockAutoIndexingService struct {
	// QueueIndexesForPbckbgeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method QueueIndexesForPbckbge.
	QueueIndexesForPbckbgeFunc *AutoIndexingServiceQueueIndexesForPbckbgeFunc
}

// NewMockAutoIndexingService crebtes b new mock of the AutoIndexingService
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockAutoIndexingService() *MockAutoIndexingService {
	return &MockAutoIndexingService{
		QueueIndexesForPbckbgeFunc: &AutoIndexingServiceQueueIndexesForPbckbgeFunc{
			defbultHook: func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockAutoIndexingService crebtes b new mock of the
// AutoIndexingService interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockAutoIndexingService() *MockAutoIndexingService {
	return &MockAutoIndexingService{
		QueueIndexesForPbckbgeFunc: &AutoIndexingServiceQueueIndexesForPbckbgeFunc{
			defbultHook: func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
				pbnic("unexpected invocbtion of MockAutoIndexingService.QueueIndexesForPbckbge")
			},
		},
	}
}

// NewMockAutoIndexingServiceFrom crebtes b new mock of the
// MockAutoIndexingService interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockAutoIndexingServiceFrom(i AutoIndexingService) *MockAutoIndexingService {
	return &MockAutoIndexingService{
		QueueIndexesForPbckbgeFunc: &AutoIndexingServiceQueueIndexesForPbckbgeFunc{
			defbultHook: i.QueueIndexesForPbckbge,
		},
	}
}

// AutoIndexingServiceQueueIndexesForPbckbgeFunc describes the behbvior when
// the QueueIndexesForPbckbge method of the pbrent MockAutoIndexingService
// instbnce is invoked.
type AutoIndexingServiceQueueIndexesForPbckbgeFunc struct {
	defbultHook func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error
	hooks       []func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error
	history     []AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll
	mutex       sync.Mutex
}

// QueueIndexesForPbckbge delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAutoIndexingService) QueueIndexesForPbckbge(v0 context.Context, v1 shbred.MinimiblVersionedPbckbgeRepo, v2 bool) error {
	r0 := m.QueueIndexesForPbckbgeFunc.nextHook()(v0, v1, v2)
	m.QueueIndexesForPbckbgeFunc.bppendCbll(AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// QueueIndexesForPbckbge method of the pbrent MockAutoIndexingService
// instbnce is invoked bnd the hook queue is empty.
func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) SetDefbultHook(hook func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueueIndexesForPbckbge method of the pbrent MockAutoIndexingService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) PushHook(hook func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
		return r0
	})
}

func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) nextHook() func(context.Context, shbred.MinimiblVersionedPbckbgeRepo, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) bppendCbll(r0 AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll objects describing the
// invocbtions of this function.
func (f *AutoIndexingServiceQueueIndexesForPbckbgeFunc) History() []AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll {
	f.mutex.Lock()
	history := mbke([]AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll is bn object thbt
// describes bn invocbtion of method QueueIndexesForPbckbge on bn instbnce
// of MockAutoIndexingService.
type AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll struct {
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
func (c AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AutoIndexingServiceQueueIndexesForPbckbgeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockDependenciesService is b mock implementbtion of the
// DependenciesService interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/internbl/bbckground)
// used for unit testing.
type MockDependenciesService struct {
	// InsertPbckbgeRepoRefsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertPbckbgeRepoRefs.
	InsertPbckbgeRepoRefsFunc *DependenciesServiceInsertPbckbgeRepoRefsFunc
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

// MockExternblServiceStore is b mock implementbtion of the
// ExternblServiceStore interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/internbl/bbckground)
// used for unit testing.
type MockExternblServiceStore struct {
	// GetByIDFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetByID.
	GetByIDFunc *ExternblServiceStoreGetByIDFunc
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
		GetByIDFunc: &ExternblServiceStoreGetByIDFunc{
			defbultHook: func(context.Context, int64) (r0 *types.ExternblService, r1 error) {
				return
			},
		},
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
		GetByIDFunc: &ExternblServiceStoreGetByIDFunc{
			defbultHook: func(context.Context, int64) (*types.ExternblService, error) {
				pbnic("unexpected invocbtion of MockExternblServiceStore.GetByID")
			},
		},
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
		GetByIDFunc: &ExternblServiceStoreGetByIDFunc{
			defbultHook: i.GetByID,
		},
		ListFunc: &ExternblServiceStoreListFunc{
			defbultHook: i.List,
		},
		UpsertFunc: &ExternblServiceStoreUpsertFunc{
			defbultHook: i.Upsert,
		},
	}
}

// ExternblServiceStoreGetByIDFunc describes the behbvior when the GetByID
// method of the pbrent MockExternblServiceStore instbnce is invoked.
type ExternblServiceStoreGetByIDFunc struct {
	defbultHook func(context.Context, int64) (*types.ExternblService, error)
	hooks       []func(context.Context, int64) (*types.ExternblService, error)
	history     []ExternblServiceStoreGetByIDFuncCbll
	mutex       sync.Mutex
}

// GetByID delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockExternblServiceStore) GetByID(v0 context.Context, v1 int64) (*types.ExternblService, error) {
	r0, r1 := m.GetByIDFunc.nextHook()(v0, v1)
	m.GetByIDFunc.bppendCbll(ExternblServiceStoreGetByIDFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByID method of
// the pbrent MockExternblServiceStore instbnce is invoked bnd the hook
// queue is empty.
func (f *ExternblServiceStoreGetByIDFunc) SetDefbultHook(hook func(context.Context, int64) (*types.ExternblService, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByID method of the pbrent MockExternblServiceStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ExternblServiceStoreGetByIDFunc) PushHook(hook func(context.Context, int64) (*types.ExternblService, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ExternblServiceStoreGetByIDFunc) SetDefbultReturn(r0 *types.ExternblService, r1 error) {
	f.SetDefbultHook(func(context.Context, int64) (*types.ExternblService, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ExternblServiceStoreGetByIDFunc) PushReturn(r0 *types.ExternblService, r1 error) {
	f.PushHook(func(context.Context, int64) (*types.ExternblService, error) {
		return r0, r1
	})
}

func (f *ExternblServiceStoreGetByIDFunc) nextHook() func(context.Context, int64) (*types.ExternblService, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExternblServiceStoreGetByIDFunc) bppendCbll(r0 ExternblServiceStoreGetByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ExternblServiceStoreGetByIDFuncCbll objects
// describing the invocbtions of this function.
func (f *ExternblServiceStoreGetByIDFunc) History() []ExternblServiceStoreGetByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]ExternblServiceStoreGetByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExternblServiceStoreGetByIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetByID on bn instbnce of MockExternblServiceStore.
type ExternblServiceStoreGetByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.ExternblService
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ExternblServiceStoreGetByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ExternblServiceStoreGetByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
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
