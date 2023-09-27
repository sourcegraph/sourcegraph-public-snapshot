// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge buth

import (
	"context"
	"sync"

	github "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
)

// MockGitHubClient is b mock implementbtion of the GitHubClient interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/http/buth)
// used for unit testing.
type MockGitHubClient struct {
	// GetRepositoryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRepository.
	GetRepositoryFunc *GitHubClientGetRepositoryFunc
	// ListInstbllbtionRepositoriesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListInstbllbtionRepositories.
	ListInstbllbtionRepositoriesFunc *GitHubClientListInstbllbtionRepositoriesFunc
}

// NewMockGitHubClient crebtes b new mock of the GitHubClient interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		GetRepositoryFunc: &GitHubClientGetRepositoryFunc{
			defbultHook: func(context.Context, string, string) (r0 *github.Repository, r1 error) {
				return
			},
		},
		ListInstbllbtionRepositoriesFunc: &GitHubClientListInstbllbtionRepositoriesFunc{
			defbultHook: func(context.Context, int) (r0 []*github.Repository, r1 bool, r2 int, r3 error) {
				return
			},
		},
	}
}

// NewStrictMockGitHubClient crebtes b new mock of the GitHubClient
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		GetRepositoryFunc: &GitHubClientGetRepositoryFunc{
			defbultHook: func(context.Context, string, string) (*github.Repository, error) {
				pbnic("unexpected invocbtion of MockGitHubClient.GetRepository")
			},
		},
		ListInstbllbtionRepositoriesFunc: &GitHubClientListInstbllbtionRepositoriesFunc{
			defbultHook: func(context.Context, int) ([]*github.Repository, bool, int, error) {
				pbnic("unexpected invocbtion of MockGitHubClient.ListInstbllbtionRepositories")
			},
		},
	}
}

// NewMockGitHubClientFrom crebtes b new mock of the MockGitHubClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGitHubClientFrom(i GitHubClient) *MockGitHubClient {
	return &MockGitHubClient{
		GetRepositoryFunc: &GitHubClientGetRepositoryFunc{
			defbultHook: i.GetRepository,
		},
		ListInstbllbtionRepositoriesFunc: &GitHubClientListInstbllbtionRepositoriesFunc{
			defbultHook: i.ListInstbllbtionRepositories,
		},
	}
}

// GitHubClientGetRepositoryFunc describes the behbvior when the
// GetRepository method of the pbrent MockGitHubClient instbnce is invoked.
type GitHubClientGetRepositoryFunc struct {
	defbultHook func(context.Context, string, string) (*github.Repository, error)
	hooks       []func(context.Context, string, string) (*github.Repository, error)
	history     []GitHubClientGetRepositoryFuncCbll
	mutex       sync.Mutex
}

// GetRepository delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubClient) GetRepository(v0 context.Context, v1 string, v2 string) (*github.Repository, error) {
	r0, r1 := m.GetRepositoryFunc.nextHook()(v0, v1, v2)
	m.GetRepositoryFunc.bppendCbll(GitHubClientGetRepositoryFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetRepository method
// of the pbrent MockGitHubClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubClientGetRepositoryFunc) SetDefbultHook(hook func(context.Context, string, string) (*github.Repository, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepository method of the pbrent MockGitHubClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitHubClientGetRepositoryFunc) PushHook(hook func(context.Context, string, string) (*github.Repository, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubClientGetRepositoryFunc) SetDefbultReturn(r0 *github.Repository, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string) (*github.Repository, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubClientGetRepositoryFunc) PushReturn(r0 *github.Repository, r1 error) {
	f.PushHook(func(context.Context, string, string) (*github.Repository, error) {
		return r0, r1
	})
}

func (f *GitHubClientGetRepositoryFunc) nextHook() func(context.Context, string, string) (*github.Repository, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubClientGetRepositoryFunc) bppendCbll(r0 GitHubClientGetRepositoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubClientGetRepositoryFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubClientGetRepositoryFunc) History() []GitHubClientGetRepositoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubClientGetRepositoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubClientGetRepositoryFuncCbll is bn object thbt describes bn
// invocbtion of method GetRepository on bn instbnce of MockGitHubClient.
type GitHubClientGetRepositoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *github.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubClientGetRepositoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubClientGetRepositoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubClientListInstbllbtionRepositoriesFunc describes the behbvior when
// the ListInstbllbtionRepositories method of the pbrent MockGitHubClient
// instbnce is invoked.
type GitHubClientListInstbllbtionRepositoriesFunc struct {
	defbultHook func(context.Context, int) ([]*github.Repository, bool, int, error)
	hooks       []func(context.Context, int) ([]*github.Repository, bool, int, error)
	history     []GitHubClientListInstbllbtionRepositoriesFuncCbll
	mutex       sync.Mutex
}

// ListInstbllbtionRepositories delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubClient) ListInstbllbtionRepositories(v0 context.Context, v1 int) ([]*github.Repository, bool, int, error) {
	r0, r1, r2, r3 := m.ListInstbllbtionRepositoriesFunc.nextHook()(v0, v1)
	m.ListInstbllbtionRepositoriesFunc.bppendCbll(GitHubClientListInstbllbtionRepositoriesFuncCbll{v0, v1, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// ListInstbllbtionRepositories method of the pbrent MockGitHubClient
// instbnce is invoked bnd the hook queue is empty.
func (f *GitHubClientListInstbllbtionRepositoriesFunc) SetDefbultHook(hook func(context.Context, int) ([]*github.Repository, bool, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListInstbllbtionRepositories method of the pbrent MockGitHubClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *GitHubClientListInstbllbtionRepositoriesFunc) PushHook(hook func(context.Context, int) ([]*github.Repository, bool, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubClientListInstbllbtionRepositoriesFunc) SetDefbultReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, int) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubClientListInstbllbtionRepositoriesFunc) PushReturn(r0 []*github.Repository, r1 bool, r2 int, r3 error) {
	f.PushHook(func(context.Context, int) ([]*github.Repository, bool, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *GitHubClientListInstbllbtionRepositoriesFunc) nextHook() func(context.Context, int) ([]*github.Repository, bool, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubClientListInstbllbtionRepositoriesFunc) bppendCbll(r0 GitHubClientListInstbllbtionRepositoriesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitHubClientListInstbllbtionRepositoriesFuncCbll objects describing the
// invocbtions of this function.
func (f *GitHubClientListInstbllbtionRepositoriesFunc) History() []GitHubClientListInstbllbtionRepositoriesFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubClientListInstbllbtionRepositoriesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubClientListInstbllbtionRepositoriesFuncCbll is bn object thbt
// describes bn invocbtion of method ListInstbllbtionRepositories on bn
// instbnce of MockGitHubClient.
type GitHubClientListInstbllbtionRepositoriesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*github.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubClientListInstbllbtionRepositoriesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubClientListInstbllbtionRepositoriesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}
