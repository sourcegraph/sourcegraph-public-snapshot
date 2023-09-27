// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge store

import (
	"context"
	"sync"

	log "github.com/sourcegrbph/log"
	encryption "github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	types "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	types1 "github.com/sourcegrbph/sourcegrbph/internbl/types"
	errors "github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// MockGitHubAppsStore is b mock implementbtion of the GitHubAppsStore
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store) used for
// unit testing.
type MockGitHubAppsStore struct {
	// BulkRemoveInstbllbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method BulkRemoveInstbllbtions.
	BulkRemoveInstbllbtionsFunc *GitHubAppsStoreBulkRemoveInstbllbtionsFunc
	// CrebteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Crebte.
	CrebteFunc *GitHubAppsStoreCrebteFunc
	// DeleteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Delete.
	DeleteFunc *GitHubAppsStoreDeleteFunc
	// GetByAppIDFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetByAppID.
	GetByAppIDFunc *GitHubAppsStoreGetByAppIDFunc
	// GetByDombinFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetByDombin.
	GetByDombinFunc *GitHubAppsStoreGetByDombinFunc
	// GetByIDFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetByID.
	GetByIDFunc *GitHubAppsStoreGetByIDFunc
	// GetBySlugFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetBySlug.
	GetBySlugFunc *GitHubAppsStoreGetBySlugFunc
	// GetInstbllIDFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetInstbllID.
	GetInstbllIDFunc *GitHubAppsStoreGetInstbllIDFunc
	// GetInstbllbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetInstbllbtions.
	GetInstbllbtionsFunc *GitHubAppsStoreGetInstbllbtionsFunc
	// InstbllFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Instbll.
	InstbllFunc *GitHubAppsStoreInstbllFunc
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *GitHubAppsStoreListFunc
	// SyncInstbllbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SyncInstbllbtions.
	SyncInstbllbtionsFunc *GitHubAppsStoreSyncInstbllbtionsFunc
	// UpdbteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Updbte.
	UpdbteFunc *GitHubAppsStoreUpdbteFunc
	// WithEncryptionKeyFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithEncryptionKey.
	WithEncryptionKeyFunc *GitHubAppsStoreWithEncryptionKeyFunc
}

// NewMockGitHubAppsStore crebtes b new mock of the GitHubAppsStore
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGitHubAppsStore() *MockGitHubAppsStore {
	return &MockGitHubAppsStore{
		BulkRemoveInstbllbtionsFunc: &GitHubAppsStoreBulkRemoveInstbllbtionsFunc{
			defbultHook: func(context.Context, int, []int) (r0 error) {
				return
			},
		},
		CrebteFunc: &GitHubAppsStoreCrebteFunc{
			defbultHook: func(context.Context, *types.GitHubApp) (r0 int, r1 error) {
				return
			},
		},
		DeleteFunc: &GitHubAppsStoreDeleteFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
		GetByAppIDFunc: &GitHubAppsStoreGetByAppIDFunc{
			defbultHook: func(context.Context, int, string) (r0 *types.GitHubApp, r1 error) {
				return
			},
		},
		GetByDombinFunc: &GitHubAppsStoreGetByDombinFunc{
			defbultHook: func(context.Context, types1.GitHubAppDombin, string) (r0 *types.GitHubApp, r1 error) {
				return
			},
		},
		GetByIDFunc: &GitHubAppsStoreGetByIDFunc{
			defbultHook: func(context.Context, int) (r0 *types.GitHubApp, r1 error) {
				return
			},
		},
		GetBySlugFunc: &GitHubAppsStoreGetBySlugFunc{
			defbultHook: func(context.Context, string, string) (r0 *types.GitHubApp, r1 error) {
				return
			},
		},
		GetInstbllIDFunc: &GitHubAppsStoreGetInstbllIDFunc{
			defbultHook: func(context.Context, int, string) (r0 int, r1 error) {
				return
			},
		},
		GetInstbllbtionsFunc: &GitHubAppsStoreGetInstbllbtionsFunc{
			defbultHook: func(context.Context, int) (r0 []*types.GitHubAppInstbllbtion, r1 error) {
				return
			},
		},
		InstbllFunc: &GitHubAppsStoreInstbllFunc{
			defbultHook: func(context.Context, types.GitHubAppInstbllbtion) (r0 *types.GitHubAppInstbllbtion, r1 error) {
				return
			},
		},
		ListFunc: &GitHubAppsStoreListFunc{
			defbultHook: func(context.Context, *types1.GitHubAppDombin) (r0 []*types.GitHubApp, r1 error) {
				return
			},
		},
		SyncInstbllbtionsFunc: &GitHubAppsStoreSyncInstbllbtionsFunc{
			defbultHook: func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) (r0 errors.MultiError) {
				return
			},
		},
		UpdbteFunc: &GitHubAppsStoreUpdbteFunc{
			defbultHook: func(context.Context, int, *types.GitHubApp) (r0 *types.GitHubApp, r1 error) {
				return
			},
		},
		WithEncryptionKeyFunc: &GitHubAppsStoreWithEncryptionKeyFunc{
			defbultHook: func(encryption.Key) (r0 GitHubAppsStore) {
				return
			},
		},
	}
}

// NewStrictMockGitHubAppsStore crebtes b new mock of the GitHubAppsStore
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGitHubAppsStore() *MockGitHubAppsStore {
	return &MockGitHubAppsStore{
		BulkRemoveInstbllbtionsFunc: &GitHubAppsStoreBulkRemoveInstbllbtionsFunc{
			defbultHook: func(context.Context, int, []int) error {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.BulkRemoveInstbllbtions")
			},
		},
		CrebteFunc: &GitHubAppsStoreCrebteFunc{
			defbultHook: func(context.Context, *types.GitHubApp) (int, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.Crebte")
			},
		},
		DeleteFunc: &GitHubAppsStoreDeleteFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.Delete")
			},
		},
		GetByAppIDFunc: &GitHubAppsStoreGetByAppIDFunc{
			defbultHook: func(context.Context, int, string) (*types.GitHubApp, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.GetByAppID")
			},
		},
		GetByDombinFunc: &GitHubAppsStoreGetByDombinFunc{
			defbultHook: func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.GetByDombin")
			},
		},
		GetByIDFunc: &GitHubAppsStoreGetByIDFunc{
			defbultHook: func(context.Context, int) (*types.GitHubApp, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.GetByID")
			},
		},
		GetBySlugFunc: &GitHubAppsStoreGetBySlugFunc{
			defbultHook: func(context.Context, string, string) (*types.GitHubApp, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.GetBySlug")
			},
		},
		GetInstbllIDFunc: &GitHubAppsStoreGetInstbllIDFunc{
			defbultHook: func(context.Context, int, string) (int, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.GetInstbllID")
			},
		},
		GetInstbllbtionsFunc: &GitHubAppsStoreGetInstbllbtionsFunc{
			defbultHook: func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.GetInstbllbtions")
			},
		},
		InstbllFunc: &GitHubAppsStoreInstbllFunc{
			defbultHook: func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.Instbll")
			},
		},
		ListFunc: &GitHubAppsStoreListFunc{
			defbultHook: func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.List")
			},
		},
		SyncInstbllbtionsFunc: &GitHubAppsStoreSyncInstbllbtionsFunc{
			defbultHook: func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.SyncInstbllbtions")
			},
		},
		UpdbteFunc: &GitHubAppsStoreUpdbteFunc{
			defbultHook: func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error) {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.Updbte")
			},
		},
		WithEncryptionKeyFunc: &GitHubAppsStoreWithEncryptionKeyFunc{
			defbultHook: func(encryption.Key) GitHubAppsStore {
				pbnic("unexpected invocbtion of MockGitHubAppsStore.WithEncryptionKey")
			},
		},
	}
}

// NewMockGitHubAppsStoreFrom crebtes b new mock of the MockGitHubAppsStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGitHubAppsStoreFrom(i GitHubAppsStore) *MockGitHubAppsStore {
	return &MockGitHubAppsStore{
		BulkRemoveInstbllbtionsFunc: &GitHubAppsStoreBulkRemoveInstbllbtionsFunc{
			defbultHook: i.BulkRemoveInstbllbtions,
		},
		CrebteFunc: &GitHubAppsStoreCrebteFunc{
			defbultHook: i.Crebte,
		},
		DeleteFunc: &GitHubAppsStoreDeleteFunc{
			defbultHook: i.Delete,
		},
		GetByAppIDFunc: &GitHubAppsStoreGetByAppIDFunc{
			defbultHook: i.GetByAppID,
		},
		GetByDombinFunc: &GitHubAppsStoreGetByDombinFunc{
			defbultHook: i.GetByDombin,
		},
		GetByIDFunc: &GitHubAppsStoreGetByIDFunc{
			defbultHook: i.GetByID,
		},
		GetBySlugFunc: &GitHubAppsStoreGetBySlugFunc{
			defbultHook: i.GetBySlug,
		},
		GetInstbllIDFunc: &GitHubAppsStoreGetInstbllIDFunc{
			defbultHook: i.GetInstbllID,
		},
		GetInstbllbtionsFunc: &GitHubAppsStoreGetInstbllbtionsFunc{
			defbultHook: i.GetInstbllbtions,
		},
		InstbllFunc: &GitHubAppsStoreInstbllFunc{
			defbultHook: i.Instbll,
		},
		ListFunc: &GitHubAppsStoreListFunc{
			defbultHook: i.List,
		},
		SyncInstbllbtionsFunc: &GitHubAppsStoreSyncInstbllbtionsFunc{
			defbultHook: i.SyncInstbllbtions,
		},
		UpdbteFunc: &GitHubAppsStoreUpdbteFunc{
			defbultHook: i.Updbte,
		},
		WithEncryptionKeyFunc: &GitHubAppsStoreWithEncryptionKeyFunc{
			defbultHook: i.WithEncryptionKey,
		},
	}
}

// GitHubAppsStoreBulkRemoveInstbllbtionsFunc describes the behbvior when
// the BulkRemoveInstbllbtions method of the pbrent MockGitHubAppsStore
// instbnce is invoked.
type GitHubAppsStoreBulkRemoveInstbllbtionsFunc struct {
	defbultHook func(context.Context, int, []int) error
	hooks       []func(context.Context, int, []int) error
	history     []GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll
	mutex       sync.Mutex
}

// BulkRemoveInstbllbtions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) BulkRemoveInstbllbtions(v0 context.Context, v1 int, v2 []int) error {
	r0 := m.BulkRemoveInstbllbtionsFunc.nextHook()(v0, v1, v2)
	m.BulkRemoveInstbllbtionsFunc.bppendCbll(GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// BulkRemoveInstbllbtions method of the pbrent MockGitHubAppsStore instbnce
// is invoked bnd the hook queue is empty.
func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) SetDefbultHook(hook func(context.Context, int, []int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BulkRemoveInstbllbtions method of the pbrent MockGitHubAppsStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) PushHook(hook func(context.Context, int, []int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, []int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, []int) error {
		return r0
	})
}

func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) nextHook() func(context.Context, int, []int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) bppendCbll(r0 GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll objects describing the
// invocbtions of this function.
func (f *GitHubAppsStoreBulkRemoveInstbllbtionsFunc) History() []GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll is bn object thbt
// describes bn invocbtion of method BulkRemoveInstbllbtions on bn instbnce
// of MockGitHubAppsStore.
type GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreBulkRemoveInstbllbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitHubAppsStoreCrebteFunc describes the behbvior when the Crebte method
// of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreCrebteFunc struct {
	defbultHook func(context.Context, *types.GitHubApp) (int, error)
	hooks       []func(context.Context, *types.GitHubApp) (int, error)
	history     []GitHubAppsStoreCrebteFuncCbll
	mutex       sync.Mutex
}

// Crebte delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) Crebte(v0 context.Context, v1 *types.GitHubApp) (int, error) {
	r0, r1 := m.CrebteFunc.nextHook()(v0, v1)
	m.CrebteFunc.bppendCbll(GitHubAppsStoreCrebteFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Crebte method of the
// pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreCrebteFunc) SetDefbultHook(hook func(context.Context, *types.GitHubApp) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Crebte method of the pbrent MockGitHubAppsStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreCrebteFunc) PushHook(hook func(context.Context, *types.GitHubApp) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreCrebteFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.GitHubApp) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreCrebteFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, *types.GitHubApp) (int, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreCrebteFunc) nextHook() func(context.Context, *types.GitHubApp) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreCrebteFunc) bppendCbll(r0 GitHubAppsStoreCrebteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreCrebteFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreCrebteFunc) History() []GitHubAppsStoreCrebteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreCrebteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreCrebteFuncCbll is bn object thbt describes bn invocbtion
// of method Crebte on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreCrebteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.GitHubApp
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreCrebteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreCrebteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreDeleteFunc describes the behbvior when the Delete method
// of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreDeleteFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []GitHubAppsStoreDeleteFuncCbll
	mutex       sync.Mutex
}

// Delete delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) Delete(v0 context.Context, v1 int) error {
	r0 := m.DeleteFunc.nextHook()(v0, v1)
	m.DeleteFunc.bppendCbll(GitHubAppsStoreDeleteFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Delete method of the
// pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreDeleteFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Delete method of the pbrent MockGitHubAppsStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreDeleteFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreDeleteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreDeleteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *GitHubAppsStoreDeleteFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreDeleteFunc) bppendCbll(r0 GitHubAppsStoreDeleteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreDeleteFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreDeleteFunc) History() []GitHubAppsStoreDeleteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreDeleteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreDeleteFuncCbll is bn object thbt describes bn invocbtion
// of method Delete on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreDeleteFuncCbll struct {
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
func (c GitHubAppsStoreDeleteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreDeleteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitHubAppsStoreGetByAppIDFunc describes the behbvior when the GetByAppID
// method of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreGetByAppIDFunc struct {
	defbultHook func(context.Context, int, string) (*types.GitHubApp, error)
	hooks       []func(context.Context, int, string) (*types.GitHubApp, error)
	history     []GitHubAppsStoreGetByAppIDFuncCbll
	mutex       sync.Mutex
}

// GetByAppID delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) GetByAppID(v0 context.Context, v1 int, v2 string) (*types.GitHubApp, error) {
	r0, r1 := m.GetByAppIDFunc.nextHook()(v0, v1, v2)
	m.GetByAppIDFunc.bppendCbll(GitHubAppsStoreGetByAppIDFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByAppID method of
// the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreGetByAppIDFunc) SetDefbultHook(hook func(context.Context, int, string) (*types.GitHubApp, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByAppID method of the pbrent MockGitHubAppsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreGetByAppIDFunc) PushHook(hook func(context.Context, int, string) (*types.GitHubApp, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreGetByAppIDFunc) SetDefbultReturn(r0 *types.GitHubApp, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string) (*types.GitHubApp, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreGetByAppIDFunc) PushReturn(r0 *types.GitHubApp, r1 error) {
	f.PushHook(func(context.Context, int, string) (*types.GitHubApp, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreGetByAppIDFunc) nextHook() func(context.Context, int, string) (*types.GitHubApp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreGetByAppIDFunc) bppendCbll(r0 GitHubAppsStoreGetByAppIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreGetByAppIDFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreGetByAppIDFunc) History() []GitHubAppsStoreGetByAppIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreGetByAppIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreGetByAppIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetByAppID on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreGetByAppIDFuncCbll struct {
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
	Result0 *types.GitHubApp
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreGetByAppIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreGetByAppIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreGetByDombinFunc describes the behbvior when the
// GetByDombin method of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreGetByDombinFunc struct {
	defbultHook func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error)
	hooks       []func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error)
	history     []GitHubAppsStoreGetByDombinFuncCbll
	mutex       sync.Mutex
}

// GetByDombin delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) GetByDombin(v0 context.Context, v1 types1.GitHubAppDombin, v2 string) (*types.GitHubApp, error) {
	r0, r1 := m.GetByDombinFunc.nextHook()(v0, v1, v2)
	m.GetByDombinFunc.bppendCbll(GitHubAppsStoreGetByDombinFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByDombin method
// of the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue
// is empty.
func (f *GitHubAppsStoreGetByDombinFunc) SetDefbultHook(hook func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByDombin method of the pbrent MockGitHubAppsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreGetByDombinFunc) PushHook(hook func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreGetByDombinFunc) SetDefbultReturn(r0 *types.GitHubApp, r1 error) {
	f.SetDefbultHook(func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreGetByDombinFunc) PushReturn(r0 *types.GitHubApp, r1 error) {
	f.PushHook(func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreGetByDombinFunc) nextHook() func(context.Context, types1.GitHubAppDombin, string) (*types.GitHubApp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreGetByDombinFunc) bppendCbll(r0 GitHubAppsStoreGetByDombinFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreGetByDombinFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreGetByDombinFunc) History() []GitHubAppsStoreGetByDombinFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreGetByDombinFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreGetByDombinFuncCbll is bn object thbt describes bn
// invocbtion of method GetByDombin on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreGetByDombinFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types1.GitHubAppDombin
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.GitHubApp
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreGetByDombinFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreGetByDombinFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreGetByIDFunc describes the behbvior when the GetByID method
// of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreGetByIDFunc struct {
	defbultHook func(context.Context, int) (*types.GitHubApp, error)
	hooks       []func(context.Context, int) (*types.GitHubApp, error)
	history     []GitHubAppsStoreGetByIDFuncCbll
	mutex       sync.Mutex
}

// GetByID delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) GetByID(v0 context.Context, v1 int) (*types.GitHubApp, error) {
	r0, r1 := m.GetByIDFunc.nextHook()(v0, v1)
	m.GetByIDFunc.bppendCbll(GitHubAppsStoreGetByIDFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByID method of
// the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreGetByIDFunc) SetDefbultHook(hook func(context.Context, int) (*types.GitHubApp, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByID method of the pbrent MockGitHubAppsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreGetByIDFunc) PushHook(hook func(context.Context, int) (*types.GitHubApp, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreGetByIDFunc) SetDefbultReturn(r0 *types.GitHubApp, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (*types.GitHubApp, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreGetByIDFunc) PushReturn(r0 *types.GitHubApp, r1 error) {
	f.PushHook(func(context.Context, int) (*types.GitHubApp, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreGetByIDFunc) nextHook() func(context.Context, int) (*types.GitHubApp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreGetByIDFunc) bppendCbll(r0 GitHubAppsStoreGetByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreGetByIDFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreGetByIDFunc) History() []GitHubAppsStoreGetByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreGetByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreGetByIDFuncCbll is bn object thbt describes bn invocbtion
// of method GetByID on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreGetByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.GitHubApp
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreGetByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreGetByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreGetBySlugFunc describes the behbvior when the GetBySlug
// method of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreGetBySlugFunc struct {
	defbultHook func(context.Context, string, string) (*types.GitHubApp, error)
	hooks       []func(context.Context, string, string) (*types.GitHubApp, error)
	history     []GitHubAppsStoreGetBySlugFuncCbll
	mutex       sync.Mutex
}

// GetBySlug delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) GetBySlug(v0 context.Context, v1 string, v2 string) (*types.GitHubApp, error) {
	r0, r1 := m.GetBySlugFunc.nextHook()(v0, v1, v2)
	m.GetBySlugFunc.bppendCbll(GitHubAppsStoreGetBySlugFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBySlug method of
// the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreGetBySlugFunc) SetDefbultHook(hook func(context.Context, string, string) (*types.GitHubApp, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBySlug method of the pbrent MockGitHubAppsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreGetBySlugFunc) PushHook(hook func(context.Context, string, string) (*types.GitHubApp, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreGetBySlugFunc) SetDefbultReturn(r0 *types.GitHubApp, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string) (*types.GitHubApp, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreGetBySlugFunc) PushReturn(r0 *types.GitHubApp, r1 error) {
	f.PushHook(func(context.Context, string, string) (*types.GitHubApp, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreGetBySlugFunc) nextHook() func(context.Context, string, string) (*types.GitHubApp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreGetBySlugFunc) bppendCbll(r0 GitHubAppsStoreGetBySlugFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreGetBySlugFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreGetBySlugFunc) History() []GitHubAppsStoreGetBySlugFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreGetBySlugFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreGetBySlugFuncCbll is bn object thbt describes bn
// invocbtion of method GetBySlug on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreGetBySlugFuncCbll struct {
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
	Result0 *types.GitHubApp
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreGetBySlugFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreGetBySlugFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreGetInstbllIDFunc describes the behbvior when the
// GetInstbllID method of the pbrent MockGitHubAppsStore instbnce is
// invoked.
type GitHubAppsStoreGetInstbllIDFunc struct {
	defbultHook func(context.Context, int, string) (int, error)
	hooks       []func(context.Context, int, string) (int, error)
	history     []GitHubAppsStoreGetInstbllIDFuncCbll
	mutex       sync.Mutex
}

// GetInstbllID delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) GetInstbllID(v0 context.Context, v1 int, v2 string) (int, error) {
	r0, r1 := m.GetInstbllIDFunc.nextHook()(v0, v1, v2)
	m.GetInstbllIDFunc.bppendCbll(GitHubAppsStoreGetInstbllIDFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetInstbllID method
// of the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue
// is empty.
func (f *GitHubAppsStoreGetInstbllIDFunc) SetDefbultHook(hook func(context.Context, int, string) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetInstbllID method of the pbrent MockGitHubAppsStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreGetInstbllIDFunc) PushHook(hook func(context.Context, int, string) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreGetInstbllIDFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreGetInstbllIDFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, int, string) (int, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreGetInstbllIDFunc) nextHook() func(context.Context, int, string) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreGetInstbllIDFunc) bppendCbll(r0 GitHubAppsStoreGetInstbllIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreGetInstbllIDFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreGetInstbllIDFunc) History() []GitHubAppsStoreGetInstbllIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreGetInstbllIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreGetInstbllIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetInstbllID on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreGetInstbllIDFuncCbll struct {
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
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreGetInstbllIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreGetInstbllIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreGetInstbllbtionsFunc describes the behbvior when the
// GetInstbllbtions method of the pbrent MockGitHubAppsStore instbnce is
// invoked.
type GitHubAppsStoreGetInstbllbtionsFunc struct {
	defbultHook func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error)
	hooks       []func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error)
	history     []GitHubAppsStoreGetInstbllbtionsFuncCbll
	mutex       sync.Mutex
}

// GetInstbllbtions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) GetInstbllbtions(v0 context.Context, v1 int) ([]*types.GitHubAppInstbllbtion, error) {
	r0, r1 := m.GetInstbllbtionsFunc.nextHook()(v0, v1)
	m.GetInstbllbtionsFunc.bppendCbll(GitHubAppsStoreGetInstbllbtionsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetInstbllbtions
// method of the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook
// queue is empty.
func (f *GitHubAppsStoreGetInstbllbtionsFunc) SetDefbultHook(hook func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetInstbllbtions method of the pbrent MockGitHubAppsStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitHubAppsStoreGetInstbllbtionsFunc) PushHook(hook func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreGetInstbllbtionsFunc) SetDefbultReturn(r0 []*types.GitHubAppInstbllbtion, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreGetInstbllbtionsFunc) PushReturn(r0 []*types.GitHubAppInstbllbtion, r1 error) {
	f.PushHook(func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreGetInstbllbtionsFunc) nextHook() func(context.Context, int) ([]*types.GitHubAppInstbllbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreGetInstbllbtionsFunc) bppendCbll(r0 GitHubAppsStoreGetInstbllbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreGetInstbllbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *GitHubAppsStoreGetInstbllbtionsFunc) History() []GitHubAppsStoreGetInstbllbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreGetInstbllbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreGetInstbllbtionsFuncCbll is bn object thbt describes bn
// invocbtion of method GetInstbllbtions on bn instbnce of
// MockGitHubAppsStore.
type GitHubAppsStoreGetInstbllbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.GitHubAppInstbllbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreGetInstbllbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreGetInstbllbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreInstbllFunc describes the behbvior when the Instbll method
// of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreInstbllFunc struct {
	defbultHook func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error)
	hooks       []func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error)
	history     []GitHubAppsStoreInstbllFuncCbll
	mutex       sync.Mutex
}

// Instbll delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) Instbll(v0 context.Context, v1 types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error) {
	r0, r1 := m.InstbllFunc.nextHook()(v0, v1)
	m.InstbllFunc.bppendCbll(GitHubAppsStoreInstbllFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Instbll method of
// the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreInstbllFunc) SetDefbultHook(hook func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Instbll method of the pbrent MockGitHubAppsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreInstbllFunc) PushHook(hook func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreInstbllFunc) SetDefbultReturn(r0 *types.GitHubAppInstbllbtion, r1 error) {
	f.SetDefbultHook(func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreInstbllFunc) PushReturn(r0 *types.GitHubAppInstbllbtion, r1 error) {
	f.PushHook(func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreInstbllFunc) nextHook() func(context.Context, types.GitHubAppInstbllbtion) (*types.GitHubAppInstbllbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreInstbllFunc) bppendCbll(r0 GitHubAppsStoreInstbllFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreInstbllFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreInstbllFunc) History() []GitHubAppsStoreInstbllFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreInstbllFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreInstbllFuncCbll is bn object thbt describes bn invocbtion
// of method Instbll on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreInstbllFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.GitHubAppInstbllbtion
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.GitHubAppInstbllbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreInstbllFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreInstbllFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreListFunc describes the behbvior when the List method of
// the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreListFunc struct {
	defbultHook func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error)
	hooks       []func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error)
	history     []GitHubAppsStoreListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) List(v0 context.Context, v1 *types1.GitHubAppDombin) ([]*types.GitHubApp, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.bppendCbll(GitHubAppsStoreListFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreListFunc) SetDefbultHook(hook func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockGitHubAppsStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreListFunc) PushHook(hook func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreListFunc) SetDefbultReturn(r0 []*types.GitHubApp, r1 error) {
	f.SetDefbultHook(func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreListFunc) PushReturn(r0 []*types.GitHubApp, r1 error) {
	f.PushHook(func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreListFunc) nextHook() func(context.Context, *types1.GitHubAppDombin) ([]*types.GitHubApp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreListFunc) bppendCbll(r0 GitHubAppsStoreListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreListFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreListFunc) History() []GitHubAppsStoreListFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreListFuncCbll is bn object thbt describes bn invocbtion of
// method List on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types1.GitHubAppDombin
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.GitHubApp
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreSyncInstbllbtionsFunc describes the behbvior when the
// SyncInstbllbtions method of the pbrent MockGitHubAppsStore instbnce is
// invoked.
type GitHubAppsStoreSyncInstbllbtionsFunc struct {
	defbultHook func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError
	hooks       []func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError
	history     []GitHubAppsStoreSyncInstbllbtionsFuncCbll
	mutex       sync.Mutex
}

// SyncInstbllbtions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) SyncInstbllbtions(v0 context.Context, v1 types.GitHubApp, v2 log.Logger, v3 types.GitHubAppClient) errors.MultiError {
	r0 := m.SyncInstbllbtionsFunc.nextHook()(v0, v1, v2, v3)
	m.SyncInstbllbtionsFunc.bppendCbll(GitHubAppsStoreSyncInstbllbtionsFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SyncInstbllbtions
// method of the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook
// queue is empty.
func (f *GitHubAppsStoreSyncInstbllbtionsFunc) SetDefbultHook(hook func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SyncInstbllbtions method of the pbrent MockGitHubAppsStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitHubAppsStoreSyncInstbllbtionsFunc) PushHook(hook func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreSyncInstbllbtionsFunc) SetDefbultReturn(r0 errors.MultiError) {
	f.SetDefbultHook(func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreSyncInstbllbtionsFunc) PushReturn(r0 errors.MultiError) {
	f.PushHook(func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError {
		return r0
	})
}

func (f *GitHubAppsStoreSyncInstbllbtionsFunc) nextHook() func(context.Context, types.GitHubApp, log.Logger, types.GitHubAppClient) errors.MultiError {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreSyncInstbllbtionsFunc) bppendCbll(r0 GitHubAppsStoreSyncInstbllbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreSyncInstbllbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *GitHubAppsStoreSyncInstbllbtionsFunc) History() []GitHubAppsStoreSyncInstbllbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreSyncInstbllbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreSyncInstbllbtionsFuncCbll is bn object thbt describes bn
// invocbtion of method SyncInstbllbtions on bn instbnce of
// MockGitHubAppsStore.
type GitHubAppsStoreSyncInstbllbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.GitHubApp
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 log.Logger
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 types.GitHubAppClient
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 errors.MultiError
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreSyncInstbllbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreSyncInstbllbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitHubAppsStoreUpdbteFunc describes the behbvior when the Updbte method
// of the pbrent MockGitHubAppsStore instbnce is invoked.
type GitHubAppsStoreUpdbteFunc struct {
	defbultHook func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error)
	hooks       []func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error)
	history     []GitHubAppsStoreUpdbteFuncCbll
	mutex       sync.Mutex
}

// Updbte delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) Updbte(v0 context.Context, v1 int, v2 *types.GitHubApp) (*types.GitHubApp, error) {
	r0, r1 := m.UpdbteFunc.nextHook()(v0, v1, v2)
	m.UpdbteFunc.bppendCbll(GitHubAppsStoreUpdbteFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Updbte method of the
// pbrent MockGitHubAppsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *GitHubAppsStoreUpdbteFunc) SetDefbultHook(hook func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Updbte method of the pbrent MockGitHubAppsStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitHubAppsStoreUpdbteFunc) PushHook(hook func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreUpdbteFunc) SetDefbultReturn(r0 *types.GitHubApp, r1 error) {
	f.SetDefbultHook(func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreUpdbteFunc) PushReturn(r0 *types.GitHubApp, r1 error) {
	f.PushHook(func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error) {
		return r0, r1
	})
}

func (f *GitHubAppsStoreUpdbteFunc) nextHook() func(context.Context, int, *types.GitHubApp) (*types.GitHubApp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreUpdbteFunc) bppendCbll(r0 GitHubAppsStoreUpdbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreUpdbteFuncCbll objects
// describing the invocbtions of this function.
func (f *GitHubAppsStoreUpdbteFunc) History() []GitHubAppsStoreUpdbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreUpdbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreUpdbteFuncCbll is bn object thbt describes bn invocbtion
// of method Updbte on bn instbnce of MockGitHubAppsStore.
type GitHubAppsStoreUpdbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *types.GitHubApp
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.GitHubApp
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreUpdbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreUpdbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitHubAppsStoreWithEncryptionKeyFunc describes the behbvior when the
// WithEncryptionKey method of the pbrent MockGitHubAppsStore instbnce is
// invoked.
type GitHubAppsStoreWithEncryptionKeyFunc struct {
	defbultHook func(encryption.Key) GitHubAppsStore
	hooks       []func(encryption.Key) GitHubAppsStore
	history     []GitHubAppsStoreWithEncryptionKeyFuncCbll
	mutex       sync.Mutex
}

// WithEncryptionKey delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitHubAppsStore) WithEncryptionKey(v0 encryption.Key) GitHubAppsStore {
	r0 := m.WithEncryptionKeyFunc.nextHook()(v0)
	m.WithEncryptionKeyFunc.bppendCbll(GitHubAppsStoreWithEncryptionKeyFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithEncryptionKey
// method of the pbrent MockGitHubAppsStore instbnce is invoked bnd the hook
// queue is empty.
func (f *GitHubAppsStoreWithEncryptionKeyFunc) SetDefbultHook(hook func(encryption.Key) GitHubAppsStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithEncryptionKey method of the pbrent MockGitHubAppsStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitHubAppsStoreWithEncryptionKeyFunc) PushHook(hook func(encryption.Key) GitHubAppsStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitHubAppsStoreWithEncryptionKeyFunc) SetDefbultReturn(r0 GitHubAppsStore) {
	f.SetDefbultHook(func(encryption.Key) GitHubAppsStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitHubAppsStoreWithEncryptionKeyFunc) PushReturn(r0 GitHubAppsStore) {
	f.PushHook(func(encryption.Key) GitHubAppsStore {
		return r0
	})
}

func (f *GitHubAppsStoreWithEncryptionKeyFunc) nextHook() func(encryption.Key) GitHubAppsStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitHubAppsStoreWithEncryptionKeyFunc) bppendCbll(r0 GitHubAppsStoreWithEncryptionKeyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitHubAppsStoreWithEncryptionKeyFuncCbll
// objects describing the invocbtions of this function.
func (f *GitHubAppsStoreWithEncryptionKeyFunc) History() []GitHubAppsStoreWithEncryptionKeyFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitHubAppsStoreWithEncryptionKeyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitHubAppsStoreWithEncryptionKeyFuncCbll is bn object thbt describes bn
// invocbtion of method WithEncryptionKey on bn instbnce of
// MockGitHubAppsStore.
type GitHubAppsStoreWithEncryptionKeyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 encryption.Key
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 GitHubAppsStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitHubAppsStoreWithEncryptionKeyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitHubAppsStoreWithEncryptionKeyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
