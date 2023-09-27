// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge bpi

import (
	"context"
	"io"
	"sync"

	gitserver "github.com/sourcegrbph/sourcegrbph/cmd/symbols/gitserver"
	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	gitdombin "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	types "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// MockGitserverClient is b mock implementbtion of the GitserverClient
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/symbols/gitserver) used for unit
// testing.
type MockGitserverClient struct {
	// FetchTbrFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method FetchTbr.
	FetchTbrFunc *GitserverClientFetchTbrFunc
	// GitDiffFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GitDiff.
	GitDiffFunc *GitserverClientGitDiffFunc
	// LogReverseEbchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method LogReverseEbch.
	LogReverseEbchFunc *GitserverClientLogReverseEbchFunc
	// RebdFileFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RebdFile.
	RebdFileFunc *GitserverClientRebdFileFunc
	// RevListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RevList.
	RevListFunc *GitserverClientRevListFunc
}

// NewMockGitserverClient crebtes b new mock of the GitserverClient
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		FetchTbrFunc: &GitserverClientFetchTbrFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		GitDiffFunc: &GitserverClientGitDiffFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (r0 gitserver.Chbnges, r1 error) {
				return
			},
		},
		LogReverseEbchFunc: &GitserverClientLogReverseEbchFunc{
			defbultHook: func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) (r0 error) {
				return
			},
		},
		RebdFileFunc: &GitserverClientRebdFileFunc{
			defbultHook: func(context.Context, types.RepoCommitPbth) (r0 []byte, r1 error) {
				return
			},
		},
		RevListFunc: &GitserverClientRevListFunc{
			defbultHook: func(context.Context, string, string, func(commit string) (bool, error)) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockGitserverClient crebtes b new mock of the GitserverClient
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		FetchTbrFunc: &GitserverClientFetchTbrFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.FetchTbr")
			},
		},
		GitDiffFunc: &GitserverClientGitDiffFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.GitDiff")
			},
		},
		LogReverseEbchFunc: &GitserverClientLogReverseEbchFunc{
			defbultHook: func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
				pbnic("unexpected invocbtion of MockGitserverClient.LogReverseEbch")
			},
		},
		RebdFileFunc: &GitserverClientRebdFileFunc{
			defbultHook: func(context.Context, types.RepoCommitPbth) ([]byte, error) {
				pbnic("unexpected invocbtion of MockGitserverClient.RebdFile")
			},
		},
		RevListFunc: &GitserverClientRevListFunc{
			defbultHook: func(context.Context, string, string, func(commit string) (bool, error)) error {
				pbnic("unexpected invocbtion of MockGitserverClient.RevList")
			},
		},
	}
}

// NewMockGitserverClientFrom crebtes b new mock of the MockGitserverClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGitserverClientFrom(i gitserver.GitserverClient) *MockGitserverClient {
	return &MockGitserverClient{
		FetchTbrFunc: &GitserverClientFetchTbrFunc{
			defbultHook: i.FetchTbr,
		},
		GitDiffFunc: &GitserverClientGitDiffFunc{
			defbultHook: i.GitDiff,
		},
		LogReverseEbchFunc: &GitserverClientLogReverseEbchFunc{
			defbultHook: i.LogReverseEbch,
		},
		RebdFileFunc: &GitserverClientRebdFileFunc{
			defbultHook: i.RebdFile,
		},
		RevListFunc: &GitserverClientRevListFunc{
			defbultHook: i.RevList,
		},
	}
}

// GitserverClientFetchTbrFunc describes the behbvior when the FetchTbr
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientFetchTbrFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error)
	hooks       []func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error)
	history     []GitserverClientFetchTbrFuncCbll
	mutex       sync.Mutex
}

// FetchTbr delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) FetchTbr(v0 context.Context, v1 bpi.RepoNbme, v2 bpi.CommitID, v3 []string) (io.RebdCloser, error) {
	r0, r1 := m.FetchTbrFunc.nextHook()(v0, v1, v2, v3)
	m.FetchTbrFunc.bppendCbll(GitserverClientFetchTbrFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the FetchTbr method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientFetchTbrFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// FetchTbr method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientFetchTbrFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientFetchTbrFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientFetchTbrFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *GitserverClientFetchTbrFunc) nextHook() func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientFetchTbrFunc) bppendCbll(r0 GitserverClientFetchTbrFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientFetchTbrFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientFetchTbrFunc) History() []GitserverClientFetchTbrFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientFetchTbrFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientFetchTbrFuncCbll is bn object thbt describes bn invocbtion
// of method FetchTbr on bn instbnce of MockGitserverClient.
type GitserverClientFetchTbrFuncCbll struct {
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
	Arg3 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientFetchTbrFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientFetchTbrFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitserverClientGitDiffFunc describes the behbvior when the GitDiff method
// of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientGitDiffFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error)
	hooks       []func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error)
	history     []GitserverClientGitDiffFuncCbll
	mutex       sync.Mutex
}

// GitDiff delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) GitDiff(v0 context.Context, v1 bpi.RepoNbme, v2 bpi.CommitID, v3 bpi.CommitID) (gitserver.Chbnges, error) {
	r0, r1 := m.GitDiffFunc.nextHook()(v0, v1, v2, v3)
	m.GitDiffFunc.bppendCbll(GitserverClientGitDiffFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GitDiff method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientGitDiffFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GitDiff method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientGitDiffFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientGitDiffFunc) SetDefbultReturn(r0 gitserver.Chbnges, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientGitDiffFunc) PushReturn(r0 gitserver.Chbnges, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error) {
		return r0, r1
	})
}

func (f *GitserverClientGitDiffFunc) nextHook() func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (gitserver.Chbnges, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientGitDiffFunc) bppendCbll(r0 GitserverClientGitDiffFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitserverClientGitDiffFuncCbll objects
// describing the invocbtions of this function.
func (f *GitserverClientGitDiffFunc) History() []GitserverClientGitDiffFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitserverClientGitDiffFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientGitDiffFuncCbll is bn object thbt describes bn invocbtion
// of method GitDiff on bn instbnce of MockGitserverClient.
type GitserverClientGitDiffFuncCbll struct {
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
	Result0 gitserver.Chbnges
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitserverClientGitDiffFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientGitDiffFuncCbll) Results() []interfbce{} {
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

// GitserverClientRebdFileFunc describes the behbvior when the RebdFile
// method of the pbrent MockGitserverClient instbnce is invoked.
type GitserverClientRebdFileFunc struct {
	defbultHook func(context.Context, types.RepoCommitPbth) ([]byte, error)
	hooks       []func(context.Context, types.RepoCommitPbth) ([]byte, error)
	history     []GitserverClientRebdFileFuncCbll
	mutex       sync.Mutex
}

// RebdFile delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitserverClient) RebdFile(v0 context.Context, v1 types.RepoCommitPbth) ([]byte, error) {
	r0, r1 := m.RebdFileFunc.nextHook()(v0, v1)
	m.RebdFileFunc.bppendCbll(GitserverClientRebdFileFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RebdFile method of
// the pbrent MockGitserverClient instbnce is invoked bnd the hook queue is
// empty.
func (f *GitserverClientRebdFileFunc) SetDefbultHook(hook func(context.Context, types.RepoCommitPbth) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RebdFile method of the pbrent MockGitserverClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GitserverClientRebdFileFunc) PushHook(hook func(context.Context, types.RepoCommitPbth) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitserverClientRebdFileFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, types.RepoCommitPbth) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitserverClientRebdFileFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, types.RepoCommitPbth) ([]byte, error) {
		return r0, r1
	})
}

func (f *GitserverClientRebdFileFunc) nextHook() func(context.Context, types.RepoCommitPbth) ([]byte, error) {
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
	Arg1 types.RepoCommitPbth
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
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitserverClientRebdFileFuncCbll) Results() []interfbce{} {
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
