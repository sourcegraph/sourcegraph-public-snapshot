// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge inference

import (
	"context"
	"io"
	"sync"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	gitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	gitdombin "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	lubsbndbox "github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox"
)

// MockGitService is b mock implementbtion of the GitService interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference)
// used for unit testing.
type MockGitService struct {
	// ArchiveFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Archive.
	ArchiveFunc *GitServiceArchiveFunc
	// LsFilesFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LsFiles.
	LsFilesFunc *GitServiceLsFilesFunc
}

// NewMockGitService crebtes b new mock of the GitService interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockGitService() *MockGitService {
	return &MockGitService{
		ArchiveFunc: &GitServiceArchiveFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		LsFilesFunc: &GitServiceLsFilesFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) (r0 []string, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockGitService crebtes b new mock of the GitService interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGitService() *MockGitService {
	return &MockGitService{
		ArchiveFunc: &GitServiceArchiveFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockGitService.Archive")
			},
		},
		LsFilesFunc: &GitServiceLsFilesFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error) {
				pbnic("unexpected invocbtion of MockGitService.LsFiles")
			},
		},
	}
}

// NewMockGitServiceFrom crebtes b new mock of the MockGitService interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockGitServiceFrom(i GitService) *MockGitService {
	return &MockGitService{
		ArchiveFunc: &GitServiceArchiveFunc{
			defbultHook: i.Archive,
		},
		LsFilesFunc: &GitServiceLsFilesFunc{
			defbultHook: i.LsFiles,
		},
	}
}

// GitServiceArchiveFunc describes the behbvior when the Archive method of
// the pbrent MockGitService instbnce is invoked.
type GitServiceArchiveFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)
	hooks       []func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)
	history     []GitServiceArchiveFuncCbll
	mutex       sync.Mutex
}

// Archive delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitService) Archive(v0 context.Context, v1 bpi.RepoNbme, v2 gitserver.ArchiveOptions) (io.RebdCloser, error) {
	r0, r1 := m.ArchiveFunc.nextHook()(v0, v1, v2)
	m.ArchiveFunc.bppendCbll(GitServiceArchiveFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Archive method of
// the pbrent MockGitService instbnce is invoked bnd the hook queue is
// empty.
func (f *GitServiceArchiveFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Archive method of the pbrent MockGitService instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitServiceArchiveFunc) PushHook(hook func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitServiceArchiveFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitServiceArchiveFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *GitServiceArchiveFunc) nextHook() func(context.Context, bpi.RepoNbme, gitserver.ArchiveOptions) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitServiceArchiveFunc) bppendCbll(r0 GitServiceArchiveFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitServiceArchiveFuncCbll objects
// describing the invocbtions of this function.
func (f *GitServiceArchiveFunc) History() []GitServiceArchiveFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitServiceArchiveFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitServiceArchiveFuncCbll is bn object thbt describes bn invocbtion of
// method Archive on bn instbnce of MockGitService.
type GitServiceArchiveFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 gitserver.ArchiveOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitServiceArchiveFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitServiceArchiveFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GitServiceLsFilesFunc describes the behbvior when the LsFiles method of
// the pbrent MockGitService instbnce is invoked.
type GitServiceLsFilesFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error)
	hooks       []func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error)
	history     []GitServiceLsFilesFuncCbll
	mutex       sync.Mutex
}

// LsFiles delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitService) LsFiles(v0 context.Context, v1 bpi.RepoNbme, v2 string, v3 ...gitdombin.Pbthspec) ([]string, error) {
	r0, r1 := m.LsFilesFunc.nextHook()(v0, v1, v2, v3...)
	m.LsFilesFunc.bppendCbll(GitServiceLsFilesFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the LsFiles method of
// the pbrent MockGitService instbnce is invoked bnd the hook queue is
// empty.
func (f *GitServiceLsFilesFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LsFiles method of the pbrent MockGitService instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GitServiceLsFilesFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitServiceLsFilesFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitServiceLsFilesFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error) {
		return r0, r1
	})
}

func (f *GitServiceLsFilesFunc) nextHook() func(context.Context, bpi.RepoNbme, string, ...gitdombin.Pbthspec) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitServiceLsFilesFunc) bppendCbll(r0 GitServiceLsFilesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GitServiceLsFilesFuncCbll objects
// describing the invocbtions of this function.
func (f *GitServiceLsFilesFunc) History() []GitServiceLsFilesFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitServiceLsFilesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitServiceLsFilesFuncCbll is bn object thbt describes bn invocbtion of
// method LsFiles on bn instbnce of MockGitService.
type GitServiceLsFilesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg3 []gitdombin.Pbthspec
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
func (c GitServiceLsFilesFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg3 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitServiceLsFilesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockSbndboxService is b mock implementbtion of the SbndboxService
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference)
// used for unit testing.
type MockSbndboxService struct {
	// CrebteSbndboxFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteSbndbox.
	CrebteSbndboxFunc *SbndboxServiceCrebteSbndboxFunc
}

// NewMockSbndboxService crebtes b new mock of the SbndboxService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockSbndboxService() *MockSbndboxService {
	return &MockSbndboxService{
		CrebteSbndboxFunc: &SbndboxServiceCrebteSbndboxFunc{
			defbultHook: func(context.Context, lubsbndbox.CrebteOptions) (r0 *lubsbndbox.Sbndbox, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockSbndboxService crebtes b new mock of the SbndboxService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockSbndboxService() *MockSbndboxService {
	return &MockSbndboxService{
		CrebteSbndboxFunc: &SbndboxServiceCrebteSbndboxFunc{
			defbultHook: func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error) {
				pbnic("unexpected invocbtion of MockSbndboxService.CrebteSbndbox")
			},
		},
	}
}

// NewMockSbndboxServiceFrom crebtes b new mock of the MockSbndboxService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockSbndboxServiceFrom(i SbndboxService) *MockSbndboxService {
	return &MockSbndboxService{
		CrebteSbndboxFunc: &SbndboxServiceCrebteSbndboxFunc{
			defbultHook: i.CrebteSbndbox,
		},
	}
}

// SbndboxServiceCrebteSbndboxFunc describes the behbvior when the
// CrebteSbndbox method of the pbrent MockSbndboxService instbnce is
// invoked.
type SbndboxServiceCrebteSbndboxFunc struct {
	defbultHook func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error)
	hooks       []func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error)
	history     []SbndboxServiceCrebteSbndboxFuncCbll
	mutex       sync.Mutex
}

// CrebteSbndbox delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSbndboxService) CrebteSbndbox(v0 context.Context, v1 lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error) {
	r0, r1 := m.CrebteSbndboxFunc.nextHook()(v0, v1)
	m.CrebteSbndboxFunc.bppendCbll(SbndboxServiceCrebteSbndboxFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CrebteSbndbox method
// of the pbrent MockSbndboxService instbnce is invoked bnd the hook queue
// is empty.
func (f *SbndboxServiceCrebteSbndboxFunc) SetDefbultHook(hook func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteSbndbox method of the pbrent MockSbndboxService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SbndboxServiceCrebteSbndboxFunc) PushHook(hook func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SbndboxServiceCrebteSbndboxFunc) SetDefbultReturn(r0 *lubsbndbox.Sbndbox, r1 error) {
	f.SetDefbultHook(func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SbndboxServiceCrebteSbndboxFunc) PushReturn(r0 *lubsbndbox.Sbndbox, r1 error) {
	f.PushHook(func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error) {
		return r0, r1
	})
}

func (f *SbndboxServiceCrebteSbndboxFunc) nextHook() func(context.Context, lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SbndboxServiceCrebteSbndboxFunc) bppendCbll(r0 SbndboxServiceCrebteSbndboxFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SbndboxServiceCrebteSbndboxFuncCbll objects
// describing the invocbtions of this function.
func (f *SbndboxServiceCrebteSbndboxFunc) History() []SbndboxServiceCrebteSbndboxFuncCbll {
	f.mutex.Lock()
	history := mbke([]SbndboxServiceCrebteSbndboxFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SbndboxServiceCrebteSbndboxFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteSbndbox on bn instbnce of MockSbndboxService.
type SbndboxServiceCrebteSbndboxFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 lubsbndbox.CrebteOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *lubsbndbox.Sbndbox
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SbndboxServiceCrebteSbndboxFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SbndboxServiceCrebteSbndboxFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
