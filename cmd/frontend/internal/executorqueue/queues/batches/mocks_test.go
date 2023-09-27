// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge bbtches

import (
	"context"
	"sync"

	store "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	types "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// MockBbtchesStore is b mock implementbtion of the BbtchesStore interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/queues/bbtches)
// used for unit testing.
type MockBbtchesStore struct {
	// DbtbbbseDBFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DbtbbbseDB.
	DbtbbbseDBFunc *BbtchesStoreDbtbbbseDBFunc
	// GetBbtchSpecFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetBbtchSpec.
	GetBbtchSpecFunc *BbtchesStoreGetBbtchSpecFunc
	// GetBbtchSpecWorkspbceFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetBbtchSpecWorkspbce.
	GetBbtchSpecWorkspbceFunc *BbtchesStoreGetBbtchSpecWorkspbceFunc
	// ListBbtchSpecWorkspbceFilesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListBbtchSpecWorkspbceFiles.
	ListBbtchSpecWorkspbceFilesFunc *BbtchesStoreListBbtchSpecWorkspbceFilesFunc
}

// NewMockBbtchesStore crebtes b new mock of the BbtchesStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockBbtchesStore() *MockBbtchesStore {
	return &MockBbtchesStore{
		DbtbbbseDBFunc: &BbtchesStoreDbtbbbseDBFunc{
			defbultHook: func() (r0 dbtbbbse.DB) {
				return
			},
		},
		GetBbtchSpecFunc: &BbtchesStoreGetBbtchSpecFunc{
			defbultHook: func(context.Context, store.GetBbtchSpecOpts) (r0 *types.BbtchSpec, r1 error) {
				return
			},
		},
		GetBbtchSpecWorkspbceFunc: &BbtchesStoreGetBbtchSpecWorkspbceFunc{
			defbultHook: func(context.Context, store.GetBbtchSpecWorkspbceOpts) (r0 *types.BbtchSpecWorkspbce, r1 error) {
				return
			},
		},
		ListBbtchSpecWorkspbceFilesFunc: &BbtchesStoreListBbtchSpecWorkspbceFilesFunc{
			defbultHook: func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) (r0 []*types.BbtchSpecWorkspbceFile, r1 int64, r2 error) {
				return
			},
		},
	}
}

// NewStrictMockBbtchesStore crebtes b new mock of the BbtchesStore
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockBbtchesStore() *MockBbtchesStore {
	return &MockBbtchesStore{
		DbtbbbseDBFunc: &BbtchesStoreDbtbbbseDBFunc{
			defbultHook: func() dbtbbbse.DB {
				pbnic("unexpected invocbtion of MockBbtchesStore.DbtbbbseDB")
			},
		},
		GetBbtchSpecFunc: &BbtchesStoreGetBbtchSpecFunc{
			defbultHook: func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error) {
				pbnic("unexpected invocbtion of MockBbtchesStore.GetBbtchSpec")
			},
		},
		GetBbtchSpecWorkspbceFunc: &BbtchesStoreGetBbtchSpecWorkspbceFunc{
			defbultHook: func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error) {
				pbnic("unexpected invocbtion of MockBbtchesStore.GetBbtchSpecWorkspbce")
			},
		},
		ListBbtchSpecWorkspbceFilesFunc: &BbtchesStoreListBbtchSpecWorkspbceFilesFunc{
			defbultHook: func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error) {
				pbnic("unexpected invocbtion of MockBbtchesStore.ListBbtchSpecWorkspbceFiles")
			},
		},
	}
}

// NewMockBbtchesStoreFrom crebtes b new mock of the MockBbtchesStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockBbtchesStoreFrom(i BbtchesStore) *MockBbtchesStore {
	return &MockBbtchesStore{
		DbtbbbseDBFunc: &BbtchesStoreDbtbbbseDBFunc{
			defbultHook: i.DbtbbbseDB,
		},
		GetBbtchSpecFunc: &BbtchesStoreGetBbtchSpecFunc{
			defbultHook: i.GetBbtchSpec,
		},
		GetBbtchSpecWorkspbceFunc: &BbtchesStoreGetBbtchSpecWorkspbceFunc{
			defbultHook: i.GetBbtchSpecWorkspbce,
		},
		ListBbtchSpecWorkspbceFilesFunc: &BbtchesStoreListBbtchSpecWorkspbceFilesFunc{
			defbultHook: i.ListBbtchSpecWorkspbceFiles,
		},
	}
}

// BbtchesStoreDbtbbbseDBFunc describes the behbvior when the DbtbbbseDB
// method of the pbrent MockBbtchesStore instbnce is invoked.
type BbtchesStoreDbtbbbseDBFunc struct {
	defbultHook func() dbtbbbse.DB
	hooks       []func() dbtbbbse.DB
	history     []BbtchesStoreDbtbbbseDBFuncCbll
	mutex       sync.Mutex
}

// DbtbbbseDB delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBbtchesStore) DbtbbbseDB() dbtbbbse.DB {
	r0 := m.DbtbbbseDBFunc.nextHook()()
	m.DbtbbbseDBFunc.bppendCbll(BbtchesStoreDbtbbbseDBFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DbtbbbseDB method of
// the pbrent MockBbtchesStore instbnce is invoked bnd the hook queue is
// empty.
func (f *BbtchesStoreDbtbbbseDBFunc) SetDefbultHook(hook func() dbtbbbse.DB) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DbtbbbseDB method of the pbrent MockBbtchesStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *BbtchesStoreDbtbbbseDBFunc) PushHook(hook func() dbtbbbse.DB) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BbtchesStoreDbtbbbseDBFunc) SetDefbultReturn(r0 dbtbbbse.DB) {
	f.SetDefbultHook(func() dbtbbbse.DB {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BbtchesStoreDbtbbbseDBFunc) PushReturn(r0 dbtbbbse.DB) {
	f.PushHook(func() dbtbbbse.DB {
		return r0
	})
}

func (f *BbtchesStoreDbtbbbseDBFunc) nextHook() func() dbtbbbse.DB {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BbtchesStoreDbtbbbseDBFunc) bppendCbll(r0 BbtchesStoreDbtbbbseDBFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BbtchesStoreDbtbbbseDBFuncCbll objects
// describing the invocbtions of this function.
func (f *BbtchesStoreDbtbbbseDBFunc) History() []BbtchesStoreDbtbbbseDBFuncCbll {
	f.mutex.Lock()
	history := mbke([]BbtchesStoreDbtbbbseDBFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BbtchesStoreDbtbbbseDBFuncCbll is bn object thbt describes bn invocbtion
// of method DbtbbbseDB on bn instbnce of MockBbtchesStore.
type BbtchesStoreDbtbbbseDBFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.DB
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BbtchesStoreDbtbbbseDBFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BbtchesStoreDbtbbbseDBFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// BbtchesStoreGetBbtchSpecFunc describes the behbvior when the GetBbtchSpec
// method of the pbrent MockBbtchesStore instbnce is invoked.
type BbtchesStoreGetBbtchSpecFunc struct {
	defbultHook func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error)
	hooks       []func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error)
	history     []BbtchesStoreGetBbtchSpecFuncCbll
	mutex       sync.Mutex
}

// GetBbtchSpec delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBbtchesStore) GetBbtchSpec(v0 context.Context, v1 store.GetBbtchSpecOpts) (*types.BbtchSpec, error) {
	r0, r1 := m.GetBbtchSpecFunc.nextHook()(v0, v1)
	m.GetBbtchSpecFunc.bppendCbll(BbtchesStoreGetBbtchSpecFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBbtchSpec method
// of the pbrent MockBbtchesStore instbnce is invoked bnd the hook queue is
// empty.
func (f *BbtchesStoreGetBbtchSpecFunc) SetDefbultHook(hook func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBbtchSpec method of the pbrent MockBbtchesStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *BbtchesStoreGetBbtchSpecFunc) PushHook(hook func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BbtchesStoreGetBbtchSpecFunc) SetDefbultReturn(r0 *types.BbtchSpec, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BbtchesStoreGetBbtchSpecFunc) PushReturn(r0 *types.BbtchSpec, r1 error) {
	f.PushHook(func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error) {
		return r0, r1
	})
}

func (f *BbtchesStoreGetBbtchSpecFunc) nextHook() func(context.Context, store.GetBbtchSpecOpts) (*types.BbtchSpec, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BbtchesStoreGetBbtchSpecFunc) bppendCbll(r0 BbtchesStoreGetBbtchSpecFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BbtchesStoreGetBbtchSpecFuncCbll objects
// describing the invocbtions of this function.
func (f *BbtchesStoreGetBbtchSpecFunc) History() []BbtchesStoreGetBbtchSpecFuncCbll {
	f.mutex.Lock()
	history := mbke([]BbtchesStoreGetBbtchSpecFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BbtchesStoreGetBbtchSpecFuncCbll is bn object thbt describes bn
// invocbtion of method GetBbtchSpec on bn instbnce of MockBbtchesStore.
type BbtchesStoreGetBbtchSpecFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetBbtchSpecOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.BbtchSpec
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BbtchesStoreGetBbtchSpecFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BbtchesStoreGetBbtchSpecFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BbtchesStoreGetBbtchSpecWorkspbceFunc describes the behbvior when the
// GetBbtchSpecWorkspbce method of the pbrent MockBbtchesStore instbnce is
// invoked.
type BbtchesStoreGetBbtchSpecWorkspbceFunc struct {
	defbultHook func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error)
	hooks       []func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error)
	history     []BbtchesStoreGetBbtchSpecWorkspbceFuncCbll
	mutex       sync.Mutex
}

// GetBbtchSpecWorkspbce delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBbtchesStore) GetBbtchSpecWorkspbce(v0 context.Context, v1 store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error) {
	r0, r1 := m.GetBbtchSpecWorkspbceFunc.nextHook()(v0, v1)
	m.GetBbtchSpecWorkspbceFunc.bppendCbll(BbtchesStoreGetBbtchSpecWorkspbceFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetBbtchSpecWorkspbce method of the pbrent MockBbtchesStore instbnce is
// invoked bnd the hook queue is empty.
func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) SetDefbultHook(hook func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBbtchSpecWorkspbce method of the pbrent MockBbtchesStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) PushHook(hook func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) SetDefbultReturn(r0 *types.BbtchSpecWorkspbce, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) PushReturn(r0 *types.BbtchSpecWorkspbce, r1 error) {
	f.PushHook(func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error) {
		return r0, r1
	})
}

func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) nextHook() func(context.Context, store.GetBbtchSpecWorkspbceOpts) (*types.BbtchSpecWorkspbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) bppendCbll(r0 BbtchesStoreGetBbtchSpecWorkspbceFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BbtchesStoreGetBbtchSpecWorkspbceFuncCbll
// objects describing the invocbtions of this function.
func (f *BbtchesStoreGetBbtchSpecWorkspbceFunc) History() []BbtchesStoreGetBbtchSpecWorkspbceFuncCbll {
	f.mutex.Lock()
	history := mbke([]BbtchesStoreGetBbtchSpecWorkspbceFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BbtchesStoreGetBbtchSpecWorkspbceFuncCbll is bn object thbt describes bn
// invocbtion of method GetBbtchSpecWorkspbce on bn instbnce of
// MockBbtchesStore.
type BbtchesStoreGetBbtchSpecWorkspbceFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetBbtchSpecWorkspbceOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.BbtchSpecWorkspbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BbtchesStoreGetBbtchSpecWorkspbceFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BbtchesStoreGetBbtchSpecWorkspbceFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BbtchesStoreListBbtchSpecWorkspbceFilesFunc describes the behbvior when
// the ListBbtchSpecWorkspbceFiles method of the pbrent MockBbtchesStore
// instbnce is invoked.
type BbtchesStoreListBbtchSpecWorkspbceFilesFunc struct {
	defbultHook func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error)
	hooks       []func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error)
	history     []BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll
	mutex       sync.Mutex
}

// ListBbtchSpecWorkspbceFiles delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBbtchesStore) ListBbtchSpecWorkspbceFiles(v0 context.Context, v1 store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error) {
	r0, r1, r2 := m.ListBbtchSpecWorkspbceFilesFunc.nextHook()(v0, v1)
	m.ListBbtchSpecWorkspbceFilesFunc.bppendCbll(BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ListBbtchSpecWorkspbceFiles method of the pbrent MockBbtchesStore
// instbnce is invoked bnd the hook queue is empty.
func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) SetDefbultHook(hook func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListBbtchSpecWorkspbceFiles method of the pbrent MockBbtchesStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) PushHook(hook func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) SetDefbultReturn(r0 []*types.BbtchSpecWorkspbceFile, r1 int64, r2 error) {
	f.SetDefbultHook(func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) PushReturn(r0 []*types.BbtchSpecWorkspbceFile, r1 int64, r2 error) {
	f.PushHook(func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error) {
		return r0, r1, r2
	})
}

func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) nextHook() func(context.Context, store.ListBbtchSpecWorkspbceFileOpts) ([]*types.BbtchSpecWorkspbceFile, int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) bppendCbll(r0 BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll objects describing the
// invocbtions of this function.
func (f *BbtchesStoreListBbtchSpecWorkspbceFilesFunc) History() []BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll {
	f.mutex.Lock()
	history := mbke([]BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll is bn object thbt
// describes bn invocbtion of method ListBbtchSpecWorkspbceFiles on bn
// instbnce of MockBbtchesStore.
type BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.ListBbtchSpecWorkspbceFileOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.BbtchSpecWorkspbceFile
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int64
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BbtchesStoreListBbtchSpecWorkspbceFilesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}
