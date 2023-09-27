// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge sources

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"sync"
	"time"

	diff "github.com/sourcegrbph/go-diff/diff"
	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	buthz "github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	store "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	types1 "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	buth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	bzuredevops "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	bitbucketcloud "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	gerrit "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	store1 "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	gitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	gitdombin "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	protocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// MockChbngesetSource is b mock implementbtion of the ChbngesetSource
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources) used for
// unit testing.
type MockChbngesetSource struct {
	// BuildCommitOptsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method BuildCommitOpts.
	BuildCommitOptsFunc *ChbngesetSourceBuildCommitOptsFunc
	// CloseChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CloseChbngeset.
	CloseChbngesetFunc *ChbngesetSourceCloseChbngesetFunc
	// CrebteChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteChbngeset.
	CrebteChbngesetFunc *ChbngesetSourceCrebteChbngesetFunc
	// CrebteCommentFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteComment.
	CrebteCommentFunc *ChbngesetSourceCrebteCommentFunc
	// GitserverPushConfigFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GitserverPushConfig.
	GitserverPushConfigFunc *ChbngesetSourceGitserverPushConfigFunc
	// LobdChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method LobdChbngeset.
	LobdChbngesetFunc *ChbngesetSourceLobdChbngesetFunc
	// MergeChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method MergeChbngeset.
	MergeChbngesetFunc *ChbngesetSourceMergeChbngesetFunc
	// ReopenChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReopenChbngeset.
	ReopenChbngesetFunc *ChbngesetSourceReopenChbngesetFunc
	// UpdbteChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteChbngeset.
	UpdbteChbngesetFunc *ChbngesetSourceUpdbteChbngesetFunc
	// VblidbteAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method VblidbteAuthenticbtor.
	VblidbteAuthenticbtorFunc *ChbngesetSourceVblidbteAuthenticbtorFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *ChbngesetSourceWithAuthenticbtorFunc
}

// NewMockChbngesetSource crebtes b new mock of the ChbngesetSource
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockChbngesetSource() *MockChbngesetSource {
	return &MockChbngesetSource{
		BuildCommitOptsFunc: &ChbngesetSourceBuildCommitOptsFunc{
			defbultHook: func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) (r0 protocol.CrebteCommitFromPbtchRequest) {
				return
			},
		},
		CloseChbngesetFunc: &ChbngesetSourceCloseChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		CrebteChbngesetFunc: &ChbngesetSourceCrebteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 bool, r1 error) {
				return
			},
		},
		CrebteCommentFunc: &ChbngesetSourceCrebteCommentFunc{
			defbultHook: func(context.Context, *Chbngeset, string) (r0 error) {
				return
			},
		},
		GitserverPushConfigFunc: &ChbngesetSourceGitserverPushConfigFunc{
			defbultHook: func(*types.Repo) (r0 *protocol.PushConfig, r1 error) {
				return
			},
		},
		LobdChbngesetFunc: &ChbngesetSourceLobdChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		MergeChbngesetFunc: &ChbngesetSourceMergeChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset, bool) (r0 error) {
				return
			},
		},
		ReopenChbngesetFunc: &ChbngesetSourceReopenChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		UpdbteChbngesetFunc: &ChbngesetSourceUpdbteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		VblidbteAuthenticbtorFunc: &ChbngesetSourceVblidbteAuthenticbtorFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		WithAuthenticbtorFunc: &ChbngesetSourceWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 ChbngesetSource, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockChbngesetSource crebtes b new mock of the ChbngesetSource
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockChbngesetSource() *MockChbngesetSource {
	return &MockChbngesetSource{
		BuildCommitOptsFunc: &ChbngesetSourceBuildCommitOptsFunc{
			defbultHook: func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
				pbnic("unexpected invocbtion of MockChbngesetSource.BuildCommitOpts")
			},
		},
		CloseChbngesetFunc: &ChbngesetSourceCloseChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.CloseChbngeset")
			},
		},
		CrebteChbngesetFunc: &ChbngesetSourceCrebteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (bool, error) {
				pbnic("unexpected invocbtion of MockChbngesetSource.CrebteChbngeset")
			},
		},
		CrebteCommentFunc: &ChbngesetSourceCrebteCommentFunc{
			defbultHook: func(context.Context, *Chbngeset, string) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.CrebteComment")
			},
		},
		GitserverPushConfigFunc: &ChbngesetSourceGitserverPushConfigFunc{
			defbultHook: func(*types.Repo) (*protocol.PushConfig, error) {
				pbnic("unexpected invocbtion of MockChbngesetSource.GitserverPushConfig")
			},
		},
		LobdChbngesetFunc: &ChbngesetSourceLobdChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.LobdChbngeset")
			},
		},
		MergeChbngesetFunc: &ChbngesetSourceMergeChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset, bool) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.MergeChbngeset")
			},
		},
		ReopenChbngesetFunc: &ChbngesetSourceReopenChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.ReopenChbngeset")
			},
		},
		UpdbteChbngesetFunc: &ChbngesetSourceUpdbteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.UpdbteChbngeset")
			},
		},
		VblidbteAuthenticbtorFunc: &ChbngesetSourceVblidbteAuthenticbtorFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockChbngesetSource.VblidbteAuthenticbtor")
			},
		},
		WithAuthenticbtorFunc: &ChbngesetSourceWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (ChbngesetSource, error) {
				pbnic("unexpected invocbtion of MockChbngesetSource.WithAuthenticbtor")
			},
		},
	}
}

// NewMockChbngesetSourceFrom crebtes b new mock of the MockChbngesetSource
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockChbngesetSourceFrom(i ChbngesetSource) *MockChbngesetSource {
	return &MockChbngesetSource{
		BuildCommitOptsFunc: &ChbngesetSourceBuildCommitOptsFunc{
			defbultHook: i.BuildCommitOpts,
		},
		CloseChbngesetFunc: &ChbngesetSourceCloseChbngesetFunc{
			defbultHook: i.CloseChbngeset,
		},
		CrebteChbngesetFunc: &ChbngesetSourceCrebteChbngesetFunc{
			defbultHook: i.CrebteChbngeset,
		},
		CrebteCommentFunc: &ChbngesetSourceCrebteCommentFunc{
			defbultHook: i.CrebteComment,
		},
		GitserverPushConfigFunc: &ChbngesetSourceGitserverPushConfigFunc{
			defbultHook: i.GitserverPushConfig,
		},
		LobdChbngesetFunc: &ChbngesetSourceLobdChbngesetFunc{
			defbultHook: i.LobdChbngeset,
		},
		MergeChbngesetFunc: &ChbngesetSourceMergeChbngesetFunc{
			defbultHook: i.MergeChbngeset,
		},
		ReopenChbngesetFunc: &ChbngesetSourceReopenChbngesetFunc{
			defbultHook: i.ReopenChbngeset,
		},
		UpdbteChbngesetFunc: &ChbngesetSourceUpdbteChbngesetFunc{
			defbultHook: i.UpdbteChbngeset,
		},
		VblidbteAuthenticbtorFunc: &ChbngesetSourceVblidbteAuthenticbtorFunc{
			defbultHook: i.VblidbteAuthenticbtor,
		},
		WithAuthenticbtorFunc: &ChbngesetSourceWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
	}
}

// ChbngesetSourceBuildCommitOptsFunc describes the behbvior when the
// BuildCommitOpts method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceBuildCommitOptsFunc struct {
	defbultHook func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest
	hooks       []func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest
	history     []ChbngesetSourceBuildCommitOptsFuncCbll
	mutex       sync.Mutex
}

// BuildCommitOpts delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) BuildCommitOpts(v0 *types.Repo, v1 *types1.Chbngeset, v2 *types1.ChbngesetSpec, v3 *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	r0 := m.BuildCommitOptsFunc.nextHook()(v0, v1, v2, v3)
	m.BuildCommitOptsFunc.bppendCbll(ChbngesetSourceBuildCommitOptsFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the BuildCommitOpts
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceBuildCommitOptsFunc) SetDefbultHook(hook func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BuildCommitOpts method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceBuildCommitOptsFunc) PushHook(hook func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceBuildCommitOptsFunc) SetDefbultReturn(r0 protocol.CrebteCommitFromPbtchRequest) {
	f.SetDefbultHook(func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceBuildCommitOptsFunc) PushReturn(r0 protocol.CrebteCommitFromPbtchRequest) {
	f.PushHook(func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
		return r0
	})
}

func (f *ChbngesetSourceBuildCommitOptsFunc) nextHook() func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceBuildCommitOptsFunc) bppendCbll(r0 ChbngesetSourceBuildCommitOptsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceBuildCommitOptsFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceBuildCommitOptsFunc) History() []ChbngesetSourceBuildCommitOptsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceBuildCommitOptsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceBuildCommitOptsFuncCbll is bn object thbt describes bn
// invocbtion of method BuildCommitOpts on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceBuildCommitOptsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *types.Repo
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types1.Chbngeset
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *types1.ChbngesetSpec
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *protocol.PushConfig
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 protocol.CrebteCommitFromPbtchRequest
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceBuildCommitOptsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceBuildCommitOptsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceCloseChbngesetFunc describes the behbvior when the
// CloseChbngeset method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceCloseChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ChbngesetSourceCloseChbngesetFuncCbll
	mutex       sync.Mutex
}

// CloseChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) CloseChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.CloseChbngesetFunc.nextHook()(v0, v1)
	m.CloseChbngesetFunc.bppendCbll(ChbngesetSourceCloseChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the CloseChbngeset
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceCloseChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CloseChbngeset method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceCloseChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceCloseChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceCloseChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ChbngesetSourceCloseChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceCloseChbngesetFunc) bppendCbll(r0 ChbngesetSourceCloseChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceCloseChbngesetFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceCloseChbngesetFunc) History() []ChbngesetSourceCloseChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceCloseChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceCloseChbngesetFuncCbll is bn object thbt describes bn
// invocbtion of method CloseChbngeset on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceCloseChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceCloseChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceCloseChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceCrebteChbngesetFunc describes the behbvior when the
// CrebteChbngeset method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceCrebteChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) (bool, error)
	hooks       []func(context.Context, *Chbngeset) (bool, error)
	history     []ChbngesetSourceCrebteChbngesetFuncCbll
	mutex       sync.Mutex
}

// CrebteChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) CrebteChbngeset(v0 context.Context, v1 *Chbngeset) (bool, error) {
	r0, r1 := m.CrebteChbngesetFunc.nextHook()(v0, v1)
	m.CrebteChbngesetFunc.bppendCbll(ChbngesetSourceCrebteChbngesetFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CrebteChbngeset
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceCrebteChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteChbngeset method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceCrebteChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceCrebteChbngesetFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceCrebteChbngesetFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, *Chbngeset) (bool, error) {
		return r0, r1
	})
}

func (f *ChbngesetSourceCrebteChbngesetFunc) nextHook() func(context.Context, *Chbngeset) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceCrebteChbngesetFunc) bppendCbll(r0 ChbngesetSourceCrebteChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceCrebteChbngesetFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceCrebteChbngesetFunc) History() []ChbngesetSourceCrebteChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceCrebteChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceCrebteChbngesetFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteChbngeset on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceCrebteChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceCrebteChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceCrebteChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ChbngesetSourceCrebteCommentFunc describes the behbvior when the
// CrebteComment method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceCrebteCommentFunc struct {
	defbultHook func(context.Context, *Chbngeset, string) error
	hooks       []func(context.Context, *Chbngeset, string) error
	history     []ChbngesetSourceCrebteCommentFuncCbll
	mutex       sync.Mutex
}

// CrebteComment delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) CrebteComment(v0 context.Context, v1 *Chbngeset, v2 string) error {
	r0 := m.CrebteCommentFunc.nextHook()(v0, v1, v2)
	m.CrebteCommentFunc.bppendCbll(ChbngesetSourceCrebteCommentFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the CrebteComment method
// of the pbrent MockChbngesetSource instbnce is invoked bnd the hook queue
// is empty.
func (f *ChbngesetSourceCrebteCommentFunc) SetDefbultHook(hook func(context.Context, *Chbngeset, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteComment method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceCrebteCommentFunc) PushHook(hook func(context.Context, *Chbngeset, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceCrebteCommentFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceCrebteCommentFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset, string) error {
		return r0
	})
}

func (f *ChbngesetSourceCrebteCommentFunc) nextHook() func(context.Context, *Chbngeset, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceCrebteCommentFunc) bppendCbll(r0 ChbngesetSourceCrebteCommentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceCrebteCommentFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceCrebteCommentFunc) History() []ChbngesetSourceCrebteCommentFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceCrebteCommentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceCrebteCommentFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteComment on bn instbnce of MockChbngesetSource.
type ChbngesetSourceCrebteCommentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceCrebteCommentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceCrebteCommentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceGitserverPushConfigFunc describes the behbvior when the
// GitserverPushConfig method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceGitserverPushConfigFunc struct {
	defbultHook func(*types.Repo) (*protocol.PushConfig, error)
	hooks       []func(*types.Repo) (*protocol.PushConfig, error)
	history     []ChbngesetSourceGitserverPushConfigFuncCbll
	mutex       sync.Mutex
}

// GitserverPushConfig delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) GitserverPushConfig(v0 *types.Repo) (*protocol.PushConfig, error) {
	r0, r1 := m.GitserverPushConfigFunc.nextHook()(v0)
	m.GitserverPushConfigFunc.bppendCbll(ChbngesetSourceGitserverPushConfigFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GitserverPushConfig
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceGitserverPushConfigFunc) SetDefbultHook(hook func(*types.Repo) (*protocol.PushConfig, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GitserverPushConfig method of the pbrent MockChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ChbngesetSourceGitserverPushConfigFunc) PushHook(hook func(*types.Repo) (*protocol.PushConfig, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceGitserverPushConfigFunc) SetDefbultReturn(r0 *protocol.PushConfig, r1 error) {
	f.SetDefbultHook(func(*types.Repo) (*protocol.PushConfig, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceGitserverPushConfigFunc) PushReturn(r0 *protocol.PushConfig, r1 error) {
	f.PushHook(func(*types.Repo) (*protocol.PushConfig, error) {
		return r0, r1
	})
}

func (f *ChbngesetSourceGitserverPushConfigFunc) nextHook() func(*types.Repo) (*protocol.PushConfig, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceGitserverPushConfigFunc) bppendCbll(r0 ChbngesetSourceGitserverPushConfigFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceGitserverPushConfigFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceGitserverPushConfigFunc) History() []ChbngesetSourceGitserverPushConfigFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceGitserverPushConfigFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceGitserverPushConfigFuncCbll is bn object thbt describes bn
// invocbtion of method GitserverPushConfig on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceGitserverPushConfigFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *types.Repo
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.PushConfig
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceGitserverPushConfigFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceGitserverPushConfigFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ChbngesetSourceLobdChbngesetFunc describes the behbvior when the
// LobdChbngeset method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceLobdChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ChbngesetSourceLobdChbngesetFuncCbll
	mutex       sync.Mutex
}

// LobdChbngeset delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) LobdChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.LobdChbngesetFunc.nextHook()(v0, v1)
	m.LobdChbngesetFunc.bppendCbll(ChbngesetSourceLobdChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LobdChbngeset method
// of the pbrent MockChbngesetSource instbnce is invoked bnd the hook queue
// is empty.
func (f *ChbngesetSourceLobdChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LobdChbngeset method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceLobdChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceLobdChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceLobdChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ChbngesetSourceLobdChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceLobdChbngesetFunc) bppendCbll(r0 ChbngesetSourceLobdChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceLobdChbngesetFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceLobdChbngesetFunc) History() []ChbngesetSourceLobdChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceLobdChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceLobdChbngesetFuncCbll is bn object thbt describes bn
// invocbtion of method LobdChbngeset on bn instbnce of MockChbngesetSource.
type ChbngesetSourceLobdChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceLobdChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceLobdChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceMergeChbngesetFunc describes the behbvior when the
// MergeChbngeset method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceMergeChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset, bool) error
	hooks       []func(context.Context, *Chbngeset, bool) error
	history     []ChbngesetSourceMergeChbngesetFuncCbll
	mutex       sync.Mutex
}

// MergeChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) MergeChbngeset(v0 context.Context, v1 *Chbngeset, v2 bool) error {
	r0 := m.MergeChbngesetFunc.nextHook()(v0, v1, v2)
	m.MergeChbngesetFunc.bppendCbll(ChbngesetSourceMergeChbngesetFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the MergeChbngeset
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceMergeChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MergeChbngeset method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceMergeChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceMergeChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceMergeChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset, bool) error {
		return r0
	})
}

func (f *ChbngesetSourceMergeChbngesetFunc) nextHook() func(context.Context, *Chbngeset, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceMergeChbngesetFunc) bppendCbll(r0 ChbngesetSourceMergeChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceMergeChbngesetFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceMergeChbngesetFunc) History() []ChbngesetSourceMergeChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceMergeChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceMergeChbngesetFuncCbll is bn object thbt describes bn
// invocbtion of method MergeChbngeset on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceMergeChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceMergeChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceMergeChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceReopenChbngesetFunc describes the behbvior when the
// ReopenChbngeset method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceReopenChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ChbngesetSourceReopenChbngesetFuncCbll
	mutex       sync.Mutex
}

// ReopenChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) ReopenChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.ReopenChbngesetFunc.nextHook()(v0, v1)
	m.ReopenChbngesetFunc.bppendCbll(ChbngesetSourceReopenChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ReopenChbngeset
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceReopenChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReopenChbngeset method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceReopenChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceReopenChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceReopenChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ChbngesetSourceReopenChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceReopenChbngesetFunc) bppendCbll(r0 ChbngesetSourceReopenChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceReopenChbngesetFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceReopenChbngesetFunc) History() []ChbngesetSourceReopenChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceReopenChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceReopenChbngesetFuncCbll is bn object thbt describes bn
// invocbtion of method ReopenChbngeset on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceReopenChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceReopenChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceReopenChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceUpdbteChbngesetFunc describes the behbvior when the
// UpdbteChbngeset method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceUpdbteChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ChbngesetSourceUpdbteChbngesetFuncCbll
	mutex       sync.Mutex
}

// UpdbteChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) UpdbteChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.UpdbteChbngesetFunc.nextHook()(v0, v1)
	m.UpdbteChbngesetFunc.bppendCbll(ChbngesetSourceUpdbteChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UpdbteChbngeset
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceUpdbteChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteChbngeset method of the pbrent MockChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ChbngesetSourceUpdbteChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceUpdbteChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceUpdbteChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ChbngesetSourceUpdbteChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceUpdbteChbngesetFunc) bppendCbll(r0 ChbngesetSourceUpdbteChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceUpdbteChbngesetFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceUpdbteChbngesetFunc) History() []ChbngesetSourceUpdbteChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceUpdbteChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceUpdbteChbngesetFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteChbngeset on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceUpdbteChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceUpdbteChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceUpdbteChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceVblidbteAuthenticbtorFunc describes the behbvior when the
// VblidbteAuthenticbtor method of the pbrent MockChbngesetSource instbnce
// is invoked.
type ChbngesetSourceVblidbteAuthenticbtorFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []ChbngesetSourceVblidbteAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// VblidbteAuthenticbtor delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) VblidbteAuthenticbtor(v0 context.Context) error {
	r0 := m.VblidbteAuthenticbtorFunc.nextHook()(v0)
	m.VblidbteAuthenticbtorFunc.bppendCbll(ChbngesetSourceVblidbteAuthenticbtorFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// VblidbteAuthenticbtor method of the pbrent MockChbngesetSource instbnce
// is invoked bnd the hook queue is empty.
func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VblidbteAuthenticbtor method of the pbrent MockChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) bppendCbll(r0 ChbngesetSourceVblidbteAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ChbngesetSourceVblidbteAuthenticbtorFuncCbll objects describing the
// invocbtions of this function.
func (f *ChbngesetSourceVblidbteAuthenticbtorFunc) History() []ChbngesetSourceVblidbteAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceVblidbteAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceVblidbteAuthenticbtorFuncCbll is bn object thbt describes
// bn invocbtion of method VblidbteAuthenticbtor on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceVblidbteAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceVblidbteAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceVblidbteAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ChbngesetSourceWithAuthenticbtorFunc describes the behbvior when the
// WithAuthenticbtor method of the pbrent MockChbngesetSource instbnce is
// invoked.
type ChbngesetSourceWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) (ChbngesetSource, error)
	hooks       []func(buth.Authenticbtor) (ChbngesetSource, error)
	history     []ChbngesetSourceWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockChbngesetSource) WithAuthenticbtor(v0 buth.Authenticbtor) (ChbngesetSource, error) {
	r0, r1 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(ChbngesetSourceWithAuthenticbtorFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ChbngesetSourceWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) (ChbngesetSource, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ChbngesetSourceWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) (ChbngesetSource, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ChbngesetSourceWithAuthenticbtorFunc) SetDefbultReturn(r0 ChbngesetSource, r1 error) {
	f.SetDefbultHook(func(buth.Authenticbtor) (ChbngesetSource, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ChbngesetSourceWithAuthenticbtorFunc) PushReturn(r0 ChbngesetSource, r1 error) {
	f.PushHook(func(buth.Authenticbtor) (ChbngesetSource, error) {
		return r0, r1
	})
}

func (f *ChbngesetSourceWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) (ChbngesetSource, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ChbngesetSourceWithAuthenticbtorFunc) bppendCbll(r0 ChbngesetSourceWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ChbngesetSourceWithAuthenticbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *ChbngesetSourceWithAuthenticbtorFunc) History() []ChbngesetSourceWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]ChbngesetSourceWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ChbngesetSourceWithAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method WithAuthenticbtor on bn instbnce of
// MockChbngesetSource.
type ChbngesetSourceWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 ChbngesetSource
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ChbngesetSourceWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ChbngesetSourceWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockForkbbleChbngesetSource is b mock implementbtion of the
// ForkbbleChbngesetSource interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources) used for
// unit testing.
type MockForkbbleChbngesetSource struct {
	// BuildCommitOptsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method BuildCommitOpts.
	BuildCommitOptsFunc *ForkbbleChbngesetSourceBuildCommitOptsFunc
	// CloseChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CloseChbngeset.
	CloseChbngesetFunc *ForkbbleChbngesetSourceCloseChbngesetFunc
	// CrebteChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteChbngeset.
	CrebteChbngesetFunc *ForkbbleChbngesetSourceCrebteChbngesetFunc
	// CrebteCommentFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteComment.
	CrebteCommentFunc *ForkbbleChbngesetSourceCrebteCommentFunc
	// GetForkFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetFork.
	GetForkFunc *ForkbbleChbngesetSourceGetForkFunc
	// GitserverPushConfigFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GitserverPushConfig.
	GitserverPushConfigFunc *ForkbbleChbngesetSourceGitserverPushConfigFunc
	// LobdChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method LobdChbngeset.
	LobdChbngesetFunc *ForkbbleChbngesetSourceLobdChbngesetFunc
	// MergeChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method MergeChbngeset.
	MergeChbngesetFunc *ForkbbleChbngesetSourceMergeChbngesetFunc
	// ReopenChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReopenChbngeset.
	ReopenChbngesetFunc *ForkbbleChbngesetSourceReopenChbngesetFunc
	// UpdbteChbngesetFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteChbngeset.
	UpdbteChbngesetFunc *ForkbbleChbngesetSourceUpdbteChbngesetFunc
	// VblidbteAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method VblidbteAuthenticbtor.
	VblidbteAuthenticbtorFunc *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *ForkbbleChbngesetSourceWithAuthenticbtorFunc
}

// NewMockForkbbleChbngesetSource crebtes b new mock of the
// ForkbbleChbngesetSource interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockForkbbleChbngesetSource() *MockForkbbleChbngesetSource {
	return &MockForkbbleChbngesetSource{
		BuildCommitOptsFunc: &ForkbbleChbngesetSourceBuildCommitOptsFunc{
			defbultHook: func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) (r0 protocol.CrebteCommitFromPbtchRequest) {
				return
			},
		},
		CloseChbngesetFunc: &ForkbbleChbngesetSourceCloseChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		CrebteChbngesetFunc: &ForkbbleChbngesetSourceCrebteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 bool, r1 error) {
				return
			},
		},
		CrebteCommentFunc: &ForkbbleChbngesetSourceCrebteCommentFunc{
			defbultHook: func(context.Context, *Chbngeset, string) (r0 error) {
				return
			},
		},
		GetForkFunc: &ForkbbleChbngesetSourceGetForkFunc{
			defbultHook: func(context.Context, *types.Repo, *string, *string) (r0 *types.Repo, r1 error) {
				return
			},
		},
		GitserverPushConfigFunc: &ForkbbleChbngesetSourceGitserverPushConfigFunc{
			defbultHook: func(*types.Repo) (r0 *protocol.PushConfig, r1 error) {
				return
			},
		},
		LobdChbngesetFunc: &ForkbbleChbngesetSourceLobdChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		MergeChbngesetFunc: &ForkbbleChbngesetSourceMergeChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset, bool) (r0 error) {
				return
			},
		},
		ReopenChbngesetFunc: &ForkbbleChbngesetSourceReopenChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		UpdbteChbngesetFunc: &ForkbbleChbngesetSourceUpdbteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (r0 error) {
				return
			},
		},
		VblidbteAuthenticbtorFunc: &ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		WithAuthenticbtorFunc: &ForkbbleChbngesetSourceWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 ChbngesetSource, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockForkbbleChbngesetSource crebtes b new mock of the
// ForkbbleChbngesetSource interfbce. All methods pbnic on invocbtion,
// unless overwritten.
func NewStrictMockForkbbleChbngesetSource() *MockForkbbleChbngesetSource {
	return &MockForkbbleChbngesetSource{
		BuildCommitOptsFunc: &ForkbbleChbngesetSourceBuildCommitOptsFunc{
			defbultHook: func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.BuildCommitOpts")
			},
		},
		CloseChbngesetFunc: &ForkbbleChbngesetSourceCloseChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.CloseChbngeset")
			},
		},
		CrebteChbngesetFunc: &ForkbbleChbngesetSourceCrebteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) (bool, error) {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.CrebteChbngeset")
			},
		},
		CrebteCommentFunc: &ForkbbleChbngesetSourceCrebteCommentFunc{
			defbultHook: func(context.Context, *Chbngeset, string) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.CrebteComment")
			},
		},
		GetForkFunc: &ForkbbleChbngesetSourceGetForkFunc{
			defbultHook: func(context.Context, *types.Repo, *string, *string) (*types.Repo, error) {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.GetFork")
			},
		},
		GitserverPushConfigFunc: &ForkbbleChbngesetSourceGitserverPushConfigFunc{
			defbultHook: func(*types.Repo) (*protocol.PushConfig, error) {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.GitserverPushConfig")
			},
		},
		LobdChbngesetFunc: &ForkbbleChbngesetSourceLobdChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.LobdChbngeset")
			},
		},
		MergeChbngesetFunc: &ForkbbleChbngesetSourceMergeChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset, bool) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.MergeChbngeset")
			},
		},
		ReopenChbngesetFunc: &ForkbbleChbngesetSourceReopenChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.ReopenChbngeset")
			},
		},
		UpdbteChbngesetFunc: &ForkbbleChbngesetSourceUpdbteChbngesetFunc{
			defbultHook: func(context.Context, *Chbngeset) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.UpdbteChbngeset")
			},
		},
		VblidbteAuthenticbtorFunc: &ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.VblidbteAuthenticbtor")
			},
		},
		WithAuthenticbtorFunc: &ForkbbleChbngesetSourceWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (ChbngesetSource, error) {
				pbnic("unexpected invocbtion of MockForkbbleChbngesetSource.WithAuthenticbtor")
			},
		},
	}
}

// NewMockForkbbleChbngesetSourceFrom crebtes b new mock of the
// MockForkbbleChbngesetSource interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockForkbbleChbngesetSourceFrom(i ForkbbleChbngesetSource) *MockForkbbleChbngesetSource {
	return &MockForkbbleChbngesetSource{
		BuildCommitOptsFunc: &ForkbbleChbngesetSourceBuildCommitOptsFunc{
			defbultHook: i.BuildCommitOpts,
		},
		CloseChbngesetFunc: &ForkbbleChbngesetSourceCloseChbngesetFunc{
			defbultHook: i.CloseChbngeset,
		},
		CrebteChbngesetFunc: &ForkbbleChbngesetSourceCrebteChbngesetFunc{
			defbultHook: i.CrebteChbngeset,
		},
		CrebteCommentFunc: &ForkbbleChbngesetSourceCrebteCommentFunc{
			defbultHook: i.CrebteComment,
		},
		GetForkFunc: &ForkbbleChbngesetSourceGetForkFunc{
			defbultHook: i.GetFork,
		},
		GitserverPushConfigFunc: &ForkbbleChbngesetSourceGitserverPushConfigFunc{
			defbultHook: i.GitserverPushConfig,
		},
		LobdChbngesetFunc: &ForkbbleChbngesetSourceLobdChbngesetFunc{
			defbultHook: i.LobdChbngeset,
		},
		MergeChbngesetFunc: &ForkbbleChbngesetSourceMergeChbngesetFunc{
			defbultHook: i.MergeChbngeset,
		},
		ReopenChbngesetFunc: &ForkbbleChbngesetSourceReopenChbngesetFunc{
			defbultHook: i.ReopenChbngeset,
		},
		UpdbteChbngesetFunc: &ForkbbleChbngesetSourceUpdbteChbngesetFunc{
			defbultHook: i.UpdbteChbngeset,
		},
		VblidbteAuthenticbtorFunc: &ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc{
			defbultHook: i.VblidbteAuthenticbtor,
		},
		WithAuthenticbtorFunc: &ForkbbleChbngesetSourceWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
	}
}

// ForkbbleChbngesetSourceBuildCommitOptsFunc describes the behbvior when
// the BuildCommitOpts method of the pbrent MockForkbbleChbngesetSource
// instbnce is invoked.
type ForkbbleChbngesetSourceBuildCommitOptsFunc struct {
	defbultHook func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest
	hooks       []func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest
	history     []ForkbbleChbngesetSourceBuildCommitOptsFuncCbll
	mutex       sync.Mutex
}

// BuildCommitOpts delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) BuildCommitOpts(v0 *types.Repo, v1 *types1.Chbngeset, v2 *types1.ChbngesetSpec, v3 *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	r0 := m.BuildCommitOptsFunc.nextHook()(v0, v1, v2, v3)
	m.BuildCommitOptsFunc.bppendCbll(ForkbbleChbngesetSourceBuildCommitOptsFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the BuildCommitOpts
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) SetDefbultHook(hook func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BuildCommitOpts method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) PushHook(hook func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) SetDefbultReturn(r0 protocol.CrebteCommitFromPbtchRequest) {
	f.SetDefbultHook(func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) PushReturn(r0 protocol.CrebteCommitFromPbtchRequest) {
	f.PushHook(func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) nextHook() func(*types.Repo, *types1.Chbngeset, *types1.ChbngesetSpec, *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) bppendCbll(r0 ForkbbleChbngesetSourceBuildCommitOptsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceBuildCommitOptsFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceBuildCommitOptsFunc) History() []ForkbbleChbngesetSourceBuildCommitOptsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceBuildCommitOptsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceBuildCommitOptsFuncCbll is bn object thbt
// describes bn invocbtion of method BuildCommitOpts on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceBuildCommitOptsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *types.Repo
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types1.Chbngeset
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *types1.ChbngesetSpec
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *protocol.PushConfig
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 protocol.CrebteCommitFromPbtchRequest
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceBuildCommitOptsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceBuildCommitOptsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceCloseChbngesetFunc describes the behbvior when the
// CloseChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// is invoked.
type ForkbbleChbngesetSourceCloseChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ForkbbleChbngesetSourceCloseChbngesetFuncCbll
	mutex       sync.Mutex
}

// CloseChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) CloseChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.CloseChbngesetFunc.nextHook()(v0, v1)
	m.CloseChbngesetFunc.bppendCbll(ForkbbleChbngesetSourceCloseChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the CloseChbngeset
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CloseChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) bppendCbll(r0 ForkbbleChbngesetSourceCloseChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceCloseChbngesetFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceCloseChbngesetFunc) History() []ForkbbleChbngesetSourceCloseChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceCloseChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceCloseChbngesetFuncCbll is bn object thbt describes
// bn invocbtion of method CloseChbngeset on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceCloseChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceCloseChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceCloseChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceCrebteChbngesetFunc describes the behbvior when
// the CrebteChbngeset method of the pbrent MockForkbbleChbngesetSource
// instbnce is invoked.
type ForkbbleChbngesetSourceCrebteChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) (bool, error)
	hooks       []func(context.Context, *Chbngeset) (bool, error)
	history     []ForkbbleChbngesetSourceCrebteChbngesetFuncCbll
	mutex       sync.Mutex
}

// CrebteChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) CrebteChbngeset(v0 context.Context, v1 *Chbngeset) (bool, error) {
	r0, r1 := m.CrebteChbngesetFunc.nextHook()(v0, v1)
	m.CrebteChbngesetFunc.bppendCbll(ForkbbleChbngesetSourceCrebteChbngesetFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CrebteChbngeset
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, *Chbngeset) (bool, error) {
		return r0, r1
	})
}

func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) nextHook() func(context.Context, *Chbngeset) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) bppendCbll(r0 ForkbbleChbngesetSourceCrebteChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceCrebteChbngesetFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceCrebteChbngesetFunc) History() []ForkbbleChbngesetSourceCrebteChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceCrebteChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceCrebteChbngesetFuncCbll is bn object thbt
// describes bn invocbtion of method CrebteChbngeset on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceCrebteChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceCrebteChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceCrebteChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ForkbbleChbngesetSourceCrebteCommentFunc describes the behbvior when the
// CrebteComment method of the pbrent MockForkbbleChbngesetSource instbnce
// is invoked.
type ForkbbleChbngesetSourceCrebteCommentFunc struct {
	defbultHook func(context.Context, *Chbngeset, string) error
	hooks       []func(context.Context, *Chbngeset, string) error
	history     []ForkbbleChbngesetSourceCrebteCommentFuncCbll
	mutex       sync.Mutex
}

// CrebteComment delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) CrebteComment(v0 context.Context, v1 *Chbngeset, v2 string) error {
	r0 := m.CrebteCommentFunc.nextHook()(v0, v1, v2)
	m.CrebteCommentFunc.bppendCbll(ForkbbleChbngesetSourceCrebteCommentFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the CrebteComment method
// of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd the
// hook queue is empty.
func (f *ForkbbleChbngesetSourceCrebteCommentFunc) SetDefbultHook(hook func(context.Context, *Chbngeset, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteComment method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceCrebteCommentFunc) PushHook(hook func(context.Context, *Chbngeset, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceCrebteCommentFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceCrebteCommentFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset, string) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceCrebteCommentFunc) nextHook() func(context.Context, *Chbngeset, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceCrebteCommentFunc) bppendCbll(r0 ForkbbleChbngesetSourceCrebteCommentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceCrebteCommentFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceCrebteCommentFunc) History() []ForkbbleChbngesetSourceCrebteCommentFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceCrebteCommentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceCrebteCommentFuncCbll is bn object thbt describes
// bn invocbtion of method CrebteComment on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceCrebteCommentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceCrebteCommentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceCrebteCommentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceGetForkFunc describes the behbvior when the
// GetFork method of the pbrent MockForkbbleChbngesetSource instbnce is
// invoked.
type ForkbbleChbngesetSourceGetForkFunc struct {
	defbultHook func(context.Context, *types.Repo, *string, *string) (*types.Repo, error)
	hooks       []func(context.Context, *types.Repo, *string, *string) (*types.Repo, error)
	history     []ForkbbleChbngesetSourceGetForkFuncCbll
	mutex       sync.Mutex
}

// GetFork delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) GetFork(v0 context.Context, v1 *types.Repo, v2 *string, v3 *string) (*types.Repo, error) {
	r0, r1 := m.GetForkFunc.nextHook()(v0, v1, v2, v3)
	m.GetForkFunc.bppendCbll(ForkbbleChbngesetSourceGetForkFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetFork method of
// the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd the hook
// queue is empty.
func (f *ForkbbleChbngesetSourceGetForkFunc) SetDefbultHook(hook func(context.Context, *types.Repo, *string, *string) (*types.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetFork method of the pbrent MockForkbbleChbngesetSource instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ForkbbleChbngesetSourceGetForkFunc) PushHook(hook func(context.Context, *types.Repo, *string, *string) (*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceGetForkFunc) SetDefbultReturn(r0 *types.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, *types.Repo, *string, *string) (*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceGetForkFunc) PushReturn(r0 *types.Repo, r1 error) {
	f.PushHook(func(context.Context, *types.Repo, *string, *string) (*types.Repo, error) {
		return r0, r1
	})
}

func (f *ForkbbleChbngesetSourceGetForkFunc) nextHook() func(context.Context, *types.Repo, *string, *string) (*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceGetForkFunc) bppendCbll(r0 ForkbbleChbngesetSourceGetForkFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ForkbbleChbngesetSourceGetForkFuncCbll
// objects describing the invocbtions of this function.
func (f *ForkbbleChbngesetSourceGetForkFunc) History() []ForkbbleChbngesetSourceGetForkFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceGetForkFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceGetForkFuncCbll is bn object thbt describes bn
// invocbtion of method GetFork on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceGetForkFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceGetForkFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceGetForkFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ForkbbleChbngesetSourceGitserverPushConfigFunc describes the behbvior
// when the GitserverPushConfig method of the pbrent
// MockForkbbleChbngesetSource instbnce is invoked.
type ForkbbleChbngesetSourceGitserverPushConfigFunc struct {
	defbultHook func(*types.Repo) (*protocol.PushConfig, error)
	hooks       []func(*types.Repo) (*protocol.PushConfig, error)
	history     []ForkbbleChbngesetSourceGitserverPushConfigFuncCbll
	mutex       sync.Mutex
}

// GitserverPushConfig delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) GitserverPushConfig(v0 *types.Repo) (*protocol.PushConfig, error) {
	r0, r1 := m.GitserverPushConfigFunc.nextHook()(v0)
	m.GitserverPushConfigFunc.bppendCbll(ForkbbleChbngesetSourceGitserverPushConfigFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GitserverPushConfig
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) SetDefbultHook(hook func(*types.Repo) (*protocol.PushConfig, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GitserverPushConfig method of the pbrent MockForkbbleChbngesetSource
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) PushHook(hook func(*types.Repo) (*protocol.PushConfig, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) SetDefbultReturn(r0 *protocol.PushConfig, r1 error) {
	f.SetDefbultHook(func(*types.Repo) (*protocol.PushConfig, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) PushReturn(r0 *protocol.PushConfig, r1 error) {
	f.PushHook(func(*types.Repo) (*protocol.PushConfig, error) {
		return r0, r1
	})
}

func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) nextHook() func(*types.Repo) (*protocol.PushConfig, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) bppendCbll(r0 ForkbbleChbngesetSourceGitserverPushConfigFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceGitserverPushConfigFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceGitserverPushConfigFunc) History() []ForkbbleChbngesetSourceGitserverPushConfigFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceGitserverPushConfigFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceGitserverPushConfigFuncCbll is bn object thbt
// describes bn invocbtion of method GitserverPushConfig on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceGitserverPushConfigFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *types.Repo
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.PushConfig
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceGitserverPushConfigFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceGitserverPushConfigFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ForkbbleChbngesetSourceLobdChbngesetFunc describes the behbvior when the
// LobdChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// is invoked.
type ForkbbleChbngesetSourceLobdChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ForkbbleChbngesetSourceLobdChbngesetFuncCbll
	mutex       sync.Mutex
}

// LobdChbngeset delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) LobdChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.LobdChbngesetFunc.nextHook()(v0, v1)
	m.LobdChbngesetFunc.bppendCbll(ForkbbleChbngesetSourceLobdChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LobdChbngeset method
// of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd the
// hook queue is empty.
func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LobdChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) bppendCbll(r0 ForkbbleChbngesetSourceLobdChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceLobdChbngesetFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceLobdChbngesetFunc) History() []ForkbbleChbngesetSourceLobdChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceLobdChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceLobdChbngesetFuncCbll is bn object thbt describes
// bn invocbtion of method LobdChbngeset on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceLobdChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceLobdChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceLobdChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceMergeChbngesetFunc describes the behbvior when the
// MergeChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// is invoked.
type ForkbbleChbngesetSourceMergeChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset, bool) error
	hooks       []func(context.Context, *Chbngeset, bool) error
	history     []ForkbbleChbngesetSourceMergeChbngesetFuncCbll
	mutex       sync.Mutex
}

// MergeChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) MergeChbngeset(v0 context.Context, v1 *Chbngeset, v2 bool) error {
	r0 := m.MergeChbngesetFunc.nextHook()(v0, v1, v2)
	m.MergeChbngesetFunc.bppendCbll(ForkbbleChbngesetSourceMergeChbngesetFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the MergeChbngeset
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MergeChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset, bool) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) nextHook() func(context.Context, *Chbngeset, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) bppendCbll(r0 ForkbbleChbngesetSourceMergeChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceMergeChbngesetFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceMergeChbngesetFunc) History() []ForkbbleChbngesetSourceMergeChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceMergeChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceMergeChbngesetFuncCbll is bn object thbt describes
// bn invocbtion of method MergeChbngeset on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceMergeChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceMergeChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceMergeChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceReopenChbngesetFunc describes the behbvior when
// the ReopenChbngeset method of the pbrent MockForkbbleChbngesetSource
// instbnce is invoked.
type ForkbbleChbngesetSourceReopenChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ForkbbleChbngesetSourceReopenChbngesetFuncCbll
	mutex       sync.Mutex
}

// ReopenChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) ReopenChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.ReopenChbngesetFunc.nextHook()(v0, v1)
	m.ReopenChbngesetFunc.bppendCbll(ForkbbleChbngesetSourceReopenChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ReopenChbngeset
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReopenChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) bppendCbll(r0 ForkbbleChbngesetSourceReopenChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceReopenChbngesetFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceReopenChbngesetFunc) History() []ForkbbleChbngesetSourceReopenChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceReopenChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceReopenChbngesetFuncCbll is bn object thbt
// describes bn invocbtion of method ReopenChbngeset on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceReopenChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceReopenChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceReopenChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceUpdbteChbngesetFunc describes the behbvior when
// the UpdbteChbngeset method of the pbrent MockForkbbleChbngesetSource
// instbnce is invoked.
type ForkbbleChbngesetSourceUpdbteChbngesetFunc struct {
	defbultHook func(context.Context, *Chbngeset) error
	hooks       []func(context.Context, *Chbngeset) error
	history     []ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll
	mutex       sync.Mutex
}

// UpdbteChbngeset delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) UpdbteChbngeset(v0 context.Context, v1 *Chbngeset) error {
	r0 := m.UpdbteChbngesetFunc.nextHook()(v0, v1)
	m.UpdbteChbngesetFunc.bppendCbll(ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UpdbteChbngeset
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) SetDefbultHook(hook func(context.Context, *Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteChbngeset method of the pbrent MockForkbbleChbngesetSource instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) PushHook(hook func(context.Context, *Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Chbngeset) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) nextHook() func(context.Context, *Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) bppendCbll(r0 ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceUpdbteChbngesetFunc) History() []ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll is bn object thbt
// describes bn invocbtion of method UpdbteChbngeset on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceUpdbteChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc describes the behbvior
// when the VblidbteAuthenticbtor method of the pbrent
// MockForkbbleChbngesetSource instbnce is invoked.
type ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// VblidbteAuthenticbtor delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) VblidbteAuthenticbtor(v0 context.Context) error {
	r0 := m.VblidbteAuthenticbtorFunc.nextHook()(v0)
	m.VblidbteAuthenticbtorFunc.bppendCbll(ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// VblidbteAuthenticbtor method of the pbrent MockForkbbleChbngesetSource
// instbnce is invoked bnd the hook queue is empty.
func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VblidbteAuthenticbtor method of the pbrent MockForkbbleChbngesetSource
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) bppendCbll(r0 ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll objects describing
// the invocbtions of this function.
func (f *ForkbbleChbngesetSourceVblidbteAuthenticbtorFunc) History() []ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll is bn object thbt
// describes bn invocbtion of method VblidbteAuthenticbtor on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceVblidbteAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ForkbbleChbngesetSourceWithAuthenticbtorFunc describes the behbvior when
// the WithAuthenticbtor method of the pbrent MockForkbbleChbngesetSource
// instbnce is invoked.
type ForkbbleChbngesetSourceWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) (ChbngesetSource, error)
	hooks       []func(buth.Authenticbtor) (ChbngesetSource, error)
	history     []ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockForkbbleChbngesetSource) WithAuthenticbtor(v0 buth.Authenticbtor) (ChbngesetSource, error) {
	r0, r1 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockForkbbleChbngesetSource instbnce is invoked bnd
// the hook queue is empty.
func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) (ChbngesetSource, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockForkbbleChbngesetSource
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) (ChbngesetSource, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) SetDefbultReturn(r0 ChbngesetSource, r1 error) {
	f.SetDefbultHook(func(buth.Authenticbtor) (ChbngesetSource, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) PushReturn(r0 ChbngesetSource, r1 error) {
	f.PushHook(func(buth.Authenticbtor) (ChbngesetSource, error) {
		return r0, r1
	})
}

func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) (ChbngesetSource, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) bppendCbll(r0 ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll objects describing the
// invocbtions of this function.
func (f *ForkbbleChbngesetSourceWithAuthenticbtorFunc) History() []ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll is bn object thbt
// describes bn invocbtion of method WithAuthenticbtor on bn instbnce of
// MockForkbbleChbngesetSource.
type ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 ChbngesetSource
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ForkbbleChbngesetSourceWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockSourcerStore is b mock implementbtion of the SourcerStore interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources) used for
// unit testing.
type MockSourcerStore struct {
	// DbtbbbseDBFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DbtbbbseDB.
	DbtbbbseDBFunc *SourcerStoreDbtbbbseDBFunc
	// ExternblServicesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ExternblServices.
	ExternblServicesFunc *SourcerStoreExternblServicesFunc
	// GetBbtchChbngeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetBbtchChbnge.
	GetBbtchChbngeFunc *SourcerStoreGetBbtchChbngeFunc
	// GetExternblServiceIDsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetExternblServiceIDs.
	GetExternblServiceIDsFunc *SourcerStoreGetExternblServiceIDsFunc
	// GetSiteCredentiblFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetSiteCredentibl.
	GetSiteCredentiblFunc *SourcerStoreGetSiteCredentiblFunc
	// GitHubAppsStoreFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GitHubAppsStore.
	GitHubAppsStoreFunc *SourcerStoreGitHubAppsStoreFunc
	// ReposFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Repos.
	ReposFunc *SourcerStoreReposFunc
	// UserCredentiblsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UserCredentibls.
	UserCredentiblsFunc *SourcerStoreUserCredentiblsFunc
}

// NewMockSourcerStore crebtes b new mock of the SourcerStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockSourcerStore() *MockSourcerStore {
	return &MockSourcerStore{
		DbtbbbseDBFunc: &SourcerStoreDbtbbbseDBFunc{
			defbultHook: func() (r0 dbtbbbse.DB) {
				return
			},
		},
		ExternblServicesFunc: &SourcerStoreExternblServicesFunc{
			defbultHook: func() (r0 dbtbbbse.ExternblServiceStore) {
				return
			},
		},
		GetBbtchChbngeFunc: &SourcerStoreGetBbtchChbngeFunc{
			defbultHook: func(context.Context, store.GetBbtchChbngeOpts) (r0 *types1.BbtchChbnge, r1 error) {
				return
			},
		},
		GetExternblServiceIDsFunc: &SourcerStoreGetExternblServiceIDsFunc{
			defbultHook: func(context.Context, store.GetExternblServiceIDsOpts) (r0 []int64, r1 error) {
				return
			},
		},
		GetSiteCredentiblFunc: &SourcerStoreGetSiteCredentiblFunc{
			defbultHook: func(context.Context, store.GetSiteCredentiblOpts) (r0 *types1.SiteCredentibl, r1 error) {
				return
			},
		},
		GitHubAppsStoreFunc: &SourcerStoreGitHubAppsStoreFunc{
			defbultHook: func() (r0 store1.GitHubAppsStore) {
				return
			},
		},
		ReposFunc: &SourcerStoreReposFunc{
			defbultHook: func() (r0 dbtbbbse.RepoStore) {
				return
			},
		},
		UserCredentiblsFunc: &SourcerStoreUserCredentiblsFunc{
			defbultHook: func() (r0 dbtbbbse.UserCredentiblsStore) {
				return
			},
		},
	}
}

// NewStrictMockSourcerStore crebtes b new mock of the SourcerStore
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockSourcerStore() *MockSourcerStore {
	return &MockSourcerStore{
		DbtbbbseDBFunc: &SourcerStoreDbtbbbseDBFunc{
			defbultHook: func() dbtbbbse.DB {
				pbnic("unexpected invocbtion of MockSourcerStore.DbtbbbseDB")
			},
		},
		ExternblServicesFunc: &SourcerStoreExternblServicesFunc{
			defbultHook: func() dbtbbbse.ExternblServiceStore {
				pbnic("unexpected invocbtion of MockSourcerStore.ExternblServices")
			},
		},
		GetBbtchChbngeFunc: &SourcerStoreGetBbtchChbngeFunc{
			defbultHook: func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error) {
				pbnic("unexpected invocbtion of MockSourcerStore.GetBbtchChbnge")
			},
		},
		GetExternblServiceIDsFunc: &SourcerStoreGetExternblServiceIDsFunc{
			defbultHook: func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
				pbnic("unexpected invocbtion of MockSourcerStore.GetExternblServiceIDs")
			},
		},
		GetSiteCredentiblFunc: &SourcerStoreGetSiteCredentiblFunc{
			defbultHook: func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error) {
				pbnic("unexpected invocbtion of MockSourcerStore.GetSiteCredentibl")
			},
		},
		GitHubAppsStoreFunc: &SourcerStoreGitHubAppsStoreFunc{
			defbultHook: func() store1.GitHubAppsStore {
				pbnic("unexpected invocbtion of MockSourcerStore.GitHubAppsStore")
			},
		},
		ReposFunc: &SourcerStoreReposFunc{
			defbultHook: func() dbtbbbse.RepoStore {
				pbnic("unexpected invocbtion of MockSourcerStore.Repos")
			},
		},
		UserCredentiblsFunc: &SourcerStoreUserCredentiblsFunc{
			defbultHook: func() dbtbbbse.UserCredentiblsStore {
				pbnic("unexpected invocbtion of MockSourcerStore.UserCredentibls")
			},
		},
	}
}

// NewMockSourcerStoreFrom crebtes b new mock of the MockSourcerStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockSourcerStoreFrom(i SourcerStore) *MockSourcerStore {
	return &MockSourcerStore{
		DbtbbbseDBFunc: &SourcerStoreDbtbbbseDBFunc{
			defbultHook: i.DbtbbbseDB,
		},
		ExternblServicesFunc: &SourcerStoreExternblServicesFunc{
			defbultHook: i.ExternblServices,
		},
		GetBbtchChbngeFunc: &SourcerStoreGetBbtchChbngeFunc{
			defbultHook: i.GetBbtchChbnge,
		},
		GetExternblServiceIDsFunc: &SourcerStoreGetExternblServiceIDsFunc{
			defbultHook: i.GetExternblServiceIDs,
		},
		GetSiteCredentiblFunc: &SourcerStoreGetSiteCredentiblFunc{
			defbultHook: i.GetSiteCredentibl,
		},
		GitHubAppsStoreFunc: &SourcerStoreGitHubAppsStoreFunc{
			defbultHook: i.GitHubAppsStore,
		},
		ReposFunc: &SourcerStoreReposFunc{
			defbultHook: i.Repos,
		},
		UserCredentiblsFunc: &SourcerStoreUserCredentiblsFunc{
			defbultHook: i.UserCredentibls,
		},
	}
}

// SourcerStoreDbtbbbseDBFunc describes the behbvior when the DbtbbbseDB
// method of the pbrent MockSourcerStore instbnce is invoked.
type SourcerStoreDbtbbbseDBFunc struct {
	defbultHook func() dbtbbbse.DB
	hooks       []func() dbtbbbse.DB
	history     []SourcerStoreDbtbbbseDBFuncCbll
	mutex       sync.Mutex
}

// DbtbbbseDB delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) DbtbbbseDB() dbtbbbse.DB {
	r0 := m.DbtbbbseDBFunc.nextHook()()
	m.DbtbbbseDBFunc.bppendCbll(SourcerStoreDbtbbbseDBFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DbtbbbseDB method of
// the pbrent MockSourcerStore instbnce is invoked bnd the hook queue is
// empty.
func (f *SourcerStoreDbtbbbseDBFunc) SetDefbultHook(hook func() dbtbbbse.DB) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DbtbbbseDB method of the pbrent MockSourcerStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SourcerStoreDbtbbbseDBFunc) PushHook(hook func() dbtbbbse.DB) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreDbtbbbseDBFunc) SetDefbultReturn(r0 dbtbbbse.DB) {
	f.SetDefbultHook(func() dbtbbbse.DB {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreDbtbbbseDBFunc) PushReturn(r0 dbtbbbse.DB) {
	f.PushHook(func() dbtbbbse.DB {
		return r0
	})
}

func (f *SourcerStoreDbtbbbseDBFunc) nextHook() func() dbtbbbse.DB {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreDbtbbbseDBFunc) bppendCbll(r0 SourcerStoreDbtbbbseDBFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreDbtbbbseDBFuncCbll objects
// describing the invocbtions of this function.
func (f *SourcerStoreDbtbbbseDBFunc) History() []SourcerStoreDbtbbbseDBFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreDbtbbbseDBFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreDbtbbbseDBFuncCbll is bn object thbt describes bn invocbtion
// of method DbtbbbseDB on bn instbnce of MockSourcerStore.
type SourcerStoreDbtbbbseDBFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.DB
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreDbtbbbseDBFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreDbtbbbseDBFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SourcerStoreExternblServicesFunc describes the behbvior when the
// ExternblServices method of the pbrent MockSourcerStore instbnce is
// invoked.
type SourcerStoreExternblServicesFunc struct {
	defbultHook func() dbtbbbse.ExternblServiceStore
	hooks       []func() dbtbbbse.ExternblServiceStore
	history     []SourcerStoreExternblServicesFuncCbll
	mutex       sync.Mutex
}

// ExternblServices delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) ExternblServices() dbtbbbse.ExternblServiceStore {
	r0 := m.ExternblServicesFunc.nextHook()()
	m.ExternblServicesFunc.bppendCbll(SourcerStoreExternblServicesFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ExternblServices
// method of the pbrent MockSourcerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *SourcerStoreExternblServicesFunc) SetDefbultHook(hook func() dbtbbbse.ExternblServiceStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExternblServices method of the pbrent MockSourcerStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SourcerStoreExternblServicesFunc) PushHook(hook func() dbtbbbse.ExternblServiceStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreExternblServicesFunc) SetDefbultReturn(r0 dbtbbbse.ExternblServiceStore) {
	f.SetDefbultHook(func() dbtbbbse.ExternblServiceStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreExternblServicesFunc) PushReturn(r0 dbtbbbse.ExternblServiceStore) {
	f.PushHook(func() dbtbbbse.ExternblServiceStore {
		return r0
	})
}

func (f *SourcerStoreExternblServicesFunc) nextHook() func() dbtbbbse.ExternblServiceStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreExternblServicesFunc) bppendCbll(r0 SourcerStoreExternblServicesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreExternblServicesFuncCbll
// objects describing the invocbtions of this function.
func (f *SourcerStoreExternblServicesFunc) History() []SourcerStoreExternblServicesFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreExternblServicesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreExternblServicesFuncCbll is bn object thbt describes bn
// invocbtion of method ExternblServices on bn instbnce of MockSourcerStore.
type SourcerStoreExternblServicesFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.ExternblServiceStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreExternblServicesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreExternblServicesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SourcerStoreGetBbtchChbngeFunc describes the behbvior when the
// GetBbtchChbnge method of the pbrent MockSourcerStore instbnce is invoked.
type SourcerStoreGetBbtchChbngeFunc struct {
	defbultHook func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error)
	hooks       []func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error)
	history     []SourcerStoreGetBbtchChbngeFuncCbll
	mutex       sync.Mutex
}

// GetBbtchChbnge delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) GetBbtchChbnge(v0 context.Context, v1 store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error) {
	r0, r1 := m.GetBbtchChbngeFunc.nextHook()(v0, v1)
	m.GetBbtchChbngeFunc.bppendCbll(SourcerStoreGetBbtchChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBbtchChbnge
// method of the pbrent MockSourcerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *SourcerStoreGetBbtchChbngeFunc) SetDefbultHook(hook func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBbtchChbnge method of the pbrent MockSourcerStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SourcerStoreGetBbtchChbngeFunc) PushHook(hook func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreGetBbtchChbngeFunc) SetDefbultReturn(r0 *types1.BbtchChbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreGetBbtchChbngeFunc) PushReturn(r0 *types1.BbtchChbnge, r1 error) {
	f.PushHook(func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error) {
		return r0, r1
	})
}

func (f *SourcerStoreGetBbtchChbngeFunc) nextHook() func(context.Context, store.GetBbtchChbngeOpts) (*types1.BbtchChbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreGetBbtchChbngeFunc) bppendCbll(r0 SourcerStoreGetBbtchChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreGetBbtchChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *SourcerStoreGetBbtchChbngeFunc) History() []SourcerStoreGetBbtchChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreGetBbtchChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreGetBbtchChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method GetBbtchChbnge on bn instbnce of MockSourcerStore.
type SourcerStoreGetBbtchChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetBbtchChbngeOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types1.BbtchChbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreGetBbtchChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreGetBbtchChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SourcerStoreGetExternblServiceIDsFunc describes the behbvior when the
// GetExternblServiceIDs method of the pbrent MockSourcerStore instbnce is
// invoked.
type SourcerStoreGetExternblServiceIDsFunc struct {
	defbultHook func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)
	hooks       []func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)
	history     []SourcerStoreGetExternblServiceIDsFuncCbll
	mutex       sync.Mutex
}

// GetExternblServiceIDs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) GetExternblServiceIDs(v0 context.Context, v1 store.GetExternblServiceIDsOpts) ([]int64, error) {
	r0, r1 := m.GetExternblServiceIDsFunc.nextHook()(v0, v1)
	m.GetExternblServiceIDsFunc.bppendCbll(SourcerStoreGetExternblServiceIDsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetExternblServiceIDs method of the pbrent MockSourcerStore instbnce is
// invoked bnd the hook queue is empty.
func (f *SourcerStoreGetExternblServiceIDsFunc) SetDefbultHook(hook func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetExternblServiceIDs method of the pbrent MockSourcerStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SourcerStoreGetExternblServiceIDsFunc) PushHook(hook func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreGetExternblServiceIDsFunc) SetDefbultReturn(r0 []int64, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreGetExternblServiceIDsFunc) PushReturn(r0 []int64, r1 error) {
	f.PushHook(func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
		return r0, r1
	})
}

func (f *SourcerStoreGetExternblServiceIDsFunc) nextHook() func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreGetExternblServiceIDsFunc) bppendCbll(r0 SourcerStoreGetExternblServiceIDsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreGetExternblServiceIDsFuncCbll
// objects describing the invocbtions of this function.
func (f *SourcerStoreGetExternblServiceIDsFunc) History() []SourcerStoreGetExternblServiceIDsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreGetExternblServiceIDsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreGetExternblServiceIDsFuncCbll is bn object thbt describes bn
// invocbtion of method GetExternblServiceIDs on bn instbnce of
// MockSourcerStore.
type SourcerStoreGetExternblServiceIDsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetExternblServiceIDsOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []int64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreGetExternblServiceIDsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreGetExternblServiceIDsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SourcerStoreGetSiteCredentiblFunc describes the behbvior when the
// GetSiteCredentibl method of the pbrent MockSourcerStore instbnce is
// invoked.
type SourcerStoreGetSiteCredentiblFunc struct {
	defbultHook func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error)
	hooks       []func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error)
	history     []SourcerStoreGetSiteCredentiblFuncCbll
	mutex       sync.Mutex
}

// GetSiteCredentibl delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) GetSiteCredentibl(v0 context.Context, v1 store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error) {
	r0, r1 := m.GetSiteCredentiblFunc.nextHook()(v0, v1)
	m.GetSiteCredentiblFunc.bppendCbll(SourcerStoreGetSiteCredentiblFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetSiteCredentibl
// method of the pbrent MockSourcerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *SourcerStoreGetSiteCredentiblFunc) SetDefbultHook(hook func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetSiteCredentibl method of the pbrent MockSourcerStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SourcerStoreGetSiteCredentiblFunc) PushHook(hook func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreGetSiteCredentiblFunc) SetDefbultReturn(r0 *types1.SiteCredentibl, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreGetSiteCredentiblFunc) PushReturn(r0 *types1.SiteCredentibl, r1 error) {
	f.PushHook(func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error) {
		return r0, r1
	})
}

func (f *SourcerStoreGetSiteCredentiblFunc) nextHook() func(context.Context, store.GetSiteCredentiblOpts) (*types1.SiteCredentibl, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreGetSiteCredentiblFunc) bppendCbll(r0 SourcerStoreGetSiteCredentiblFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreGetSiteCredentiblFuncCbll
// objects describing the invocbtions of this function.
func (f *SourcerStoreGetSiteCredentiblFunc) History() []SourcerStoreGetSiteCredentiblFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreGetSiteCredentiblFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreGetSiteCredentiblFuncCbll is bn object thbt describes bn
// invocbtion of method GetSiteCredentibl on bn instbnce of
// MockSourcerStore.
type SourcerStoreGetSiteCredentiblFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetSiteCredentiblOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types1.SiteCredentibl
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreGetSiteCredentiblFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreGetSiteCredentiblFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SourcerStoreGitHubAppsStoreFunc describes the behbvior when the
// GitHubAppsStore method of the pbrent MockSourcerStore instbnce is
// invoked.
type SourcerStoreGitHubAppsStoreFunc struct {
	defbultHook func() store1.GitHubAppsStore
	hooks       []func() store1.GitHubAppsStore
	history     []SourcerStoreGitHubAppsStoreFuncCbll
	mutex       sync.Mutex
}

// GitHubAppsStore delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) GitHubAppsStore() store1.GitHubAppsStore {
	r0 := m.GitHubAppsStoreFunc.nextHook()()
	m.GitHubAppsStoreFunc.bppendCbll(SourcerStoreGitHubAppsStoreFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GitHubAppsStore
// method of the pbrent MockSourcerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *SourcerStoreGitHubAppsStoreFunc) SetDefbultHook(hook func() store1.GitHubAppsStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GitHubAppsStore method of the pbrent MockSourcerStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SourcerStoreGitHubAppsStoreFunc) PushHook(hook func() store1.GitHubAppsStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreGitHubAppsStoreFunc) SetDefbultReturn(r0 store1.GitHubAppsStore) {
	f.SetDefbultHook(func() store1.GitHubAppsStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreGitHubAppsStoreFunc) PushReturn(r0 store1.GitHubAppsStore) {
	f.PushHook(func() store1.GitHubAppsStore {
		return r0
	})
}

func (f *SourcerStoreGitHubAppsStoreFunc) nextHook() func() store1.GitHubAppsStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreGitHubAppsStoreFunc) bppendCbll(r0 SourcerStoreGitHubAppsStoreFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreGitHubAppsStoreFuncCbll objects
// describing the invocbtions of this function.
func (f *SourcerStoreGitHubAppsStoreFunc) History() []SourcerStoreGitHubAppsStoreFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreGitHubAppsStoreFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreGitHubAppsStoreFuncCbll is bn object thbt describes bn
// invocbtion of method GitHubAppsStore on bn instbnce of MockSourcerStore.
type SourcerStoreGitHubAppsStoreFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 store1.GitHubAppsStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreGitHubAppsStoreFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreGitHubAppsStoreFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SourcerStoreReposFunc describes the behbvior when the Repos method of the
// pbrent MockSourcerStore instbnce is invoked.
type SourcerStoreReposFunc struct {
	defbultHook func() dbtbbbse.RepoStore
	hooks       []func() dbtbbbse.RepoStore
	history     []SourcerStoreReposFuncCbll
	mutex       sync.Mutex
}

// Repos delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) Repos() dbtbbbse.RepoStore {
	r0 := m.ReposFunc.nextHook()()
	m.ReposFunc.bppendCbll(SourcerStoreReposFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Repos method of the
// pbrent MockSourcerStore instbnce is invoked bnd the hook queue is empty.
func (f *SourcerStoreReposFunc) SetDefbultHook(hook func() dbtbbbse.RepoStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Repos method of the pbrent MockSourcerStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *SourcerStoreReposFunc) PushHook(hook func() dbtbbbse.RepoStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreReposFunc) SetDefbultReturn(r0 dbtbbbse.RepoStore) {
	f.SetDefbultHook(func() dbtbbbse.RepoStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreReposFunc) PushReturn(r0 dbtbbbse.RepoStore) {
	f.PushHook(func() dbtbbbse.RepoStore {
		return r0
	})
}

func (f *SourcerStoreReposFunc) nextHook() func() dbtbbbse.RepoStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreReposFunc) bppendCbll(r0 SourcerStoreReposFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreReposFuncCbll objects
// describing the invocbtions of this function.
func (f *SourcerStoreReposFunc) History() []SourcerStoreReposFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreReposFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreReposFuncCbll is bn object thbt describes bn invocbtion of
// method Repos on bn instbnce of MockSourcerStore.
type SourcerStoreReposFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.RepoStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreReposFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreReposFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SourcerStoreUserCredentiblsFunc describes the behbvior when the
// UserCredentibls method of the pbrent MockSourcerStore instbnce is
// invoked.
type SourcerStoreUserCredentiblsFunc struct {
	defbultHook func() dbtbbbse.UserCredentiblsStore
	hooks       []func() dbtbbbse.UserCredentiblsStore
	history     []SourcerStoreUserCredentiblsFuncCbll
	mutex       sync.Mutex
}

// UserCredentibls delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSourcerStore) UserCredentibls() dbtbbbse.UserCredentiblsStore {
	r0 := m.UserCredentiblsFunc.nextHook()()
	m.UserCredentiblsFunc.bppendCbll(SourcerStoreUserCredentiblsFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UserCredentibls
// method of the pbrent MockSourcerStore instbnce is invoked bnd the hook
// queue is empty.
func (f *SourcerStoreUserCredentiblsFunc) SetDefbultHook(hook func() dbtbbbse.UserCredentiblsStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UserCredentibls method of the pbrent MockSourcerStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SourcerStoreUserCredentiblsFunc) PushHook(hook func() dbtbbbse.UserCredentiblsStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SourcerStoreUserCredentiblsFunc) SetDefbultReturn(r0 dbtbbbse.UserCredentiblsStore) {
	f.SetDefbultHook(func() dbtbbbse.UserCredentiblsStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SourcerStoreUserCredentiblsFunc) PushReturn(r0 dbtbbbse.UserCredentiblsStore) {
	f.PushHook(func() dbtbbbse.UserCredentiblsStore {
		return r0
	})
}

func (f *SourcerStoreUserCredentiblsFunc) nextHook() func() dbtbbbse.UserCredentiblsStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SourcerStoreUserCredentiblsFunc) bppendCbll(r0 SourcerStoreUserCredentiblsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SourcerStoreUserCredentiblsFuncCbll objects
// describing the invocbtions of this function.
func (f *SourcerStoreUserCredentiblsFunc) History() []SourcerStoreUserCredentiblsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SourcerStoreUserCredentiblsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SourcerStoreUserCredentiblsFuncCbll is bn object thbt describes bn
// invocbtion of method UserCredentibls on bn instbnce of MockSourcerStore.
type SourcerStoreUserCredentiblsFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.UserCredentiblsStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SourcerStoreUserCredentiblsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SourcerStoreUserCredentiblsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockBitbucketCloudClient is b mock implementbtion of the Client interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud) used
// for unit testing.
type MockBitbucketCloudClient struct {
	// AllCurrentUserEmbilsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AllCurrentUserEmbils.
	AllCurrentUserEmbilsFunc *BitbucketCloudClientAllCurrentUserEmbilsFunc
	// AuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method Authenticbtor.
	AuthenticbtorFunc *BitbucketCloudClientAuthenticbtorFunc
	// CrebtePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebtePullRequest.
	CrebtePullRequestFunc *BitbucketCloudClientCrebtePullRequestFunc
	// CrebtePullRequestCommentFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebtePullRequestComment.
	CrebtePullRequestCommentFunc *BitbucketCloudClientCrebtePullRequestCommentFunc
	// CurrentUserFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CurrentUser.
	CurrentUserFunc *BitbucketCloudClientCurrentUserFunc
	// CurrentUserEmbilsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CurrentUserEmbils.
	CurrentUserEmbilsFunc *BitbucketCloudClientCurrentUserEmbilsFunc
	// DeclinePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeclinePullRequest.
	DeclinePullRequestFunc *BitbucketCloudClientDeclinePullRequestFunc
	// ForkRepositoryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ForkRepository.
	ForkRepositoryFunc *BitbucketCloudClientForkRepositoryFunc
	// GetPullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPullRequest.
	GetPullRequestFunc *BitbucketCloudClientGetPullRequestFunc
	// GetPullRequestStbtusesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPullRequestStbtuses.
	GetPullRequestStbtusesFunc *BitbucketCloudClientGetPullRequestStbtusesFunc
	// ListExplicitUserPermsForRepoFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListExplicitUserPermsForRepo.
	ListExplicitUserPermsForRepoFunc *BitbucketCloudClientListExplicitUserPermsForRepoFunc
	// MergePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method MergePullRequest.
	MergePullRequestFunc *BitbucketCloudClientMergePullRequestFunc
	// PingFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Ping.
	PingFunc *BitbucketCloudClientPingFunc
	// RepoFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Repo.
	RepoFunc *BitbucketCloudClientRepoFunc
	// ReposFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Repos.
	ReposFunc *BitbucketCloudClientReposFunc
	// UpdbtePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbtePullRequest.
	UpdbtePullRequestFunc *BitbucketCloudClientUpdbtePullRequestFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *BitbucketCloudClientWithAuthenticbtorFunc
}

// NewMockBitbucketCloudClient crebtes b new mock of the Client interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockBitbucketCloudClient() *MockBitbucketCloudClient {
	return &MockBitbucketCloudClient{
		AllCurrentUserEmbilsFunc: &BitbucketCloudClientAllCurrentUserEmbilsFunc{
			defbultHook: func(context.Context) (r0 []*bitbucketcloud.UserEmbil, r1 error) {
				return
			},
		},
		AuthenticbtorFunc: &BitbucketCloudClientAuthenticbtorFunc{
			defbultHook: func() (r0 buth.Authenticbtor) {
				return
			},
		},
		CrebtePullRequestFunc: &BitbucketCloudClientCrebtePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (r0 *bitbucketcloud.PullRequest, r1 error) {
				return
			},
		},
		CrebtePullRequestCommentFunc: &BitbucketCloudClientCrebtePullRequestCommentFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (r0 *bitbucketcloud.Comment, r1 error) {
				return
			},
		},
		CurrentUserFunc: &BitbucketCloudClientCurrentUserFunc{
			defbultHook: func(context.Context) (r0 *bitbucketcloud.User, r1 error) {
				return
			},
		},
		CurrentUserEmbilsFunc: &BitbucketCloudClientCurrentUserEmbilsFunc{
			defbultHook: func(context.Context, *bitbucketcloud.PbgeToken) (r0 []*bitbucketcloud.UserEmbil, r1 *bitbucketcloud.PbgeToken, r2 error) {
				return
			},
		},
		DeclinePullRequestFunc: &BitbucketCloudClientDeclinePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64) (r0 *bitbucketcloud.PullRequest, r1 error) {
				return
			},
		},
		ForkRepositoryFunc: &BitbucketCloudClientForkRepositoryFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (r0 *bitbucketcloud.Repo, r1 error) {
				return
			},
		},
		GetPullRequestFunc: &BitbucketCloudClientGetPullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64) (r0 *bitbucketcloud.PullRequest, r1 error) {
				return
			},
		},
		GetPullRequestStbtusesFunc: &BitbucketCloudClientGetPullRequestStbtusesFunc{
			defbultHook: func(*bitbucketcloud.Repo, int64) (r0 *bitbucketcloud.PbginbtedResultSet, r1 error) {
				return
			},
		},
		ListExplicitUserPermsForRepoFunc: &BitbucketCloudClientListExplicitUserPermsForRepoFunc{
			defbultHook: func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) (r0 []*bitbucketcloud.Account, r1 *bitbucketcloud.PbgeToken, r2 error) {
				return
			},
		},
		MergePullRequestFunc: &BitbucketCloudClientMergePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (r0 *bitbucketcloud.PullRequest, r1 error) {
				return
			},
		},
		PingFunc: &BitbucketCloudClientPingFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		RepoFunc: &BitbucketCloudClientRepoFunc{
			defbultHook: func(context.Context, string, string) (r0 *bitbucketcloud.Repo, r1 error) {
				return
			},
		},
		ReposFunc: &BitbucketCloudClientReposFunc{
			defbultHook: func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) (r0 []*bitbucketcloud.Repo, r1 *bitbucketcloud.PbgeToken, r2 error) {
				return
			},
		},
		UpdbtePullRequestFunc: &BitbucketCloudClientUpdbtePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (r0 *bitbucketcloud.PullRequest, r1 error) {
				return
			},
		},
		WithAuthenticbtorFunc: &BitbucketCloudClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 bitbucketcloud.Client) {
				return
			},
		},
	}
}

// NewStrictMockBitbucketCloudClient crebtes b new mock of the Client
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockBitbucketCloudClient() *MockBitbucketCloudClient {
	return &MockBitbucketCloudClient{
		AllCurrentUserEmbilsFunc: &BitbucketCloudClientAllCurrentUserEmbilsFunc{
			defbultHook: func(context.Context) ([]*bitbucketcloud.UserEmbil, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.AllCurrentUserEmbils")
			},
		},
		AuthenticbtorFunc: &BitbucketCloudClientAuthenticbtorFunc{
			defbultHook: func() buth.Authenticbtor {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.Authenticbtor")
			},
		},
		CrebtePullRequestFunc: &BitbucketCloudClientCrebtePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.CrebtePullRequest")
			},
		},
		CrebtePullRequestCommentFunc: &BitbucketCloudClientCrebtePullRequestCommentFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.CrebtePullRequestComment")
			},
		},
		CurrentUserFunc: &BitbucketCloudClientCurrentUserFunc{
			defbultHook: func(context.Context) (*bitbucketcloud.User, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.CurrentUser")
			},
		},
		CurrentUserEmbilsFunc: &BitbucketCloudClientCurrentUserEmbilsFunc{
			defbultHook: func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.CurrentUserEmbils")
			},
		},
		DeclinePullRequestFunc: &BitbucketCloudClientDeclinePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.DeclinePullRequest")
			},
		},
		ForkRepositoryFunc: &BitbucketCloudClientForkRepositoryFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.ForkRepository")
			},
		},
		GetPullRequestFunc: &BitbucketCloudClientGetPullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.GetPullRequest")
			},
		},
		GetPullRequestStbtusesFunc: &BitbucketCloudClientGetPullRequestStbtusesFunc{
			defbultHook: func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.GetPullRequestStbtuses")
			},
		},
		ListExplicitUserPermsForRepoFunc: &BitbucketCloudClientListExplicitUserPermsForRepoFunc{
			defbultHook: func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.ListExplicitUserPermsForRepo")
			},
		},
		MergePullRequestFunc: &BitbucketCloudClientMergePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.MergePullRequest")
			},
		},
		PingFunc: &BitbucketCloudClientPingFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.Ping")
			},
		},
		RepoFunc: &BitbucketCloudClientRepoFunc{
			defbultHook: func(context.Context, string, string) (*bitbucketcloud.Repo, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.Repo")
			},
		},
		ReposFunc: &BitbucketCloudClientReposFunc{
			defbultHook: func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.Repos")
			},
		},
		UpdbtePullRequestFunc: &BitbucketCloudClientUpdbtePullRequestFunc{
			defbultHook: func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.UpdbtePullRequest")
			},
		},
		WithAuthenticbtorFunc: &BitbucketCloudClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) bitbucketcloud.Client {
				pbnic("unexpected invocbtion of MockBitbucketCloudClient.WithAuthenticbtor")
			},
		},
	}
}

// NewMockBitbucketCloudClientFrom crebtes b new mock of the
// MockBitbucketCloudClient interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockBitbucketCloudClientFrom(i bitbucketcloud.Client) *MockBitbucketCloudClient {
	return &MockBitbucketCloudClient{
		AllCurrentUserEmbilsFunc: &BitbucketCloudClientAllCurrentUserEmbilsFunc{
			defbultHook: i.AllCurrentUserEmbils,
		},
		AuthenticbtorFunc: &BitbucketCloudClientAuthenticbtorFunc{
			defbultHook: i.Authenticbtor,
		},
		CrebtePullRequestFunc: &BitbucketCloudClientCrebtePullRequestFunc{
			defbultHook: i.CrebtePullRequest,
		},
		CrebtePullRequestCommentFunc: &BitbucketCloudClientCrebtePullRequestCommentFunc{
			defbultHook: i.CrebtePullRequestComment,
		},
		CurrentUserFunc: &BitbucketCloudClientCurrentUserFunc{
			defbultHook: i.CurrentUser,
		},
		CurrentUserEmbilsFunc: &BitbucketCloudClientCurrentUserEmbilsFunc{
			defbultHook: i.CurrentUserEmbils,
		},
		DeclinePullRequestFunc: &BitbucketCloudClientDeclinePullRequestFunc{
			defbultHook: i.DeclinePullRequest,
		},
		ForkRepositoryFunc: &BitbucketCloudClientForkRepositoryFunc{
			defbultHook: i.ForkRepository,
		},
		GetPullRequestFunc: &BitbucketCloudClientGetPullRequestFunc{
			defbultHook: i.GetPullRequest,
		},
		GetPullRequestStbtusesFunc: &BitbucketCloudClientGetPullRequestStbtusesFunc{
			defbultHook: i.GetPullRequestStbtuses,
		},
		ListExplicitUserPermsForRepoFunc: &BitbucketCloudClientListExplicitUserPermsForRepoFunc{
			defbultHook: i.ListExplicitUserPermsForRepo,
		},
		MergePullRequestFunc: &BitbucketCloudClientMergePullRequestFunc{
			defbultHook: i.MergePullRequest,
		},
		PingFunc: &BitbucketCloudClientPingFunc{
			defbultHook: i.Ping,
		},
		RepoFunc: &BitbucketCloudClientRepoFunc{
			defbultHook: i.Repo,
		},
		ReposFunc: &BitbucketCloudClientReposFunc{
			defbultHook: i.Repos,
		},
		UpdbtePullRequestFunc: &BitbucketCloudClientUpdbtePullRequestFunc{
			defbultHook: i.UpdbtePullRequest,
		},
		WithAuthenticbtorFunc: &BitbucketCloudClientWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
	}
}

// BitbucketCloudClientAllCurrentUserEmbilsFunc describes the behbvior when
// the AllCurrentUserEmbils method of the pbrent MockBitbucketCloudClient
// instbnce is invoked.
type BitbucketCloudClientAllCurrentUserEmbilsFunc struct {
	defbultHook func(context.Context) ([]*bitbucketcloud.UserEmbil, error)
	hooks       []func(context.Context) ([]*bitbucketcloud.UserEmbil, error)
	history     []BitbucketCloudClientAllCurrentUserEmbilsFuncCbll
	mutex       sync.Mutex
}

// AllCurrentUserEmbils delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) AllCurrentUserEmbils(v0 context.Context) ([]*bitbucketcloud.UserEmbil, error) {
	r0, r1 := m.AllCurrentUserEmbilsFunc.nextHook()(v0)
	m.AllCurrentUserEmbilsFunc.bppendCbll(BitbucketCloudClientAllCurrentUserEmbilsFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AllCurrentUserEmbils
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) SetDefbultHook(hook func(context.Context) ([]*bitbucketcloud.UserEmbil, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AllCurrentUserEmbils method of the pbrent MockBitbucketCloudClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) PushHook(hook func(context.Context) ([]*bitbucketcloud.UserEmbil, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) SetDefbultReturn(r0 []*bitbucketcloud.UserEmbil, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]*bitbucketcloud.UserEmbil, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) PushReturn(r0 []*bitbucketcloud.UserEmbil, r1 error) {
	f.PushHook(func(context.Context) ([]*bitbucketcloud.UserEmbil, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) nextHook() func(context.Context) ([]*bitbucketcloud.UserEmbil, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) bppendCbll(r0 BitbucketCloudClientAllCurrentUserEmbilsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientAllCurrentUserEmbilsFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientAllCurrentUserEmbilsFunc) History() []BitbucketCloudClientAllCurrentUserEmbilsFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientAllCurrentUserEmbilsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientAllCurrentUserEmbilsFuncCbll is bn object thbt
// describes bn invocbtion of method AllCurrentUserEmbils on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientAllCurrentUserEmbilsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*bitbucketcloud.UserEmbil
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientAllCurrentUserEmbilsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientAllCurrentUserEmbilsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientAuthenticbtorFunc describes the behbvior when the
// Authenticbtor method of the pbrent MockBitbucketCloudClient instbnce is
// invoked.
type BitbucketCloudClientAuthenticbtorFunc struct {
	defbultHook func() buth.Authenticbtor
	hooks       []func() buth.Authenticbtor
	history     []BitbucketCloudClientAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// Authenticbtor delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) Authenticbtor() buth.Authenticbtor {
	r0 := m.AuthenticbtorFunc.nextHook()()
	m.AuthenticbtorFunc.bppendCbll(BitbucketCloudClientAuthenticbtorFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Authenticbtor method
// of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the hook
// queue is empty.
func (f *BitbucketCloudClientAuthenticbtorFunc) SetDefbultHook(hook func() buth.Authenticbtor) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Authenticbtor method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientAuthenticbtorFunc) PushHook(hook func() buth.Authenticbtor) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientAuthenticbtorFunc) SetDefbultReturn(r0 buth.Authenticbtor) {
	f.SetDefbultHook(func() buth.Authenticbtor {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientAuthenticbtorFunc) PushReturn(r0 buth.Authenticbtor) {
	f.PushHook(func() buth.Authenticbtor {
		return r0
	})
}

func (f *BitbucketCloudClientAuthenticbtorFunc) nextHook() func() buth.Authenticbtor {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientAuthenticbtorFunc) bppendCbll(r0 BitbucketCloudClientAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientAuthenticbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *BitbucketCloudClientAuthenticbtorFunc) History() []BitbucketCloudClientAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method Authenticbtor on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientAuthenticbtorFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 buth.Authenticbtor
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// BitbucketCloudClientCrebtePullRequestFunc describes the behbvior when the
// CrebtePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// is invoked.
type BitbucketCloudClientCrebtePullRequestFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)
	history     []BitbucketCloudClientCrebtePullRequestFuncCbll
	mutex       sync.Mutex
}

// CrebtePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) CrebtePullRequest(v0 context.Context, v1 *bitbucketcloud.Repo, v2 bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
	r0, r1 := m.CrebtePullRequestFunc.nextHook()(v0, v1, v2)
	m.CrebtePullRequestFunc.bppendCbll(BitbucketCloudClientCrebtePullRequestFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CrebtePullRequest
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientCrebtePullRequestFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebtePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientCrebtePullRequestFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientCrebtePullRequestFunc) SetDefbultReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientCrebtePullRequestFunc) PushReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientCrebtePullRequestFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientCrebtePullRequestFunc) bppendCbll(r0 BitbucketCloudClientCrebtePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientCrebtePullRequestFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientCrebtePullRequestFunc) History() []BitbucketCloudClientCrebtePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientCrebtePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientCrebtePullRequestFuncCbll is bn object thbt describes
// bn invocbtion of method CrebtePullRequest on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientCrebtePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bitbucketcloud.PullRequestInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientCrebtePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientCrebtePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientCrebtePullRequestCommentFunc describes the behbvior
// when the CrebtePullRequestComment method of the pbrent
// MockBitbucketCloudClient instbnce is invoked.
type BitbucketCloudClientCrebtePullRequestCommentFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error)
	history     []BitbucketCloudClientCrebtePullRequestCommentFuncCbll
	mutex       sync.Mutex
}

// CrebtePullRequestComment delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) CrebtePullRequestComment(v0 context.Context, v1 *bitbucketcloud.Repo, v2 int64, v3 bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
	r0, r1 := m.CrebtePullRequestCommentFunc.nextHook()(v0, v1, v2, v3)
	m.CrebtePullRequestCommentFunc.bppendCbll(BitbucketCloudClientCrebtePullRequestCommentFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebtePullRequestComment method of the pbrent MockBitbucketCloudClient
// instbnce is invoked bnd the hook queue is empty.
func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebtePullRequestComment method of the pbrent MockBitbucketCloudClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) SetDefbultReturn(r0 *bitbucketcloud.Comment, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) PushReturn(r0 *bitbucketcloud.Comment, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) bppendCbll(r0 BitbucketCloudClientCrebtePullRequestCommentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientCrebtePullRequestCommentFuncCbll objects describing
// the invocbtions of this function.
func (f *BitbucketCloudClientCrebtePullRequestCommentFunc) History() []BitbucketCloudClientCrebtePullRequestCommentFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientCrebtePullRequestCommentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientCrebtePullRequestCommentFuncCbll is bn object thbt
// describes bn invocbtion of method CrebtePullRequestComment on bn instbnce
// of MockBitbucketCloudClient.
type BitbucketCloudClientCrebtePullRequestCommentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int64
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bitbucketcloud.CommentInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.Comment
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientCrebtePullRequestCommentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientCrebtePullRequestCommentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientCurrentUserFunc describes the behbvior when the
// CurrentUser method of the pbrent MockBitbucketCloudClient instbnce is
// invoked.
type BitbucketCloudClientCurrentUserFunc struct {
	defbultHook func(context.Context) (*bitbucketcloud.User, error)
	hooks       []func(context.Context) (*bitbucketcloud.User, error)
	history     []BitbucketCloudClientCurrentUserFuncCbll
	mutex       sync.Mutex
}

// CurrentUser delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) CurrentUser(v0 context.Context) (*bitbucketcloud.User, error) {
	r0, r1 := m.CurrentUserFunc.nextHook()(v0)
	m.CurrentUserFunc.bppendCbll(BitbucketCloudClientCurrentUserFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CurrentUser method
// of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the hook
// queue is empty.
func (f *BitbucketCloudClientCurrentUserFunc) SetDefbultHook(hook func(context.Context) (*bitbucketcloud.User, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CurrentUser method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientCurrentUserFunc) PushHook(hook func(context.Context) (*bitbucketcloud.User, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientCurrentUserFunc) SetDefbultReturn(r0 *bitbucketcloud.User, r1 error) {
	f.SetDefbultHook(func(context.Context) (*bitbucketcloud.User, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientCurrentUserFunc) PushReturn(r0 *bitbucketcloud.User, r1 error) {
	f.PushHook(func(context.Context) (*bitbucketcloud.User, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientCurrentUserFunc) nextHook() func(context.Context) (*bitbucketcloud.User, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientCurrentUserFunc) bppendCbll(r0 BitbucketCloudClientCurrentUserFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientCurrentUserFuncCbll
// objects describing the invocbtions of this function.
func (f *BitbucketCloudClientCurrentUserFunc) History() []BitbucketCloudClientCurrentUserFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientCurrentUserFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientCurrentUserFuncCbll is bn object thbt describes bn
// invocbtion of method CurrentUser on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientCurrentUserFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.User
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientCurrentUserFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientCurrentUserFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientCurrentUserEmbilsFunc describes the behbvior when the
// CurrentUserEmbils method of the pbrent MockBitbucketCloudClient instbnce
// is invoked.
type BitbucketCloudClientCurrentUserEmbilsFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error)
	hooks       []func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error)
	history     []BitbucketCloudClientCurrentUserEmbilsFuncCbll
	mutex       sync.Mutex
}

// CurrentUserEmbils delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) CurrentUserEmbils(v0 context.Context, v1 *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error) {
	r0, r1, r2 := m.CurrentUserEmbilsFunc.nextHook()(v0, v1)
	m.CurrentUserEmbilsFunc.bppendCbll(BitbucketCloudClientCurrentUserEmbilsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the CurrentUserEmbils
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientCurrentUserEmbilsFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CurrentUserEmbils method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientCurrentUserEmbilsFunc) PushHook(hook func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientCurrentUserEmbilsFunc) SetDefbultReturn(r0 []*bitbucketcloud.UserEmbil, r1 *bitbucketcloud.PbgeToken, r2 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientCurrentUserEmbilsFunc) PushReturn(r0 []*bitbucketcloud.UserEmbil, r1 *bitbucketcloud.PbgeToken, r2 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error) {
		return r0, r1, r2
	})
}

func (f *BitbucketCloudClientCurrentUserEmbilsFunc) nextHook() func(context.Context, *bitbucketcloud.PbgeToken) ([]*bitbucketcloud.UserEmbil, *bitbucketcloud.PbgeToken, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientCurrentUserEmbilsFunc) bppendCbll(r0 BitbucketCloudClientCurrentUserEmbilsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientCurrentUserEmbilsFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientCurrentUserEmbilsFunc) History() []BitbucketCloudClientCurrentUserEmbilsFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientCurrentUserEmbilsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientCurrentUserEmbilsFuncCbll is bn object thbt describes
// bn invocbtion of method CurrentUserEmbils on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientCurrentUserEmbilsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.PbgeToken
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*bitbucketcloud.UserEmbil
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 *bitbucketcloud.PbgeToken
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientCurrentUserEmbilsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientCurrentUserEmbilsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// BitbucketCloudClientDeclinePullRequestFunc describes the behbvior when
// the DeclinePullRequest method of the pbrent MockBitbucketCloudClient
// instbnce is invoked.
type BitbucketCloudClientDeclinePullRequestFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)
	history     []BitbucketCloudClientDeclinePullRequestFuncCbll
	mutex       sync.Mutex
}

// DeclinePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) DeclinePullRequest(v0 context.Context, v1 *bitbucketcloud.Repo, v2 int64) (*bitbucketcloud.PullRequest, error) {
	r0, r1 := m.DeclinePullRequestFunc.nextHook()(v0, v1, v2)
	m.DeclinePullRequestFunc.bppendCbll(BitbucketCloudClientDeclinePullRequestFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DeclinePullRequest
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientDeclinePullRequestFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeclinePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientDeclinePullRequestFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientDeclinePullRequestFunc) SetDefbultReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientDeclinePullRequestFunc) PushReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientDeclinePullRequestFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientDeclinePullRequestFunc) bppendCbll(r0 BitbucketCloudClientDeclinePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientDeclinePullRequestFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientDeclinePullRequestFunc) History() []BitbucketCloudClientDeclinePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientDeclinePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientDeclinePullRequestFuncCbll is bn object thbt
// describes bn invocbtion of method DeclinePullRequest on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientDeclinePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientDeclinePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientDeclinePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientForkRepositoryFunc describes the behbvior when the
// ForkRepository method of the pbrent MockBitbucketCloudClient instbnce is
// invoked.
type BitbucketCloudClientForkRepositoryFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error)
	history     []BitbucketCloudClientForkRepositoryFuncCbll
	mutex       sync.Mutex
}

// ForkRepository delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) ForkRepository(v0 context.Context, v1 *bitbucketcloud.Repo, v2 bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
	r0, r1 := m.ForkRepositoryFunc.nextHook()(v0, v1, v2)
	m.ForkRepositoryFunc.bppendCbll(BitbucketCloudClientForkRepositoryFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ForkRepository
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientForkRepositoryFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ForkRepository method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientForkRepositoryFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientForkRepositoryFunc) SetDefbultReturn(r0 *bitbucketcloud.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientForkRepositoryFunc) PushReturn(r0 *bitbucketcloud.Repo, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientForkRepositoryFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientForkRepositoryFunc) bppendCbll(r0 BitbucketCloudClientForkRepositoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientForkRepositoryFuncCbll
// objects describing the invocbtions of this function.
func (f *BitbucketCloudClientForkRepositoryFunc) History() []BitbucketCloudClientForkRepositoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientForkRepositoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientForkRepositoryFuncCbll is bn object thbt describes bn
// invocbtion of method ForkRepository on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientForkRepositoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bitbucketcloud.ForkInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientForkRepositoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientForkRepositoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientGetPullRequestFunc describes the behbvior when the
// GetPullRequest method of the pbrent MockBitbucketCloudClient instbnce is
// invoked.
type BitbucketCloudClientGetPullRequestFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)
	history     []BitbucketCloudClientGetPullRequestFuncCbll
	mutex       sync.Mutex
}

// GetPullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) GetPullRequest(v0 context.Context, v1 *bitbucketcloud.Repo, v2 int64) (*bitbucketcloud.PullRequest, error) {
	r0, r1 := m.GetPullRequestFunc.nextHook()(v0, v1, v2)
	m.GetPullRequestFunc.bppendCbll(BitbucketCloudClientGetPullRequestFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetPullRequest
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientGetPullRequestFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPullRequest method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientGetPullRequestFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientGetPullRequestFunc) SetDefbultReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientGetPullRequestFunc) PushReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientGetPullRequestFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, int64) (*bitbucketcloud.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientGetPullRequestFunc) bppendCbll(r0 BitbucketCloudClientGetPullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientGetPullRequestFuncCbll
// objects describing the invocbtions of this function.
func (f *BitbucketCloudClientGetPullRequestFunc) History() []BitbucketCloudClientGetPullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientGetPullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientGetPullRequestFuncCbll is bn object thbt describes bn
// invocbtion of method GetPullRequest on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientGetPullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientGetPullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientGetPullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientGetPullRequestStbtusesFunc describes the behbvior
// when the GetPullRequestStbtuses method of the pbrent
// MockBitbucketCloudClient instbnce is invoked.
type BitbucketCloudClientGetPullRequestStbtusesFunc struct {
	defbultHook func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error)
	hooks       []func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error)
	history     []BitbucketCloudClientGetPullRequestStbtusesFuncCbll
	mutex       sync.Mutex
}

// GetPullRequestStbtuses delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) GetPullRequestStbtuses(v0 *bitbucketcloud.Repo, v1 int64) (*bitbucketcloud.PbginbtedResultSet, error) {
	r0, r1 := m.GetPullRequestStbtusesFunc.nextHook()(v0, v1)
	m.GetPullRequestStbtusesFunc.bppendCbll(BitbucketCloudClientGetPullRequestStbtusesFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetPullRequestStbtuses method of the pbrent MockBitbucketCloudClient
// instbnce is invoked bnd the hook queue is empty.
func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) SetDefbultHook(hook func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPullRequestStbtuses method of the pbrent MockBitbucketCloudClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) PushHook(hook func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) SetDefbultReturn(r0 *bitbucketcloud.PbginbtedResultSet, r1 error) {
	f.SetDefbultHook(func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) PushReturn(r0 *bitbucketcloud.PbginbtedResultSet, r1 error) {
	f.PushHook(func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) nextHook() func(*bitbucketcloud.Repo, int64) (*bitbucketcloud.PbginbtedResultSet, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) bppendCbll(r0 BitbucketCloudClientGetPullRequestStbtusesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientGetPullRequestStbtusesFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientGetPullRequestStbtusesFunc) History() []BitbucketCloudClientGetPullRequestStbtusesFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientGetPullRequestStbtusesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientGetPullRequestStbtusesFuncCbll is bn object thbt
// describes bn invocbtion of method GetPullRequestStbtuses on bn instbnce
// of MockBitbucketCloudClient.
type BitbucketCloudClientGetPullRequestStbtusesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *bitbucketcloud.Repo
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.PbginbtedResultSet
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientGetPullRequestStbtusesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientGetPullRequestStbtusesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientListExplicitUserPermsForRepoFunc describes the
// behbvior when the ListExplicitUserPermsForRepo method of the pbrent
// MockBitbucketCloudClient instbnce is invoked.
type BitbucketCloudClientListExplicitUserPermsForRepoFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error)
	hooks       []func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error)
	history     []BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll
	mutex       sync.Mutex
}

// ListExplicitUserPermsForRepo delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) ListExplicitUserPermsForRepo(v0 context.Context, v1 *bitbucketcloud.PbgeToken, v2 string, v3 string, v4 *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error) {
	r0, r1, r2 := m.ListExplicitUserPermsForRepoFunc.nextHook()(v0, v1, v2, v3, v4)
	m.ListExplicitUserPermsForRepoFunc.bppendCbll(BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ListExplicitUserPermsForRepo method of the pbrent
// MockBitbucketCloudClient instbnce is invoked bnd the hook queue is empty.
func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListExplicitUserPermsForRepo method of the pbrent
// MockBitbucketCloudClient instbnce invokes the hook bt the front of the
// queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) PushHook(hook func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) SetDefbultReturn(r0 []*bitbucketcloud.Account, r1 *bitbucketcloud.PbgeToken, r2 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) PushReturn(r0 []*bitbucketcloud.Account, r1 *bitbucketcloud.PbgeToken, r2 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error) {
		return r0, r1, r2
	})
}

func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) nextHook() func(context.Context, *bitbucketcloud.PbgeToken, string, string, *bitbucketcloud.RequestOptions) ([]*bitbucketcloud.Account, *bitbucketcloud.PbgeToken, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) bppendCbll(r0 BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll objects
// describing the invocbtions of this function.
func (f *BitbucketCloudClientListExplicitUserPermsForRepoFunc) History() []BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll is bn object
// thbt describes bn invocbtion of method ListExplicitUserPermsForRepo on bn
// instbnce of MockBitbucketCloudClient.
type BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.PbgeToken
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 *bitbucketcloud.RequestOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*bitbucketcloud.Account
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 *bitbucketcloud.PbgeToken
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientListExplicitUserPermsForRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// BitbucketCloudClientMergePullRequestFunc describes the behbvior when the
// MergePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// is invoked.
type BitbucketCloudClientMergePullRequestFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error)
	history     []BitbucketCloudClientMergePullRequestFuncCbll
	mutex       sync.Mutex
}

// MergePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) MergePullRequest(v0 context.Context, v1 *bitbucketcloud.Repo, v2 int64, v3 bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
	r0, r1 := m.MergePullRequestFunc.nextHook()(v0, v1, v2, v3)
	m.MergePullRequestFunc.bppendCbll(BitbucketCloudClientMergePullRequestFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MergePullRequest
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientMergePullRequestFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MergePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientMergePullRequestFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientMergePullRequestFunc) SetDefbultReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientMergePullRequestFunc) PushReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientMergePullRequestFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientMergePullRequestFunc) bppendCbll(r0 BitbucketCloudClientMergePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientMergePullRequestFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientMergePullRequestFunc) History() []BitbucketCloudClientMergePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientMergePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientMergePullRequestFuncCbll is bn object thbt describes
// bn invocbtion of method MergePullRequest on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientMergePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int64
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bitbucketcloud.MergePullRequestOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientMergePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientMergePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientPingFunc describes the behbvior when the Ping method
// of the pbrent MockBitbucketCloudClient instbnce is invoked.
type BitbucketCloudClientPingFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []BitbucketCloudClientPingFuncCbll
	mutex       sync.Mutex
}

// Ping delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) Ping(v0 context.Context) error {
	r0 := m.PingFunc.nextHook()(v0)
	m.PingFunc.bppendCbll(BitbucketCloudClientPingFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Ping method of the
// pbrent MockBitbucketCloudClient instbnce is invoked bnd the hook queue is
// empty.
func (f *BitbucketCloudClientPingFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Ping method of the pbrent MockBitbucketCloudClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *BitbucketCloudClientPingFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientPingFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientPingFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *BitbucketCloudClientPingFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientPingFunc) bppendCbll(r0 BitbucketCloudClientPingFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientPingFuncCbll objects
// describing the invocbtions of this function.
func (f *BitbucketCloudClientPingFunc) History() []BitbucketCloudClientPingFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientPingFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientPingFuncCbll is bn object thbt describes bn
// invocbtion of method Ping on bn instbnce of MockBitbucketCloudClient.
type BitbucketCloudClientPingFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientPingFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientPingFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// BitbucketCloudClientRepoFunc describes the behbvior when the Repo method
// of the pbrent MockBitbucketCloudClient instbnce is invoked.
type BitbucketCloudClientRepoFunc struct {
	defbultHook func(context.Context, string, string) (*bitbucketcloud.Repo, error)
	hooks       []func(context.Context, string, string) (*bitbucketcloud.Repo, error)
	history     []BitbucketCloudClientRepoFuncCbll
	mutex       sync.Mutex
}

// Repo delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) Repo(v0 context.Context, v1 string, v2 string) (*bitbucketcloud.Repo, error) {
	r0, r1 := m.RepoFunc.nextHook()(v0, v1, v2)
	m.RepoFunc.bppendCbll(BitbucketCloudClientRepoFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Repo method of the
// pbrent MockBitbucketCloudClient instbnce is invoked bnd the hook queue is
// empty.
func (f *BitbucketCloudClientRepoFunc) SetDefbultHook(hook func(context.Context, string, string) (*bitbucketcloud.Repo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Repo method of the pbrent MockBitbucketCloudClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *BitbucketCloudClientRepoFunc) PushHook(hook func(context.Context, string, string) (*bitbucketcloud.Repo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientRepoFunc) SetDefbultReturn(r0 *bitbucketcloud.Repo, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string) (*bitbucketcloud.Repo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientRepoFunc) PushReturn(r0 *bitbucketcloud.Repo, r1 error) {
	f.PushHook(func(context.Context, string, string) (*bitbucketcloud.Repo, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientRepoFunc) nextHook() func(context.Context, string, string) (*bitbucketcloud.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientRepoFunc) bppendCbll(r0 BitbucketCloudClientRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientRepoFuncCbll objects
// describing the invocbtions of this function.
func (f *BitbucketCloudClientRepoFunc) History() []BitbucketCloudClientRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientRepoFuncCbll is bn object thbt describes bn
// invocbtion of method Repo on bn instbnce of MockBitbucketCloudClient.
type BitbucketCloudClientRepoFuncCbll struct {
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
	Result0 *bitbucketcloud.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientReposFunc describes the behbvior when the Repos
// method of the pbrent MockBitbucketCloudClient instbnce is invoked.
type BitbucketCloudClientReposFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error)
	hooks       []func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error)
	history     []BitbucketCloudClientReposFuncCbll
	mutex       sync.Mutex
}

// Repos delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) Repos(v0 context.Context, v1 *bitbucketcloud.PbgeToken, v2 string, v3 *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error) {
	r0, r1, r2 := m.ReposFunc.nextHook()(v0, v1, v2, v3)
	m.ReposFunc.bppendCbll(BitbucketCloudClientReposFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Repos method of the
// pbrent MockBitbucketCloudClient instbnce is invoked bnd the hook queue is
// empty.
func (f *BitbucketCloudClientReposFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Repos method of the pbrent MockBitbucketCloudClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *BitbucketCloudClientReposFunc) PushHook(hook func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientReposFunc) SetDefbultReturn(r0 []*bitbucketcloud.Repo, r1 *bitbucketcloud.PbgeToken, r2 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientReposFunc) PushReturn(r0 []*bitbucketcloud.Repo, r1 *bitbucketcloud.PbgeToken, r2 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error) {
		return r0, r1, r2
	})
}

func (f *BitbucketCloudClientReposFunc) nextHook() func(context.Context, *bitbucketcloud.PbgeToken, string, *bitbucketcloud.ReposOptions) ([]*bitbucketcloud.Repo, *bitbucketcloud.PbgeToken, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientReposFunc) bppendCbll(r0 BitbucketCloudClientReposFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BitbucketCloudClientReposFuncCbll objects
// describing the invocbtions of this function.
func (f *BitbucketCloudClientReposFunc) History() []BitbucketCloudClientReposFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientReposFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientReposFuncCbll is bn object thbt describes bn
// invocbtion of method Repos on bn instbnce of MockBitbucketCloudClient.
type BitbucketCloudClientReposFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.PbgeToken
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *bitbucketcloud.ReposOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*bitbucketcloud.Repo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 *bitbucketcloud.PbgeToken
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientReposFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientReposFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// BitbucketCloudClientUpdbtePullRequestFunc describes the behbvior when the
// UpdbtePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// is invoked.
type BitbucketCloudClientUpdbtePullRequestFunc struct {
	defbultHook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)
	hooks       []func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)
	history     []BitbucketCloudClientUpdbtePullRequestFuncCbll
	mutex       sync.Mutex
}

// UpdbtePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) UpdbtePullRequest(v0 context.Context, v1 *bitbucketcloud.Repo, v2 int64, v3 bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
	r0, r1 := m.UpdbtePullRequestFunc.nextHook()(v0, v1, v2, v3)
	m.UpdbtePullRequestFunc.bppendCbll(BitbucketCloudClientUpdbtePullRequestFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the UpdbtePullRequest
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientUpdbtePullRequestFunc) SetDefbultHook(hook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbtePullRequest method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientUpdbtePullRequestFunc) PushHook(hook func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientUpdbtePullRequestFunc) SetDefbultReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientUpdbtePullRequestFunc) PushReturn(r0 *bitbucketcloud.PullRequest, r1 error) {
	f.PushHook(func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
		return r0, r1
	})
}

func (f *BitbucketCloudClientUpdbtePullRequestFunc) nextHook() func(context.Context, *bitbucketcloud.Repo, int64, bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientUpdbtePullRequestFunc) bppendCbll(r0 BitbucketCloudClientUpdbtePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientUpdbtePullRequestFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientUpdbtePullRequestFunc) History() []BitbucketCloudClientUpdbtePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientUpdbtePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientUpdbtePullRequestFuncCbll is bn object thbt describes
// bn invocbtion of method UpdbtePullRequest on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientUpdbtePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *bitbucketcloud.Repo
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int64
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bitbucketcloud.PullRequestInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *bitbucketcloud.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientUpdbtePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientUpdbtePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BitbucketCloudClientWithAuthenticbtorFunc describes the behbvior when the
// WithAuthenticbtor method of the pbrent MockBitbucketCloudClient instbnce
// is invoked.
type BitbucketCloudClientWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) bitbucketcloud.Client
	hooks       []func(buth.Authenticbtor) bitbucketcloud.Client
	history     []BitbucketCloudClientWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBitbucketCloudClient) WithAuthenticbtor(v0 buth.Authenticbtor) bitbucketcloud.Client {
	r0 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(BitbucketCloudClientWithAuthenticbtorFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockBitbucketCloudClient instbnce is invoked bnd the
// hook queue is empty.
func (f *BitbucketCloudClientWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) bitbucketcloud.Client) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockBitbucketCloudClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *BitbucketCloudClientWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) bitbucketcloud.Client) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BitbucketCloudClientWithAuthenticbtorFunc) SetDefbultReturn(r0 bitbucketcloud.Client) {
	f.SetDefbultHook(func(buth.Authenticbtor) bitbucketcloud.Client {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BitbucketCloudClientWithAuthenticbtorFunc) PushReturn(r0 bitbucketcloud.Client) {
	f.PushHook(func(buth.Authenticbtor) bitbucketcloud.Client {
		return r0
	})
}

func (f *BitbucketCloudClientWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) bitbucketcloud.Client {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BitbucketCloudClientWithAuthenticbtorFunc) bppendCbll(r0 BitbucketCloudClientWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// BitbucketCloudClientWithAuthenticbtorFuncCbll objects describing the
// invocbtions of this function.
func (f *BitbucketCloudClientWithAuthenticbtorFunc) History() []BitbucketCloudClientWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]BitbucketCloudClientWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BitbucketCloudClientWithAuthenticbtorFuncCbll is bn object thbt describes
// bn invocbtion of method WithAuthenticbtor on bn instbnce of
// MockBitbucketCloudClient.
type BitbucketCloudClientWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bitbucketcloud.Client
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BitbucketCloudClientWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BitbucketCloudClientWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockAzureDevOpsClient is b mock implementbtion of the Client interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops) used for
// unit testing.
type MockAzureDevOpsClient struct {
	// AbbndonPullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AbbndonPullRequest.
	AbbndonPullRequestFunc *AzureDevOpsClientAbbndonPullRequestFunc
	// AuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method Authenticbtor.
	AuthenticbtorFunc *AzureDevOpsClientAuthenticbtorFunc
	// CompletePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CompletePullRequest.
	CompletePullRequestFunc *AzureDevOpsClientCompletePullRequestFunc
	// CrebtePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebtePullRequest.
	CrebtePullRequestFunc *AzureDevOpsClientCrebtePullRequestFunc
	// CrebtePullRequestCommentThrebdFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// CrebtePullRequestCommentThrebd.
	CrebtePullRequestCommentThrebdFunc *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc
	// ForkRepositoryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ForkRepository.
	ForkRepositoryFunc *AzureDevOpsClientForkRepositoryFunc
	// GetAuthorizedProfileFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetAuthorizedProfile.
	GetAuthorizedProfileFunc *AzureDevOpsClientGetAuthorizedProfileFunc
	// GetProjectFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetProject.
	GetProjectFunc *AzureDevOpsClientGetProjectFunc
	// GetPullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPullRequest.
	GetPullRequestFunc *AzureDevOpsClientGetPullRequestFunc
	// GetPullRequestStbtusesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPullRequestStbtuses.
	GetPullRequestStbtusesFunc *AzureDevOpsClientGetPullRequestStbtusesFunc
	// GetRepoFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetRepo.
	GetRepoFunc *AzureDevOpsClientGetRepoFunc
	// GetRepositoryBrbnchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRepositoryBrbnch.
	GetRepositoryBrbnchFunc *AzureDevOpsClientGetRepositoryBrbnchFunc
	// GetURLFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetURL.
	GetURLFunc *AzureDevOpsClientGetURLFunc
	// IsAzureDevOpsServicesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method IsAzureDevOpsServices.
	IsAzureDevOpsServicesFunc *AzureDevOpsClientIsAzureDevOpsServicesFunc
	// ListAuthorizedUserOrgbnizbtionsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListAuthorizedUserOrgbnizbtions.
	ListAuthorizedUserOrgbnizbtionsFunc *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc
	// ListRepositoriesByProjectOrOrgFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// ListRepositoriesByProjectOrOrg.
	ListRepositoriesByProjectOrOrgFunc *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc
	// SetWbitForRbteLimitFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetWbitForRbteLimit.
	SetWbitForRbteLimitFunc *AzureDevOpsClientSetWbitForRbteLimitFunc
	// UpdbtePullRequestFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbtePullRequest.
	UpdbtePullRequestFunc *AzureDevOpsClientUpdbtePullRequestFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *AzureDevOpsClientWithAuthenticbtorFunc
}

// NewMockAzureDevOpsClient crebtes b new mock of the Client interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockAzureDevOpsClient() *MockAzureDevOpsClient {
	return &MockAzureDevOpsClient{
		AbbndonPullRequestFunc: &AzureDevOpsClientAbbndonPullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs) (r0 bzuredevops.PullRequest, r1 error) {
				return
			},
		},
		AuthenticbtorFunc: &AzureDevOpsClientAuthenticbtorFunc{
			defbultHook: func() (r0 buth.Authenticbtor) {
				return
			},
		},
		CompletePullRequestFunc: &AzureDevOpsClientCompletePullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (r0 bzuredevops.PullRequest, r1 error) {
				return
			},
		},
		CrebtePullRequestFunc: &AzureDevOpsClientCrebtePullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (r0 bzuredevops.PullRequest, r1 error) {
				return
			},
		},
		CrebtePullRequestCommentThrebdFunc: &AzureDevOpsClientCrebtePullRequestCommentThrebdFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (r0 bzuredevops.PullRequestCommentResponse, r1 error) {
				return
			},
		},
		ForkRepositoryFunc: &AzureDevOpsClientForkRepositoryFunc{
			defbultHook: func(context.Context, string, bzuredevops.ForkRepositoryInput) (r0 bzuredevops.Repository, r1 error) {
				return
			},
		},
		GetAuthorizedProfileFunc: &AzureDevOpsClientGetAuthorizedProfileFunc{
			defbultHook: func(context.Context) (r0 bzuredevops.Profile, r1 error) {
				return
			},
		},
		GetProjectFunc: &AzureDevOpsClientGetProjectFunc{
			defbultHook: func(context.Context, string, string) (r0 bzuredevops.Project, r1 error) {
				return
			},
		},
		GetPullRequestFunc: &AzureDevOpsClientGetPullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs) (r0 bzuredevops.PullRequest, r1 error) {
				return
			},
		},
		GetPullRequestStbtusesFunc: &AzureDevOpsClientGetPullRequestStbtusesFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs) (r0 []bzuredevops.PullRequestBuildStbtus, r1 error) {
				return
			},
		},
		GetRepoFunc: &AzureDevOpsClientGetRepoFunc{
			defbultHook: func(context.Context, bzuredevops.OrgProjectRepoArgs) (r0 bzuredevops.Repository, r1 error) {
				return
			},
		},
		GetRepositoryBrbnchFunc: &AzureDevOpsClientGetRepositoryBrbnchFunc{
			defbultHook: func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (r0 bzuredevops.Ref, r1 error) {
				return
			},
		},
		GetURLFunc: &AzureDevOpsClientGetURLFunc{
			defbultHook: func() (r0 *url.URL) {
				return
			},
		},
		IsAzureDevOpsServicesFunc: &AzureDevOpsClientIsAzureDevOpsServicesFunc{
			defbultHook: func() (r0 bool) {
				return
			},
		},
		ListAuthorizedUserOrgbnizbtionsFunc: &AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc{
			defbultHook: func(context.Context, bzuredevops.Profile) (r0 []bzuredevops.Org, r1 error) {
				return
			},
		},
		ListRepositoriesByProjectOrOrgFunc: &AzureDevOpsClientListRepositoriesByProjectOrOrgFunc{
			defbultHook: func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) (r0 []bzuredevops.Repository, r1 error) {
				return
			},
		},
		SetWbitForRbteLimitFunc: &AzureDevOpsClientSetWbitForRbteLimitFunc{
			defbultHook: func(bool) {
				return
			},
		},
		UpdbtePullRequestFunc: &AzureDevOpsClientUpdbtePullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (r0 bzuredevops.PullRequest, r1 error) {
				return
			},
		},
		WithAuthenticbtorFunc: &AzureDevOpsClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 bzuredevops.Client, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockAzureDevOpsClient crebtes b new mock of the Client
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockAzureDevOpsClient() *MockAzureDevOpsClient {
	return &MockAzureDevOpsClient{
		AbbndonPullRequestFunc: &AzureDevOpsClientAbbndonPullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.AbbndonPullRequest")
			},
		},
		AuthenticbtorFunc: &AzureDevOpsClientAuthenticbtorFunc{
			defbultHook: func() buth.Authenticbtor {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.Authenticbtor")
			},
		},
		CompletePullRequestFunc: &AzureDevOpsClientCompletePullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.CompletePullRequest")
			},
		},
		CrebtePullRequestFunc: &AzureDevOpsClientCrebtePullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.CrebtePullRequest")
			},
		},
		CrebtePullRequestCommentThrebdFunc: &AzureDevOpsClientCrebtePullRequestCommentThrebdFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.CrebtePullRequestCommentThrebd")
			},
		},
		ForkRepositoryFunc: &AzureDevOpsClientForkRepositoryFunc{
			defbultHook: func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.ForkRepository")
			},
		},
		GetAuthorizedProfileFunc: &AzureDevOpsClientGetAuthorizedProfileFunc{
			defbultHook: func(context.Context) (bzuredevops.Profile, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetAuthorizedProfile")
			},
		},
		GetProjectFunc: &AzureDevOpsClientGetProjectFunc{
			defbultHook: func(context.Context, string, string) (bzuredevops.Project, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetProject")
			},
		},
		GetPullRequestFunc: &AzureDevOpsClientGetPullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetPullRequest")
			},
		},
		GetPullRequestStbtusesFunc: &AzureDevOpsClientGetPullRequestStbtusesFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetPullRequestStbtuses")
			},
		},
		GetRepoFunc: &AzureDevOpsClientGetRepoFunc{
			defbultHook: func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetRepo")
			},
		},
		GetRepositoryBrbnchFunc: &AzureDevOpsClientGetRepositoryBrbnchFunc{
			defbultHook: func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetRepositoryBrbnch")
			},
		},
		GetURLFunc: &AzureDevOpsClientGetURLFunc{
			defbultHook: func() *url.URL {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.GetURL")
			},
		},
		IsAzureDevOpsServicesFunc: &AzureDevOpsClientIsAzureDevOpsServicesFunc{
			defbultHook: func() bool {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.IsAzureDevOpsServices")
			},
		},
		ListAuthorizedUserOrgbnizbtionsFunc: &AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc{
			defbultHook: func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.ListAuthorizedUserOrgbnizbtions")
			},
		},
		ListRepositoriesByProjectOrOrgFunc: &AzureDevOpsClientListRepositoriesByProjectOrOrgFunc{
			defbultHook: func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.ListRepositoriesByProjectOrOrg")
			},
		},
		SetWbitForRbteLimitFunc: &AzureDevOpsClientSetWbitForRbteLimitFunc{
			defbultHook: func(bool) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.SetWbitForRbteLimit")
			},
		},
		UpdbtePullRequestFunc: &AzureDevOpsClientUpdbtePullRequestFunc{
			defbultHook: func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.UpdbtePullRequest")
			},
		},
		WithAuthenticbtorFunc: &AzureDevOpsClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (bzuredevops.Client, error) {
				pbnic("unexpected invocbtion of MockAzureDevOpsClient.WithAuthenticbtor")
			},
		},
	}
}

// NewMockAzureDevOpsClientFrom crebtes b new mock of the
// MockAzureDevOpsClient interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockAzureDevOpsClientFrom(i bzuredevops.Client) *MockAzureDevOpsClient {
	return &MockAzureDevOpsClient{
		AbbndonPullRequestFunc: &AzureDevOpsClientAbbndonPullRequestFunc{
			defbultHook: i.AbbndonPullRequest,
		},
		AuthenticbtorFunc: &AzureDevOpsClientAuthenticbtorFunc{
			defbultHook: i.Authenticbtor,
		},
		CompletePullRequestFunc: &AzureDevOpsClientCompletePullRequestFunc{
			defbultHook: i.CompletePullRequest,
		},
		CrebtePullRequestFunc: &AzureDevOpsClientCrebtePullRequestFunc{
			defbultHook: i.CrebtePullRequest,
		},
		CrebtePullRequestCommentThrebdFunc: &AzureDevOpsClientCrebtePullRequestCommentThrebdFunc{
			defbultHook: i.CrebtePullRequestCommentThrebd,
		},
		ForkRepositoryFunc: &AzureDevOpsClientForkRepositoryFunc{
			defbultHook: i.ForkRepository,
		},
		GetAuthorizedProfileFunc: &AzureDevOpsClientGetAuthorizedProfileFunc{
			defbultHook: i.GetAuthorizedProfile,
		},
		GetProjectFunc: &AzureDevOpsClientGetProjectFunc{
			defbultHook: i.GetProject,
		},
		GetPullRequestFunc: &AzureDevOpsClientGetPullRequestFunc{
			defbultHook: i.GetPullRequest,
		},
		GetPullRequestStbtusesFunc: &AzureDevOpsClientGetPullRequestStbtusesFunc{
			defbultHook: i.GetPullRequestStbtuses,
		},
		GetRepoFunc: &AzureDevOpsClientGetRepoFunc{
			defbultHook: i.GetRepo,
		},
		GetRepositoryBrbnchFunc: &AzureDevOpsClientGetRepositoryBrbnchFunc{
			defbultHook: i.GetRepositoryBrbnch,
		},
		GetURLFunc: &AzureDevOpsClientGetURLFunc{
			defbultHook: i.GetURL,
		},
		IsAzureDevOpsServicesFunc: &AzureDevOpsClientIsAzureDevOpsServicesFunc{
			defbultHook: i.IsAzureDevOpsServices,
		},
		ListAuthorizedUserOrgbnizbtionsFunc: &AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc{
			defbultHook: i.ListAuthorizedUserOrgbnizbtions,
		},
		ListRepositoriesByProjectOrOrgFunc: &AzureDevOpsClientListRepositoriesByProjectOrOrgFunc{
			defbultHook: i.ListRepositoriesByProjectOrOrg,
		},
		SetWbitForRbteLimitFunc: &AzureDevOpsClientSetWbitForRbteLimitFunc{
			defbultHook: i.SetWbitForRbteLimit,
		},
		UpdbtePullRequestFunc: &AzureDevOpsClientUpdbtePullRequestFunc{
			defbultHook: i.UpdbtePullRequest,
		},
		WithAuthenticbtorFunc: &AzureDevOpsClientWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
	}
}

// AzureDevOpsClientAbbndonPullRequestFunc describes the behbvior when the
// AbbndonPullRequest method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientAbbndonPullRequestFunc struct {
	defbultHook func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)
	hooks       []func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)
	history     []AzureDevOpsClientAbbndonPullRequestFuncCbll
	mutex       sync.Mutex
}

// AbbndonPullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) AbbndonPullRequest(v0 context.Context, v1 bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
	r0, r1 := m.AbbndonPullRequestFunc.nextHook()(v0, v1)
	m.AbbndonPullRequestFunc.bppendCbll(AzureDevOpsClientAbbndonPullRequestFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AbbndonPullRequest
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientAbbndonPullRequestFunc) SetDefbultHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AbbndonPullRequest method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientAbbndonPullRequestFunc) PushHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientAbbndonPullRequestFunc) SetDefbultReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientAbbndonPullRequestFunc) PushReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientAbbndonPullRequestFunc) nextHook() func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientAbbndonPullRequestFunc) bppendCbll(r0 AzureDevOpsClientAbbndonPullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientAbbndonPullRequestFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientAbbndonPullRequestFunc) History() []AzureDevOpsClientAbbndonPullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientAbbndonPullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientAbbndonPullRequestFuncCbll is bn object thbt describes
// bn invocbtion of method AbbndonPullRequest on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientAbbndonPullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.PullRequestCommonArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientAbbndonPullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientAbbndonPullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientAuthenticbtorFunc describes the behbvior when the
// Authenticbtor method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientAuthenticbtorFunc struct {
	defbultHook func() buth.Authenticbtor
	hooks       []func() buth.Authenticbtor
	history     []AzureDevOpsClientAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// Authenticbtor delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) Authenticbtor() buth.Authenticbtor {
	r0 := m.AuthenticbtorFunc.nextHook()()
	m.AuthenticbtorFunc.bppendCbll(AzureDevOpsClientAuthenticbtorFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Authenticbtor method
// of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the hook
// queue is empty.
func (f *AzureDevOpsClientAuthenticbtorFunc) SetDefbultHook(hook func() buth.Authenticbtor) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Authenticbtor method of the pbrent MockAzureDevOpsClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *AzureDevOpsClientAuthenticbtorFunc) PushHook(hook func() buth.Authenticbtor) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientAuthenticbtorFunc) SetDefbultReturn(r0 buth.Authenticbtor) {
	f.SetDefbultHook(func() buth.Authenticbtor {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientAuthenticbtorFunc) PushReturn(r0 buth.Authenticbtor) {
	f.PushHook(func() buth.Authenticbtor {
		return r0
	})
}

func (f *AzureDevOpsClientAuthenticbtorFunc) nextHook() func() buth.Authenticbtor {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientAuthenticbtorFunc) bppendCbll(r0 AzureDevOpsClientAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientAuthenticbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientAuthenticbtorFunc) History() []AzureDevOpsClientAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method Authenticbtor on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientAuthenticbtorFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 buth.Authenticbtor
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// AzureDevOpsClientCompletePullRequestFunc describes the behbvior when the
// CompletePullRequest method of the pbrent MockAzureDevOpsClient instbnce
// is invoked.
type AzureDevOpsClientCompletePullRequestFunc struct {
	defbultHook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error)
	hooks       []func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error)
	history     []AzureDevOpsClientCompletePullRequestFuncCbll
	mutex       sync.Mutex
}

// CompletePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) CompletePullRequest(v0 context.Context, v1 bzuredevops.PullRequestCommonArgs, v2 bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
	r0, r1 := m.CompletePullRequestFunc.nextHook()(v0, v1, v2)
	m.CompletePullRequestFunc.bppendCbll(AzureDevOpsClientCompletePullRequestFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CompletePullRequest
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientCompletePullRequestFunc) SetDefbultHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CompletePullRequest method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientCompletePullRequestFunc) PushHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientCompletePullRequestFunc) SetDefbultReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientCompletePullRequestFunc) PushReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientCompletePullRequestFunc) nextHook() func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientCompletePullRequestFunc) bppendCbll(r0 AzureDevOpsClientCompletePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientCompletePullRequestFuncCbll objects describing the
// invocbtions of this function.
func (f *AzureDevOpsClientCompletePullRequestFunc) History() []AzureDevOpsClientCompletePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientCompletePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientCompletePullRequestFuncCbll is bn object thbt describes
// bn invocbtion of method CompletePullRequest on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientCompletePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.PullRequestCommonArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bzuredevops.PullRequestCompleteInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientCompletePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientCompletePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientCrebtePullRequestFunc describes the behbvior when the
// CrebtePullRequest method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientCrebtePullRequestFunc struct {
	defbultHook func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error)
	hooks       []func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error)
	history     []AzureDevOpsClientCrebtePullRequestFuncCbll
	mutex       sync.Mutex
}

// CrebtePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) CrebtePullRequest(v0 context.Context, v1 bzuredevops.OrgProjectRepoArgs, v2 bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
	r0, r1 := m.CrebtePullRequestFunc.nextHook()(v0, v1, v2)
	m.CrebtePullRequestFunc.bppendCbll(AzureDevOpsClientCrebtePullRequestFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CrebtePullRequest
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientCrebtePullRequestFunc) SetDefbultHook(hook func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebtePullRequest method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientCrebtePullRequestFunc) PushHook(hook func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientCrebtePullRequestFunc) SetDefbultReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientCrebtePullRequestFunc) PushReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientCrebtePullRequestFunc) nextHook() func(context.Context, bzuredevops.OrgProjectRepoArgs, bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientCrebtePullRequestFunc) bppendCbll(r0 AzureDevOpsClientCrebtePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientCrebtePullRequestFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientCrebtePullRequestFunc) History() []AzureDevOpsClientCrebtePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientCrebtePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientCrebtePullRequestFuncCbll is bn object thbt describes bn
// invocbtion of method CrebtePullRequest on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientCrebtePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.OrgProjectRepoArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bzuredevops.CrebtePullRequestInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientCrebtePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientCrebtePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientCrebtePullRequestCommentThrebdFunc describes the
// behbvior when the CrebtePullRequestCommentThrebd method of the pbrent
// MockAzureDevOpsClient instbnce is invoked.
type AzureDevOpsClientCrebtePullRequestCommentThrebdFunc struct {
	defbultHook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error)
	hooks       []func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error)
	history     []AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll
	mutex       sync.Mutex
}

// CrebtePullRequestCommentThrebd delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) CrebtePullRequestCommentThrebd(v0 context.Context, v1 bzuredevops.PullRequestCommonArgs, v2 bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
	r0, r1 := m.CrebtePullRequestCommentThrebdFunc.nextHook()(v0, v1, v2)
	m.CrebtePullRequestCommentThrebdFunc.bppendCbll(AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebtePullRequestCommentThrebd method of the pbrent MockAzureDevOpsClient
// instbnce is invoked bnd the hook queue is empty.
func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) SetDefbultHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebtePullRequestCommentThrebd method of the pbrent MockAzureDevOpsClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) PushHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) SetDefbultReturn(r0 bzuredevops.PullRequestCommentResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) PushReturn(r0 bzuredevops.PullRequestCommentResponse, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) nextHook() func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) bppendCbll(r0 AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll objects
// describing the invocbtions of this function.
func (f *AzureDevOpsClientCrebtePullRequestCommentThrebdFunc) History() []AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll is bn object thbt
// describes bn invocbtion of method CrebtePullRequestCommentThrebd on bn
// instbnce of MockAzureDevOpsClient.
type AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.PullRequestCommonArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bzuredevops.PullRequestCommentInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.PullRequestCommentResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientCrebtePullRequestCommentThrebdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientForkRepositoryFunc describes the behbvior when the
// ForkRepository method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientForkRepositoryFunc struct {
	defbultHook func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error)
	hooks       []func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error)
	history     []AzureDevOpsClientForkRepositoryFuncCbll
	mutex       sync.Mutex
}

// ForkRepository delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) ForkRepository(v0 context.Context, v1 string, v2 bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
	r0, r1 := m.ForkRepositoryFunc.nextHook()(v0, v1, v2)
	m.ForkRepositoryFunc.bppendCbll(AzureDevOpsClientForkRepositoryFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ForkRepository
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientForkRepositoryFunc) SetDefbultHook(hook func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ForkRepository method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientForkRepositoryFunc) PushHook(hook func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientForkRepositoryFunc) SetDefbultReturn(r0 bzuredevops.Repository, r1 error) {
	f.SetDefbultHook(func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientForkRepositoryFunc) PushReturn(r0 bzuredevops.Repository, r1 error) {
	f.PushHook(func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientForkRepositoryFunc) nextHook() func(context.Context, string, bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientForkRepositoryFunc) bppendCbll(r0 AzureDevOpsClientForkRepositoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientForkRepositoryFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientForkRepositoryFunc) History() []AzureDevOpsClientForkRepositoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientForkRepositoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientForkRepositoryFuncCbll is bn object thbt describes bn
// invocbtion of method ForkRepository on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientForkRepositoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bzuredevops.ForkRepositoryInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientForkRepositoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientForkRepositoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetAuthorizedProfileFunc describes the behbvior when the
// GetAuthorizedProfile method of the pbrent MockAzureDevOpsClient instbnce
// is invoked.
type AzureDevOpsClientGetAuthorizedProfileFunc struct {
	defbultHook func(context.Context) (bzuredevops.Profile, error)
	hooks       []func(context.Context) (bzuredevops.Profile, error)
	history     []AzureDevOpsClientGetAuthorizedProfileFuncCbll
	mutex       sync.Mutex
}

// GetAuthorizedProfile delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetAuthorizedProfile(v0 context.Context) (bzuredevops.Profile, error) {
	r0, r1 := m.GetAuthorizedProfileFunc.nextHook()(v0)
	m.GetAuthorizedProfileFunc.bppendCbll(AzureDevOpsClientGetAuthorizedProfileFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetAuthorizedProfile
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientGetAuthorizedProfileFunc) SetDefbultHook(hook func(context.Context) (bzuredevops.Profile, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuthorizedProfile method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientGetAuthorizedProfileFunc) PushHook(hook func(context.Context) (bzuredevops.Profile, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetAuthorizedProfileFunc) SetDefbultReturn(r0 bzuredevops.Profile, r1 error) {
	f.SetDefbultHook(func(context.Context) (bzuredevops.Profile, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetAuthorizedProfileFunc) PushReturn(r0 bzuredevops.Profile, r1 error) {
	f.PushHook(func(context.Context) (bzuredevops.Profile, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientGetAuthorizedProfileFunc) nextHook() func(context.Context) (bzuredevops.Profile, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetAuthorizedProfileFunc) bppendCbll(r0 AzureDevOpsClientGetAuthorizedProfileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientGetAuthorizedProfileFuncCbll objects describing the
// invocbtions of this function.
func (f *AzureDevOpsClientGetAuthorizedProfileFunc) History() []AzureDevOpsClientGetAuthorizedProfileFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetAuthorizedProfileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetAuthorizedProfileFuncCbll is bn object thbt describes
// bn invocbtion of method GetAuthorizedProfile on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientGetAuthorizedProfileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.Profile
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetAuthorizedProfileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetAuthorizedProfileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetProjectFunc describes the behbvior when the
// GetProject method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientGetProjectFunc struct {
	defbultHook func(context.Context, string, string) (bzuredevops.Project, error)
	hooks       []func(context.Context, string, string) (bzuredevops.Project, error)
	history     []AzureDevOpsClientGetProjectFuncCbll
	mutex       sync.Mutex
}

// GetProject delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetProject(v0 context.Context, v1 string, v2 string) (bzuredevops.Project, error) {
	r0, r1 := m.GetProjectFunc.nextHook()(v0, v1, v2)
	m.GetProjectFunc.bppendCbll(AzureDevOpsClientGetProjectFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetProject method of
// the pbrent MockAzureDevOpsClient instbnce is invoked bnd the hook queue
// is empty.
func (f *AzureDevOpsClientGetProjectFunc) SetDefbultHook(hook func(context.Context, string, string) (bzuredevops.Project, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetProject method of the pbrent MockAzureDevOpsClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *AzureDevOpsClientGetProjectFunc) PushHook(hook func(context.Context, string, string) (bzuredevops.Project, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetProjectFunc) SetDefbultReturn(r0 bzuredevops.Project, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string) (bzuredevops.Project, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetProjectFunc) PushReturn(r0 bzuredevops.Project, r1 error) {
	f.PushHook(func(context.Context, string, string) (bzuredevops.Project, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientGetProjectFunc) nextHook() func(context.Context, string, string) (bzuredevops.Project, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetProjectFunc) bppendCbll(r0 AzureDevOpsClientGetProjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientGetProjectFuncCbll objects
// describing the invocbtions of this function.
func (f *AzureDevOpsClientGetProjectFunc) History() []AzureDevOpsClientGetProjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetProjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetProjectFuncCbll is bn object thbt describes bn
// invocbtion of method GetProject on bn instbnce of MockAzureDevOpsClient.
type AzureDevOpsClientGetProjectFuncCbll struct {
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
	Result0 bzuredevops.Project
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetProjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetProjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetPullRequestFunc describes the behbvior when the
// GetPullRequest method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientGetPullRequestFunc struct {
	defbultHook func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)
	hooks       []func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)
	history     []AzureDevOpsClientGetPullRequestFuncCbll
	mutex       sync.Mutex
}

// GetPullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetPullRequest(v0 context.Context, v1 bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
	r0, r1 := m.GetPullRequestFunc.nextHook()(v0, v1)
	m.GetPullRequestFunc.bppendCbll(AzureDevOpsClientGetPullRequestFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetPullRequest
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientGetPullRequestFunc) SetDefbultHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPullRequest method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientGetPullRequestFunc) PushHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetPullRequestFunc) SetDefbultReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetPullRequestFunc) PushReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientGetPullRequestFunc) nextHook() func(context.Context, bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetPullRequestFunc) bppendCbll(r0 AzureDevOpsClientGetPullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientGetPullRequestFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientGetPullRequestFunc) History() []AzureDevOpsClientGetPullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetPullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetPullRequestFuncCbll is bn object thbt describes bn
// invocbtion of method GetPullRequest on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientGetPullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.PullRequestCommonArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetPullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetPullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetPullRequestStbtusesFunc describes the behbvior when
// the GetPullRequestStbtuses method of the pbrent MockAzureDevOpsClient
// instbnce is invoked.
type AzureDevOpsClientGetPullRequestStbtusesFunc struct {
	defbultHook func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error)
	hooks       []func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error)
	history     []AzureDevOpsClientGetPullRequestStbtusesFuncCbll
	mutex       sync.Mutex
}

// GetPullRequestStbtuses delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetPullRequestStbtuses(v0 context.Context, v1 bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
	r0, r1 := m.GetPullRequestStbtusesFunc.nextHook()(v0, v1)
	m.GetPullRequestStbtusesFunc.bppendCbll(AzureDevOpsClientGetPullRequestStbtusesFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetPullRequestStbtuses method of the pbrent MockAzureDevOpsClient
// instbnce is invoked bnd the hook queue is empty.
func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) SetDefbultHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPullRequestStbtuses method of the pbrent MockAzureDevOpsClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) PushHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) SetDefbultReturn(r0 []bzuredevops.PullRequestBuildStbtus, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) PushReturn(r0 []bzuredevops.PullRequestBuildStbtus, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) nextHook() func(context.Context, bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) bppendCbll(r0 AzureDevOpsClientGetPullRequestStbtusesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientGetPullRequestStbtusesFuncCbll objects describing the
// invocbtions of this function.
func (f *AzureDevOpsClientGetPullRequestStbtusesFunc) History() []AzureDevOpsClientGetPullRequestStbtusesFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetPullRequestStbtusesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetPullRequestStbtusesFuncCbll is bn object thbt
// describes bn invocbtion of method GetPullRequestStbtuses on bn instbnce
// of MockAzureDevOpsClient.
type AzureDevOpsClientGetPullRequestStbtusesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.PullRequestCommonArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bzuredevops.PullRequestBuildStbtus
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetPullRequestStbtusesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetPullRequestStbtusesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetRepoFunc describes the behbvior when the GetRepo
// method of the pbrent MockAzureDevOpsClient instbnce is invoked.
type AzureDevOpsClientGetRepoFunc struct {
	defbultHook func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error)
	hooks       []func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error)
	history     []AzureDevOpsClientGetRepoFuncCbll
	mutex       sync.Mutex
}

// GetRepo delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetRepo(v0 context.Context, v1 bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
	r0, r1 := m.GetRepoFunc.nextHook()(v0, v1)
	m.GetRepoFunc.bppendCbll(AzureDevOpsClientGetRepoFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetRepo method of
// the pbrent MockAzureDevOpsClient instbnce is invoked bnd the hook queue
// is empty.
func (f *AzureDevOpsClientGetRepoFunc) SetDefbultHook(hook func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepo method of the pbrent MockAzureDevOpsClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *AzureDevOpsClientGetRepoFunc) PushHook(hook func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetRepoFunc) SetDefbultReturn(r0 bzuredevops.Repository, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetRepoFunc) PushReturn(r0 bzuredevops.Repository, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientGetRepoFunc) nextHook() func(context.Context, bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetRepoFunc) bppendCbll(r0 AzureDevOpsClientGetRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientGetRepoFuncCbll objects
// describing the invocbtions of this function.
func (f *AzureDevOpsClientGetRepoFunc) History() []AzureDevOpsClientGetRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetRepoFuncCbll is bn object thbt describes bn
// invocbtion of method GetRepo on bn instbnce of MockAzureDevOpsClient.
type AzureDevOpsClientGetRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.OrgProjectRepoArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetRepositoryBrbnchFunc describes the behbvior when the
// GetRepositoryBrbnch method of the pbrent MockAzureDevOpsClient instbnce
// is invoked.
type AzureDevOpsClientGetRepositoryBrbnchFunc struct {
	defbultHook func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error)
	hooks       []func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error)
	history     []AzureDevOpsClientGetRepositoryBrbnchFuncCbll
	mutex       sync.Mutex
}

// GetRepositoryBrbnch delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetRepositoryBrbnch(v0 context.Context, v1 bzuredevops.OrgProjectRepoArgs, v2 string) (bzuredevops.Ref, error) {
	r0, r1 := m.GetRepositoryBrbnchFunc.nextHook()(v0, v1, v2)
	m.GetRepositoryBrbnchFunc.bppendCbll(AzureDevOpsClientGetRepositoryBrbnchFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetRepositoryBrbnch
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) SetDefbultHook(hook func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepositoryBrbnch method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) PushHook(hook func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) SetDefbultReturn(r0 bzuredevops.Ref, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) PushReturn(r0 bzuredevops.Ref, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) nextHook() func(context.Context, bzuredevops.OrgProjectRepoArgs, string) (bzuredevops.Ref, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) bppendCbll(r0 AzureDevOpsClientGetRepositoryBrbnchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientGetRepositoryBrbnchFuncCbll objects describing the
// invocbtions of this function.
func (f *AzureDevOpsClientGetRepositoryBrbnchFunc) History() []AzureDevOpsClientGetRepositoryBrbnchFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetRepositoryBrbnchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetRepositoryBrbnchFuncCbll is bn object thbt describes
// bn invocbtion of method GetRepositoryBrbnch on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientGetRepositoryBrbnchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.OrgProjectRepoArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.Ref
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetRepositoryBrbnchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetRepositoryBrbnchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientGetURLFunc describes the behbvior when the GetURL method
// of the pbrent MockAzureDevOpsClient instbnce is invoked.
type AzureDevOpsClientGetURLFunc struct {
	defbultHook func() *url.URL
	hooks       []func() *url.URL
	history     []AzureDevOpsClientGetURLFuncCbll
	mutex       sync.Mutex
}

// GetURL delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) GetURL() *url.URL {
	r0 := m.GetURLFunc.nextHook()()
	m.GetURLFunc.bppendCbll(AzureDevOpsClientGetURLFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GetURL method of the
// pbrent MockAzureDevOpsClient instbnce is invoked bnd the hook queue is
// empty.
func (f *AzureDevOpsClientGetURLFunc) SetDefbultHook(hook func() *url.URL) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetURL method of the pbrent MockAzureDevOpsClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *AzureDevOpsClientGetURLFunc) PushHook(hook func() *url.URL) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientGetURLFunc) SetDefbultReturn(r0 *url.URL) {
	f.SetDefbultHook(func() *url.URL {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientGetURLFunc) PushReturn(r0 *url.URL) {
	f.PushHook(func() *url.URL {
		return r0
	})
}

func (f *AzureDevOpsClientGetURLFunc) nextHook() func() *url.URL {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientGetURLFunc) bppendCbll(r0 AzureDevOpsClientGetURLFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientGetURLFuncCbll objects
// describing the invocbtions of this function.
func (f *AzureDevOpsClientGetURLFunc) History() []AzureDevOpsClientGetURLFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientGetURLFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientGetURLFuncCbll is bn object thbt describes bn invocbtion
// of method GetURL on bn instbnce of MockAzureDevOpsClient.
type AzureDevOpsClientGetURLFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *url.URL
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientGetURLFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientGetURLFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// AzureDevOpsClientIsAzureDevOpsServicesFunc describes the behbvior when
// the IsAzureDevOpsServices method of the pbrent MockAzureDevOpsClient
// instbnce is invoked.
type AzureDevOpsClientIsAzureDevOpsServicesFunc struct {
	defbultHook func() bool
	hooks       []func() bool
	history     []AzureDevOpsClientIsAzureDevOpsServicesFuncCbll
	mutex       sync.Mutex
}

// IsAzureDevOpsServices delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) IsAzureDevOpsServices() bool {
	r0 := m.IsAzureDevOpsServicesFunc.nextHook()()
	m.IsAzureDevOpsServicesFunc.bppendCbll(AzureDevOpsClientIsAzureDevOpsServicesFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// IsAzureDevOpsServices method of the pbrent MockAzureDevOpsClient instbnce
// is invoked bnd the hook queue is empty.
func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) SetDefbultHook(hook func() bool) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IsAzureDevOpsServices method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) PushHook(hook func() bool) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) SetDefbultReturn(r0 bool) {
	f.SetDefbultHook(func() bool {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) PushReturn(r0 bool) {
	f.PushHook(func() bool {
		return r0
	})
}

func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) nextHook() func() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) bppendCbll(r0 AzureDevOpsClientIsAzureDevOpsServicesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientIsAzureDevOpsServicesFuncCbll objects describing the
// invocbtions of this function.
func (f *AzureDevOpsClientIsAzureDevOpsServicesFunc) History() []AzureDevOpsClientIsAzureDevOpsServicesFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientIsAzureDevOpsServicesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientIsAzureDevOpsServicesFuncCbll is bn object thbt
// describes bn invocbtion of method IsAzureDevOpsServices on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientIsAzureDevOpsServicesFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientIsAzureDevOpsServicesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientIsAzureDevOpsServicesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc describes the
// behbvior when the ListAuthorizedUserOrgbnizbtions method of the pbrent
// MockAzureDevOpsClient instbnce is invoked.
type AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc struct {
	defbultHook func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error)
	hooks       []func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error)
	history     []AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll
	mutex       sync.Mutex
}

// ListAuthorizedUserOrgbnizbtions delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) ListAuthorizedUserOrgbnizbtions(v0 context.Context, v1 bzuredevops.Profile) ([]bzuredevops.Org, error) {
	r0, r1 := m.ListAuthorizedUserOrgbnizbtionsFunc.nextHook()(v0, v1)
	m.ListAuthorizedUserOrgbnizbtionsFunc.bppendCbll(AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ListAuthorizedUserOrgbnizbtions method of the pbrent
// MockAzureDevOpsClient instbnce is invoked bnd the hook queue is empty.
func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) SetDefbultHook(hook func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListAuthorizedUserOrgbnizbtions method of the pbrent
// MockAzureDevOpsClient instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) PushHook(hook func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) SetDefbultReturn(r0 []bzuredevops.Org, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) PushReturn(r0 []bzuredevops.Org, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) nextHook() func(context.Context, bzuredevops.Profile) ([]bzuredevops.Org, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) bppendCbll(r0 AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll objects
// describing the invocbtions of this function.
func (f *AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFunc) History() []AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll is bn object
// thbt describes bn invocbtion of method ListAuthorizedUserOrgbnizbtions on
// bn instbnce of MockAzureDevOpsClient.
type AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.Profile
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bzuredevops.Org
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientListAuthorizedUserOrgbnizbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientListRepositoriesByProjectOrOrgFunc describes the
// behbvior when the ListRepositoriesByProjectOrOrg method of the pbrent
// MockAzureDevOpsClient instbnce is invoked.
type AzureDevOpsClientListRepositoriesByProjectOrOrgFunc struct {
	defbultHook func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error)
	hooks       []func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error)
	history     []AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll
	mutex       sync.Mutex
}

// ListRepositoriesByProjectOrOrg delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) ListRepositoriesByProjectOrOrg(v0 context.Context, v1 bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error) {
	r0, r1 := m.ListRepositoriesByProjectOrOrgFunc.nextHook()(v0, v1)
	m.ListRepositoriesByProjectOrOrgFunc.bppendCbll(AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ListRepositoriesByProjectOrOrg method of the pbrent MockAzureDevOpsClient
// instbnce is invoked bnd the hook queue is empty.
func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) SetDefbultHook(hook func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListRepositoriesByProjectOrOrg method of the pbrent MockAzureDevOpsClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) PushHook(hook func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) SetDefbultReturn(r0 []bzuredevops.Repository, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) PushReturn(r0 []bzuredevops.Repository, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) nextHook() func(context.Context, bzuredevops.ListRepositoriesByProjectOrOrgArgs) ([]bzuredevops.Repository, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) bppendCbll(r0 AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll objects
// describing the invocbtions of this function.
func (f *AzureDevOpsClientListRepositoriesByProjectOrOrgFunc) History() []AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll is bn object thbt
// describes bn invocbtion of method ListRepositoriesByProjectOrOrg on bn
// instbnce of MockAzureDevOpsClient.
type AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.ListRepositoriesByProjectOrOrgArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bzuredevops.Repository
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientListRepositoriesByProjectOrOrgFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientSetWbitForRbteLimitFunc describes the behbvior when the
// SetWbitForRbteLimit method of the pbrent MockAzureDevOpsClient instbnce
// is invoked.
type AzureDevOpsClientSetWbitForRbteLimitFunc struct {
	defbultHook func(bool)
	hooks       []func(bool)
	history     []AzureDevOpsClientSetWbitForRbteLimitFuncCbll
	mutex       sync.Mutex
}

// SetWbitForRbteLimit delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) SetWbitForRbteLimit(v0 bool) {
	m.SetWbitForRbteLimitFunc.nextHook()(v0)
	m.SetWbitForRbteLimitFunc.bppendCbll(AzureDevOpsClientSetWbitForRbteLimitFuncCbll{v0})
	return
}

// SetDefbultHook sets function thbt is cblled when the SetWbitForRbteLimit
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) SetDefbultHook(hook func(bool)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetWbitForRbteLimit method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) PushHook(hook func(bool)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(bool) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) PushReturn() {
	f.PushHook(func(bool) {
		return
	})
}

func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) nextHook() func(bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) bppendCbll(r0 AzureDevOpsClientSetWbitForRbteLimitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// AzureDevOpsClientSetWbitForRbteLimitFuncCbll objects describing the
// invocbtions of this function.
func (f *AzureDevOpsClientSetWbitForRbteLimitFunc) History() []AzureDevOpsClientSetWbitForRbteLimitFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientSetWbitForRbteLimitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientSetWbitForRbteLimitFuncCbll is bn object thbt describes
// bn invocbtion of method SetWbitForRbteLimit on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientSetWbitForRbteLimitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientSetWbitForRbteLimitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientSetWbitForRbteLimitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// AzureDevOpsClientUpdbtePullRequestFunc describes the behbvior when the
// UpdbtePullRequest method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientUpdbtePullRequestFunc struct {
	defbultHook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error)
	hooks       []func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error)
	history     []AzureDevOpsClientUpdbtePullRequestFuncCbll
	mutex       sync.Mutex
}

// UpdbtePullRequest delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) UpdbtePullRequest(v0 context.Context, v1 bzuredevops.PullRequestCommonArgs, v2 bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
	r0, r1 := m.UpdbtePullRequestFunc.nextHook()(v0, v1, v2)
	m.UpdbtePullRequestFunc.bppendCbll(AzureDevOpsClientUpdbtePullRequestFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the UpdbtePullRequest
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientUpdbtePullRequestFunc) SetDefbultHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbtePullRequest method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientUpdbtePullRequestFunc) PushHook(hook func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientUpdbtePullRequestFunc) SetDefbultReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.SetDefbultHook(func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientUpdbtePullRequestFunc) PushReturn(r0 bzuredevops.PullRequest, r1 error) {
	f.PushHook(func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientUpdbtePullRequestFunc) nextHook() func(context.Context, bzuredevops.PullRequestCommonArgs, bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientUpdbtePullRequestFunc) bppendCbll(r0 AzureDevOpsClientUpdbtePullRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientUpdbtePullRequestFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientUpdbtePullRequestFunc) History() []AzureDevOpsClientUpdbtePullRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientUpdbtePullRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientUpdbtePullRequestFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbtePullRequest on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientUpdbtePullRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bzuredevops.PullRequestCommonArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bzuredevops.PullRequestUpdbteInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.PullRequest
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientUpdbtePullRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientUpdbtePullRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// AzureDevOpsClientWithAuthenticbtorFunc describes the behbvior when the
// WithAuthenticbtor method of the pbrent MockAzureDevOpsClient instbnce is
// invoked.
type AzureDevOpsClientWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) (bzuredevops.Client, error)
	hooks       []func(buth.Authenticbtor) (bzuredevops.Client, error)
	history     []AzureDevOpsClientWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAzureDevOpsClient) WithAuthenticbtor(v0 buth.Authenticbtor) (bzuredevops.Client, error) {
	r0, r1 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(AzureDevOpsClientWithAuthenticbtorFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockAzureDevOpsClient instbnce is invoked bnd the
// hook queue is empty.
func (f *AzureDevOpsClientWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) (bzuredevops.Client, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockAzureDevOpsClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AzureDevOpsClientWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) (bzuredevops.Client, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AzureDevOpsClientWithAuthenticbtorFunc) SetDefbultReturn(r0 bzuredevops.Client, r1 error) {
	f.SetDefbultHook(func(buth.Authenticbtor) (bzuredevops.Client, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AzureDevOpsClientWithAuthenticbtorFunc) PushReturn(r0 bzuredevops.Client, r1 error) {
	f.PushHook(func(buth.Authenticbtor) (bzuredevops.Client, error) {
		return r0, r1
	})
}

func (f *AzureDevOpsClientWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) (bzuredevops.Client, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AzureDevOpsClientWithAuthenticbtorFunc) bppendCbll(r0 AzureDevOpsClientWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AzureDevOpsClientWithAuthenticbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *AzureDevOpsClientWithAuthenticbtorFunc) History() []AzureDevOpsClientWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]AzureDevOpsClientWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AzureDevOpsClientWithAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method WithAuthenticbtor on bn instbnce of
// MockAzureDevOpsClient.
type AzureDevOpsClientWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bzuredevops.Client
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AzureDevOpsClientWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AzureDevOpsClientWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockGerritClient is b mock implementbtion of the Client interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit)
// used for unit testing.
type MockGerritClient struct {
	// AbbndonChbngeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AbbndonChbnge.
	AbbndonChbngeFunc *GerritClientAbbndonChbngeFunc
	// AuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method Authenticbtor.
	AuthenticbtorFunc *GerritClientAuthenticbtorFunc
	// DeleteChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DeleteChbnge.
	DeleteChbngeFunc *GerritClientDeleteChbngeFunc
	// GetAuthenticbtedUserAccountFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetAuthenticbtedUserAccount.
	GetAuthenticbtedUserAccountFunc *GerritClientGetAuthenticbtedUserAccountFunc
	// GetChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetChbnge.
	GetChbngeFunc *GerritClientGetChbngeFunc
	// GetChbngeReviewsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetChbngeReviews.
	GetChbngeReviewsFunc *GerritClientGetChbngeReviewsFunc
	// GetGroupFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetGroup.
	GetGroupFunc *GerritClientGetGroupFunc
	// GetURLFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetURL.
	GetURLFunc *GerritClientGetURLFunc
	// ListProjectsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ListProjects.
	ListProjectsFunc *GerritClientListProjectsFunc
	// MoveChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MoveChbnge.
	MoveChbngeFunc *GerritClientMoveChbngeFunc
	// RestoreChbngeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RestoreChbnge.
	RestoreChbngeFunc *GerritClientRestoreChbngeFunc
	// SetCommitMessbgeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetCommitMessbge.
	SetCommitMessbgeFunc *GerritClientSetCommitMessbgeFunc
	// SetRebdyForReviewFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetRebdyForReview.
	SetRebdyForReviewFunc *GerritClientSetRebdyForReviewFunc
	// SetWIPFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method SetWIP.
	SetWIPFunc *GerritClientSetWIPFunc
	// SubmitChbngeFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SubmitChbnge.
	SubmitChbngeFunc *GerritClientSubmitChbngeFunc
	// WithAuthenticbtorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithAuthenticbtor.
	WithAuthenticbtorFunc *GerritClientWithAuthenticbtorFunc
	// WriteReviewCommentFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WriteReviewComment.
	WriteReviewCommentFunc *GerritClientWriteReviewCommentFunc
}

// NewMockGerritClient crebtes b new mock of the Client interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockGerritClient() *MockGerritClient {
	return &MockGerritClient{
		AbbndonChbngeFunc: &GerritClientAbbndonChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		AuthenticbtorFunc: &GerritClientAuthenticbtorFunc{
			defbultHook: func() (r0 buth.Authenticbtor) {
				return
			},
		},
		DeleteChbngeFunc: &GerritClientDeleteChbngeFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		GetAuthenticbtedUserAccountFunc: &GerritClientGetAuthenticbtedUserAccountFunc{
			defbultHook: func(context.Context) (r0 *gerrit.Account, r1 error) {
				return
			},
		},
		GetChbngeFunc: &GerritClientGetChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		GetChbngeReviewsFunc: &GerritClientGetChbngeReviewsFunc{
			defbultHook: func(context.Context, string) (r0 *[]gerrit.Reviewer, r1 error) {
				return
			},
		},
		GetGroupFunc: &GerritClientGetGroupFunc{
			defbultHook: func(context.Context, string) (r0 gerrit.Group, r1 error) {
				return
			},
		},
		GetURLFunc: &GerritClientGetURLFunc{
			defbultHook: func() (r0 *url.URL) {
				return
			},
		},
		ListProjectsFunc: &GerritClientListProjectsFunc{
			defbultHook: func(context.Context, gerrit.ListProjectsArgs) (r0 gerrit.ListProjectsResponse, r1 bool, r2 error) {
				return
			},
		},
		MoveChbngeFunc: &GerritClientMoveChbngeFunc{
			defbultHook: func(context.Context, string, gerrit.MoveChbngePbylobd) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		RestoreChbngeFunc: &GerritClientRestoreChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		SetCommitMessbgeFunc: &GerritClientSetCommitMessbgeFunc{
			defbultHook: func(context.Context, string, gerrit.SetCommitMessbgePbylobd) (r0 error) {
				return
			},
		},
		SetRebdyForReviewFunc: &GerritClientSetRebdyForReviewFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		SetWIPFunc: &GerritClientSetWIPFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		SubmitChbngeFunc: &GerritClientSubmitChbngeFunc{
			defbultHook: func(context.Context, string) (r0 *gerrit.Chbnge, r1 error) {
				return
			},
		},
		WithAuthenticbtorFunc: &GerritClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (r0 gerrit.Client, r1 error) {
				return
			},
		},
		WriteReviewCommentFunc: &GerritClientWriteReviewCommentFunc{
			defbultHook: func(context.Context, string, gerrit.ChbngeReviewComment) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockGerritClient crebtes b new mock of the Client interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGerritClient() *MockGerritClient {
	return &MockGerritClient{
		AbbndonChbngeFunc: &GerritClientAbbndonChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.AbbndonChbnge")
			},
		},
		AuthenticbtorFunc: &GerritClientAuthenticbtorFunc{
			defbultHook: func() buth.Authenticbtor {
				pbnic("unexpected invocbtion of MockGerritClient.Authenticbtor")
			},
		},
		DeleteChbngeFunc: &GerritClientDeleteChbngeFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockGerritClient.DeleteChbnge")
			},
		},
		GetAuthenticbtedUserAccountFunc: &GerritClientGetAuthenticbtedUserAccountFunc{
			defbultHook: func(context.Context) (*gerrit.Account, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetAuthenticbtedUserAccount")
			},
		},
		GetChbngeFunc: &GerritClientGetChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetChbnge")
			},
		},
		GetChbngeReviewsFunc: &GerritClientGetChbngeReviewsFunc{
			defbultHook: func(context.Context, string) (*[]gerrit.Reviewer, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetChbngeReviews")
			},
		},
		GetGroupFunc: &GerritClientGetGroupFunc{
			defbultHook: func(context.Context, string) (gerrit.Group, error) {
				pbnic("unexpected invocbtion of MockGerritClient.GetGroup")
			},
		},
		GetURLFunc: &GerritClientGetURLFunc{
			defbultHook: func() *url.URL {
				pbnic("unexpected invocbtion of MockGerritClient.GetURL")
			},
		},
		ListProjectsFunc: &GerritClientListProjectsFunc{
			defbultHook: func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
				pbnic("unexpected invocbtion of MockGerritClient.ListProjects")
			},
		},
		MoveChbngeFunc: &GerritClientMoveChbngeFunc{
			defbultHook: func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.MoveChbnge")
			},
		},
		RestoreChbngeFunc: &GerritClientRestoreChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.RestoreChbnge")
			},
		},
		SetCommitMessbgeFunc: &GerritClientSetCommitMessbgeFunc{
			defbultHook: func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
				pbnic("unexpected invocbtion of MockGerritClient.SetCommitMessbge")
			},
		},
		SetRebdyForReviewFunc: &GerritClientSetRebdyForReviewFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockGerritClient.SetRebdyForReview")
			},
		},
		SetWIPFunc: &GerritClientSetWIPFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockGerritClient.SetWIP")
			},
		},
		SubmitChbngeFunc: &GerritClientSubmitChbngeFunc{
			defbultHook: func(context.Context, string) (*gerrit.Chbnge, error) {
				pbnic("unexpected invocbtion of MockGerritClient.SubmitChbnge")
			},
		},
		WithAuthenticbtorFunc: &GerritClientWithAuthenticbtorFunc{
			defbultHook: func(buth.Authenticbtor) (gerrit.Client, error) {
				pbnic("unexpected invocbtion of MockGerritClient.WithAuthenticbtor")
			},
		},
		WriteReviewCommentFunc: &GerritClientWriteReviewCommentFunc{
			defbultHook: func(context.Context, string, gerrit.ChbngeReviewComment) error {
				pbnic("unexpected invocbtion of MockGerritClient.WriteReviewComment")
			},
		},
	}
}

// NewMockGerritClientFrom crebtes b new mock of the MockGerritClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGerritClientFrom(i gerrit.Client) *MockGerritClient {
	return &MockGerritClient{
		AbbndonChbngeFunc: &GerritClientAbbndonChbngeFunc{
			defbultHook: i.AbbndonChbnge,
		},
		AuthenticbtorFunc: &GerritClientAuthenticbtorFunc{
			defbultHook: i.Authenticbtor,
		},
		DeleteChbngeFunc: &GerritClientDeleteChbngeFunc{
			defbultHook: i.DeleteChbnge,
		},
		GetAuthenticbtedUserAccountFunc: &GerritClientGetAuthenticbtedUserAccountFunc{
			defbultHook: i.GetAuthenticbtedUserAccount,
		},
		GetChbngeFunc: &GerritClientGetChbngeFunc{
			defbultHook: i.GetChbnge,
		},
		GetChbngeReviewsFunc: &GerritClientGetChbngeReviewsFunc{
			defbultHook: i.GetChbngeReviews,
		},
		GetGroupFunc: &GerritClientGetGroupFunc{
			defbultHook: i.GetGroup,
		},
		GetURLFunc: &GerritClientGetURLFunc{
			defbultHook: i.GetURL,
		},
		ListProjectsFunc: &GerritClientListProjectsFunc{
			defbultHook: i.ListProjects,
		},
		MoveChbngeFunc: &GerritClientMoveChbngeFunc{
			defbultHook: i.MoveChbnge,
		},
		RestoreChbngeFunc: &GerritClientRestoreChbngeFunc{
			defbultHook: i.RestoreChbnge,
		},
		SetCommitMessbgeFunc: &GerritClientSetCommitMessbgeFunc{
			defbultHook: i.SetCommitMessbge,
		},
		SetRebdyForReviewFunc: &GerritClientSetRebdyForReviewFunc{
			defbultHook: i.SetRebdyForReview,
		},
		SetWIPFunc: &GerritClientSetWIPFunc{
			defbultHook: i.SetWIP,
		},
		SubmitChbngeFunc: &GerritClientSubmitChbngeFunc{
			defbultHook: i.SubmitChbnge,
		},
		WithAuthenticbtorFunc: &GerritClientWithAuthenticbtorFunc{
			defbultHook: i.WithAuthenticbtor,
		},
		WriteReviewCommentFunc: &GerritClientWriteReviewCommentFunc{
			defbultHook: i.WriteReviewComment,
		},
	}
}

// GerritClientAbbndonChbngeFunc describes the behbvior when the
// AbbndonChbnge method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientAbbndonChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientAbbndonChbngeFuncCbll
	mutex       sync.Mutex
}

// AbbndonChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) AbbndonChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.AbbndonChbngeFunc.nextHook()(v0, v1)
	m.AbbndonChbngeFunc.bppendCbll(GerritClientAbbndonChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AbbndonChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientAbbndonChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AbbndonChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientAbbndonChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientAbbndonChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientAbbndonChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientAbbndonChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientAbbndonChbngeFunc) bppendCbll(r0 GerritClientAbbndonChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientAbbndonChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientAbbndonChbngeFunc) History() []GerritClientAbbndonChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientAbbndonChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientAbbndonChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method AbbndonChbnge on bn instbnce of MockGerritClient.
type GerritClientAbbndonChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientAbbndonChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientAbbndonChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientAuthenticbtorFunc describes the behbvior when the
// Authenticbtor method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientAuthenticbtorFunc struct {
	defbultHook func() buth.Authenticbtor
	hooks       []func() buth.Authenticbtor
	history     []GerritClientAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// Authenticbtor delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) Authenticbtor() buth.Authenticbtor {
	r0 := m.AuthenticbtorFunc.nextHook()()
	m.AuthenticbtorFunc.bppendCbll(GerritClientAuthenticbtorFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Authenticbtor method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientAuthenticbtorFunc) SetDefbultHook(hook func() buth.Authenticbtor) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Authenticbtor method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientAuthenticbtorFunc) PushHook(hook func() buth.Authenticbtor) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientAuthenticbtorFunc) SetDefbultReturn(r0 buth.Authenticbtor) {
	f.SetDefbultHook(func() buth.Authenticbtor {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientAuthenticbtorFunc) PushReturn(r0 buth.Authenticbtor) {
	f.PushHook(func() buth.Authenticbtor {
		return r0
	})
}

func (f *GerritClientAuthenticbtorFunc) nextHook() func() buth.Authenticbtor {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientAuthenticbtorFunc) bppendCbll(r0 GerritClientAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientAuthenticbtorFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientAuthenticbtorFunc) History() []GerritClientAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method Authenticbtor on bn instbnce of MockGerritClient.
type GerritClientAuthenticbtorFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 buth.Authenticbtor
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientDeleteChbngeFunc describes the behbvior when the DeleteChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientDeleteChbngeFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []GerritClientDeleteChbngeFuncCbll
	mutex       sync.Mutex
}

// DeleteChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) DeleteChbnge(v0 context.Context, v1 string) error {
	r0 := m.DeleteChbngeFunc.nextHook()(v0, v1)
	m.DeleteChbngeFunc.bppendCbll(GerritClientDeleteChbngeFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DeleteChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientDeleteChbngeFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientDeleteChbngeFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientDeleteChbngeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientDeleteChbngeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *GerritClientDeleteChbngeFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientDeleteChbngeFunc) bppendCbll(r0 GerritClientDeleteChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientDeleteChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientDeleteChbngeFunc) History() []GerritClientDeleteChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientDeleteChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientDeleteChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteChbnge on bn instbnce of MockGerritClient.
type GerritClientDeleteChbngeFuncCbll struct {
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
func (c GerritClientDeleteChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientDeleteChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientGetAuthenticbtedUserAccountFunc describes the behbvior when
// the GetAuthenticbtedUserAccount method of the pbrent MockGerritClient
// instbnce is invoked.
type GerritClientGetAuthenticbtedUserAccountFunc struct {
	defbultHook func(context.Context) (*gerrit.Account, error)
	hooks       []func(context.Context) (*gerrit.Account, error)
	history     []GerritClientGetAuthenticbtedUserAccountFuncCbll
	mutex       sync.Mutex
}

// GetAuthenticbtedUserAccount delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetAuthenticbtedUserAccount(v0 context.Context) (*gerrit.Account, error) {
	r0, r1 := m.GetAuthenticbtedUserAccountFunc.nextHook()(v0)
	m.GetAuthenticbtedUserAccountFunc.bppendCbll(GerritClientGetAuthenticbtedUserAccountFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetAuthenticbtedUserAccount method of the pbrent MockGerritClient
// instbnce is invoked bnd the hook queue is empty.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) SetDefbultHook(hook func(context.Context) (*gerrit.Account, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuthenticbtedUserAccount method of the pbrent MockGerritClient
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) PushHook(hook func(context.Context) (*gerrit.Account, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) SetDefbultReturn(r0 *gerrit.Account, r1 error) {
	f.SetDefbultHook(func(context.Context) (*gerrit.Account, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) PushReturn(r0 *gerrit.Account, r1 error) {
	f.PushHook(func(context.Context) (*gerrit.Account, error) {
		return r0, r1
	})
}

func (f *GerritClientGetAuthenticbtedUserAccountFunc) nextHook() func(context.Context) (*gerrit.Account, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetAuthenticbtedUserAccountFunc) bppendCbll(r0 GerritClientGetAuthenticbtedUserAccountFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GerritClientGetAuthenticbtedUserAccountFuncCbll objects describing the
// invocbtions of this function.
func (f *GerritClientGetAuthenticbtedUserAccountFunc) History() []GerritClientGetAuthenticbtedUserAccountFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetAuthenticbtedUserAccountFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetAuthenticbtedUserAccountFuncCbll is bn object thbt
// describes bn invocbtion of method GetAuthenticbtedUserAccount on bn
// instbnce of MockGerritClient.
type GerritClientGetAuthenticbtedUserAccountFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Account
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetAuthenticbtedUserAccountFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetAuthenticbtedUserAccountFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetChbngeFunc describes the behbvior when the GetChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientGetChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientGetChbngeFuncCbll
	mutex       sync.Mutex
}

// GetChbnge delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.GetChbngeFunc.nextHook()(v0, v1)
	m.GetChbngeFunc.bppendCbll(GerritClientGetChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetChbnge method of
// the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientGetChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetChbnge method of the pbrent MockGerritClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientGetChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientGetChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetChbngeFunc) bppendCbll(r0 GerritClientGetChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientGetChbngeFunc) History() []GerritClientGetChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetChbngeFuncCbll is bn object thbt describes bn invocbtion
// of method GetChbnge on bn instbnce of MockGerritClient.
type GerritClientGetChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetChbngeReviewsFunc describes the behbvior when the
// GetChbngeReviews method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientGetChbngeReviewsFunc struct {
	defbultHook func(context.Context, string) (*[]gerrit.Reviewer, error)
	hooks       []func(context.Context, string) (*[]gerrit.Reviewer, error)
	history     []GerritClientGetChbngeReviewsFuncCbll
	mutex       sync.Mutex
}

// GetChbngeReviews delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetChbngeReviews(v0 context.Context, v1 string) (*[]gerrit.Reviewer, error) {
	r0, r1 := m.GetChbngeReviewsFunc.nextHook()(v0, v1)
	m.GetChbngeReviewsFunc.bppendCbll(GerritClientGetChbngeReviewsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetChbngeReviews
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientGetChbngeReviewsFunc) SetDefbultHook(hook func(context.Context, string) (*[]gerrit.Reviewer, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetChbngeReviews method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientGetChbngeReviewsFunc) PushHook(hook func(context.Context, string) (*[]gerrit.Reviewer, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetChbngeReviewsFunc) SetDefbultReturn(r0 *[]gerrit.Reviewer, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*[]gerrit.Reviewer, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetChbngeReviewsFunc) PushReturn(r0 *[]gerrit.Reviewer, r1 error) {
	f.PushHook(func(context.Context, string) (*[]gerrit.Reviewer, error) {
		return r0, r1
	})
}

func (f *GerritClientGetChbngeReviewsFunc) nextHook() func(context.Context, string) (*[]gerrit.Reviewer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetChbngeReviewsFunc) bppendCbll(r0 GerritClientGetChbngeReviewsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetChbngeReviewsFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientGetChbngeReviewsFunc) History() []GerritClientGetChbngeReviewsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetChbngeReviewsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetChbngeReviewsFuncCbll is bn object thbt describes bn
// invocbtion of method GetChbngeReviews on bn instbnce of MockGerritClient.
type GerritClientGetChbngeReviewsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *[]gerrit.Reviewer
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetChbngeReviewsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetChbngeReviewsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetGroupFunc describes the behbvior when the GetGroup method
// of the pbrent MockGerritClient instbnce is invoked.
type GerritClientGetGroupFunc struct {
	defbultHook func(context.Context, string) (gerrit.Group, error)
	hooks       []func(context.Context, string) (gerrit.Group, error)
	history     []GerritClientGetGroupFuncCbll
	mutex       sync.Mutex
}

// GetGroup delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetGroup(v0 context.Context, v1 string) (gerrit.Group, error) {
	r0, r1 := m.GetGroupFunc.nextHook()(v0, v1)
	m.GetGroupFunc.bppendCbll(GerritClientGetGroupFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetGroup method of
// the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientGetGroupFunc) SetDefbultHook(hook func(context.Context, string) (gerrit.Group, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetGroup method of the pbrent MockGerritClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientGetGroupFunc) PushHook(hook func(context.Context, string) (gerrit.Group, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetGroupFunc) SetDefbultReturn(r0 gerrit.Group, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (gerrit.Group, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetGroupFunc) PushReturn(r0 gerrit.Group, r1 error) {
	f.PushHook(func(context.Context, string) (gerrit.Group, error) {
		return r0, r1
	})
}

func (f *GerritClientGetGroupFunc) nextHook() func(context.Context, string) (gerrit.Group, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetGroupFunc) bppendCbll(r0 GerritClientGetGroupFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetGroupFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientGetGroupFunc) History() []GerritClientGetGroupFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetGroupFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetGroupFuncCbll is bn object thbt describes bn invocbtion of
// method GetGroup on bn instbnce of MockGerritClient.
type GerritClientGetGroupFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gerrit.Group
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetGroupFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetGroupFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientGetURLFunc describes the behbvior when the GetURL method of
// the pbrent MockGerritClient instbnce is invoked.
type GerritClientGetURLFunc struct {
	defbultHook func() *url.URL
	hooks       []func() *url.URL
	history     []GerritClientGetURLFuncCbll
	mutex       sync.Mutex
}

// GetURL delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) GetURL() *url.URL {
	r0 := m.GetURLFunc.nextHook()()
	m.GetURLFunc.bppendCbll(GerritClientGetURLFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GetURL method of the
// pbrent MockGerritClient instbnce is invoked bnd the hook queue is empty.
func (f *GerritClientGetURLFunc) SetDefbultHook(hook func() *url.URL) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetURL method of the pbrent MockGerritClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientGetURLFunc) PushHook(hook func() *url.URL) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientGetURLFunc) SetDefbultReturn(r0 *url.URL) {
	f.SetDefbultHook(func() *url.URL {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientGetURLFunc) PushReturn(r0 *url.URL) {
	f.PushHook(func() *url.URL {
		return r0
	})
}

func (f *GerritClientGetURLFunc) nextHook() func() *url.URL {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientGetURLFunc) bppendCbll(r0 GerritClientGetURLFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientGetURLFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientGetURLFunc) History() []GerritClientGetURLFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientGetURLFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientGetURLFuncCbll is bn object thbt describes bn invocbtion of
// method GetURL on bn instbnce of MockGerritClient.
type GerritClientGetURLFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *url.URL
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientGetURLFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientGetURLFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientListProjectsFunc describes the behbvior when the ListProjects
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientListProjectsFunc struct {
	defbultHook func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)
	hooks       []func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)
	history     []GerritClientListProjectsFuncCbll
	mutex       sync.Mutex
}

// ListProjects delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) ListProjects(v0 context.Context, v1 gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
	r0, r1, r2 := m.ListProjectsFunc.nextHook()(v0, v1)
	m.ListProjectsFunc.bppendCbll(GerritClientListProjectsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ListProjects method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientListProjectsFunc) SetDefbultHook(hook func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListProjects method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientListProjectsFunc) PushHook(hook func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientListProjectsFunc) SetDefbultReturn(r0 gerrit.ListProjectsResponse, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientListProjectsFunc) PushReturn(r0 gerrit.ListProjectsResponse, r1 bool, r2 error) {
	f.PushHook(func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
		return r0, r1, r2
	})
}

func (f *GerritClientListProjectsFunc) nextHook() func(context.Context, gerrit.ListProjectsArgs) (gerrit.ListProjectsResponse, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientListProjectsFunc) bppendCbll(r0 GerritClientListProjectsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientListProjectsFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientListProjectsFunc) History() []GerritClientListProjectsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientListProjectsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientListProjectsFuncCbll is bn object thbt describes bn
// invocbtion of method ListProjects on bn instbnce of MockGerritClient.
type GerritClientListProjectsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 gerrit.ListProjectsArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gerrit.ListProjectsResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientListProjectsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientListProjectsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// GerritClientMoveChbngeFunc describes the behbvior when the MoveChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientMoveChbngeFunc struct {
	defbultHook func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)
	history     []GerritClientMoveChbngeFuncCbll
	mutex       sync.Mutex
}

// MoveChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) MoveChbnge(v0 context.Context, v1 string, v2 gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
	r0, r1 := m.MoveChbngeFunc.nextHook()(v0, v1, v2)
	m.MoveChbngeFunc.bppendCbll(GerritClientMoveChbngeFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MoveChbnge method of
// the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientMoveChbngeFunc) SetDefbultHook(hook func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MoveChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientMoveChbngeFunc) PushHook(hook func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientMoveChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientMoveChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientMoveChbngeFunc) nextHook() func(context.Context, string, gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientMoveChbngeFunc) bppendCbll(r0 GerritClientMoveChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientMoveChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientMoveChbngeFunc) History() []GerritClientMoveChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientMoveChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientMoveChbngeFuncCbll is bn object thbt describes bn invocbtion
// of method MoveChbnge on bn instbnce of MockGerritClient.
type GerritClientMoveChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gerrit.MoveChbngePbylobd
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientMoveChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientMoveChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientRestoreChbngeFunc describes the behbvior when the
// RestoreChbnge method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientRestoreChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientRestoreChbngeFuncCbll
	mutex       sync.Mutex
}

// RestoreChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) RestoreChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.RestoreChbngeFunc.nextHook()(v0, v1)
	m.RestoreChbngeFunc.bppendCbll(GerritClientRestoreChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RestoreChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientRestoreChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RestoreChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientRestoreChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientRestoreChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientRestoreChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientRestoreChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientRestoreChbngeFunc) bppendCbll(r0 GerritClientRestoreChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientRestoreChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientRestoreChbngeFunc) History() []GerritClientRestoreChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientRestoreChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientRestoreChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method RestoreChbnge on bn instbnce of MockGerritClient.
type GerritClientRestoreChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientRestoreChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientRestoreChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientSetCommitMessbgeFunc describes the behbvior when the
// SetCommitMessbge method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientSetCommitMessbgeFunc struct {
	defbultHook func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error
	hooks       []func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error
	history     []GerritClientSetCommitMessbgeFuncCbll
	mutex       sync.Mutex
}

// SetCommitMessbge delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SetCommitMessbge(v0 context.Context, v1 string, v2 gerrit.SetCommitMessbgePbylobd) error {
	r0 := m.SetCommitMessbgeFunc.nextHook()(v0, v1, v2)
	m.SetCommitMessbgeFunc.bppendCbll(GerritClientSetCommitMessbgeFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetCommitMessbge
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientSetCommitMessbgeFunc) SetDefbultHook(hook func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetCommitMessbge method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientSetCommitMessbgeFunc) PushHook(hook func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSetCommitMessbgeFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSetCommitMessbgeFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
		return r0
	})
}

func (f *GerritClientSetCommitMessbgeFunc) nextHook() func(context.Context, string, gerrit.SetCommitMessbgePbylobd) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSetCommitMessbgeFunc) bppendCbll(r0 GerritClientSetCommitMessbgeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSetCommitMessbgeFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientSetCommitMessbgeFunc) History() []GerritClientSetCommitMessbgeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSetCommitMessbgeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSetCommitMessbgeFuncCbll is bn object thbt describes bn
// invocbtion of method SetCommitMessbge on bn instbnce of MockGerritClient.
type GerritClientSetCommitMessbgeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gerrit.SetCommitMessbgePbylobd
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientSetCommitMessbgeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSetCommitMessbgeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientSetRebdyForReviewFunc describes the behbvior when the
// SetRebdyForReview method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientSetRebdyForReviewFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []GerritClientSetRebdyForReviewFuncCbll
	mutex       sync.Mutex
}

// SetRebdyForReview delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SetRebdyForReview(v0 context.Context, v1 string) error {
	r0 := m.SetRebdyForReviewFunc.nextHook()(v0, v1)
	m.SetRebdyForReviewFunc.bppendCbll(GerritClientSetRebdyForReviewFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetRebdyForReview
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientSetRebdyForReviewFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetRebdyForReview method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientSetRebdyForReviewFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSetRebdyForReviewFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSetRebdyForReviewFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *GerritClientSetRebdyForReviewFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSetRebdyForReviewFunc) bppendCbll(r0 GerritClientSetRebdyForReviewFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSetRebdyForReviewFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientSetRebdyForReviewFunc) History() []GerritClientSetRebdyForReviewFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSetRebdyForReviewFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSetRebdyForReviewFuncCbll is bn object thbt describes bn
// invocbtion of method SetRebdyForReview on bn instbnce of
// MockGerritClient.
type GerritClientSetRebdyForReviewFuncCbll struct {
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
func (c GerritClientSetRebdyForReviewFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSetRebdyForReviewFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientSetWIPFunc describes the behbvior when the SetWIP method of
// the pbrent MockGerritClient instbnce is invoked.
type GerritClientSetWIPFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []GerritClientSetWIPFuncCbll
	mutex       sync.Mutex
}

// SetWIP delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SetWIP(v0 context.Context, v1 string) error {
	r0 := m.SetWIPFunc.nextHook()(v0, v1)
	m.SetWIPFunc.bppendCbll(GerritClientSetWIPFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetWIP method of the
// pbrent MockGerritClient instbnce is invoked bnd the hook queue is empty.
func (f *GerritClientSetWIPFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetWIP method of the pbrent MockGerritClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GerritClientSetWIPFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSetWIPFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSetWIPFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *GerritClientSetWIPFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSetWIPFunc) bppendCbll(r0 GerritClientSetWIPFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSetWIPFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientSetWIPFunc) History() []GerritClientSetWIPFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSetWIPFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSetWIPFuncCbll is bn object thbt describes bn invocbtion of
// method SetWIP on bn instbnce of MockGerritClient.
type GerritClientSetWIPFuncCbll struct {
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
func (c GerritClientSetWIPFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSetWIPFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GerritClientSubmitChbngeFunc describes the behbvior when the SubmitChbnge
// method of the pbrent MockGerritClient instbnce is invoked.
type GerritClientSubmitChbngeFunc struct {
	defbultHook func(context.Context, string) (*gerrit.Chbnge, error)
	hooks       []func(context.Context, string) (*gerrit.Chbnge, error)
	history     []GerritClientSubmitChbngeFuncCbll
	mutex       sync.Mutex
}

// SubmitChbnge delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) SubmitChbnge(v0 context.Context, v1 string) (*gerrit.Chbnge, error) {
	r0, r1 := m.SubmitChbngeFunc.nextHook()(v0, v1)
	m.SubmitChbngeFunc.bppendCbll(GerritClientSubmitChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SubmitChbnge method
// of the pbrent MockGerritClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GerritClientSubmitChbngeFunc) SetDefbultHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SubmitChbnge method of the pbrent MockGerritClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GerritClientSubmitChbngeFunc) PushHook(hook func(context.Context, string) (*gerrit.Chbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientSubmitChbngeFunc) SetDefbultReturn(r0 *gerrit.Chbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientSubmitChbngeFunc) PushReturn(r0 *gerrit.Chbnge, r1 error) {
	f.PushHook(func(context.Context, string) (*gerrit.Chbnge, error) {
		return r0, r1
	})
}

func (f *GerritClientSubmitChbngeFunc) nextHook() func(context.Context, string) (*gerrit.Chbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientSubmitChbngeFunc) bppendCbll(r0 GerritClientSubmitChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientSubmitChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GerritClientSubmitChbngeFunc) History() []GerritClientSubmitChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientSubmitChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientSubmitChbngeFuncCbll is bn object thbt describes bn
// invocbtion of method SubmitChbnge on bn instbnce of MockGerritClient.
type GerritClientSubmitChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gerrit.Chbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientSubmitChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientSubmitChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientWithAuthenticbtorFunc describes the behbvior when the
// WithAuthenticbtor method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientWithAuthenticbtorFunc struct {
	defbultHook func(buth.Authenticbtor) (gerrit.Client, error)
	hooks       []func(buth.Authenticbtor) (gerrit.Client, error)
	history     []GerritClientWithAuthenticbtorFuncCbll
	mutex       sync.Mutex
}

// WithAuthenticbtor delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) WithAuthenticbtor(v0 buth.Authenticbtor) (gerrit.Client, error) {
	r0, r1 := m.WithAuthenticbtorFunc.nextHook()(v0)
	m.WithAuthenticbtorFunc.bppendCbll(GerritClientWithAuthenticbtorFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the WithAuthenticbtor
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientWithAuthenticbtorFunc) SetDefbultHook(hook func(buth.Authenticbtor) (gerrit.Client, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithAuthenticbtor method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientWithAuthenticbtorFunc) PushHook(hook func(buth.Authenticbtor) (gerrit.Client, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientWithAuthenticbtorFunc) SetDefbultReturn(r0 gerrit.Client, r1 error) {
	f.SetDefbultHook(func(buth.Authenticbtor) (gerrit.Client, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientWithAuthenticbtorFunc) PushReturn(r0 gerrit.Client, r1 error) {
	f.PushHook(func(buth.Authenticbtor) (gerrit.Client, error) {
		return r0, r1
	})
}

func (f *GerritClientWithAuthenticbtorFunc) nextHook() func(buth.Authenticbtor) (gerrit.Client, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientWithAuthenticbtorFunc) bppendCbll(r0 GerritClientWithAuthenticbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientWithAuthenticbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientWithAuthenticbtorFunc) History() []GerritClientWithAuthenticbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientWithAuthenticbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientWithAuthenticbtorFuncCbll is bn object thbt describes bn
// invocbtion of method WithAuthenticbtor on bn instbnce of
// MockGerritClient.
type GerritClientWithAuthenticbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 buth.Authenticbtor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gerrit.Client
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientWithAuthenticbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientWithAuthenticbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GerritClientWriteReviewCommentFunc describes the behbvior when the
// WriteReviewComment method of the pbrent MockGerritClient instbnce is
// invoked.
type GerritClientWriteReviewCommentFunc struct {
	defbultHook func(context.Context, string, gerrit.ChbngeReviewComment) error
	hooks       []func(context.Context, string, gerrit.ChbngeReviewComment) error
	history     []GerritClientWriteReviewCommentFuncCbll
	mutex       sync.Mutex
}

// WriteReviewComment delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGerritClient) WriteReviewComment(v0 context.Context, v1 string, v2 gerrit.ChbngeReviewComment) error {
	r0 := m.WriteReviewCommentFunc.nextHook()(v0, v1, v2)
	m.WriteReviewCommentFunc.bppendCbll(GerritClientWriteReviewCommentFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WriteReviewComment
// method of the pbrent MockGerritClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GerritClientWriteReviewCommentFunc) SetDefbultHook(hook func(context.Context, string, gerrit.ChbngeReviewComment) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WriteReviewComment method of the pbrent MockGerritClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GerritClientWriteReviewCommentFunc) PushHook(hook func(context.Context, string, gerrit.ChbngeReviewComment) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GerritClientWriteReviewCommentFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, gerrit.ChbngeReviewComment) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GerritClientWriteReviewCommentFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, gerrit.ChbngeReviewComment) error {
		return r0
	})
}

func (f *GerritClientWriteReviewCommentFunc) nextHook() func(context.Context, string, gerrit.ChbngeReviewComment) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GerritClientWriteReviewCommentFunc) bppendCbll(r0 GerritClientWriteReviewCommentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GerritClientWriteReviewCommentFuncCbll
// objects describing the invocbtions of this function.
func (f *GerritClientWriteReviewCommentFunc) History() []GerritClientWriteReviewCommentFuncCbll {
	f.mutex.Lock()
	history := mbke([]GerritClientWriteReviewCommentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GerritClientWriteReviewCommentFuncCbll is bn object thbt describes bn
// invocbtion of method WriteReviewComment on bn instbnce of
// MockGerritClient.
type GerritClientWriteReviewCommentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gerrit.ChbngeReviewComment
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GerritClientWriteReviewCommentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GerritClientWriteReviewCommentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockGitserverClient is b mock implementbtion of the Client interfbce
// (from the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/gitserver)
// used for unit testing.
type MockGitserverClient struct {
	// AddrForRepoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method AddrForRepo.
	AddrForRepoFunc *GitserverClientAddrForRepoFunc
	// AddrsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Addrs.
	AddrsFunc *GitserverClientAddrsFunc
	// ArchiveRebderFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ArchiveRebder.
	ArchiveRebderFunc *GitserverClientArchiveRebderFunc
	// BbtchLogFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method BbtchLog.
	BbtchLogFunc *GitserverClientBbtchLogFunc
	// BlbmeFileFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method BlbmeFile.
	BlbmeFileFunc *GitserverClientBlbmeFileFunc
	// BrbnchesContbiningFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method BrbnchesContbining.
	BrbnchesContbiningFunc *GitserverClientBrbnchesContbiningFunc
	// CommitDbteFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitDbte.
	CommitDbteFunc *GitserverClientCommitDbteFunc
	// CommitExistsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitExists.
	CommitExistsFunc *GitserverClientCommitExistsFunc
	// CommitGrbphFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitGrbph.
	CommitGrbphFunc *GitserverClientCommitGrbphFunc
	// CommitLogFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitLog.
	CommitLogFunc *GitserverClientCommitLogFunc
	// CommitsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Commits.
	CommitsFunc *GitserverClientCommitsFunc
	// CommitsExistFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitsExist.
	CommitsExistFunc *GitserverClientCommitsExistFunc
	// CommitsUniqueToBrbnchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CommitsUniqueToBrbnch.
	CommitsUniqueToBrbnchFunc *GitserverClientCommitsUniqueToBrbnchFunc
	// ContributorCountFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ContributorCount.
	ContributorCountFunc *GitserverClientContributorCountFunc
	// CrebteCommitFromPbtchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteCommitFromPbtch.
	CrebteCommitFromPbtchFunc *GitserverClientCrebteCommitFromPbtchFunc
	// DiffFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Diff.
	DiffFunc *GitserverClientDiffFunc
	// DiffPbthFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method DiffPbth.
	DiffPbthFunc *GitserverClientDiffPbthFunc
	// DiffSymbolsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DiffSymbols.
	DiffSymbolsFunc *GitserverClientDiffSymbolsFunc
	// FirstEverCommitFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method FirstEverCommit.
	FirstEverCommitFunc *GitserverClientFirstEverCommitFunc
	// GetBehindAhebdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetBehindAhebd.
	GetBehindAhebdFunc *GitserverClientGetBehindAhebdFunc
	// GetCommitFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetCommit.
	GetCommitFunc *GitserverClientGetCommitFunc
	// GetCommitsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetCommits.
	GetCommitsFunc *GitserverClientGetCommitsFunc
	// GetDefbultBrbnchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDefbultBrbnch.
	GetDefbultBrbnchFunc *GitserverClientGetDefbultBrbnchFunc
	// GetObjectFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetObject.
	GetObjectFunc *GitserverClientGetObjectFunc
	// HbsCommitAfterFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method HbsCommitAfter.
	HbsCommitAfterFunc *GitserverClientHbsCommitAfterFunc
	// HebdFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hebd.
	HebdFunc *GitserverClientHebdFunc
	// IsRepoClonebbleFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method IsRepoClonebble.
	IsRepoClonebbleFunc *GitserverClientIsRepoClonebbleFunc
	// ListBrbnchesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ListBrbnches.
	ListBrbnchesFunc *GitserverClientListBrbnchesFunc
	// ListDirectoryChildrenFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListDirectoryChildren.
	ListDirectoryChildrenFunc *GitserverClientListDirectoryChildrenFunc
	// ListRefsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method ListRefs.
	ListRefsFunc *GitserverClientListRefsFunc
	// ListTbgsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method ListTbgs.
	ListTbgsFunc *GitserverClientListTbgsFunc
	// LogReverseEbchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method LogReverseEbch.
	LogReverseEbchFunc *GitserverClientLogReverseEbchFunc
	// LsFilesFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LsFiles.
	LsFilesFunc *GitserverClientLsFilesFunc
	// MergeBbseFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MergeBbse.
	MergeBbseFunc *GitserverClientMergeBbseFunc
	// NewFileRebderFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewFileRebder.
	NewFileRebderFunc *GitserverClientNewFileRebderFunc
	// P4ExecFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method P4Exec.
	P4ExecFunc *GitserverClientP4ExecFunc
	// P4GetChbngelistFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method P4GetChbngelist.
	P4GetChbngelistFunc *GitserverClientP4GetChbngelistFunc
	// RebdDirFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RebdDir.
	RebdDirFunc *GitserverClientRebdDirFunc
	// RebdFileFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RebdFile.
	RebdFileFunc *GitserverClientRebdFileFunc
	// RefDescriptionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RefDescriptions.
	RefDescriptionsFunc *GitserverClientRefDescriptionsFunc
	// RemoveFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Remove.
	RemoveFunc *GitserverClientRemoveFunc
	// RepoCloneProgressFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RepoCloneProgress.
	RepoCloneProgressFunc *GitserverClientRepoCloneProgressFunc
	// RequestRepoCloneFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RequestRepoClone.
	RequestRepoCloneFunc *GitserverClientRequestRepoCloneFunc
	// RequestRepoUpdbteFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RequestRepoUpdbte.
	RequestRepoUpdbteFunc *GitserverClientRequestRepoUpdbteFunc
	// ResolveRevisionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveRevision.
	ResolveRevisionFunc *GitserverClientResolveRevisionFunc
	// ResolveRevisionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveRevisions.
	ResolveRevisionsFunc *GitserverClientResolveRevisionsFunc
	// RevListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RevList.
	RevListFunc *GitserverClientRevListFunc
	// SebrchFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Sebrch.
	SebrchFunc *GitserverClientSebrchFunc
	// StbtFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Stbt.
	StbtFunc *GitserverClientStbtFunc
	// StrebmBlbmeFileFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method StrebmBlbmeFile.
	StrebmBlbmeFileFunc *GitserverClientStrebmBlbmeFileFunc
	// SystemInfoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SystemInfo.
	SystemInfoFunc *GitserverClientSystemInfoFunc
	// SystemsInfoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SystemsInfo.
	SystemsInfoFunc *GitserverClientSystemsInfoFunc
}

// NewMockGitserverClient crebtes b new mock of the Client interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 string) {
				return
			},
		},
		AddrsFunc: &GitserverClientAddrsFunc{
			defbultHook: func() (r0 []string) {
				return
			},
		},
		ArchiveRebderFunc: &GitserverClientArchiveRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		BbtchLogFunc: &GitserverClientBbtchLogFunc{
			defbultHook: func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) (r0 error) {
				return
			},
		},
		BlbmeFileFunc: &GitserverClientBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (r0 []*gitserver.Hunk, r1 error) {
				return
			},
		},
		BrbnchesContbiningFunc: &GitserverClientBrbnchesContbiningFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (r0 []string, r1 error) {
				return
			},
		},
		CommitDbteFunc: &GitserverClientCommitDbteFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (r0 string, r1 time.Time, r2 bool, r3 error) {
				return
			},
		},
		CommitExistsFunc: &GitserverClientCommitExistsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (r0 bool, r1 error) {
				return
			},
		},
		CommitGrbphFunc: &GitserverClientCommitGrbphFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (r0 *gitdombin.CommitGrbph, r1 error) {
				return
			},
		},
		CommitLogFunc: &GitserverClientCommitLogFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Time) (r0 []gitserver.CommitLog, r1 error) {
				return
			},
		},
		CommitsFunc: &GitserverClientCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) (r0 []*gitdombin.Commit, r1 error) {
				return
			},
		},
		CommitsExistFunc: &GitserverClientCommitsExistFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) (r0 []bool, r1 error) {
				return
			},
		},
		CommitsUniqueToBrbnchFunc: &GitserverClientCommitsUniqueToBrbnchFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (r0 mbp[string]time.Time, r1 error) {
				return
			},
		},
		ContributorCountFunc: &GitserverClientContributorCountFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) (r0 []*gitdombin.ContributorCount, r1 error) {
				return
			},
		},
		CrebteCommitFromPbtchFunc: &GitserverClientCrebteCommitFromPbtchFunc{
			defbultHook: func(context.Context, protocol.CrebteCommitFromPbtchRequest) (r0 *protocol.CrebteCommitFromPbtchResponse, r1 error) {
				return
			},
		},
		DiffFunc: &GitserverClientDiffFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (r0 *gitserver.DiffFileIterbtor, r1 error) {
				return
			},
		},
		DiffPbthFunc: &GitserverClientDiffPbthFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) (r0 []*diff.Hunk, r1 error) {
				return
			},
		},
		DiffSymbolsFunc: &GitserverClientDiffSymbolsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (r0 []byte, r1 error) {
				return
			},
		},
		FirstEverCommitFunc: &GitserverClientFirstEverCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (r0 *gitdombin.Commit, r1 error) {
				return
			},
		},
		GetBehindAhebdFunc: &GitserverClientGetBehindAhebdFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, string) (r0 *gitdombin.BehindAhebd, r1 error) {
				return
			},
		},
		GetCommitFunc: &GitserverClientGetCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (r0 *gitdombin.Commit, r1 error) {
				return
			},
		},
		GetCommitsFunc: &GitserverClientGetCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) (r0 []*gitdombin.Commit, r1 error) {
				return
			},
		},
		GetDefbultBrbnchFunc: &GitserverClientGetDefbultBrbnchFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bool) (r0 string, r1 bpi.CommitID, r2 error) {
				return
			},
		},
		GetObjectFunc: &GitserverClientGetObjectFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string) (r0 *gitdombin.GitObject, r1 error) {
				return
			},
		},
		HbsCommitAfterFunc: &GitserverClientHbsCommitAfterFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (r0 bool, r1 error) {
				return
			},
		},
		HebdFunc: &GitserverClientHebdFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (r0 string, r1 bool, r2 error) {
				return
			},
		},
		IsRepoClonebbleFunc: &GitserverClientIsRepoClonebbleFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 error) {
				return
			},
		},
		ListBrbnchesFunc: &GitserverClientListBrbnchesFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) (r0 []*gitdombin.Brbnch, r1 error) {
				return
			},
		},
		ListDirectoryChildrenFunc: &GitserverClientListDirectoryChildrenFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (r0 mbp[string][]string, r1 error) {
				return
			},
		},
		ListRefsFunc: &GitserverClientListRefsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 []gitdombin.Ref, r1 error) {
				return
			},
		},
		ListTbgsFunc: &GitserverClientListTbgsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, ...string) (r0 []*gitdombin.Tbg, r1 error) {
				return
			},
		},
		LogReverseEbchFunc: &GitserverClientLogReverseEbchFunc{
			defbultHook: func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) (r0 error) {
				return
			},
		},
		LsFilesFunc: &GitserverClientLsFilesFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) (r0 []string, r1 error) {
				return
			},
		},
		MergeBbseFunc: &GitserverClientMergeBbseFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (r0 bpi.CommitID, r1 error) {
				return
			},
		},
		NewFileRebderFunc: &GitserverClientNewFileRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		P4ExecFunc: &GitserverClientP4ExecFunc{
			defbultHook: func(context.Context, string, string, string, ...string) (r0 io.RebdCloser, r1 http.Hebder, r2 error) {
				return
			},
		},
		P4GetChbngelistFunc: &GitserverClientP4GetChbngelistFunc{
			defbultHook: func(context.Context, string, gitserver.PerforceCredentibls) (r0 *protocol.PerforceChbngelist, r1 error) {
				return
			},
		},
		RebdDirFunc: &GitserverClientRebdDirFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) (r0 []fs.FileInfo, r1 error) {
				return
			},
		},
		RebdFileFunc: &GitserverClientRebdFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (r0 []byte, r1 error) {
				return
			},
		},
		RefDescriptionsFunc: &GitserverClientRefDescriptionsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (r0 mbp[string][]gitdombin.RefDescription, r1 error) {
				return
			},
		},
		RemoveFunc: &GitserverClientRemoveFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 error) {
				return
			},
		},
		RepoCloneProgressFunc: &GitserverClientRepoCloneProgressFunc{
			defbultHook: func(context.Context, ...bpi.RepoNbme) (r0 *protocol.RepoCloneProgressResponse, r1 error) {
				return
			},
		},
		RequestRepoCloneFunc: &GitserverClientRequestRepoCloneFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 *protocol.RepoCloneResponse, r1 error) {
				return
			},
		},
		RequestRepoUpdbteFunc: &GitserverClientRequestRepoUpdbteFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Durbtion) (r0 *protocol.RepoUpdbteResponse, r1 error) {
				return
			},
		},
		ResolveRevisionFunc: &GitserverClientResolveRevisionFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (r0 bpi.CommitID, r1 error) {
				return
			},
		},
		ResolveRevisionsFunc: &GitserverClientResolveRevisionsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) (r0 []string, r1 error) {
				return
			},
		},
		RevListFunc: &GitserverClientRevListFunc{
			defbultHook: func(context.Context, string, string, func(commit string) (bool, error)) (r0 error) {
				return
			},
		},
		SebrchFunc: &GitserverClientSebrchFunc{
			defbultHook: func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (r0 bool, r1 error) {
				return
			},
		},
		StbtFunc: &GitserverClientStbtFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (r0 fs.FileInfo, r1 error) {
				return
			},
		},
		StrebmBlbmeFileFunc: &GitserverClientStrebmBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (r0 gitserver.HunkRebder, r1 error) {
				return
			},
		},
		SystemInfoFunc: &GitserverClientSystemInfoFunc{
			defbultHook: func(context.Context, string) (r0 gitserver.SystemInfo, r1 error) {
				return
			},
		},
		SystemsInfoFunc: &GitserverClientSystemsInfoFunc{
			defbultHook: func(context.Context) (r0 []gitserver.SystemInfo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockGitserverClient crebtes b new mock of the Client interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) string {
				pbnic("unexpected invocbtion of MockGitserverClient.AddrForRepo")
			},
		},
		AddrsFunc: &GitserverClientAddrsFunc{
			defbultHook: func() []string {
				pbnic("unexpected invocbtion of MockGitserverClient.Addrs")
			},
		},
		ArchiveRebderFunc: &GitserverClientArchiveRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ArchiveRebder")
			},
		},
		BbtchLogFunc: &GitserverClientBbtchLogFunc{
			defbultHook: func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error {
				pbnic("unexpected invocbtion of MockGitserverClient.BbtchLog")
			},
		},
		BlbmeFileFunc: &GitserverClientBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.BlbmeFile")
			},
		},
		BrbnchesContbiningFunc: &GitserverClientBrbnchesContbiningFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.BrbnchesContbining")
			},
		},
		CommitDbteFunc: &GitserverClientCommitDbteFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CommitDbte")
			},
		},
		CommitExistsFunc: &GitserverClientCommitExistsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CommitExists")
			},
		},
		CommitGrbphFunc: &GitserverClientCommitGrbphFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CommitGrbph")
			},
		},
		CommitLogFunc: &GitserverClientCommitLogFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CommitLog")
			},
		},
		CommitsFunc: &GitserverClientCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.Commits")
			},
		},
		CommitsExistFunc: &GitserverClientCommitsExistFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CommitsExist")
			},
		},
		CommitsUniqueToBrbnchFunc: &GitserverClientCommitsUniqueToBrbnchFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CommitsUniqueToBrbnch")
			},
		},
		ContributorCountFunc: &GitserverClientContributorCountFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ContributorCount")
			},
		},
		CrebteCommitFromPbtchFunc: &GitserverClientCrebteCommitFromPbtchFunc{
			defbultHook: func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.CrebteCommitFromPbtch")
			},
		},
		DiffFunc: &GitserverClientDiffFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.Diff")
			},
		},
		DiffPbthFunc: &GitserverClientDiffPbthFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.DiffPbth")
			},
		},
		DiffSymbolsFunc: &GitserverClientDiffSymbolsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.DiffSymbols")
			},
		},
		FirstEverCommitFunc: &GitserverClientFirstEverCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.FirstEverCommit")
			},
		},
		GetBehindAhebdFunc: &GitserverClientGetBehindAhebdFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.GetBehindAhebd")
			},
		},
		GetCommitFunc: &GitserverClientGetCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.GetCommit")
			},
		},
		GetCommitsFunc: &GitserverClientGetCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.GetCommits")
			},
		},
		GetDefbultBrbnchFunc: &GitserverClientGetDefbultBrbnchFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.GetDefbultBrbnch")
			},
		},
		GetObjectFunc: &GitserverClientGetObjectFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.GetObject")
			},
		},
		HbsCommitAfterFunc: &GitserverClientHbsCommitAfterFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.HbsCommitAfter")
			},
		},
		HebdFunc: &GitserverClientHebdFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.Hebd")
			},
		},
		IsRepoClonebbleFunc: &GitserverClientIsRepoClonebbleFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) error {
				pbnic("unexpected invocbtion of MockGitserverClient.IsRepoClonebble")
			},
		},
		ListBrbnchesFunc: &GitserverClientListBrbnchesFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ListBrbnches")
			},
		},
		ListDirectoryChildrenFunc: &GitserverClientListDirectoryChildrenFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ListDirectoryChildren")
			},
		},
		ListRefsFunc: &GitserverClientListRefsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ListRefs")
			},
		},
		ListTbgsFunc: &GitserverClientListTbgsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ListTbgs")
			},
		},
		LogReverseEbchFunc: &GitserverClientLogReverseEbchFunc{
			defbultHook: func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
				pbnic("unexpected invocbtion of MockGitserverClient.LogReverseEbch")
			},
		},
		LsFilesFunc: &GitserverClientLsFilesFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.LsFiles")
			},
		},
		MergeBbseFunc: &GitserverClientMergeBbseFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.MergeBbse")
			},
		},
		NewFileRebderFunc: &GitserverClientNewFileRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.NewFileRebder")
			},
		},
		P4ExecFunc: &GitserverClientP4ExecFunc{
			defbultHook: func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.P4Exec")
			},
		},
		P4GetChbngelistFunc: &GitserverClientP4GetChbngelistFunc{
			defbultHook: func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.P4GetChbngelist")
			},
		},
		RebdDirFunc: &GitserverClientRebdDirFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RebdDir")
			},
		},
		RebdFileFunc: &GitserverClientRebdFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RebdFile")
			},
		},
		RefDescriptionsFunc: &GitserverClientRefDescriptionsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RefDescriptions")
			},
		},
		RemoveFunc: &GitserverClientRemoveFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) error {
				pbnic("unexpected invocbtion of MockGitserverClient.Remove")
			},
		},
		RepoCloneProgressFunc: &GitserverClientRepoCloneProgressFunc{
			defbultHook: func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RepoCloneProgress")
			},
		},
		RequestRepoCloneFunc: &GitserverClientRequestRepoCloneFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RequestRepoClone")
			},
		},
		RequestRepoUpdbteFunc: &GitserverClientRequestRepoUpdbteFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RequestRepoUpdbte")
			},
		},
		ResolveRevisionFunc: &GitserverClientResolveRevisionFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ResolveRevision")
			},
		},
		ResolveRevisionsFunc: &GitserverClientResolveRevisionsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.ResolveRevisions")
			},
		},
		RevListFunc: &GitserverClientRevListFunc{
			defbultHook: func(context.Context, string, string, func(commit string) (bool, error)) error {
				pbnic("unexpected invocbtion of MockGitserverClient.RevList")
			},
		},
		SebrchFunc: &GitserverClientSebrchFunc{
			defbultHook: func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.Sebrch")
			},
		},
		StbtFunc: &GitserverClientStbtFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.Stbt")
			},
		},
		StrebmBlbmeFileFunc: &GitserverClientStrebmBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.StrebmBlbmeFile")
			},
		},
		SystemInfoFunc: &GitserverClientSystemInfoFunc{
			defbultHook: func(context.Context, string) (gitserver.SystemInfo, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.SystemInfo")
			},
		},
		SystemsInfoFunc: &GitserverClientSystemsInfoFunc{
			defbultHook: func(context.Context) ([]gitserver.SystemInfo, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.SystemsInfo")
			},
		},
	}
}

// NewMockGitserverClientFrom crebtes b new mock of the MockGitserverClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGitserverClientFrom(i gitserver.Client) *MockGitserverClient {
	return &MockGitserverClient{
		AddrForRepoFunc: &GitserverClientAddrForRepoFunc{
			defbultHook: i.AddrForRepo,
		},
		AddrsFunc: &GitserverClientAddrsFunc{
			defbultHook: i.Addrs,
		},
		ArchiveRebderFunc: &GitserverClientArchiveRebderFunc{
			defbultHook: i.ArchiveRebder,
		},
		BbtchLogFunc: &GitserverClientBbtchLogFunc{
			defbultHook: i.BbtchLog,
		},
		BlbmeFileFunc: &GitserverClientBlbmeFileFunc{
			defbultHook: i.BlbmeFile,
		},
		BrbnchesContbiningFunc: &GitserverClientBrbnchesContbiningFunc{
			defbultHook: i.BrbnchesContbining,
		},
		CommitDbteFunc: &GitserverClientCommitDbteFunc{
			defbultHook: i.CommitDbte,
		},
		CommitExistsFunc: &GitserverClientCommitExistsFunc{
			defbultHook: i.CommitExists,
		},
		CommitGrbphFunc: &GitserverClientCommitGrbphFunc{
			defbultHook: i.CommitGrbph,
		},
		CommitLogFunc: &GitserverClientCommitLogFunc{
			defbultHook: i.CommitLog,
		},
		CommitsFunc: &GitserverClientCommitsFunc{
			defbultHook: i.Commits,
		},
		CommitsExistFunc: &GitserverClientCommitsExistFunc{
			defbultHook: i.CommitsExist,
		},
		CommitsUniqueToBrbnchFunc: &GitserverClientCommitsUniqueToBrbnchFunc{
			defbultHook: i.CommitsUniqueToBrbnch,
		},
		ContributorCountFunc: &GitserverClientContributorCountFunc{
			defbultHook: i.ContributorCount,
		},
		CrebteCommitFromPbtchFunc: &GitserverClientCrebteCommitFromPbtchFunc{
			defbultHook: i.CrebteCommitFromPbtch,
		},
		DiffFunc: &GitserverClientDiffFunc{
			defbultHook: i.Diff,
		},
		DiffPbthFunc: &GitserverClientDiffPbthFunc{
			defbultHook: i.DiffPbth,
		},
		DiffSymbolsFunc: &GitserverClientDiffSymbolsFunc{
			defbultHook: i.DiffSymbols,
		},
		FirstEverCommitFunc: &GitserverClientFirstEverCommitFunc{
			defbultHook: i.FirstEverCommit,
		},
		GetBehindAhebdFunc: &GitserverClientGetBehindAhebdFunc{
			defbultHook: i.GetBehindAhebd,
		},
		GetCommitFunc: &GitserverClientGetCommitFunc{
			defbultHook: i.GetCommit,
		},
		GetCommitsFunc: &GitserverClientGetCommitsFunc{
			defbultHook: i.GetCommits,
		},
		GetDefbultBrbnchFunc: &GitserverClientGetDefbultBrbnchFunc{
			defbultHook: i.GetDefbultBrbnch,
		},
		GetObjectFunc: &GitserverClientGetObjectFunc{
			defbultHook: i.GetObject,
		},
		HbsCommitAfterFunc: &GitserverClientHbsCommitAfterFunc{
			defbultHook: i.HbsCommitAfter,
		},
		HebdFunc: &GitserverClientHebdFunc{
			defbultHook: i.Hebd,
		},
		IsRepoClonebbleFunc: &GitserverClientIsRepoClonebbleFunc{
			defbultHook: i.IsRepoClonebble,
		},
		ListBrbnchesFunc: &GitserverClientListBrbnchesFunc{
			defbultHook: i.ListBrbnches,
		},
		ListDirectoryChildrenFunc: &GitserverClientListDirectoryChildrenFunc{
			defbultHook: i.ListDirectoryChildren,
		},
		ListRefsFunc: &GitserverClientListRefsFunc{
			defbultHook: i.ListRefs,
		},
		ListTbgsFunc: &GitserverClientListTbgsFunc{
			defbultHook: i.ListTbgs,
		},
		LogReverseEbchFunc: &GitserverClientLogReverseEbchFunc{
			defbultHook: i.LogReverseEbch,
		},
		LsFilesFunc: &GitserverClientLsFilesFunc{
			defbultHook: i.LsFiles,
		},
		MergeBbseFunc: &GitserverClientMergeBbseFunc{
			defbultHook: i.MergeBbse,
		},
		NewFileRebderFunc: &GitserverClientNewFileRebderFunc{
			defbultHook: i.NewFileRebder,
		},
		P4ExecFunc: &GitserverClientP4ExecFunc{
			defbultHook: i.P4Exec,
		},
		P4GetChbngelistFunc: &GitserverClientP4GetChbngelistFunc{
			defbultHook: i.P4GetChbngelist,
		},
		RebdDirFunc: &GitserverClientRebdDirFunc{
			defbultHook: i.RebdDir,
		},
		RebdFileFunc: &GitserverClientRebdFileFunc{
			defbultHook: i.RebdFile,
		},
		RefDescriptionsFunc: &GitserverClientRefDescriptionsFunc{
			defbultHook: i.RefDescriptions,
		},
		RemoveFunc: &GitserverClientRemoveFunc{
			defbultHook: i.Remove,
		},
		RepoCloneProgressFunc: &GitserverClientRepoCloneProgressFunc{
			defbultHook: i.RepoCloneProgress,
		},
		RequestRepoCloneFunc: &GitserverClientRequestRepoCloneFunc{
			defbultHook: i.RequestRepoClone,
		},
		RequestRepoUpdbteFunc: &GitserverClientRequestRepoUpdbteFunc{
			defbultHook: i.RequestRepoUpdbte,
		},
		ResolveRevisionFunc: &GitserverClientResolveRevisionFunc{
			defbultHook: i.ResolveRevision,
		},
		ResolveRevisionsFunc: &GitserverClientResolveRevisionsFunc{
			defbultHook: i.ResolveRevisions,
		},
		RevListFunc: &GitserverClientRevListFunc{
			defbultHook: i.RevList,
		},
		SebrchFunc: &GitserverClientSebrchFunc{
			defbultHook: i.Sebrch,
		},
		StbtFunc: &GitserverClientStbtFunc{
			defbultHook: i.Stbt,
		},
		StrebmBlbmeFileFunc: &GitserverClientStrebmBlbmeFileFunc{
			defbultHook: i.StrebmBlbmeFile,
		},
		SystemInfoFunc: &GitserverClientSystemInfoFunc{
			defbultHook: i.SystemInfo,
		},
		SystemsInfoFunc: &GitserverClientSystemsInfoFunc{
			defbultHook: i.SystemsInfo,
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

// GitserverClientAddrsFunc describes the behbvior when the Addrs method of
// the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientAddrsFunc struct {
	defbultHook func() []string
	hooks       []func() []string
	history     []GitserverClientAddrsFuncCbll
	mutex       sync.Mutex
}

// Addrs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Addrs() []string {
	r0 := m.AddrsFunc.nextHook()()
	m.AddrsFunc.bppendCbll(GitserverClientAddrsFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Addrs method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientAddrsFunc) SetDefbultHook(hook func() []string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Addrs method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientAddrsFunc) PushHook(hook func() []string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientAddrsFunc) SetDefbultReturn(r0 []string) {
	f.SetDefbultHook(func() []string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientAddrsFunc) PushReturn(r0 []string) {
	f.PushHook(func() []string {
		return r0
	})
}

func (f *GitserverClientAddrsFunc) nextHook() func() []string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientAddrsFunc) bppendCbll(r0 GitserverClientAddrsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientAddrsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientAddrsFunc) History() []GitserverClientAddrsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientAddrsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientAddrsFuncCbll is bn object thbt describes bn invocbtion of
// method Addrs on bn instbnce of MockGitserverClient.
type GitserverClientAddrsFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientAddrsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientAddrsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitserverClientArchiveRebderFunc describes the behbvior when the
// ArchiveRebder method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientArchiveRebderFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)
	history     []GitserverClientArchiveRebderFuncCbll
	mutex       sync.Mutex
}

// ArchiveRebder delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ArchiveRebder(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 gitserver.ArchiveOptions) (io.RebdCloser, error) {
	r0, r1 := m.ArchiveRebderFunc.nextHook()(v0, v1, v2, v3)
	m.ArchiveRebderFunc.bppendCbll(GitserverClientArchiveRebderFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ArchiveRebder method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientArchiveRebderFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ArchiveRebder method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientArchiveRebderFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientArchiveRebderFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientArchiveRebderFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *GitserverClientArchiveRebderFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientArchiveRebderFunc) bppendCbll(r0 GitserverClientArchiveRebderFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientArchiveRebderFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientArchiveRebderFunc) History() []GitserverClientArchiveRebderFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientArchiveRebderFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientArchiveRebderFuncCbll is bn object thbt describes bn
// invocbtion of method ArchiveRebder on bn instbnce of MockGitserverClient.
type GitserverClientArchiveRebderFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 gitserver.ArchiveOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientArchiveRebderFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientArchiveRebderFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientBbtchLogFunc describes the behbvior when the BbtchLog
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientBbtchLogFunc struct {
	defbultHook func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error
	hooks       []func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error
	history     []GitserverClientBbtchLogFuncCbll
	mutex       sync.Mutex
}

// BbtchLog delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) BbtchLog(v0 context.Context, v1 gitserver.BbtchLogOptions, v2 gitserver.BbtchLogCbllbbck) error {
	r0 := m.BbtchLogFunc.nextHook()(v0, v1, v2)
	m.BbtchLogFunc.bppendCbll(GitserverClientBbtchLogFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the BbtchLog method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientBbtchLogFunc) SetDefbultHook(hook func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BbtchLog method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientBbtchLogFunc) PushHook(hook func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientBbtchLogFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientBbtchLogFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error {
		return r0
	})
}

func (f *GitserverClientBbtchLogFunc) nextHook() func(context.Context, gitserver.BbtchLogOptions, gitserver.BbtchLogCbllbbck) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientBbtchLogFunc) bppendCbll(r0 GitserverClientBbtchLogFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientBbtchLogFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientBbtchLogFunc) History() []GitserverClientBbtchLogFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientBbtchLogFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientBbtchLogFuncCbll is bn object thbt describes bn invocbtion
// of method BbtchLog on bn instbnce of MockGitserverClient.
type GitserverClientBbtchLogFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 gitserver.BbtchLogOptions
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.BbtchLogCbllbbck
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientBbtchLogFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientBbtchLogFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitserverClientBlbmeFileFunc describes the behbvior when the BlbmeFile
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientBlbmeFileFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error)
	history     []GitserverClientBlbmeFileFuncCbll
	mutex       sync.Mutex
}

// BlbmeFile delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) BlbmeFile(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error) {
	r0, r1 := m.BlbmeFileFunc.nextHook()(v0, v1, v2, v3, v4)
	m.BlbmeFileFunc.bppendCbll(GitserverClientBlbmeFileFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the BlbmeFile method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientBlbmeFileFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BlbmeFile method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientBlbmeFileFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientBlbmeFileFunc) SetDefbultReturn(r0 []*gitserver.Hunk, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientBlbmeFileFunc) PushReturn(r0 []*gitserver.Hunk, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error) {
		return r0, r1
	})
}

func (f *GitserverClientBlbmeFileFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) ([]*gitserver.Hunk, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientBlbmeFileFunc) bppendCbll(r0 GitserverClientBlbmeFileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientBlbmeFileFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientBlbmeFileFunc) History() []GitserverClientBlbmeFileFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientBlbmeFileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientBlbmeFileFuncCbll is bn object thbt describes bn
// invocbtion of method BlbmeFile on bn instbnce of MockGitserverClient.
type GitserverClientBlbmeFileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 *gitserver.BlbmeOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitserver.Hunk
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientBlbmeFileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientBlbmeFileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientBrbnchesContbiningFunc describes the behbvior when the
// BrbnchesContbining method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientBrbnchesContbiningFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)
	history     []GitserverClientBrbnchesContbiningFuncCbll
	mutex       sync.Mutex
}

// BrbnchesContbining delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) BrbnchesContbining(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID) ([]string, error) {
	r0, r1 := m.BrbnchesContbiningFunc.nextHook()(v0, v1, v2, v3)
	m.BrbnchesContbiningFunc.bppendCbll(GitserverClientBrbnchesContbiningFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the BrbnchesContbining
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientBrbnchesContbiningFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BrbnchesContbining method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientBrbnchesContbiningFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientBrbnchesContbiningFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientBrbnchesContbiningFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
		return r0, r1
	})
}

func (f *GitserverClientBrbnchesContbiningFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientBrbnchesContbiningFunc) bppendCbll(r0 GitserverClientBrbnchesContbiningFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientBrbnchesContbiningFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientBrbnchesContbiningFunc) History() []GitserverClientBrbnchesContbiningFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientBrbnchesContbiningFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientBrbnchesContbiningFuncCbll is bn object thbt describes bn
// invocbtion of method BrbnchesContbining on bn instbnce of
// MockGitserverClient.
type GitserverClientBrbnchesContbiningFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientBrbnchesContbiningFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientBrbnchesContbiningFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCommitDbteFunc describes the behbvior when the CommitDbte
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientCommitDbteFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)
	history     []GitserverClientCommitDbteFuncCbll
	mutex       sync.Mutex
}

// CommitDbte delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CommitDbte(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID) (string, time.Time, bool, error) {
	r0, r1, r2, r3 := m.CommitDbteFunc.nextHook()(v0, v1, v2, v3)
	m.CommitDbteFunc.bppendCbll(GitserverClientCommitDbteFuncCbll{v0, v1, v2, v3, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the CommitDbte method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientCommitDbteFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitDbte method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientCommitDbteFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitDbteFunc) SetDefbultReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitDbteFunc) PushReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *GitserverClientCommitDbteFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitDbteFunc) bppendCbll(r0 GitserverClientCommitDbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientCommitDbteFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientCommitDbteFunc) History() []GitserverClientCommitDbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitDbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitDbteFuncCbll is bn object thbt describes bn
// invocbtion of method CommitDbte on bn instbnce of MockGitserverClient.
type GitserverClientCommitDbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 time.Time
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitDbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitDbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// GitserverClientCommitExistsFunc describes the behbvior when the
// CommitExists method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientCommitExistsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)
	history     []GitserverClientCommitExistsFuncCbll
	mutex       sync.Mutex
}

// CommitExists delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CommitExists(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID) (bool, error) {
	r0, r1 := m.CommitExistsFunc.nextHook()(v0, v1, v2, v3)
	m.CommitExistsFunc.bppendCbll(GitserverClientCommitExistsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitExists method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientCommitExistsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitExists method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientCommitExistsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitExistsFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitExistsFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitExistsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitExistsFunc) bppendCbll(r0 GitserverClientCommitExistsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientCommitExistsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientCommitExistsFunc) History() []GitserverClientCommitExistsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitExistsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitExistsFuncCbll is bn object thbt describes bn
// invocbtion of method CommitExists on bn instbnce of MockGitserverClient.
type GitserverClientCommitExistsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitExistsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitExistsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCommitGrbphFunc describes the behbvior when the
// CommitGrbph method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientCommitGrbphFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error)
	hooks       []func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error)
	history     []GitserverClientCommitGrbphFuncCbll
	mutex       sync.Mutex
}

// CommitGrbph delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CommitGrbph(v0 context.Context, v1 bpi.RepoNbme, v2 gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
	r0, r1 := m.CommitGrbphFunc.nextHook()(v0, v1, v2)
	m.CommitGrbphFunc.bppendCbll(GitserverClientCommitGrbphFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitGrbph method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientCommitGrbphFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitGrbph method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientCommitGrbphFunc) PushHook(hook func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitGrbphFunc) SetDefbultReturn(r0 *gitdombin.CommitGrbph, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitGrbphFunc) PushReturn(r0 *gitdombin.CommitGrbph, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitGrbphFunc) nextHook() func(context.Context, bpi.RepoNbme, gitserver.CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitGrbphFunc) bppendCbll(r0 GitserverClientCommitGrbphFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientCommitGrbphFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientCommitGrbphFunc) History() []GitserverClientCommitGrbphFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitGrbphFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitGrbphFuncCbll is bn object thbt describes bn
// invocbtion of method CommitGrbph on bn instbnce of MockGitserverClient.
type GitserverClientCommitGrbphFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.CommitGrbphOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.CommitGrbph
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitGrbphFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitGrbphFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCommitLogFunc describes the behbvior when the CommitLog
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientCommitLogFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error)
	hooks       []func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error)
	history     []GitserverClientCommitLogFuncCbll
	mutex       sync.Mutex
}

// CommitLog delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CommitLog(v0 context.Context, v1 bpi.RepoNbme, v2 time.Time) ([]gitserver.CommitLog, error) {
	r0, r1 := m.CommitLogFunc.nextHook()(v0, v1, v2)
	m.CommitLogFunc.bppendCbll(GitserverClientCommitLogFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitLog method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientCommitLogFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitLog method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientCommitLogFunc) PushHook(hook func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitLogFunc) SetDefbultReturn(r0 []gitserver.CommitLog, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitLogFunc) PushReturn(r0 []gitserver.CommitLog, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitLogFunc) nextHook() func(context.Context, bpi.RepoNbme, time.Time) ([]gitserver.CommitLog, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitLogFunc) bppendCbll(r0 GitserverClientCommitLogFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientCommitLogFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientCommitLogFunc) History() []GitserverClientCommitLogFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitLogFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitLogFuncCbll is bn object thbt describes bn
// invocbtion of method CommitLog on bn instbnce of MockGitserverClient.
type GitserverClientCommitLogFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []gitserver.CommitLog
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitLogFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitLogFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCommitsFunc describes the behbvior when the Commits method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientCommitsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error)
	history     []GitserverClientCommitsFuncCbll
	mutex       sync.Mutex
}

// Commits delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Commits(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
	r0, r1 := m.CommitsFunc.nextHook()(v0, v1, v2, v3)
	m.CommitsFunc.bppendCbll(GitserverClientCommitsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Commits method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientCommitsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Commits method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientCommitsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitsFunc) SetDefbultReturn(r0 []*gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitsFunc) PushReturn(r0 []*gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, gitserver.CommitsOptions) ([]*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitsFunc) bppendCbll(r0 GitserverClientCommitsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientCommitsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientCommitsFunc) History() []GitserverClientCommitsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitsFuncCbll is bn object thbt describes bn invocbtion
// of method Commits on bn instbnce of MockGitserverClient.
type GitserverClientCommitsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 gitserver.CommitsOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCommitsExistFunc describes the behbvior when the
// CommitsExist method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientCommitsExistFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)
	history     []GitserverClientCommitsExistFuncCbll
	mutex       sync.Mutex
}

// CommitsExist delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CommitsExist(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 []bpi.RepoCommit) ([]bool, error) {
	r0, r1 := m.CommitsExistFunc.nextHook()(v0, v1, v2)
	m.CommitsExistFunc.bppendCbll(GitserverClientCommitsExistFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitsExist method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientCommitsExistFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitsExist method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientCommitsExistFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitsExistFunc) SetDefbultReturn(r0 []bool, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitsExistFunc) PushReturn(r0 []bool, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitsExistFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitsExistFunc) bppendCbll(r0 GitserverClientCommitsExistFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientCommitsExistFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientCommitsExistFunc) History() []GitserverClientCommitsExistFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitsExistFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitsExistFuncCbll is bn object thbt describes bn
// invocbtion of method CommitsExist on bn instbnce of MockGitserverClient.
type GitserverClientCommitsExistFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []bpi.RepoCommit
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitsExistFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitsExistFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCommitsUniqueToBrbnchFunc describes the behbvior when the
// CommitsUniqueToBrbnch method of the pbrent MockGitserverClient instbnce
// is invoked.
type GitserverClientCommitsUniqueToBrbnchFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)
	history     []GitserverClientCommitsUniqueToBrbnchFuncCbll
	mutex       sync.Mutex
}

// CommitsUniqueToBrbnch delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CommitsUniqueToBrbnch(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 bool, v5 *time.Time) (mbp[string]time.Time, error) {
	r0, r1 := m.CommitsUniqueToBrbnchFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.CommitsUniqueToBrbnchFunc.bppendCbll(GitserverClientCommitsUniqueToBrbnchFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CommitsUniqueToBrbnch method of the pbrent MockGitserverClient instbnce
// is invoked bnd the hook queue is empty.
func (f *GitserverClientCommitsUniqueToBrbnchFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitsUniqueToBrbnch method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientCommitsUniqueToBrbnchFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCommitsUniqueToBrbnchFunc) SetDefbultReturn(r0 mbp[string]time.Time, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCommitsUniqueToBrbnchFunc) PushReturn(r0 mbp[string]time.Time, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitsUniqueToBrbnchFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitsUniqueToBrbnchFunc) bppendCbll(r0 GitserverClientCommitsUniqueToBrbnchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitserverClientCommitsUniqueToBrbnchFuncCbll objects describing the
// invocbtions of this function.
func (f *GitserverClientCommitsUniqueToBrbnchFunc) History() []GitserverClientCommitsUniqueToBrbnchFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCommitsUniqueToBrbnchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitsUniqueToBrbnchFuncCbll is bn object thbt describes
// bn invocbtion of method CommitsUniqueToBrbnch on bn instbnce of
// MockGitserverClient.
type GitserverClientCommitsUniqueToBrbnchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 bool
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 *time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]time.Time
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCommitsUniqueToBrbnchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCommitsUniqueToBrbnchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientContributorCountFunc describes the behbvior when the
// ContributorCount method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientContributorCountFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error)
	hooks       []func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error)
	history     []GitserverClientContributorCountFuncCbll
	mutex       sync.Mutex
}

// ContributorCount delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ContributorCount(v0 context.Context, v1 bpi.RepoNbme, v2 gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error) {
	r0, r1 := m.ContributorCountFunc.nextHook()(v0, v1, v2)
	m.ContributorCountFunc.bppendCbll(GitserverClientContributorCountFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ContributorCount
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientContributorCountFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ContributorCount method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientContributorCountFunc) PushHook(hook func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientContributorCountFunc) SetDefbultReturn(r0 []*gitdombin.ContributorCount, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientContributorCountFunc) PushReturn(r0 []*gitdombin.ContributorCount, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error) {
		return r0, r1
	})
}

func (f *GitserverClientContributorCountFunc) nextHook() func(context.Context, bpi.RepoNbme, gitserver.ContributorOptions) ([]*gitdombin.ContributorCount, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientContributorCountFunc) bppendCbll(r0 GitserverClientContributorCountFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientContributorCountFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientContributorCountFunc) History() []GitserverClientContributorCountFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientContributorCountFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientContributorCountFuncCbll is bn object thbt describes bn
// invocbtion of method ContributorCount on bn instbnce of
// MockGitserverClient.
type GitserverClientContributorCountFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.ContributorOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.ContributorCount
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientContributorCountFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientContributorCountFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientCrebteCommitFromPbtchFunc describes the behbvior when the
// CrebteCommitFromPbtch method of the pbrent MockGitserverClient instbnce
// is invoked.
type GitserverClientCrebteCommitFromPbtchFunc struct {
	defbultHook func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)
	hooks       []func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)
	history     []GitserverClientCrebteCommitFromPbtchFuncCbll
	mutex       sync.Mutex
}

// CrebteCommitFromPbtch delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) CrebteCommitFromPbtch(v0 context.Context, v1 protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
	r0, r1 := m.CrebteCommitFromPbtchFunc.nextHook()(v0, v1)
	m.CrebteCommitFromPbtchFunc.bppendCbll(GitserverClientCrebteCommitFromPbtchFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebteCommitFromPbtch method of the pbrent MockGitserverClient instbnce
// is invoked bnd the hook queue is empty.
func (f *GitserverClientCrebteCommitFromPbtchFunc) SetDefbultHook(hook func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteCommitFromPbtch method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientCrebteCommitFromPbtchFunc) PushHook(hook func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientCrebteCommitFromPbtchFunc) SetDefbultReturn(r0 *protocol.CrebteCommitFromPbtchResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientCrebteCommitFromPbtchFunc) PushReturn(r0 *protocol.CrebteCommitFromPbtchResponse, r1 error) {
	f.PushHook(func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
		return r0, r1
	})
}

func (f *GitserverClientCrebteCommitFromPbtchFunc) nextHook() func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCrebteCommitFromPbtchFunc) bppendCbll(r0 GitserverClientCrebteCommitFromPbtchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitserverClientCrebteCommitFromPbtchFuncCbll objects describing the
// invocbtions of this function.
func (f *GitserverClientCrebteCommitFromPbtchFunc) History() []GitserverClientCrebteCommitFromPbtchFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientCrebteCommitFromPbtchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCrebteCommitFromPbtchFuncCbll is bn object thbt describes
// bn invocbtion of method CrebteCommitFromPbtch on bn instbnce of
// MockGitserverClient.
type GitserverClientCrebteCommitFromPbtchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 protocol.CrebteCommitFromPbtchRequest
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.CrebteCommitFromPbtchResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientCrebteCommitFromPbtchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientCrebteCommitFromPbtchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientDiffFunc describes the behbvior when the Diff method of
// the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientDiffFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error)
	history     []GitserverClientDiffFuncCbll
	mutex       sync.Mutex
}

// Diff delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Diff(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error) {
	r0, r1 := m.DiffFunc.nextHook()(v0, v1, v2)
	m.DiffFunc.bppendCbll(GitserverClientDiffFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Diff method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientDiffFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Diff method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientDiffFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientDiffFunc) SetDefbultReturn(r0 *gitserver.DiffFileIterbtor, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientDiffFunc) PushReturn(r0 *gitserver.DiffFileIterbtor, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error) {
		return r0, r1
	})
}

func (f *GitserverClientDiffFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, gitserver.DiffOptions) (*gitserver.DiffFileIterbtor, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientDiffFunc) bppendCbll(r0 GitserverClientDiffFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientDiffFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientDiffFunc) History() []GitserverClientDiffFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientDiffFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientDiffFuncCbll is bn object thbt describes bn invocbtion of
// method Diff on bn instbnce of MockGitserverClient.
type GitserverClientDiffFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.DiffOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitserver.DiffFileIterbtor
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientDiffFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientDiffFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientDiffPbthFunc describes the behbvior when the DiffPbth
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientDiffPbthFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)
	history     []GitserverClientDiffPbthFuncCbll
	mutex       sync.Mutex
}

// DiffPbth delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) DiffPbth(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 string, v5 string) ([]*diff.Hunk, error) {
	r0, r1 := m.DiffPbthFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.DiffPbthFunc.bppendCbll(GitserverClientDiffPbthFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DiffPbth method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientDiffPbthFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DiffPbth method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientDiffPbthFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientDiffPbthFunc) SetDefbultReturn(r0 []*diff.Hunk, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientDiffPbthFunc) PushReturn(r0 []*diff.Hunk, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
		return r0, r1
	})
}

func (f *GitserverClientDiffPbthFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientDiffPbthFunc) bppendCbll(r0 GitserverClientDiffPbthFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientDiffPbthFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientDiffPbthFunc) History() []GitserverClientDiffPbthFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientDiffPbthFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientDiffPbthFuncCbll is bn object thbt describes bn invocbtion
// of method DiffPbth on bn instbnce of MockGitserverClient.
type GitserverClientDiffPbthFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*diff.Hunk
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientDiffPbthFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientDiffPbthFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientDiffSymbolsFunc describes the behbvior when the
// DiffSymbols method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientDiffSymbolsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)
	hooks       []func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)
	history     []GitserverClientDiffSymbolsFuncCbll
	mutex       sync.Mutex
}

// DiffSymbols delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) DiffSymbols(v0 context.Context, v1 bpi.RepoNbme, v2 bpi.CommitID, v3 bpi.CommitID) ([]byte, error) {
	r0, r1 := m.DiffSymbolsFunc.nextHook()(v0, v1, v2, v3)
	m.DiffSymbolsFunc.bppendCbll(GitserverClientDiffSymbolsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DiffSymbols method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientDiffSymbolsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DiffSymbols method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientDiffSymbolsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientDiffSymbolsFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientDiffSymbolsFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
		return r0, r1
	})
}

func (f *GitserverClientDiffSymbolsFunc) nextHook() func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientDiffSymbolsFunc) bppendCbll(r0 GitserverClientDiffSymbolsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientDiffSymbolsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientDiffSymbolsFunc) History() []GitserverClientDiffSymbolsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientDiffSymbolsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientDiffSymbolsFuncCbll is bn object thbt describes bn
// invocbtion of method DiffSymbols on bn instbnce of MockGitserverClient.
type GitserverClientDiffSymbolsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []byte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientDiffSymbolsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientDiffSymbolsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientFirstEverCommitFunc describes the behbvior when the
// FirstEverCommit method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientFirstEverCommitFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)
	history     []GitserverClientFirstEverCommitFuncCbll
	mutex       sync.Mutex
}

// FirstEverCommit delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) FirstEverCommit(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme) (*gitdombin.Commit, error) {
	r0, r1 := m.FirstEverCommitFunc.nextHook()(v0, v1, v2)
	m.FirstEverCommitFunc.bppendCbll(GitserverClientFirstEverCommitFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the FirstEverCommit
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientFirstEverCommitFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// FirstEverCommit method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientFirstEverCommitFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientFirstEverCommitFunc) SetDefbultReturn(r0 *gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientFirstEverCommitFunc) PushReturn(r0 *gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *GitserverClientFirstEverCommitFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientFirstEverCommitFunc) bppendCbll(r0 GitserverClientFirstEverCommitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientFirstEverCommitFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientFirstEverCommitFunc) History() []GitserverClientFirstEverCommitFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientFirstEverCommitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientFirstEverCommitFuncCbll is bn object thbt describes bn
// invocbtion of method FirstEverCommit on bn instbnce of
// MockGitserverClient.
type GitserverClientFirstEverCommitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientFirstEverCommitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientFirstEverCommitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientGetBehindAhebdFunc describes the behbvior when the
// GetBehindAhebd method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientGetBehindAhebdFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)
	hooks       []func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)
	history     []GitserverClientGetBehindAhebdFuncCbll
	mutex       sync.Mutex
}

// GetBehindAhebd delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) GetBehindAhebd(v0 context.Context, v1 bpi.RepoNbme, v2 string, v3 string) (*gitdombin.BehindAhebd, error) {
	r0, r1 := m.GetBehindAhebdFunc.nextHook()(v0, v1, v2, v3)
	m.GetBehindAhebdFunc.bppendCbll(GitserverClientGetBehindAhebdFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBehindAhebd
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientGetBehindAhebdFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBehindAhebd method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientGetBehindAhebdFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientGetBehindAhebdFunc) SetDefbultReturn(r0 *gitdombin.BehindAhebd, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientGetBehindAhebdFunc) PushReturn(r0 *gitdombin.BehindAhebd, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
		return r0, r1
	})
}

func (f *GitserverClientGetBehindAhebdFunc) nextHook() func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientGetBehindAhebdFunc) bppendCbll(r0 GitserverClientGetBehindAhebdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientGetBehindAhebdFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientGetBehindAhebdFunc) History() []GitserverClientGetBehindAhebdFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientGetBehindAhebdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientGetBehindAhebdFuncCbll is bn object thbt describes bn
// invocbtion of method GetBehindAhebd on bn instbnce of
// MockGitserverClient.
type GitserverClientGetBehindAhebdFuncCbll struct {
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
	Result0 *gitdombin.BehindAhebd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientGetBehindAhebdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientGetBehindAhebdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientGetCommitFunc describes the behbvior when the GetCommit
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientGetCommitFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error)
	history     []GitserverClientGetCommitFuncCbll
	mutex       sync.Mutex
}

// GetCommit delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) GetCommit(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error) {
	r0, r1 := m.GetCommitFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetCommitFunc.bppendCbll(GitserverClientGetCommitFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetCommit method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientGetCommitFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommit method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientGetCommitFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientGetCommitFunc) SetDefbultReturn(r0 *gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientGetCommitFunc) PushReturn(r0 *gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *GitserverClientGetCommitFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientGetCommitFunc) bppendCbll(r0 GitserverClientGetCommitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientGetCommitFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientGetCommitFunc) History() []GitserverClientGetCommitFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientGetCommitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientGetCommitFuncCbll is bn object thbt describes bn
// invocbtion of method GetCommit on bn instbnce of MockGitserverClient.
type GitserverClientGetCommitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 gitserver.ResolveRevisionOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientGetCommitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientGetCommitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientGetCommitsFunc describes the behbvior when the GetCommits
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientGetCommitsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)
	history     []GitserverClientGetCommitsFuncCbll
	mutex       sync.Mutex
}

// GetCommits delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) GetCommits(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 []bpi.RepoCommit, v3 bool) ([]*gitdombin.Commit, error) {
	r0, r1 := m.GetCommitsFunc.nextHook()(v0, v1, v2, v3)
	m.GetCommitsFunc.bppendCbll(GitserverClientGetCommitsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetCommits method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientGetCommitsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommits method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientGetCommitsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientGetCommitsFunc) SetDefbultReturn(r0 []*gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientGetCommitsFunc) PushReturn(r0 []*gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *GitserverClientGetCommitsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientGetCommitsFunc) bppendCbll(r0 GitserverClientGetCommitsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientGetCommitsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientGetCommitsFunc) History() []GitserverClientGetCommitsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientGetCommitsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientGetCommitsFuncCbll is bn object thbt describes bn
// invocbtion of method GetCommits on bn instbnce of MockGitserverClient.
type GitserverClientGetCommitsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []bpi.RepoCommit
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientGetCommitsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientGetCommitsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientGetDefbultBrbnchFunc describes the behbvior when the
// GetDefbultBrbnch method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientGetDefbultBrbnchFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)
	hooks       []func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)
	history     []GitserverClientGetDefbultBrbnchFuncCbll
	mutex       sync.Mutex
}

// GetDefbultBrbnch delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) GetDefbultBrbnch(v0 context.Context, v1 bpi.RepoNbme, v2 bool) (string, bpi.CommitID, error) {
	r0, r1, r2 := m.GetDefbultBrbnchFunc.nextHook()(v0, v1, v2)
	m.GetDefbultBrbnchFunc.bppendCbll(GitserverClientGetDefbultBrbnchFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetDefbultBrbnch
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientGetDefbultBrbnchFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDefbultBrbnch method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientGetDefbultBrbnchFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientGetDefbultBrbnchFunc) SetDefbultReturn(r0 string, r1 bpi.CommitID, r2 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientGetDefbultBrbnchFunc) PushReturn(r0 string, r1 bpi.CommitID, r2 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
		return r0, r1, r2
	})
}

func (f *GitserverClientGetDefbultBrbnchFunc) nextHook() func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientGetDefbultBrbnchFunc) bppendCbll(r0 GitserverClientGetDefbultBrbnchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientGetDefbultBrbnchFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientGetDefbultBrbnchFunc) History() []GitserverClientGetDefbultBrbnchFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientGetDefbultBrbnchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientGetDefbultBrbnchFuncCbll is bn object thbt describes bn
// invocbtion of method GetDefbultBrbnch on bn instbnce of
// MockGitserverClient.
type GitserverClientGetDefbultBrbnchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bpi.CommitID
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientGetDefbultBrbnchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientGetDefbultBrbnchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// GitserverClientGetObjectFunc describes the behbvior when the GetObject
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientGetObjectFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)
	hooks       []func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)
	history     []GitserverClientGetObjectFuncCbll
	mutex       sync.Mutex
}

// GetObject delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) GetObject(v0 context.Context, v1 bpi.RepoNbme, v2 string) (*gitdombin.GitObject, error) {
	r0, r1 := m.GetObjectFunc.nextHook()(v0, v1, v2)
	m.GetObjectFunc.bppendCbll(GitserverClientGetObjectFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetObject method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientGetObjectFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetObject method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientGetObjectFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientGetObjectFunc) SetDefbultReturn(r0 *gitdombin.GitObject, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientGetObjectFunc) PushReturn(r0 *gitdombin.GitObject, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
		return r0, r1
	})
}

func (f *GitserverClientGetObjectFunc) nextHook() func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientGetObjectFunc) bppendCbll(r0 GitserverClientGetObjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientGetObjectFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientGetObjectFunc) History() []GitserverClientGetObjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientGetObjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientGetObjectFuncCbll is bn object thbt describes bn
// invocbtion of method GetObject on bn instbnce of MockGitserverClient.
type GitserverClientGetObjectFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.GitObject
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientGetObjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientGetObjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientHbsCommitAfterFunc describes the behbvior when the
// HbsCommitAfter method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientHbsCommitAfterFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)
	history     []GitserverClientHbsCommitAfterFuncCbll
	mutex       sync.Mutex
}

// HbsCommitAfter delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) HbsCommitAfter(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 string) (bool, error) {
	r0, r1 := m.HbsCommitAfterFunc.nextHook()(v0, v1, v2, v3, v4)
	m.HbsCommitAfterFunc.bppendCbll(GitserverClientHbsCommitAfterFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the HbsCommitAfter
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientHbsCommitAfterFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HbsCommitAfter method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientHbsCommitAfterFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientHbsCommitAfterFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientHbsCommitAfterFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
		return r0, r1
	})
}

func (f *GitserverClientHbsCommitAfterFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientHbsCommitAfterFunc) bppendCbll(r0 GitserverClientHbsCommitAfterFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientHbsCommitAfterFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientHbsCommitAfterFunc) History() []GitserverClientHbsCommitAfterFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientHbsCommitAfterFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientHbsCommitAfterFuncCbll is bn object thbt describes bn
// invocbtion of method HbsCommitAfter on bn instbnce of
// MockGitserverClient.
type GitserverClientHbsCommitAfterFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
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
func (c GitserverClientHbsCommitAfterFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientHbsCommitAfterFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientHebdFunc describes the behbvior when the Hebd method of
// the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientHebdFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)
	history     []GitserverClientHebdFuncCbll
	mutex       sync.Mutex
}

// Hebd delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Hebd(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme) (string, bool, error) {
	r0, r1, r2 := m.HebdFunc.nextHook()(v0, v1, v2)
	m.HebdFunc.bppendCbll(GitserverClientHebdFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Hebd method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientHebdFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hebd method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientHebdFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientHebdFunc) SetDefbultReturn(r0 string, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientHebdFunc) PushReturn(r0 string, r1 bool, r2 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
		return r0, r1, r2
	})
}

func (f *GitserverClientHebdFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientHebdFunc) bppendCbll(r0 GitserverClientHebdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientHebdFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientHebdFunc) History() []GitserverClientHebdFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientHebdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientHebdFuncCbll is bn object thbt describes bn invocbtion of
// method Hebd on bn instbnce of MockGitserverClient.
type GitserverClientHebdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientHebdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientHebdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// GitserverClientIsRepoClonebbleFunc describes the behbvior when the
// IsRepoClonebble method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientIsRepoClonebbleFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) error
	hooks       []func(context.Context, bpi.RepoNbme) error
	history     []GitserverClientIsRepoClonebbleFuncCbll
	mutex       sync.Mutex
}

// IsRepoClonebble delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) IsRepoClonebble(v0 context.Context, v1 bpi.RepoNbme) error {
	r0 := m.IsRepoClonebbleFunc.nextHook()(v0, v1)
	m.IsRepoClonebbleFunc.bppendCbll(GitserverClientIsRepoClonebbleFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the IsRepoClonebble
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientIsRepoClonebbleFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IsRepoClonebble method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientIsRepoClonebbleFunc) PushHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientIsRepoClonebbleFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientIsRepoClonebbleFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

func (f *GitserverClientIsRepoClonebbleFunc) nextHook() func(context.Context, bpi.RepoNbme) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientIsRepoClonebbleFunc) bppendCbll(r0 GitserverClientIsRepoClonebbleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientIsRepoClonebbleFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientIsRepoClonebbleFunc) History() []GitserverClientIsRepoClonebbleFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientIsRepoClonebbleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientIsRepoClonebbleFuncCbll is bn object thbt describes bn
// invocbtion of method IsRepoClonebble on bn instbnce of
// MockGitserverClient.
type GitserverClientIsRepoClonebbleFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientIsRepoClonebbleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientIsRepoClonebbleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitserverClientListBrbnchesFunc describes the behbvior when the
// ListBrbnches method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientListBrbnchesFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error)
	hooks       []func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error)
	history     []GitserverClientListBrbnchesFuncCbll
	mutex       sync.Mutex
}

// ListBrbnches delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ListBrbnches(v0 context.Context, v1 bpi.RepoNbme, v2 gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
	r0, r1 := m.ListBrbnchesFunc.nextHook()(v0, v1, v2)
	m.ListBrbnchesFunc.bppendCbll(GitserverClientListBrbnchesFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListBrbnches method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientListBrbnchesFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListBrbnches method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientListBrbnchesFunc) PushHook(hook func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientListBrbnchesFunc) SetDefbultReturn(r0 []*gitdombin.Brbnch, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientListBrbnchesFunc) PushReturn(r0 []*gitdombin.Brbnch, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
		return r0, r1
	})
}

func (f *GitserverClientListBrbnchesFunc) nextHook() func(context.Context, bpi.RepoNbme, gitserver.BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientListBrbnchesFunc) bppendCbll(r0 GitserverClientListBrbnchesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientListBrbnchesFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientListBrbnchesFunc) History() []GitserverClientListBrbnchesFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientListBrbnchesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientListBrbnchesFuncCbll is bn object thbt describes bn
// invocbtion of method ListBrbnches on bn instbnce of MockGitserverClient.
type GitserverClientListBrbnchesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.BrbnchesOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Brbnch
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientListBrbnchesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientListBrbnchesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientListDirectoryChildrenFunc describes the behbvior when the
// ListDirectoryChildren method of the pbrent MockGitserverClient instbnce
// is invoked.
type GitserverClientListDirectoryChildrenFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)
	history     []GitserverClientListDirectoryChildrenFuncCbll
	mutex       sync.Mutex
}

// ListDirectoryChildren delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ListDirectoryChildren(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 []string) (mbp[string][]string, error) {
	r0, r1 := m.ListDirectoryChildrenFunc.nextHook()(v0, v1, v2, v3, v4)
	m.ListDirectoryChildrenFunc.bppendCbll(GitserverClientListDirectoryChildrenFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ListDirectoryChildren method of the pbrent MockGitserverClient instbnce
// is invoked bnd the hook queue is empty.
func (f *GitserverClientListDirectoryChildrenFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListDirectoryChildren method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientListDirectoryChildrenFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientListDirectoryChildrenFunc) SetDefbultReturn(r0 mbp[string][]string, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientListDirectoryChildrenFunc) PushReturn(r0 mbp[string][]string, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
		return r0, r1
	})
}

func (f *GitserverClientListDirectoryChildrenFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientListDirectoryChildrenFunc) bppendCbll(r0 GitserverClientListDirectoryChildrenFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitserverClientListDirectoryChildrenFuncCbll objects describing the
// invocbtions of this function.
func (f *GitserverClientListDirectoryChildrenFunc) History() []GitserverClientListDirectoryChildrenFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientListDirectoryChildrenFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientListDirectoryChildrenFuncCbll is bn object thbt describes
// bn invocbtion of method ListDirectoryChildren on bn instbnce of
// MockGitserverClient.
type GitserverClientListDirectoryChildrenFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string][]string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientListDirectoryChildrenFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientListDirectoryChildrenFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientListRefsFunc describes the behbvior when the ListRefs
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientListRefsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)
	hooks       []func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)
	history     []GitserverClientListRefsFuncCbll
	mutex       sync.Mutex
}

// ListRefs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ListRefs(v0 context.Context, v1 bpi.RepoNbme) ([]gitdombin.Ref, error) {
	r0, r1 := m.ListRefsFunc.nextHook()(v0, v1)
	m.ListRefsFunc.bppendCbll(GitserverClientListRefsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListRefs method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientListRefsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListRefs method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientListRefsFunc) PushHook(hook func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientListRefsFunc) SetDefbultReturn(r0 []gitdombin.Ref, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientListRefsFunc) PushReturn(r0 []gitdombin.Ref, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
		return r0, r1
	})
}

func (f *GitserverClientListRefsFunc) nextHook() func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientListRefsFunc) bppendCbll(r0 GitserverClientListRefsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientListRefsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientListRefsFunc) History() []GitserverClientListRefsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientListRefsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientListRefsFuncCbll is bn object thbt describes bn invocbtion
// of method ListRefs on bn instbnce of MockGitserverClient.
type GitserverClientListRefsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []gitdombin.Ref
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientListRefsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientListRefsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientListTbgsFunc describes the behbvior when the ListTbgs
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientListTbgsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)
	hooks       []func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)
	history     []GitserverClientListTbgsFuncCbll
	mutex       sync.Mutex
}

// ListTbgs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ListTbgs(v0 context.Context, v1 bpi.RepoNbme, v2 ...string) ([]*gitdombin.Tbg, error) {
	r0, r1 := m.ListTbgsFunc.nextHook()(v0, v1, v2...)
	m.ListTbgsFunc.bppendCbll(GitserverClientListTbgsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListTbgs method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientListTbgsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListTbgs method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientListTbgsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientListTbgsFunc) SetDefbultReturn(r0 []*gitdombin.Tbg, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientListTbgsFunc) PushReturn(r0 []*gitdombin.Tbg, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
		return r0, r1
	})
}

func (f *GitserverClientListTbgsFunc) nextHook() func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientListTbgsFunc) bppendCbll(r0 GitserverClientListTbgsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientListTbgsFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientListTbgsFunc) History() []GitserverClientListTbgsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientListTbgsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientListTbgsFuncCbll is bn object thbt describes bn invocbtion
// of method ListTbgs on bn instbnce of MockGitserverClient.
type GitserverClientListTbgsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg2 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Tbg
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GitserverClientListTbgsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg2 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientListTbgsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientLogReverseEbchFunc describes the behbvior when the
// LogReverseEbch method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientLogReverseEbchFunc struct {
	defbultHook func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error
	hooks       []func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error
	history     []GitserverClientLogReverseEbchFuncCbll
	mutex       sync.Mutex
}

// LogReverseEbch delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) LogReverseEbch(v0 context.Context, v1 string, v2 string, v3 int, v4 func(entry gitdombin.LogEntry) error) error {
	r0 := m.LogReverseEbchFunc.nextHook()(v0, v1, v2, v3, v4)
	m.LogReverseEbchFunc.bppendCbll(GitserverClientLogReverseEbchFuncCbll{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LogReverseEbch
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientLogReverseEbchFunc) SetDefbultHook(hook func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LogReverseEbch method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientLogReverseEbchFunc) PushHook(hook func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientLogReverseEbchFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientLogReverseEbchFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
		return r0
	})
}

func (f *GitserverClientLogReverseEbchFunc) nextHook() func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientLogReverseEbchFunc) bppendCbll(r0 GitserverClientLogReverseEbchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientLogReverseEbchFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientLogReverseEbchFunc) History() []GitserverClientLogReverseEbchFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientLogReverseEbchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientLogReverseEbchFuncCbll is bn object thbt describes bn
// invocbtion of method LogReverseEbch on bn instbnce of
// MockGitserverClient.
type GitserverClientLogReverseEbchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 func(entry gitdombin.LogEntry) error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientLogReverseEbchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientLogReverseEbchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitserverClientLsFilesFunc describes the behbvior when the LsFiles method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientLsFilesFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)
	history     []GitserverClientLsFilesFuncCbll
	mutex       sync.Mutex
}

// LsFiles delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) LsFiles(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 ...gitdombin.Pbthspec) ([]string, error) {
	r0, r1 := m.LsFilesFunc.nextHook()(v0, v1, v2, v3, v4...)
	m.LsFilesFunc.bppendCbll(GitserverClientLsFilesFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the LsFiles method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientLsFilesFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LsFiles method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientLsFilesFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientLsFilesFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientLsFilesFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
		return r0, r1
	})
}

func (f *GitserverClientLsFilesFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientLsFilesFunc) bppendCbll(r0 GitserverClientLsFilesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientLsFilesFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientLsFilesFunc) History() []GitserverClientLsFilesFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientLsFilesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientLsFilesFuncCbll is bn object thbt describes bn invocbtion
// of method LsFiles on bn instbnce of MockGitserverClient.
type GitserverClientLsFilesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg4 []gitdombin.Pbthspec
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GitserverClientLsFilesFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg4 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientLsFilesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientMergeBbseFunc describes the behbvior when the MergeBbse
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientMergeBbseFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)
	hooks       []func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)
	history     []GitserverClientMergeBbseFuncCbll
	mutex       sync.Mutex
}

// MergeBbse delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) MergeBbse(v0 context.Context, v1 bpi.RepoNbme, v2 bpi.CommitID, v3 bpi.CommitID) (bpi.CommitID, error) {
	r0, r1 := m.MergeBbseFunc.nextHook()(v0, v1, v2, v3)
	m.MergeBbseFunc.bppendCbll(GitserverClientMergeBbseFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MergeBbse method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientMergeBbseFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MergeBbse method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientMergeBbseFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientMergeBbseFunc) SetDefbultReturn(r0 bpi.CommitID, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientMergeBbseFunc) PushReturn(r0 bpi.CommitID, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
		return r0, r1
	})
}

func (f *GitserverClientMergeBbseFunc) nextHook() func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientMergeBbseFunc) bppendCbll(r0 GitserverClientMergeBbseFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientMergeBbseFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientMergeBbseFunc) History() []GitserverClientMergeBbseFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientMergeBbseFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientMergeBbseFuncCbll is bn object thbt describes bn
// invocbtion of method MergeBbse on bn instbnce of MockGitserverClient.
type GitserverClientMergeBbseFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bpi.CommitID
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientMergeBbseFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientMergeBbseFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientNewFileRebderFunc describes the behbvior when the
// NewFileRebder method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientNewFileRebderFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)
	history     []GitserverClientNewFileRebderFuncCbll
	mutex       sync.Mutex
}

// NewFileRebder delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) NewFileRebder(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string) (io.RebdCloser, error) {
	r0, r1 := m.NewFileRebderFunc.nextHook()(v0, v1, v2, v3, v4)
	m.NewFileRebderFunc.bppendCbll(GitserverClientNewFileRebderFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the NewFileRebder method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientNewFileRebderFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewFileRebder method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientNewFileRebderFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientNewFileRebderFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientNewFileRebderFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *GitserverClientNewFileRebderFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientNewFileRebderFunc) bppendCbll(r0 GitserverClientNewFileRebderFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientNewFileRebderFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientNewFileRebderFunc) History() []GitserverClientNewFileRebderFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientNewFileRebderFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientNewFileRebderFuncCbll is bn object thbt describes bn
// invocbtion of method NewFileRebder on bn instbnce of MockGitserverClient.
type GitserverClientNewFileRebderFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientNewFileRebderFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientNewFileRebderFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientP4ExecFunc describes the behbvior when the P4Exec method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientP4ExecFunc struct {
	defbultHook func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)
	hooks       []func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)
	history     []GitserverClientP4ExecFuncCbll
	mutex       sync.Mutex
}

// P4Exec delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) P4Exec(v0 context.Context, v1 string, v2 string, v3 string, v4 ...string) (io.RebdCloser, http.Hebder, error) {
	r0, r1, r2 := m.P4ExecFunc.nextHook()(v0, v1, v2, v3, v4...)
	m.P4ExecFunc.bppendCbll(GitserverClientP4ExecFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the P4Exec method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientP4ExecFunc) SetDefbultHook(hook func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// P4Exec method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientP4ExecFunc) PushHook(hook func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientP4ExecFunc) SetDefbultReturn(r0 io.RebdCloser, r1 http.Hebder, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientP4ExecFunc) PushReturn(r0 io.RebdCloser, r1 http.Hebder, r2 error) {
	f.PushHook(func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
		return r0, r1, r2
	})
}

func (f *GitserverClientP4ExecFunc) nextHook() func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientP4ExecFunc) bppendCbll(r0 GitserverClientP4ExecFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientP4ExecFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientP4ExecFunc) History() []GitserverClientP4ExecFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientP4ExecFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientP4ExecFuncCbll is bn object thbt describes bn invocbtion
// of method P4Exec on bn instbnce of MockGitserverClient.
type GitserverClientP4ExecFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg4 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 http.Hebder
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GitserverClientP4ExecFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg4 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientP4ExecFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// GitserverClientP4GetChbngelistFunc describes the behbvior when the
// P4GetChbngelist method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientP4GetChbngelistFunc struct {
	defbultHook func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error)
	hooks       []func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error)
	history     []GitserverClientP4GetChbngelistFuncCbll
	mutex       sync.Mutex
}

// P4GetChbngelist delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) P4GetChbngelist(v0 context.Context, v1 string, v2 gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
	r0, r1 := m.P4GetChbngelistFunc.nextHook()(v0, v1, v2)
	m.P4GetChbngelistFunc.bppendCbll(GitserverClientP4GetChbngelistFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the P4GetChbngelist
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientP4GetChbngelistFunc) SetDefbultHook(hook func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// P4GetChbngelist method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientP4GetChbngelistFunc) PushHook(hook func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientP4GetChbngelistFunc) SetDefbultReturn(r0 *protocol.PerforceChbngelist, r1 error) {
	f.SetDefbultHook(func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientP4GetChbngelistFunc) PushReturn(r0 *protocol.PerforceChbngelist, r1 error) {
	f.PushHook(func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
		return r0, r1
	})
}

func (f *GitserverClientP4GetChbngelistFunc) nextHook() func(context.Context, string, gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientP4GetChbngelistFunc) bppendCbll(r0 GitserverClientP4GetChbngelistFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientP4GetChbngelistFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientP4GetChbngelistFunc) History() []GitserverClientP4GetChbngelistFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientP4GetChbngelistFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientP4GetChbngelistFuncCbll is bn object thbt describes bn
// invocbtion of method P4GetChbngelist on bn instbnce of
// MockGitserverClient.
type GitserverClientP4GetChbngelistFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.PerforceCredentibls
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.PerforceChbngelist
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientP4GetChbngelistFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientP4GetChbngelistFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRebdDirFunc describes the behbvior when the RebdDir method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientRebdDirFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)
	history     []GitserverClientRebdDirFuncCbll
	mutex       sync.Mutex
}

// RebdDir delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RebdDir(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string, v5 bool) ([]fs.FileInfo, error) {
	r0, r1 := m.RebdDirFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.RebdDirFunc.bppendCbll(GitserverClientRebdDirFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RebdDir method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientRebdDirFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RebdDir method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientRebdDirFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRebdDirFunc) SetDefbultReturn(r0 []fs.FileInfo, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRebdDirFunc) PushReturn(r0 []fs.FileInfo, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
		return r0, r1
	})
}

func (f *GitserverClientRebdDirFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRebdDirFunc) bppendCbll(r0 GitserverClientRebdDirFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRebdDirFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientRebdDirFunc) History() []GitserverClientRebdDirFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRebdDirFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRebdDirFuncCbll is bn object thbt describes bn invocbtion
// of method RebdDir on bn instbnce of MockGitserverClient.
type GitserverClientRebdDirFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []fs.FileInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientRebdDirFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRebdDirFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRebdFileFunc describes the behbvior when the RebdFile
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientRebdFileFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)
	history     []GitserverClientRebdFileFuncCbll
	mutex       sync.Mutex
}

// RebdFile delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RebdFile(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string) ([]byte, error) {
	r0, r1 := m.RebdFileFunc.nextHook()(v0, v1, v2, v3, v4)
	m.RebdFileFunc.bppendCbll(GitserverClientRebdFileFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RebdFile method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientRebdFileFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RebdFile method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientRebdFileFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRebdFileFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRebdFileFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
		return r0, r1
	})
}

func (f *GitserverClientRebdFileFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRebdFileFunc) bppendCbll(r0 GitserverClientRebdFileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRebdFileFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientRebdFileFunc) History() []GitserverClientRebdFileFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRebdFileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRebdFileFuncCbll is bn object thbt describes bn invocbtion
// of method RebdFile on bn instbnce of MockGitserverClient.
type GitserverClientRebdFileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []byte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientRebdFileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRebdFileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRefDescriptionsFunc describes the behbvior when the
// RefDescriptions method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientRefDescriptionsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)
	history     []GitserverClientRefDescriptionsFuncCbll
	mutex       sync.Mutex
}

// RefDescriptions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RefDescriptions(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 ...string) (mbp[string][]gitdombin.RefDescription, error) {
	r0, r1 := m.RefDescriptionsFunc.nextHook()(v0, v1, v2, v3...)
	m.RefDescriptionsFunc.bppendCbll(GitserverClientRefDescriptionsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RefDescriptions
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientRefDescriptionsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RefDescriptions method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientRefDescriptionsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRefDescriptionsFunc) SetDefbultReturn(r0 mbp[string][]gitdombin.RefDescription, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRefDescriptionsFunc) PushReturn(r0 mbp[string][]gitdombin.RefDescription, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
		return r0, r1
	})
}

func (f *GitserverClientRefDescriptionsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRefDescriptionsFunc) bppendCbll(r0 GitserverClientRefDescriptionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRefDescriptionsFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientRefDescriptionsFunc) History() []GitserverClientRefDescriptionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRefDescriptionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRefDescriptionsFuncCbll is bn object thbt describes bn
// invocbtion of method RefDescriptions on bn instbnce of
// MockGitserverClient.
type GitserverClientRefDescriptionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg3 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string][]gitdombin.RefDescription
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GitserverClientRefDescriptionsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg3 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRefDescriptionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRemoveFunc describes the behbvior when the Remove method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientRemoveFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) error
	hooks       []func(context.Context, bpi.RepoNbme) error
	history     []GitserverClientRemoveFuncCbll
	mutex       sync.Mutex
}

// Remove delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Remove(v0 context.Context, v1 bpi.RepoNbme) error {
	r0 := m.RemoveFunc.nextHook()(v0, v1)
	m.RemoveFunc.bppendCbll(GitserverClientRemoveFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Remove method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientRemoveFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Remove method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientRemoveFunc) PushHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRemoveFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRemoveFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

func (f *GitserverClientRemoveFunc) nextHook() func(context.Context, bpi.RepoNbme) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRemoveFunc) bppendCbll(r0 GitserverClientRemoveFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRemoveFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientRemoveFunc) History() []GitserverClientRemoveFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRemoveFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRemoveFuncCbll is bn object thbt describes bn invocbtion
// of method Remove on bn instbnce of MockGitserverClient.
type GitserverClientRemoveFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientRemoveFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRemoveFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitserverClientRepoCloneProgressFunc describes the behbvior when the
// RepoCloneProgress method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientRepoCloneProgressFunc struct {
	defbultHook func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)
	hooks       []func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)
	history     []GitserverClientRepoCloneProgressFuncCbll
	mutex       sync.Mutex
}

// RepoCloneProgress delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RepoCloneProgress(v0 context.Context, v1 ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
	r0, r1 := m.RepoCloneProgressFunc.nextHook()(v0, v1...)
	m.RepoCloneProgressFunc.bppendCbll(GitserverClientRepoCloneProgressFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RepoCloneProgress
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientRepoCloneProgressFunc) SetDefbultHook(hook func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoCloneProgress method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientRepoCloneProgressFunc) PushHook(hook func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRepoCloneProgressFunc) SetDefbultReturn(r0 *protocol.RepoCloneProgressResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRepoCloneProgressFunc) PushReturn(r0 *protocol.RepoCloneProgressResponse, r1 error) {
	f.PushHook(func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
		return r0, r1
	})
}

func (f *GitserverClientRepoCloneProgressFunc) nextHook() func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRepoCloneProgressFunc) bppendCbll(r0 GitserverClientRepoCloneProgressFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRepoCloneProgressFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientRepoCloneProgressFunc) History() []GitserverClientRepoCloneProgressFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRepoCloneProgressFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRepoCloneProgressFuncCbll is bn object thbt describes bn
// invocbtion of method RepoCloneProgress on bn instbnce of
// MockGitserverClient.
type GitserverClientRepoCloneProgressFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoCloneProgressResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GitserverClientRepoCloneProgressFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRepoCloneProgressFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRequestRepoCloneFunc describes the behbvior when the
// RequestRepoClone method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientRequestRepoCloneFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)
	hooks       []func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)
	history     []GitserverClientRequestRepoCloneFuncCbll
	mutex       sync.Mutex
}

// RequestRepoClone delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RequestRepoClone(v0 context.Context, v1 bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
	r0, r1 := m.RequestRepoCloneFunc.nextHook()(v0, v1)
	m.RequestRepoCloneFunc.bppendCbll(GitserverClientRequestRepoCloneFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RequestRepoClone
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientRequestRepoCloneFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RequestRepoClone method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientRequestRepoCloneFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRequestRepoCloneFunc) SetDefbultReturn(r0 *protocol.RepoCloneResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRequestRepoCloneFunc) PushReturn(r0 *protocol.RepoCloneResponse, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
		return r0, r1
	})
}

func (f *GitserverClientRequestRepoCloneFunc) nextHook() func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRequestRepoCloneFunc) bppendCbll(r0 GitserverClientRequestRepoCloneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRequestRepoCloneFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientRequestRepoCloneFunc) History() []GitserverClientRequestRepoCloneFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRequestRepoCloneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRequestRepoCloneFuncCbll is bn object thbt describes bn
// invocbtion of method RequestRepoClone on bn instbnce of
// MockGitserverClient.
type GitserverClientRequestRepoCloneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoCloneResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientRequestRepoCloneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRequestRepoCloneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRequestRepoUpdbteFunc describes the behbvior when the
// RequestRepoUpdbte method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientRequestRepoUpdbteFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)
	hooks       []func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)
	history     []GitserverClientRequestRepoUpdbteFuncCbll
	mutex       sync.Mutex
}

// RequestRepoUpdbte delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RequestRepoUpdbte(v0 context.Context, v1 bpi.RepoNbme, v2 time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
	r0, r1 := m.RequestRepoUpdbteFunc.nextHook()(v0, v1, v2)
	m.RequestRepoUpdbteFunc.bppendCbll(GitserverClientRequestRepoUpdbteFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RequestRepoUpdbte
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientRequestRepoUpdbteFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RequestRepoUpdbte method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientRequestRepoUpdbteFunc) PushHook(hook func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRequestRepoUpdbteFunc) SetDefbultReturn(r0 *protocol.RepoUpdbteResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRequestRepoUpdbteFunc) PushReturn(r0 *protocol.RepoUpdbteResponse, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
		return r0, r1
	})
}

func (f *GitserverClientRequestRepoUpdbteFunc) nextHook() func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRequestRepoUpdbteFunc) bppendCbll(r0 GitserverClientRequestRepoUpdbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRequestRepoUpdbteFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientRequestRepoUpdbteFunc) History() []GitserverClientRequestRepoUpdbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRequestRepoUpdbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRequestRepoUpdbteFuncCbll is bn object thbt describes bn
// invocbtion of method RequestRepoUpdbte on bn instbnce of
// MockGitserverClient.
type GitserverClientRequestRepoUpdbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Durbtion
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoUpdbteResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientRequestRepoUpdbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRequestRepoUpdbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientResolveRevisionFunc describes the behbvior when the
// ResolveRevision method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientResolveRevisionFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error)
	hooks       []func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error)
	history     []GitserverClientResolveRevisionFuncCbll
	mutex       sync.Mutex
}

// ResolveRevision delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ResolveRevision(v0 context.Context, v1 bpi.RepoNbme, v2 string, v3 gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
	r0, r1 := m.ResolveRevisionFunc.nextHook()(v0, v1, v2, v3)
	m.ResolveRevisionFunc.bppendCbll(GitserverClientResolveRevisionFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ResolveRevision
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientResolveRevisionFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveRevision method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientResolveRevisionFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientResolveRevisionFunc) SetDefbultReturn(r0 bpi.CommitID, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientResolveRevisionFunc) PushReturn(r0 bpi.CommitID, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return r0, r1
	})
}

func (f *GitserverClientResolveRevisionFunc) nextHook() func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientResolveRevisionFunc) bppendCbll(r0 GitserverClientResolveRevisionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientResolveRevisionFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientResolveRevisionFunc) History() []GitserverClientResolveRevisionFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientResolveRevisionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientResolveRevisionFuncCbll is bn object thbt describes bn
// invocbtion of method ResolveRevision on bn instbnce of
// MockGitserverClient.
type GitserverClientResolveRevisionFuncCbll struct {
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
	Arg3 gitserver.ResolveRevisionOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bpi.CommitID
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientResolveRevisionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientResolveRevisionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientResolveRevisionsFunc describes the behbvior when the
// ResolveRevisions method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientResolveRevisionsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)
	hooks       []func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)
	history     []GitserverClientResolveRevisionsFuncCbll
	mutex       sync.Mutex
}

// ResolveRevisions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) ResolveRevisions(v0 context.Context, v1 bpi.RepoNbme, v2 []protocol.RevisionSpecifier) ([]string, error) {
	r0, r1 := m.ResolveRevisionsFunc.nextHook()(v0, v1, v2)
	m.ResolveRevisionsFunc.bppendCbll(GitserverClientResolveRevisionsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ResolveRevisions
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientResolveRevisionsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveRevisions method of the pbrent MockGitserverClient instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *GitserverClientResolveRevisionsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientResolveRevisionsFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientResolveRevisionsFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
		return r0, r1
	})
}

func (f *GitserverClientResolveRevisionsFunc) nextHook() func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientResolveRevisionsFunc) bppendCbll(r0 GitserverClientResolveRevisionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientResolveRevisionsFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientResolveRevisionsFunc) History() []GitserverClientResolveRevisionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientResolveRevisionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientResolveRevisionsFuncCbll is bn object thbt describes bn
// invocbtion of method ResolveRevisions on bn instbnce of
// MockGitserverClient.
type GitserverClientResolveRevisionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []protocol.RevisionSpecifier
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientResolveRevisionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientResolveRevisionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientRevListFunc describes the behbvior when the RevList method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientRevListFunc struct {
	defbultHook func(context.Context, string, string, func(commit string) (bool, error)) error
	hooks       []func(context.Context, string, string, func(commit string) (bool, error)) error
	history     []GitserverClientRevListFuncCbll
	mutex       sync.Mutex
}

// RevList delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RevList(v0 context.Context, v1 string, v2 string, v3 func(commit string) (bool, error)) error {
	r0 := m.RevListFunc.nextHook()(v0, v1, v2, v3)
	m.RevListFunc.bppendCbll(GitserverClientRevListFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the RevList method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientRevListFunc) SetDefbultHook(hook func(context.Context, string, string, func(commit string) (bool, error)) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RevList method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientRevListFunc) PushHook(hook func(context.Context, string, string, func(commit string) (bool, error)) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRevListFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, string, func(commit string) (bool, error)) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRevListFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, string, func(commit string) (bool, error)) error {
		return r0
	})
}

func (f *GitserverClientRevListFunc) nextHook() func(context.Context, string, string, func(commit string) (bool, error)) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRevListFunc) bppendCbll(r0 GitserverClientRevListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientRevListFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientRevListFunc) History() []GitserverClientRevListFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientRevListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRevListFuncCbll is bn object thbt describes bn invocbtion
// of method RevList on bn instbnce of MockGitserverClient.
type GitserverClientRevListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 func(commit string) (bool, error)
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientRevListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRevListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GitserverClientSebrchFunc describes the behbvior when the Sebrch method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientSebrchFunc struct {
	defbultHook func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)
	hooks       []func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)
	history     []GitserverClientSebrchFuncCbll
	mutex       sync.Mutex
}

// Sebrch delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Sebrch(v0 context.Context, v1 *protocol.SebrchRequest, v2 func([]protocol.CommitMbtch)) (bool, error) {
	r0, r1 := m.SebrchFunc.nextHook()(v0, v1, v2)
	m.SebrchFunc.bppendCbll(GitserverClientSebrchFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Sebrch method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientSebrchFunc) SetDefbultHook(hook func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Sebrch method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientSebrchFunc) PushHook(hook func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientSebrchFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientSebrchFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
		return r0, r1
	})
}

func (f *GitserverClientSebrchFunc) nextHook() func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientSebrchFunc) bppendCbll(r0 GitserverClientSebrchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientSebrchFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientSebrchFunc) History() []GitserverClientSebrchFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientSebrchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientSebrchFuncCbll is bn object thbt describes bn invocbtion
// of method Sebrch on bn instbnce of MockGitserverClient.
type GitserverClientSebrchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *protocol.SebrchRequest
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 func([]protocol.CommitMbtch)
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientSebrchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientSebrchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientStbtFunc describes the behbvior when the Stbt method of
// the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientStbtFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)
	history     []GitserverClientStbtFuncCbll
	mutex       sync.Mutex
}

// Stbt delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) Stbt(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string) (fs.FileInfo, error) {
	r0, r1 := m.StbtFunc.nextHook()(v0, v1, v2, v3, v4)
	m.StbtFunc.bppendCbll(GitserverClientStbtFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Stbt method of the
// pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientStbtFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Stbt method of the pbrent MockGitserverClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitserverClientStbtFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientStbtFunc) SetDefbultReturn(r0 fs.FileInfo, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientStbtFunc) PushReturn(r0 fs.FileInfo, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
		return r0, r1
	})
}

func (f *GitserverClientStbtFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientStbtFunc) bppendCbll(r0 GitserverClientStbtFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientStbtFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientStbtFunc) History() []GitserverClientStbtFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientStbtFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientStbtFuncCbll is bn object thbt describes bn invocbtion of
// method Stbt on bn instbnce of MockGitserverClient.
type GitserverClientStbtFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 fs.FileInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientStbtFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientStbtFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientStrebmBlbmeFileFunc describes the behbvior when the
// StrebmBlbmeFile method of the pbrent MockGitserverClient instbnce is
// invoked.
type GitserverClientStrebmBlbmeFileFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error)
	history     []GitserverClientStrebmBlbmeFileFuncCbll
	mutex       sync.Mutex
}

// StrebmBlbmeFile delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) StrebmBlbmeFile(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 *gitserver.BlbmeOptions) (gitserver.HunkRebder, error) {
	r0, r1 := m.StrebmBlbmeFileFunc.nextHook()(v0, v1, v2, v3, v4)
	m.StrebmBlbmeFileFunc.bppendCbll(GitserverClientStrebmBlbmeFileFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the StrebmBlbmeFile
// method of the pbrent MockGitserverClient instbnce is invoked bnd the hook
// queue is empty.
func (f *GitserverClientStrebmBlbmeFileFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StrebmBlbmeFile method of the pbrent MockGitserverClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GitserverClientStrebmBlbmeFileFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientStrebmBlbmeFileFunc) SetDefbultReturn(r0 gitserver.HunkRebder, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientStrebmBlbmeFileFunc) PushReturn(r0 gitserver.HunkRebder, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error) {
		return r0, r1
	})
}

func (f *GitserverClientStrebmBlbmeFileFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *gitserver.BlbmeOptions) (gitserver.HunkRebder, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientStrebmBlbmeFileFunc) bppendCbll(r0 GitserverClientStrebmBlbmeFileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientStrebmBlbmeFileFuncCbll
// objects describing the invocbtions of this function.
func (f *GitserverClientStrebmBlbmeFileFunc) History() []GitserverClientStrebmBlbmeFileFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientStrebmBlbmeFileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientStrebmBlbmeFileFuncCbll is bn object thbt describes bn
// invocbtion of method StrebmBlbmeFile on bn instbnce of
// MockGitserverClient.
type GitserverClientStrebmBlbmeFileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 *gitserver.BlbmeOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gitserver.HunkRebder
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientStrebmBlbmeFileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientStrebmBlbmeFileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientSystemInfoFunc describes the behbvior when the SystemInfo
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientSystemInfoFunc struct {
	defbultHook func(context.Context, string) (gitserver.SystemInfo, error)
	hooks       []func(context.Context, string) (gitserver.SystemInfo, error)
	history     []GitserverClientSystemInfoFuncCbll
	mutex       sync.Mutex
}

// SystemInfo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) SystemInfo(v0 context.Context, v1 string) (gitserver.SystemInfo, error) {
	r0, r1 := m.SystemInfoFunc.nextHook()(v0, v1)
	m.SystemInfoFunc.bppendCbll(GitserverClientSystemInfoFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SystemInfo method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientSystemInfoFunc) SetDefbultHook(hook func(context.Context, string) (gitserver.SystemInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SystemInfo method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientSystemInfoFunc) PushHook(hook func(context.Context, string) (gitserver.SystemInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientSystemInfoFunc) SetDefbultReturn(r0 gitserver.SystemInfo, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (gitserver.SystemInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientSystemInfoFunc) PushReturn(r0 gitserver.SystemInfo, r1 error) {
	f.PushHook(func(context.Context, string) (gitserver.SystemInfo, error) {
		return r0, r1
	})
}

func (f *GitserverClientSystemInfoFunc) nextHook() func(context.Context, string) (gitserver.SystemInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientSystemInfoFunc) bppendCbll(r0 GitserverClientSystemInfoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientSystemInfoFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientSystemInfoFunc) History() []GitserverClientSystemInfoFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientSystemInfoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientSystemInfoFuncCbll is bn object thbt describes bn
// invocbtion of method SystemInfo on bn instbnce of MockGitserverClient.
type GitserverClientSystemInfoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gitserver.SystemInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientSystemInfoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientSystemInfoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientSystemsInfoFunc describes the behbvior when the
// SystemsInfo method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientSystemsInfoFunc struct {
	defbultHook func(context.Context) ([]gitserver.SystemInfo, error)
	hooks       []func(context.Context) ([]gitserver.SystemInfo, error)
	history     []GitserverClientSystemsInfoFuncCbll
	mutex       sync.Mutex
}

// SystemsInfo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) SystemsInfo(v0 context.Context) ([]gitserver.SystemInfo, error) {
	r0, r1 := m.SystemsInfoFunc.nextHook()(v0)
	m.SystemsInfoFunc.bppendCbll(GitserverClientSystemsInfoFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SystemsInfo method
// of the pbrent MockGitserverClient instbnce is invoked bnd the hook queue
// is empty.
func (f *GitserverClientSystemsInfoFunc) SetDefbultHook(hook func(context.Context) ([]gitserver.SystemInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SystemsInfo method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientSystemsInfoFunc) PushHook(hook func(context.Context) ([]gitserver.SystemInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientSystemsInfoFunc) SetDefbultReturn(r0 []gitserver.SystemInfo, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]gitserver.SystemInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientSystemsInfoFunc) PushReturn(r0 []gitserver.SystemInfo, r1 error) {
	f.PushHook(func(context.Context) ([]gitserver.SystemInfo, error) {
		return r0, r1
	})
}

func (f *GitserverClientSystemsInfoFunc) nextHook() func(context.Context) ([]gitserver.SystemInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientSystemsInfoFunc) bppendCbll(r0 GitserverClientSystemsInfoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientSystemsInfoFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientSystemsInfoFunc) History() []GitserverClientSystemsInfoFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientSystemsInfoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientSystemsInfoFuncCbll is bn object thbt describes bn
// invocbtion of method SystemsInfo on bn instbnce of MockGitserverClient.
type GitserverClientSystemsInfoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []gitserver.SystemInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientSystemsInfoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientSystemsInfoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
