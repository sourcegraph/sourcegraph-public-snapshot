// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge discovery

import (
	"context"
	"sync"

	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// MockIndexbbleReposLister is b mock implementbtion of the
// IndexbbleReposLister interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery) used for
// unit testing.
type MockIndexbbleReposLister struct {
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *IndexbbleReposListerListFunc
}

// NewMockIndexbbleReposLister crebtes b new mock of the
// IndexbbleReposLister interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockIndexbbleReposLister() *MockIndexbbleReposLister {
	return &MockIndexbbleReposLister{
		ListFunc: &IndexbbleReposListerListFunc{
			defbultHook: func(context.Context) (r0 []types.MinimblRepo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockIndexbbleReposLister crebtes b new mock of the
// IndexbbleReposLister interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockIndexbbleReposLister() *MockIndexbbleReposLister {
	return &MockIndexbbleReposLister{
		ListFunc: &IndexbbleReposListerListFunc{
			defbultHook: func(context.Context) ([]types.MinimblRepo, error) {
				pbnic("unexpected invocbtion of MockIndexbbleReposLister.List")
			},
		},
	}
}

// NewMockIndexbbleReposListerFrom crebtes b new mock of the
// MockIndexbbleReposLister interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockIndexbbleReposListerFrom(i IndexbbleReposLister) *MockIndexbbleReposLister {
	return &MockIndexbbleReposLister{
		ListFunc: &IndexbbleReposListerListFunc{
			defbultHook: i.List,
		},
	}
}

// IndexbbleReposListerListFunc describes the behbvior when the List method
// of the pbrent MockIndexbbleReposLister instbnce is invoked.
type IndexbbleReposListerListFunc struct {
	defbultHook func(context.Context) ([]types.MinimblRepo, error)
	hooks       []func(context.Context) ([]types.MinimblRepo, error)
	history     []IndexbbleReposListerListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockIndexbbleReposLister) List(v0 context.Context) ([]types.MinimblRepo, error) {
	r0, r1 := m.ListFunc.nextHook()(v0)
	m.ListFunc.bppendCbll(IndexbbleReposListerListFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockIndexbbleReposLister instbnce is invoked bnd the hook queue is
// empty.
func (f *IndexbbleReposListerListFunc) SetDefbultHook(hook func(context.Context) ([]types.MinimblRepo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockIndexbbleReposLister instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *IndexbbleReposListerListFunc) PushHook(hook func(context.Context) ([]types.MinimblRepo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *IndexbbleReposListerListFunc) SetDefbultReturn(r0 []types.MinimblRepo, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]types.MinimblRepo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *IndexbbleReposListerListFunc) PushReturn(r0 []types.MinimblRepo, r1 error) {
	f.PushHook(func(context.Context) ([]types.MinimblRepo, error) {
		return r0, r1
	})
}

func (f *IndexbbleReposListerListFunc) nextHook() func(context.Context) ([]types.MinimblRepo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *IndexbbleReposListerListFunc) bppendCbll(r0 IndexbbleReposListerListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of IndexbbleReposListerListFuncCbll objects
// describing the invocbtions of this function.
func (f *IndexbbleReposListerListFunc) History() []IndexbbleReposListerListFuncCbll {
	f.mutex.Lock()
	history := mbke([]IndexbbleReposListerListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// IndexbbleReposListerListFuncCbll is bn object thbt describes bn
// invocbtion of method List on bn instbnce of MockIndexbbleReposLister.
type IndexbbleReposListerListFuncCbll struct {
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
func (c IndexbbleReposListerListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c IndexbbleReposListerListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockRepoStore is b mock implementbtion of the RepoStore interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery) used for
// unit testing.
type MockRepoStore struct {
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *RepoStoreListFunc
}

// NewMockRepoStore crebtes b new mock of the RepoStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockRepoStore() *MockRepoStore {
	return &MockRepoStore{
		ListFunc: &RepoStoreListFunc{
			defbultHook: func(context.Context, dbtbbbse.ReposListOptions) (r0 []*types.Repo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockRepoStore crebtes b new mock of the RepoStore interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockRepoStore() *MockRepoStore {
	return &MockRepoStore{
		ListFunc: &RepoStoreListFunc{
			defbultHook: func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
				pbnic("unexpected invocbtion of MockRepoStore.List")
			},
		},
	}
}

// NewMockRepoStoreFrom crebtes b new mock of the MockRepoStore interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockRepoStoreFrom(i RepoStore) *MockRepoStore {
	return &MockRepoStore{
		ListFunc: &RepoStoreListFunc{
			defbultHook: i.List,
		},
	}
}

// RepoStoreListFunc describes the behbvior when the List method of the
// pbrent MockRepoStore instbnce is invoked.
type RepoStoreListFunc struct {
	defbultHook func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)
	hooks       []func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)
	history     []RepoStoreListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoStore) List(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.bppendCbll(RepoStoreListFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockRepoStore instbnce is invoked bnd the hook queue is empty.
func (f *RepoStoreListFunc) SetDefbultHook(hook func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockRepoStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *RepoStoreListFunc) PushHook(hook func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoStoreListFunc) SetDefbultReturn(r0 []*types.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoStoreListFunc) PushReturn(r0 []*types.Repo, r1 error) {
	f.PushHook(func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		return r0, r1
	})
}

func (f *RepoStoreListFunc) nextHook() func(context.Context, dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoStoreListFunc) bppendCbll(r0 RepoStoreListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoStoreListFuncCbll objects describing
// the invocbtions of this function.
func (f *RepoStoreListFunc) History() []RepoStoreListFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoStoreListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoStoreListFuncCbll is bn object thbt describes bn invocbtion of method
// List on bn instbnce of MockRepoStore.
type RepoStoreListFuncCbll struct {
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
func (c RepoStoreListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoStoreListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
