// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge bbckend

import (
	"context"
	"sync"
	"time"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	gitdombin "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	inventory "github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// MockReposService is b mock implementbtion of the ReposService interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend) used for unit
// testing.
type MockReposService struct {
	// DeleteRepositoryFromDiskFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteRepositoryFromDisk.
	DeleteRepositoryFromDiskFunc *ReposServiceDeleteRepositoryFromDiskFunc
	// GetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Get.
	GetFunc *ReposServiceGetFunc
	// GetByNbmeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetByNbme.
	GetByNbmeFunc *ReposServiceGetByNbmeFunc
	// GetCommitFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetCommit.
	GetCommitFunc *ReposServiceGetCommitFunc
	// GetInventoryFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetInventory.
	GetInventoryFunc *ReposServiceGetInventoryFunc
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *ReposServiceListFunc
	// ListIndexbbleFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListIndexbble.
	ListIndexbbleFunc *ReposServiceListIndexbbleFunc
	// RequestRepositoryCloneFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RequestRepositoryClone.
	RequestRepositoryCloneFunc *ReposServiceRequestRepositoryCloneFunc
	// ResolveRevFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ResolveRev.
	ResolveRevFunc *ReposServiceResolveRevFunc
}

// NewMockReposService crebtes b new mock of the ReposService interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockReposService() *MockReposService {
	return &MockReposService{
		DeleteRepositoryFromDiskFunc: &ReposServiceDeleteRepositoryFromDiskFunc{
			defbultHook: func(context.Context, bpi.RepoID) (r0 error) {
				return
			},
		},
		GetFunc: &ReposServiceGetFunc{
			defbultHook: func(context.Context, bpi.RepoID) (r0 *types.Repo, r1 error) {
				return
			},
		},
		GetByNbmeFunc: &ReposServiceGetByNbmeFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 *types.Repo, r1 error) {
				return
			},
		},
		GetCommitFunc: &ReposServiceGetCommitFunc{
			defbultHook: func(context.Context, *types.Repo, bpi.CommitID) (r0 *gitdombin.Commit, r1 error) {
				return
			},
		},
		GetInventoryFunc: &ReposServiceGetInventoryFunc{
			defbultHook: func(context.Context, *types.Repo, bpi.CommitID, bool) (r0 *inventory.Inventory, r1 error) {
				return
			},
		},
		ListFunc: &ReposServiceListFunc{
			defbultHook: func(context.Context, dbtbbbse.ReposListOptions) (r0 []*types.Repo, r1 error) {
				return
			},
		},
		ListIndexbbleFunc: &ReposServiceListIndexbbleFunc{
			defbultHook: func(context.Context) (r0 []types.MinimblRepo, r1 error) {
				return
			},
		},
		RequestRepositoryCloneFunc: &ReposServiceRequestRepositoryCloneFunc{
			defbultHook: func(context.Context, bpi.RepoID) (r0 error) {
				return
			},
		},
		ResolveRevFunc: &ReposServiceResolveRevFunc{
			defbultHook: func(context.Context, *types.Repo, string) (r0 bpi.CommitID, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockReposService crebtes b new mock of the ReposService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockReposService() *MockReposService {
	return &MockReposService{
		DeleteRepositoryFromDiskFunc: &ReposServiceDeleteRepositoryFromDiskFunc{
			defbultHook: func(context.Context, bpi.RepoID) error {
				pbnic("unexpected invocbtion of MockReposService.DeleteRepositoryFromDisk")
			},
		},
		GetFunc: &ReposServiceGetFunc{
			defbultHook: func(context.Context, bpi.RepoID) (*types.Repo, error) {
				pbnic("unexpected invocbtion of MockReposService.Get")
			},
		},
		GetByNbmeFunc: &ReposServiceGetByNbmeFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (*types.Repo, error) {
				pbnic("unexpected invocbtion of MockReposService.GetByNbme")
			},
		},
		GetCommitFunc: &ReposServiceGetCommitFunc{
			defbultHook: func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockReposService.GetCommit")
			},
		},
		GetInventoryFunc: &ReposServiceGetInventoryFunc{
			defbultHook: func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error) {
				pbnic("unexpected invocbtion of MockReposService.GetInventory")
			},
		},
		ListFunc: &ReposServiceListFunc{
			defbultHook: func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
				pbnic("unexpected invocbtion of MockReposService.List")
			},
		},
		ListIndexbbleFunc: &ReposServiceListIndexbbleFunc{
			defbultHook: func(context.Context) ([]types.MinimblRepo, error) {
				pbnic("unexpected invocbtion of MockReposService.ListIndexbble")
			},
		},
		RequestRepositoryCloneFunc: &ReposServiceRequestRepositoryCloneFunc{
			defbultHook: func(context.Context, bpi.RepoID) error {
				pbnic("unexpected invocbtion of MockReposService.RequestRepositoryClone")
			},
		},
		ResolveRevFunc: &ReposServiceResolveRevFunc{
			defbultHook: func(context.Context, *types.Repo, string) (bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockReposService.ResolveRev")
			},
		},
	}
}

// NewMockReposServiceFrom crebtes b new mock of the MockReposService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockReposServiceFrom(i ReposService) *MockReposService {
	return &MockReposService{
		DeleteRepositoryFromDiskFunc: &ReposServiceDeleteRepositoryFromDiskFunc{
			defbultHook: i.DeleteRepositoryFromDisk,
		},
		GetFunc: &ReposServiceGetFunc{
			defbultHook: i.Get,
		},
		GetByNbmeFunc: &ReposServiceGetByNbmeFunc{
			defbultHook: i.GetByNbme,
		},
		GetCommitFunc: &ReposServiceGetCommitFunc{
			defbultHook: i.GetCommit,
		},
		GetInventoryFunc: &ReposServiceGetInventoryFunc{
			defbultHook: i.GetInventory,
		},
		ListFunc: &ReposServiceListFunc{
			defbultHook: i.List,
		},
		ListIndexbbleFunc: &ReposServiceListIndexbbleFunc{
			defbultHook: i.ListIndexbble,
		},
		RequestRepositoryCloneFunc: &ReposServiceRequestRepositoryCloneFunc{
			defbultHook: i.RequestRepositoryClone,
		},
		ResolveRevFunc: &ReposServiceResolveRevFunc{
			defbultHook: i.ResolveRev,
		},
	}
}

// ReposServiceDeleteRepositoryFromDiskFunc describes the behbvior when the
// DeleteRepositoryFromDisk method of the pbrent MockReposService instbnce
// is invoked.
type ReposServiceDeleteRepositoryFromDiskFunc struct {
	defbultHook func(context.Context, bpi.RepoID) error
	hooks       []func(context.Context, bpi.RepoID) error
	history     []ReposServiceDeleteRepositoryFromDiskFuncCbll
	mutex       sync.Mutex
}

// DeleteRepositoryFromDisk delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) DeleteRepositoryFromDisk(v0 context.Context, v1 bpi.RepoID) error {
	r0 := m.DeleteRepositoryFromDiskFunc.nextHook()(v0, v1)
	m.DeleteRepositoryFromDiskFunc.bppendCbll(ReposServiceDeleteRepositoryFromDiskFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// DeleteRepositoryFromDisk method of the pbrent MockReposService instbnce
// is invoked bnd the hook queue is empty.
func (f *ReposServiceDeleteRepositoryFromDiskFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteRepositoryFromDisk method of the pbrent MockReposService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ReposServiceDeleteRepositoryFromDiskFunc) PushHook(hook func(context.Context, bpi.RepoID) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceDeleteRepositoryFromDiskFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceDeleteRepositoryFromDiskFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoID) error {
		return r0
	})
}

func (f *ReposServiceDeleteRepositoryFromDiskFunc) nextHook() func(context.Context, bpi.RepoID) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceDeleteRepositoryFromDiskFunc) bppendCbll(r0 ReposServiceDeleteRepositoryFromDiskFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ReposServiceDeleteRepositoryFromDiskFuncCbll objects describing the
// invocbtions of this function.
func (f *ReposServiceDeleteRepositoryFromDiskFunc) History() []ReposServiceDeleteRepositoryFromDiskFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceDeleteRepositoryFromDiskFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceDeleteRepositoryFromDiskFuncCbll is bn object thbt describes
// bn invocbtion of method DeleteRepositoryFromDisk on bn instbnce of
// MockReposService.
type ReposServiceDeleteRepositoryFromDiskFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceDeleteRepositoryFromDiskFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceDeleteRepositoryFromDiskFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ReposServiceGetFunc describes the behbvior when the Get method of the
// pbrent MockReposService instbnce is invoked.
type ReposServiceGetFunc struct {
	defbultHook func(context.Context, bpi.RepoID) (*types.Repo, error)
	hooks       []func(context.Context, bpi.RepoID) (*types.Repo, error)
	history     []ReposServiceGetFuncCbll
	mutex       sync.Mutex
}

// Get delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) Get(v0 context.Context, v1 bpi.RepoID) (*types.Repo, error) {
	r0, r1 := m.GetFunc.nextHook()(v0, v1)
	m.GetFunc.bppendCbll(ReposServiceGetFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Get method of the
// pbrent MockReposService instbnce is invoked bnd the hook queue is empty.
func (f *ReposServiceGetFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID) (*types.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Get method of the pbrent MockReposService instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ReposServiceGetFunc) PushHook(hook func(context.Context, bpi.RepoID) (*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceGetFunc) SetDefbultReturn(r0 *types.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID) (*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceGetFunc) PushReturn(r0 *types.Repo, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoID) (*types.Repo, error) {
		return r0, r1
	})
}

func (f *ReposServiceGetFunc) nextHook() func(context.Context, bpi.RepoID) (*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceGetFunc) bppendCbll(r0 ReposServiceGetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceGetFuncCbll objects describing
// the invocbtions of this function.
func (f *ReposServiceGetFunc) History() []ReposServiceGetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceGetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceGetFuncCbll is bn object thbt describes bn invocbtion of
// method Get on bn instbnce of MockReposService.
type ReposServiceGetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceGetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceGetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ReposServiceGetByNbmeFunc describes the behbvior when the GetByNbme
// method of the pbrent MockReposService instbnce is invoked.
type ReposServiceGetByNbmeFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (*types.Repo, error)
	hooks       []func(context.Context, bpi.RepoNbme) (*types.Repo, error)
	history     []ReposServiceGetByNbmeFuncCbll
	mutex       sync.Mutex
}

// GetByNbme delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) GetByNbme(v0 context.Context, v1 bpi.RepoNbme) (*types.Repo, error) {
	r0, r1 := m.GetByNbmeFunc.nextHook()(v0, v1)
	m.GetByNbmeFunc.bppendCbll(ReposServiceGetByNbmeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByNbme method of
// the pbrent MockReposService instbnce is invoked bnd the hook queue is
// empty.
func (f *ReposServiceGetByNbmeFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (*types.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByNbme method of the pbrent MockReposService instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ReposServiceGetByNbmeFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceGetByNbmeFunc) SetDefbultReturn(r0 *types.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceGetByNbmeFunc) PushReturn(r0 *types.Repo, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (*types.Repo, error) {
		return r0, r1
	})
}

func (f *ReposServiceGetByNbmeFunc) nextHook() func(context.Context, bpi.RepoNbme) (*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceGetByNbmeFunc) bppendCbll(r0 ReposServiceGetByNbmeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceGetByNbmeFuncCbll objects
// describing the invocbtions of this function.
func (f *ReposServiceGetByNbmeFunc) History() []ReposServiceGetByNbmeFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceGetByNbmeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceGetByNbmeFuncCbll is bn object thbt describes bn invocbtion
// of method GetByNbme on bn instbnce of MockReposService.
type ReposServiceGetByNbmeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceGetByNbmeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceGetByNbmeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ReposServiceGetCommitFunc describes the behbvior when the GetCommit
// method of the pbrent MockReposService instbnce is invoked.
type ReposServiceGetCommitFunc struct {
	defbultHook func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error)
	hooks       []func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error)
	history     []ReposServiceGetCommitFuncCbll
	mutex       sync.Mutex
}

// GetCommit delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) GetCommit(v0 context.Context, v1 *types.Repo, v2 bpi.CommitID) (*gitdombin.Commit, error) {
	r0, r1 := m.GetCommitFunc.nextHook()(v0, v1, v2)
	m.GetCommitFunc.bppendCbll(ReposServiceGetCommitFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetCommit method of
// the pbrent MockReposService instbnce is invoked bnd the hook queue is
// empty.
func (f *ReposServiceGetCommitFunc) SetDefbultHook(hook func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommit method of the pbrent MockReposService instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ReposServiceGetCommitFunc) PushHook(hook func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceGetCommitFunc) SetDefbultReturn(r0 *gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceGetCommitFunc) PushReturn(r0 *gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *ReposServiceGetCommitFunc) nextHook() func(context.Context, *types.Repo, bpi.CommitID) (*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceGetCommitFunc) bppendCbll(r0 ReposServiceGetCommitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceGetCommitFuncCbll objects
// describing the invocbtions of this function.
func (f *ReposServiceGetCommitFunc) History() []ReposServiceGetCommitFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceGetCommitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceGetCommitFuncCbll is bn object thbt describes bn invocbtion
// of method GetCommit on bn instbnce of MockReposService.
type ReposServiceGetCommitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceGetCommitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceGetCommitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ReposServiceGetInventoryFunc describes the behbvior when the GetInventory
// method of the pbrent MockReposService instbnce is invoked.
type ReposServiceGetInventoryFunc struct {
	defbultHook func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error)
	hooks       []func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error)
	history     []ReposServiceGetInventoryFuncCbll
	mutex       sync.Mutex
}

// GetInventory delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) GetInventory(v0 context.Context, v1 *types.Repo, v2 bpi.CommitID, v3 bool) (*inventory.Inventory, error) {
	r0, r1 := m.GetInventoryFunc.nextHook()(v0, v1, v2, v3)
	m.GetInventoryFunc.bppendCbll(ReposServiceGetInventoryFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetInventory method
// of the pbrent MockReposService instbnce is invoked bnd the hook queue is
// empty.
func (f *ReposServiceGetInventoryFunc) SetDefbultHook(hook func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetInventory method of the pbrent MockReposService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ReposServiceGetInventoryFunc) PushHook(hook func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceGetInventoryFunc) SetDefbultReturn(r0 *inventory.Inventory, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceGetInventoryFunc) PushReturn(r0 *inventory.Inventory, r1 error) {
	f.PushHook(func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error) {
		return r0, r1
	})
}

func (f *ReposServiceGetInventoryFunc) nextHook() func(context.Context, *types.Repo, bpi.CommitID, bool) (*inventory.Inventory, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceGetInventoryFunc) bppendCbll(r0 ReposServiceGetInventoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceGetInventoryFuncCbll objects
// describing the invocbtions of this function.
func (f *ReposServiceGetInventoryFunc) History() []ReposServiceGetInventoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceGetInventoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceGetInventoryFuncCbll is bn object thbt describes bn
// invocbtion of method GetInventory on bn instbnce of MockReposService.
type ReposServiceGetInventoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *inventory.Inventory
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceGetInventoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceGetInventoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ReposServiceListFunc describes the behbvior when the List method of the
// pbrent MockReposService instbnce is invoked.
type ReposServiceListFunc struct {
	defbultHook func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)
	hooks       []func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)
	history     []ReposServiceListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) List(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.bppendCbll(ReposServiceListFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockReposService instbnce is invoked bnd the hook queue is empty.
func (f *ReposServiceListFunc) SetDefbultHook(hook func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockReposService instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ReposServiceListFunc) PushHook(hook func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceListFunc) SetDefbultReturn(r0 []*types.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceListFunc) PushReturn(r0 []*types.Repo, r1 error) {
	f.PushHook(func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		return r0, r1
	})
}

func (f *ReposServiceListFunc) nextHook() func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceListFunc) bppendCbll(r0 ReposServiceListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceListFuncCbll objects describing
// the invocbtions of this function.
func (f *ReposServiceListFunc) History() []ReposServiceListFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceListFuncCbll is bn object thbt describes bn invocbtion of
// method List on bn instbnce of MockReposService.
type ReposServiceListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 dbtbbbse.ReposListOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ReposServiceListIndexbbleFunc describes the behbvior when the
// ListIndexbble method of the pbrent MockReposService instbnce is invoked.
type ReposServiceListIndexbbleFunc struct {
	defbultHook func(context.Context) ([]types.MinimblRepo, error)
	hooks       []func(context.Context) ([]types.MinimblRepo, error)
	history     []ReposServiceListIndexbbleFuncCbll
	mutex       sync.Mutex
}

// ListIndexbble delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) ListIndexbble(v0 context.Context) ([]types.MinimblRepo, error) {
	r0, r1 := m.ListIndexbbleFunc.nextHook()(v0)
	m.ListIndexbbleFunc.bppendCbll(ReposServiceListIndexbbleFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListIndexbble method
// of the pbrent MockReposService instbnce is invoked bnd the hook queue is
// empty.
func (f *ReposServiceListIndexbbleFunc) SetDefbultHook(hook func(context.Context) ([]types.MinimblRepo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListIndexbble method of the pbrent MockReposService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ReposServiceListIndexbbleFunc) PushHook(hook func(context.Context) ([]types.MinimblRepo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceListIndexbbleFunc) SetDefbultReturn(r0 []types.MinimblRepo, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]types.MinimblRepo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceListIndexbbleFunc) PushReturn(r0 []types.MinimblRepo, r1 error) {
	f.PushHook(func(context.Context) ([]types.MinimblRepo, error) {
		return r0, r1
	})
}

func (f *ReposServiceListIndexbbleFunc) nextHook() func(context.Context) ([]types.MinimblRepo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceListIndexbbleFunc) bppendCbll(r0 ReposServiceListIndexbbleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceListIndexbbleFuncCbll objects
// describing the invocbtions of this function.
func (f *ReposServiceListIndexbbleFunc) History() []ReposServiceListIndexbbleFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceListIndexbbleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceListIndexbbleFuncCbll is bn object thbt describes bn
// invocbtion of method ListIndexbble on bn instbnce of MockReposService.
type ReposServiceListIndexbbleFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []types.MinimblRepo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceListIndexbbleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceListIndexbbleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ReposServiceRequestRepositoryCloneFunc describes the behbvior when the
// RequestRepositoryClone method of the pbrent MockReposService instbnce is
// invoked.
type ReposServiceRequestRepositoryCloneFunc struct {
	defbultHook func(context.Context, bpi.RepoID) error
	hooks       []func(context.Context, bpi.RepoID) error
	history     []ReposServiceRequestRepositoryCloneFuncCbll
	mutex       sync.Mutex
}

// RequestRepositoryClone delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) RequestRepositoryClone(v0 context.Context, v1 bpi.RepoID) error {
	r0 := m.RequestRepositoryCloneFunc.nextHook()(v0, v1)
	m.RequestRepositoryCloneFunc.bppendCbll(ReposServiceRequestRepositoryCloneFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// RequestRepositoryClone method of the pbrent MockReposService instbnce is
// invoked bnd the hook queue is empty.
func (f *ReposServiceRequestRepositoryCloneFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RequestRepositoryClone method of the pbrent MockReposService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ReposServiceRequestRepositoryCloneFunc) PushHook(hook func(context.Context, bpi.RepoID) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceRequestRepositoryCloneFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceRequestRepositoryCloneFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoID) error {
		return r0
	})
}

func (f *ReposServiceRequestRepositoryCloneFunc) nextHook() func(context.Context, bpi.RepoID) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceRequestRepositoryCloneFunc) bppendCbll(r0 ReposServiceRequestRepositoryCloneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceRequestRepositoryCloneFuncCbll
// objects describing the invocbtions of this function.
func (f *ReposServiceRequestRepositoryCloneFunc) History() []ReposServiceRequestRepositoryCloneFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceRequestRepositoryCloneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceRequestRepositoryCloneFuncCbll is bn object thbt describes bn
// invocbtion of method RequestRepositoryClone on bn instbnce of
// MockReposService.
type ReposServiceRequestRepositoryCloneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceRequestRepositoryCloneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceRequestRepositoryCloneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ReposServiceResolveRevFunc describes the behbvior when the ResolveRev
// method of the pbrent MockReposService instbnce is invoked.
type ReposServiceResolveRevFunc struct {
	defbultHook func(context.Context, *types.Repo, string) (bpi.CommitID, error)
	hooks       []func(context.Context, *types.Repo, string) (bpi.CommitID, error)
	history     []ReposServiceResolveRevFuncCbll
	mutex       sync.Mutex
}

// ResolveRev delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockReposService) ResolveRev(v0 context.Context, v1 *types.Repo, v2 string) (bpi.CommitID, error) {
	r0, r1 := m.ResolveRevFunc.nextHook()(v0, v1, v2)
	m.ResolveRevFunc.bppendCbll(ReposServiceResolveRevFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ResolveRev method of
// the pbrent MockReposService instbnce is invoked bnd the hook queue is
// empty.
func (f *ReposServiceResolveRevFunc) SetDefbultHook(hook func(context.Context, *types.Repo, string) (bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveRev method of the pbrent MockReposService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ReposServiceResolveRevFunc) PushHook(hook func(context.Context, *types.Repo, string) (bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ReposServiceResolveRevFunc) SetDefbultReturn(r0 bpi.CommitID, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.Repo, string) (bpi.CommitID, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ReposServiceResolveRevFunc) PushReturn(r0 bpi.CommitID, r1 error) {
	f.PushHook(func(context.Context, *types.Repo, string) (bpi.CommitID, error) {
		return r0, r1
	})
}

func (f *ReposServiceResolveRevFunc) nextHook() func(context.Context, *types.Repo, string) (bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ReposServiceResolveRevFunc) bppendCbll(r0 ReposServiceResolveRevFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ReposServiceResolveRevFuncCbll objects
// describing the invocbtions of this function.
func (f *ReposServiceResolveRevFunc) History() []ReposServiceResolveRevFuncCbll {
	f.mutex.Lock()
	history := mbke([]ReposServiceResolveRevFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ReposServiceResolveRevFuncCbll is bn object thbt describes bn invocbtion
// of method ResolveRev on bn instbnce of MockReposService.
type ReposServiceResolveRevFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bpi.CommitID
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ReposServiceResolveRevFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ReposServiceResolveRevFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockUserEmbilsService is b mock implementbtion of the UserEmbilsService
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend) used for unit
// testing.
type MockUserEmbilsService struct {
	// AddFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Add.
	AddFunc *UserEmbilsServiceAddFunc
	// CurrentActorHbsVerifiedEmbilFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// CurrentActorHbsVerifiedEmbil.
	CurrentActorHbsVerifiedEmbilFunc *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc
	// HbsVerifiedEmbilFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method HbsVerifiedEmbil.
	HbsVerifiedEmbilFunc *UserEmbilsServiceHbsVerifiedEmbilFunc
	// RemoveFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Remove.
	RemoveFunc *UserEmbilsServiceRemoveFunc
	// ResendVerificbtionEmbilFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResendVerificbtionEmbil.
	ResendVerificbtionEmbilFunc *UserEmbilsServiceResendVerificbtionEmbilFunc
	// SendUserEmbilOnAccessTokenChbngeFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// SendUserEmbilOnAccessTokenChbnge.
	SendUserEmbilOnAccessTokenChbngeFunc *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc
	// SendUserEmbilOnFieldUpdbteFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// SendUserEmbilOnFieldUpdbte.
	SendUserEmbilOnFieldUpdbteFunc *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc
	// SetPrimbryEmbilFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetPrimbryEmbil.
	SetPrimbryEmbilFunc *UserEmbilsServiceSetPrimbryEmbilFunc
	// SetVerifiedFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SetVerified.
	SetVerifiedFunc *UserEmbilsServiceSetVerifiedFunc
}

// NewMockUserEmbilsService crebtes b new mock of the UserEmbilsService
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockUserEmbilsService() *MockUserEmbilsService {
	return &MockUserEmbilsService{
		AddFunc: &UserEmbilsServiceAddFunc{
			defbultHook: func(context.Context, int32, string) (r0 error) {
				return
			},
		},
		CurrentActorHbsVerifiedEmbilFunc: &UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc{
			defbultHook: func(context.Context) (r0 bool, r1 error) {
				return
			},
		},
		HbsVerifiedEmbilFunc: &UserEmbilsServiceHbsVerifiedEmbilFunc{
			defbultHook: func(context.Context, int32) (r0 bool, r1 error) {
				return
			},
		},
		RemoveFunc: &UserEmbilsServiceRemoveFunc{
			defbultHook: func(context.Context, int32, string) (r0 error) {
				return
			},
		},
		ResendVerificbtionEmbilFunc: &UserEmbilsServiceResendVerificbtionEmbilFunc{
			defbultHook: func(context.Context, int32, string, time.Time) (r0 error) {
				return
			},
		},
		SendUserEmbilOnAccessTokenChbngeFunc: &UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc{
			defbultHook: func(context.Context, int32, string, bool) (r0 error) {
				return
			},
		},
		SendUserEmbilOnFieldUpdbteFunc: &UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc{
			defbultHook: func(context.Context, int32, string) (r0 error) {
				return
			},
		},
		SetPrimbryEmbilFunc: &UserEmbilsServiceSetPrimbryEmbilFunc{
			defbultHook: func(context.Context, int32, string) (r0 error) {
				return
			},
		},
		SetVerifiedFunc: &UserEmbilsServiceSetVerifiedFunc{
			defbultHook: func(context.Context, int32, string, bool) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockUserEmbilsService crebtes b new mock of the
// UserEmbilsService interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockUserEmbilsService() *MockUserEmbilsService {
	return &MockUserEmbilsService{
		AddFunc: &UserEmbilsServiceAddFunc{
			defbultHook: func(context.Context, int32, string) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.Add")
			},
		},
		CurrentActorHbsVerifiedEmbilFunc: &UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc{
			defbultHook: func(context.Context) (bool, error) {
				pbnic("unexpected invocbtion of MockUserEmbilsService.CurrentActorHbsVerifiedEmbil")
			},
		},
		HbsVerifiedEmbilFunc: &UserEmbilsServiceHbsVerifiedEmbilFunc{
			defbultHook: func(context.Context, int32) (bool, error) {
				pbnic("unexpected invocbtion of MockUserEmbilsService.HbsVerifiedEmbil")
			},
		},
		RemoveFunc: &UserEmbilsServiceRemoveFunc{
			defbultHook: func(context.Context, int32, string) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.Remove")
			},
		},
		ResendVerificbtionEmbilFunc: &UserEmbilsServiceResendVerificbtionEmbilFunc{
			defbultHook: func(context.Context, int32, string, time.Time) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.ResendVerificbtionEmbil")
			},
		},
		SendUserEmbilOnAccessTokenChbngeFunc: &UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc{
			defbultHook: func(context.Context, int32, string, bool) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.SendUserEmbilOnAccessTokenChbnge")
			},
		},
		SendUserEmbilOnFieldUpdbteFunc: &UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc{
			defbultHook: func(context.Context, int32, string) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.SendUserEmbilOnFieldUpdbte")
			},
		},
		SetPrimbryEmbilFunc: &UserEmbilsServiceSetPrimbryEmbilFunc{
			defbultHook: func(context.Context, int32, string) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.SetPrimbryEmbil")
			},
		},
		SetVerifiedFunc: &UserEmbilsServiceSetVerifiedFunc{
			defbultHook: func(context.Context, int32, string, bool) error {
				pbnic("unexpected invocbtion of MockUserEmbilsService.SetVerified")
			},
		},
	}
}

// NewMockUserEmbilsServiceFrom crebtes b new mock of the
// MockUserEmbilsService interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockUserEmbilsServiceFrom(i UserEmbilsService) *MockUserEmbilsService {
	return &MockUserEmbilsService{
		AddFunc: &UserEmbilsServiceAddFunc{
			defbultHook: i.Add,
		},
		CurrentActorHbsVerifiedEmbilFunc: &UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc{
			defbultHook: i.CurrentActorHbsVerifiedEmbil,
		},
		HbsVerifiedEmbilFunc: &UserEmbilsServiceHbsVerifiedEmbilFunc{
			defbultHook: i.HbsVerifiedEmbil,
		},
		RemoveFunc: &UserEmbilsServiceRemoveFunc{
			defbultHook: i.Remove,
		},
		ResendVerificbtionEmbilFunc: &UserEmbilsServiceResendVerificbtionEmbilFunc{
			defbultHook: i.ResendVerificbtionEmbil,
		},
		SendUserEmbilOnAccessTokenChbngeFunc: &UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc{
			defbultHook: i.SendUserEmbilOnAccessTokenChbnge,
		},
		SendUserEmbilOnFieldUpdbteFunc: &UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc{
			defbultHook: i.SendUserEmbilOnFieldUpdbte,
		},
		SetPrimbryEmbilFunc: &UserEmbilsServiceSetPrimbryEmbilFunc{
			defbultHook: i.SetPrimbryEmbil,
		},
		SetVerifiedFunc: &UserEmbilsServiceSetVerifiedFunc{
			defbultHook: i.SetVerified,
		},
	}
}

// UserEmbilsServiceAddFunc describes the behbvior when the Add method of
// the pbrent MockUserEmbilsService instbnce is invoked.
type UserEmbilsServiceAddFunc struct {
	defbultHook func(context.Context, int32, string) error
	hooks       []func(context.Context, int32, string) error
	history     []UserEmbilsServiceAddFuncCbll
	mutex       sync.Mutex
}

// Add delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) Add(v0 context.Context, v1 int32, v2 string) error {
	r0 := m.AddFunc.nextHook()(v0, v1, v2)
	m.AddFunc.bppendCbll(UserEmbilsServiceAddFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Add method of the
// pbrent MockUserEmbilsService instbnce is invoked bnd the hook queue is
// empty.
func (f *UserEmbilsServiceAddFunc) SetDefbultHook(hook func(context.Context, int32, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Add method of the pbrent MockUserEmbilsService instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *UserEmbilsServiceAddFunc) PushHook(hook func(context.Context, int32, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceAddFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceAddFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string) error {
		return r0
	})
}

func (f *UserEmbilsServiceAddFunc) nextHook() func(context.Context, int32, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceAddFunc) bppendCbll(r0 UserEmbilsServiceAddFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UserEmbilsServiceAddFuncCbll objects
// describing the invocbtions of this function.
func (f *UserEmbilsServiceAddFunc) History() []UserEmbilsServiceAddFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceAddFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceAddFuncCbll is bn object thbt describes bn invocbtion of
// method Add on bn instbnce of MockUserEmbilsService.
type UserEmbilsServiceAddFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceAddFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceAddFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc describes the behbvior
// when the CurrentActorHbsVerifiedEmbil method of the pbrent
// MockUserEmbilsService instbnce is invoked.
type UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc struct {
	defbultHook func(context.Context) (bool, error)
	hooks       []func(context.Context) (bool, error)
	history     []UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll
	mutex       sync.Mutex
}

// CurrentActorHbsVerifiedEmbil delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) CurrentActorHbsVerifiedEmbil(v0 context.Context) (bool, error) {
	r0, r1 := m.CurrentActorHbsVerifiedEmbilFunc.nextHook()(v0)
	m.CurrentActorHbsVerifiedEmbilFunc.bppendCbll(UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CurrentActorHbsVerifiedEmbil method of the pbrent MockUserEmbilsService
// instbnce is invoked bnd the hook queue is empty.
func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) SetDefbultHook(hook func(context.Context) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CurrentActorHbsVerifiedEmbil method of the pbrent MockUserEmbilsService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) PushHook(hook func(context.Context) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context) (bool, error) {
		return r0, r1
	})
}

func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) nextHook() func(context.Context) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) bppendCbll(r0 UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll objects describing
// the invocbtions of this function.
func (f *UserEmbilsServiceCurrentActorHbsVerifiedEmbilFunc) History() []UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll is bn object thbt
// describes bn invocbtion of method CurrentActorHbsVerifiedEmbil on bn
// instbnce of MockUserEmbilsService.
type UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceCurrentActorHbsVerifiedEmbilFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UserEmbilsServiceHbsVerifiedEmbilFunc describes the behbvior when the
// HbsVerifiedEmbil method of the pbrent MockUserEmbilsService instbnce is
// invoked.
type UserEmbilsServiceHbsVerifiedEmbilFunc struct {
	defbultHook func(context.Context, int32) (bool, error)
	hooks       []func(context.Context, int32) (bool, error)
	history     []UserEmbilsServiceHbsVerifiedEmbilFuncCbll
	mutex       sync.Mutex
}

// HbsVerifiedEmbil delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) HbsVerifiedEmbil(v0 context.Context, v1 int32) (bool, error) {
	r0, r1 := m.HbsVerifiedEmbilFunc.nextHook()(v0, v1)
	m.HbsVerifiedEmbilFunc.bppendCbll(UserEmbilsServiceHbsVerifiedEmbilFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the HbsVerifiedEmbil
// method of the pbrent MockUserEmbilsService instbnce is invoked bnd the
// hook queue is empty.
func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) SetDefbultHook(hook func(context.Context, int32) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HbsVerifiedEmbil method of the pbrent MockUserEmbilsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) PushHook(hook func(context.Context, int32) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int32) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int32) (bool, error) {
		return r0, r1
	})
}

func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) nextHook() func(context.Context, int32) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) bppendCbll(r0 UserEmbilsServiceHbsVerifiedEmbilFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UserEmbilsServiceHbsVerifiedEmbilFuncCbll
// objects describing the invocbtions of this function.
func (f *UserEmbilsServiceHbsVerifiedEmbilFunc) History() []UserEmbilsServiceHbsVerifiedEmbilFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceHbsVerifiedEmbilFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceHbsVerifiedEmbilFuncCbll is bn object thbt describes bn
// invocbtion of method HbsVerifiedEmbil on bn instbnce of
// MockUserEmbilsService.
type UserEmbilsServiceHbsVerifiedEmbilFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceHbsVerifiedEmbilFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceHbsVerifiedEmbilFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UserEmbilsServiceRemoveFunc describes the behbvior when the Remove method
// of the pbrent MockUserEmbilsService instbnce is invoked.
type UserEmbilsServiceRemoveFunc struct {
	defbultHook func(context.Context, int32, string) error
	hooks       []func(context.Context, int32, string) error
	history     []UserEmbilsServiceRemoveFuncCbll
	mutex       sync.Mutex
}

// Remove delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) Remove(v0 context.Context, v1 int32, v2 string) error {
	r0 := m.RemoveFunc.nextHook()(v0, v1, v2)
	m.RemoveFunc.bppendCbll(UserEmbilsServiceRemoveFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Remove method of the
// pbrent MockUserEmbilsService instbnce is invoked bnd the hook queue is
// empty.
func (f *UserEmbilsServiceRemoveFunc) SetDefbultHook(hook func(context.Context, int32, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Remove method of the pbrent MockUserEmbilsService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UserEmbilsServiceRemoveFunc) PushHook(hook func(context.Context, int32, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceRemoveFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceRemoveFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string) error {
		return r0
	})
}

func (f *UserEmbilsServiceRemoveFunc) nextHook() func(context.Context, int32, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceRemoveFunc) bppendCbll(r0 UserEmbilsServiceRemoveFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UserEmbilsServiceRemoveFuncCbll objects
// describing the invocbtions of this function.
func (f *UserEmbilsServiceRemoveFunc) History() []UserEmbilsServiceRemoveFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceRemoveFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceRemoveFuncCbll is bn object thbt describes bn invocbtion
// of method Remove on bn instbnce of MockUserEmbilsService.
type UserEmbilsServiceRemoveFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceRemoveFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceRemoveFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UserEmbilsServiceResendVerificbtionEmbilFunc describes the behbvior when
// the ResendVerificbtionEmbil method of the pbrent MockUserEmbilsService
// instbnce is invoked.
type UserEmbilsServiceResendVerificbtionEmbilFunc struct {
	defbultHook func(context.Context, int32, string, time.Time) error
	hooks       []func(context.Context, int32, string, time.Time) error
	history     []UserEmbilsServiceResendVerificbtionEmbilFuncCbll
	mutex       sync.Mutex
}

// ResendVerificbtionEmbil delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) ResendVerificbtionEmbil(v0 context.Context, v1 int32, v2 string, v3 time.Time) error {
	r0 := m.ResendVerificbtionEmbilFunc.nextHook()(v0, v1, v2, v3)
	m.ResendVerificbtionEmbilFunc.bppendCbll(UserEmbilsServiceResendVerificbtionEmbilFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// ResendVerificbtionEmbil method of the pbrent MockUserEmbilsService
// instbnce is invoked bnd the hook queue is empty.
func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) SetDefbultHook(hook func(context.Context, int32, string, time.Time) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResendVerificbtionEmbil method of the pbrent MockUserEmbilsService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) PushHook(hook func(context.Context, int32, string, time.Time) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string, time.Time) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string, time.Time) error {
		return r0
	})
}

func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) nextHook() func(context.Context, int32, string, time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) bppendCbll(r0 UserEmbilsServiceResendVerificbtionEmbilFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UserEmbilsServiceResendVerificbtionEmbilFuncCbll objects describing the
// invocbtions of this function.
func (f *UserEmbilsServiceResendVerificbtionEmbilFunc) History() []UserEmbilsServiceResendVerificbtionEmbilFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceResendVerificbtionEmbilFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceResendVerificbtionEmbilFuncCbll is bn object thbt
// describes bn invocbtion of method ResendVerificbtionEmbil on bn instbnce
// of MockUserEmbilsService.
type UserEmbilsServiceResendVerificbtionEmbilFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceResendVerificbtionEmbilFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceResendVerificbtionEmbilFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc describes the
// behbvior when the SendUserEmbilOnAccessTokenChbnge method of the pbrent
// MockUserEmbilsService instbnce is invoked.
type UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc struct {
	defbultHook func(context.Context, int32, string, bool) error
	hooks       []func(context.Context, int32, string, bool) error
	history     []UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll
	mutex       sync.Mutex
}

// SendUserEmbilOnAccessTokenChbnge delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) SendUserEmbilOnAccessTokenChbnge(v0 context.Context, v1 int32, v2 string, v3 bool) error {
	r0 := m.SendUserEmbilOnAccessTokenChbngeFunc.nextHook()(v0, v1, v2, v3)
	m.SendUserEmbilOnAccessTokenChbngeFunc.bppendCbll(UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// SendUserEmbilOnAccessTokenChbnge method of the pbrent
// MockUserEmbilsService instbnce is invoked bnd the hook queue is empty.
func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) SetDefbultHook(hook func(context.Context, int32, string, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SendUserEmbilOnAccessTokenChbnge method of the pbrent
// MockUserEmbilsService instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) PushHook(hook func(context.Context, int32, string, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string, bool) error {
		return r0
	})
}

func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) nextHook() func(context.Context, int32, string, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) bppendCbll(r0 UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFunc) History() []UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll is bn object
// thbt describes bn invocbtion of method SendUserEmbilOnAccessTokenChbnge
// on bn instbnce of MockUserEmbilsService.
type UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceSendUserEmbilOnAccessTokenChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc describes the behbvior
// when the SendUserEmbilOnFieldUpdbte method of the pbrent
// MockUserEmbilsService instbnce is invoked.
type UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc struct {
	defbultHook func(context.Context, int32, string) error
	hooks       []func(context.Context, int32, string) error
	history     []UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll
	mutex       sync.Mutex
}

// SendUserEmbilOnFieldUpdbte delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) SendUserEmbilOnFieldUpdbte(v0 context.Context, v1 int32, v2 string) error {
	r0 := m.SendUserEmbilOnFieldUpdbteFunc.nextHook()(v0, v1, v2)
	m.SendUserEmbilOnFieldUpdbteFunc.bppendCbll(UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// SendUserEmbilOnFieldUpdbte method of the pbrent MockUserEmbilsService
// instbnce is invoked bnd the hook queue is empty.
func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) SetDefbultHook(hook func(context.Context, int32, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SendUserEmbilOnFieldUpdbte method of the pbrent MockUserEmbilsService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) PushHook(hook func(context.Context, int32, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string) error {
		return r0
	})
}

func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) nextHook() func(context.Context, int32, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) bppendCbll(r0 UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll objects describing
// the invocbtions of this function.
func (f *UserEmbilsServiceSendUserEmbilOnFieldUpdbteFunc) History() []UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll is bn object thbt
// describes bn invocbtion of method SendUserEmbilOnFieldUpdbte on bn
// instbnce of MockUserEmbilsService.
type UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceSendUserEmbilOnFieldUpdbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UserEmbilsServiceSetPrimbryEmbilFunc describes the behbvior when the
// SetPrimbryEmbil method of the pbrent MockUserEmbilsService instbnce is
// invoked.
type UserEmbilsServiceSetPrimbryEmbilFunc struct {
	defbultHook func(context.Context, int32, string) error
	hooks       []func(context.Context, int32, string) error
	history     []UserEmbilsServiceSetPrimbryEmbilFuncCbll
	mutex       sync.Mutex
}

// SetPrimbryEmbil delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) SetPrimbryEmbil(v0 context.Context, v1 int32, v2 string) error {
	r0 := m.SetPrimbryEmbilFunc.nextHook()(v0, v1, v2)
	m.SetPrimbryEmbilFunc.bppendCbll(UserEmbilsServiceSetPrimbryEmbilFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetPrimbryEmbil
// method of the pbrent MockUserEmbilsService instbnce is invoked bnd the
// hook queue is empty.
func (f *UserEmbilsServiceSetPrimbryEmbilFunc) SetDefbultHook(hook func(context.Context, int32, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetPrimbryEmbil method of the pbrent MockUserEmbilsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UserEmbilsServiceSetPrimbryEmbilFunc) PushHook(hook func(context.Context, int32, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceSetPrimbryEmbilFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceSetPrimbryEmbilFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string) error {
		return r0
	})
}

func (f *UserEmbilsServiceSetPrimbryEmbilFunc) nextHook() func(context.Context, int32, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceSetPrimbryEmbilFunc) bppendCbll(r0 UserEmbilsServiceSetPrimbryEmbilFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UserEmbilsServiceSetPrimbryEmbilFuncCbll
// objects describing the invocbtions of this function.
func (f *UserEmbilsServiceSetPrimbryEmbilFunc) History() []UserEmbilsServiceSetPrimbryEmbilFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceSetPrimbryEmbilFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceSetPrimbryEmbilFuncCbll is bn object thbt describes bn
// invocbtion of method SetPrimbryEmbil on bn instbnce of
// MockUserEmbilsService.
type UserEmbilsServiceSetPrimbryEmbilFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceSetPrimbryEmbilFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceSetPrimbryEmbilFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UserEmbilsServiceSetVerifiedFunc describes the behbvior when the
// SetVerified method of the pbrent MockUserEmbilsService instbnce is
// invoked.
type UserEmbilsServiceSetVerifiedFunc struct {
	defbultHook func(context.Context, int32, string, bool) error
	hooks       []func(context.Context, int32, string, bool) error
	history     []UserEmbilsServiceSetVerifiedFuncCbll
	mutex       sync.Mutex
}

// SetVerified delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUserEmbilsService) SetVerified(v0 context.Context, v1 int32, v2 string, v3 bool) error {
	r0 := m.SetVerifiedFunc.nextHook()(v0, v1, v2, v3)
	m.SetVerifiedFunc.bppendCbll(UserEmbilsServiceSetVerifiedFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetVerified method
// of the pbrent MockUserEmbilsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UserEmbilsServiceSetVerifiedFunc) SetDefbultHook(hook func(context.Context, int32, string, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetVerified method of the pbrent MockUserEmbilsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UserEmbilsServiceSetVerifiedFunc) PushHook(hook func(context.Context, int32, string, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UserEmbilsServiceSetVerifiedFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int32, string, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UserEmbilsServiceSetVerifiedFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, string, bool) error {
		return r0
	})
}

func (f *UserEmbilsServiceSetVerifiedFunc) nextHook() func(context.Context, int32, string, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UserEmbilsServiceSetVerifiedFunc) bppendCbll(r0 UserEmbilsServiceSetVerifiedFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UserEmbilsServiceSetVerifiedFuncCbll
// objects describing the invocbtions of this function.
func (f *UserEmbilsServiceSetVerifiedFunc) History() []UserEmbilsServiceSetVerifiedFuncCbll {
	f.mutex.Lock()
	history := mbke([]UserEmbilsServiceSetVerifiedFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UserEmbilsServiceSetVerifiedFuncCbll is bn object thbt describes bn
// invocbtion of method SetVerified on bn instbnce of MockUserEmbilsService.
type UserEmbilsServiceSetVerifiedFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UserEmbilsServiceSetVerifiedFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UserEmbilsServiceSetVerifiedFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
