// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge executorqueue

import (
	"context"
	"sync"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// MockGitserverClient is b mock implementbtion of the GitserverClient
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue)
// used for unit testing.
type MockGitserverClient struct {
	// AddrForRepoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method AddrForRepo.
	AddrForRepoFunc *GitserverClientAddrForRepoFunc
}

// NewMockGitserverClient crebtes b new mock of the GitserverClient
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockGitserverClient crebtes b new mock of the GitserverClient
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) string {
				pbnic("unexpected invocbtion of MockGitserverClient.AddrForRepo")
			},
		},
	}
}

// NewMockGitserverClientFrom crebtes b new mock of the MockGitserverClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGitserverClientFrom(i GitserverClient) *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defbultHook: i.AddrForRepo,
		},
	}
}

// GitserverClientAddrForRepoFunc describes the behbvior when the
// AddrForRepo method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientAddrForRepoFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) string
	hooks       []func(context.Context, bpi.RepoNbme) string
	history     []GitserverClientAddrForRepoFuncCbll
	mutex       sync.Mutex
}

// AddrForRepo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) AddrForRepo(v0 context.Context, v1 bpi.RepoNbme) string {
	r0 := m.AddrForRepoFunc.nextHook()(v0, v1)
	m.AddrForRepoFunc.bppendCbll(GitserverClientAddrForRepoFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the AddrForRepo method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientAddrForRepoFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddrForRepo method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientAddrForRepoFunc) PushHook(hook func(context.Context, bpi.RepoNbme) string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientAddrForRepoFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientAddrForRepoFunc) PushReturn(r0 string) {
	f.PushHook(func(context.Context, bpi.RepoNbme) string {
		return r0
	})
}

func (f *GitserverClientAddrForRepoFunc) nextHook() func(context.Context, bpi.RepoNbme) string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientAddrForRepoFunc) bppendCbll(r0 GitserverClientAddrForRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientAddrForRepoFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientAddrForRepoFunc) History() []GitserverClientAddrForRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientAddrForRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientAddrForRepoFuncCbll is bn object thbt describes bn
// invocbtion of method AddrForRepo on bn instbnce of MockGitserverClient.
type GitserverClientAddrForRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientAddrForRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientAddrForRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
