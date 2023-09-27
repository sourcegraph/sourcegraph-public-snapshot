// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge server

import (
	"context"
	"os/exec"
	"sync"

	common "github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	vcs "github.com/sourcegrbph/sourcegrbph/internbl/vcs"
)

// MockVCSSyncer is b mock implementbtion of the VCSSyncer interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server) used
// for unit testing.
type MockVCSSyncer struct {
	// CloneCommbndFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CloneCommbnd.
	CloneCommbndFunc *VCSSyncerCloneCommbndFunc
	// FetchFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Fetch.
	FetchFunc *VCSSyncerFetchFunc
	// IsClonebbleFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method IsClonebble.
	IsClonebbleFunc *VCSSyncerIsClonebbleFunc
	// RemoteShowCommbndFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RemoteShowCommbnd.
	RemoteShowCommbndFunc *VCSSyncerRemoteShowCommbndFunc
	// TypeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Type.
	TypeFunc *VCSSyncerTypeFunc
}

// NewMockVCSSyncer crebtes b new mock of the VCSSyncer interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockVCSSyncer() *MockVCSSyncer {
	return &MockVCSSyncer{
		CloneCommbndFunc: &VCSSyncerCloneCommbndFunc{
			defbultHook: func(context.Context, *vcs.URL, string) (r0 *exec.Cmd, r1 error) {
				return
			},
		},
		FetchFunc: &VCSSyncerFetchFunc{
			defbultHook: func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) (r0 []byte, r1 error) {
				return
			},
		},
		IsClonebbleFunc: &VCSSyncerIsClonebbleFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, *vcs.URL) (r0 error) {
				return
			},
		},
		RemoteShowCommbndFunc: &VCSSyncerRemoteShowCommbndFunc{
			defbultHook: func(context.Context, *vcs.URL) (r0 *exec.Cmd, r1 error) {
				return
			},
		},
		TypeFunc: &VCSSyncerTypeFunc{
			defbultHook: func() (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockVCSSyncer crebtes b new mock of the VCSSyncer interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockVCSSyncer() *MockVCSSyncer {
	return &MockVCSSyncer{
		CloneCommbndFunc: &VCSSyncerCloneCommbndFunc{
			defbultHook: func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
				pbnic("unexpected invocbtion of MockVCSSyncer.CloneCommbnd")
			},
		},
		FetchFunc: &VCSSyncerFetchFunc{
			defbultHook: func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error) {
				pbnic("unexpected invocbtion of MockVCSSyncer.Fetch")
			},
		},
		IsClonebbleFunc: &VCSSyncerIsClonebbleFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, *vcs.URL) error {
				pbnic("unexpected invocbtion of MockVCSSyncer.IsClonebble")
			},
		},
		RemoteShowCommbndFunc: &VCSSyncerRemoteShowCommbndFunc{
			defbultHook: func(context.Context, *vcs.URL) (*exec.Cmd, error) {
				pbnic("unexpected invocbtion of MockVCSSyncer.RemoteShowCommbnd")
			},
		},
		TypeFunc: &VCSSyncerTypeFunc{
			defbultHook: func() string {
				pbnic("unexpected invocbtion of MockVCSSyncer.Type")
			},
		},
	}
}

// NewMockVCSSyncerFrom crebtes b new mock of the MockVCSSyncer interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockVCSSyncerFrom(i VCSSyncer) *MockVCSSyncer {
	return &MockVCSSyncer{
		CloneCommbndFunc: &VCSSyncerCloneCommbndFunc{
			defbultHook: i.CloneCommbnd,
		},
		FetchFunc: &VCSSyncerFetchFunc{
			defbultHook: i.Fetch,
		},
		IsClonebbleFunc: &VCSSyncerIsClonebbleFunc{
			defbultHook: i.IsClonebble,
		},
		RemoteShowCommbndFunc: &VCSSyncerRemoteShowCommbndFunc{
			defbultHook: i.RemoteShowCommbnd,
		},
		TypeFunc: &VCSSyncerTypeFunc{
			defbultHook: i.Type,
		},
	}
}

// VCSSyncerCloneCommbndFunc describes the behbvior when the CloneCommbnd
// method of the pbrent MockVCSSyncer instbnce is invoked.
type VCSSyncerCloneCommbndFunc struct {
	defbultHook func(context.Context, *vcs.URL, string) (*exec.Cmd, error)
	hooks       []func(context.Context, *vcs.URL, string) (*exec.Cmd, error)
	history     []VCSSyncerCloneCommbndFuncCbll
	mutex       sync.Mutex
}

// CloneCommbnd delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockVCSSyncer) CloneCommbnd(v0 context.Context, v1 *vcs.URL, v2 string) (*exec.Cmd, error) {
	r0, r1 := m.CloneCommbndFunc.nextHook()(v0, v1, v2)
	m.CloneCommbndFunc.bppendCbll(VCSSyncerCloneCommbndFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CloneCommbnd method
// of the pbrent MockVCSSyncer instbnce is invoked bnd the hook queue is
// empty.
func (f *VCSSyncerCloneCommbndFunc) SetDefbultHook(hook func(context.Context, *vcs.URL, string) (*exec.Cmd, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CloneCommbnd method of the pbrent MockVCSSyncer instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *VCSSyncerCloneCommbndFunc) PushHook(hook func(context.Context, *vcs.URL, string) (*exec.Cmd, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *VCSSyncerCloneCommbndFunc) SetDefbultReturn(r0 *exec.Cmd, r1 error) {
	f.SetDefbultHook(func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *VCSSyncerCloneCommbndFunc) PushReturn(r0 *exec.Cmd, r1 error) {
	f.PushHook(func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
		return r0, r1
	})
}

func (f *VCSSyncerCloneCommbndFunc) nextHook() func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerCloneCommbndFunc) bppendCbll(r0 VCSSyncerCloneCommbndFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of VCSSyncerCloneCommbndFuncCbll objects
// describing the invocbtions of this function.
func (f *VCSSyncerCloneCommbndFunc) History() []VCSSyncerCloneCommbndFuncCbll {
	f.mutex.Lock()
	history := mbke([]VCSSyncerCloneCommbndFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerCloneCommbndFuncCbll is bn object thbt describes bn invocbtion
// of method CloneCommbnd on bn instbnce of MockVCSSyncer.
type VCSSyncerCloneCommbndFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *vcs.URL
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *exec.Cmd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c VCSSyncerCloneCommbndFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c VCSSyncerCloneCommbndFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// VCSSyncerFetchFunc describes the behbvior when the Fetch method of the
// pbrent MockVCSSyncer instbnce is invoked.
type VCSSyncerFetchFunc struct {
	defbultHook func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error)
	hooks       []func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error)
	history     []VCSSyncerFetchFuncCbll
	mutex       sync.Mutex
}

// Fetch delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockVCSSyncer) Fetch(v0 context.Context, v1 *vcs.URL, v2 bpi.RepoNbme, v3 common.GitDir, v4 string) ([]byte, error) {
	r0, r1 := m.FetchFunc.nextHook()(v0, v1, v2, v3, v4)
	m.FetchFunc.bppendCbll(VCSSyncerFetchFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Fetch method of the
// pbrent MockVCSSyncer instbnce is invoked bnd the hook queue is empty.
func (f *VCSSyncerFetchFunc) SetDefbultHook(hook func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Fetch method of the pbrent MockVCSSyncer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *VCSSyncerFetchFunc) PushHook(hook func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *VCSSyncerFetchFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *VCSSyncerFetchFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error) {
		return r0, r1
	})
}

func (f *VCSSyncerFetchFunc) nextHook() func(context.Context, *vcs.URL, bpi.RepoNbme, common.GitDir, string) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerFetchFunc) bppendCbll(r0 VCSSyncerFetchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of VCSSyncerFetchFuncCbll objects describing
// the invocbtions of this function.
func (f *VCSSyncerFetchFunc) History() []VCSSyncerFetchFuncCbll {
	f.mutex.Lock()
	history := mbke([]VCSSyncerFetchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerFetchFuncCbll is bn object thbt describes bn invocbtion of
// method Fetch on bn instbnce of MockVCSSyncer.
type VCSSyncerFetchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *vcs.URL
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 common.GitDir
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
func (c VCSSyncerFetchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c VCSSyncerFetchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// VCSSyncerIsClonebbleFunc describes the behbvior when the IsClonebble
// method of the pbrent MockVCSSyncer instbnce is invoked.
type VCSSyncerIsClonebbleFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, *vcs.URL) error
	hooks       []func(context.Context, bpi.RepoNbme, *vcs.URL) error
	history     []VCSSyncerIsClonebbleFuncCbll
	mutex       sync.Mutex
}

// IsClonebble delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockVCSSyncer) IsClonebble(v0 context.Context, v1 bpi.RepoNbme, v2 *vcs.URL) error {
	r0 := m.IsClonebbleFunc.nextHook()(v0, v1, v2)
	m.IsClonebbleFunc.bppendCbll(VCSSyncerIsClonebbleFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the IsClonebble method
// of the pbrent MockVCSSyncer instbnce is invoked bnd the hook queue is
// empty.
func (f *VCSSyncerIsClonebbleFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, *vcs.URL) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IsClonebble method of the pbrent MockVCSSyncer instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *VCSSyncerIsClonebbleFunc) PushHook(hook func(context.Context, bpi.RepoNbme, *vcs.URL) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *VCSSyncerIsClonebbleFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, *vcs.URL) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *VCSSyncerIsClonebbleFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, *vcs.URL) error {
		return r0
	})
}

func (f *VCSSyncerIsClonebbleFunc) nextHook() func(context.Context, bpi.RepoNbme, *vcs.URL) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerIsClonebbleFunc) bppendCbll(r0 VCSSyncerIsClonebbleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of VCSSyncerIsClonebbleFuncCbll objects
// describing the invocbtions of this function.
func (f *VCSSyncerIsClonebbleFunc) History() []VCSSyncerIsClonebbleFuncCbll {
	f.mutex.Lock()
	history := mbke([]VCSSyncerIsClonebbleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerIsClonebbleFuncCbll is bn object thbt describes bn invocbtion of
// method IsClonebble on bn instbnce of MockVCSSyncer.
type VCSSyncerIsClonebbleFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *vcs.URL
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c VCSSyncerIsClonebbleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c VCSSyncerIsClonebbleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// VCSSyncerRemoteShowCommbndFunc describes the behbvior when the
// RemoteShowCommbnd method of the pbrent MockVCSSyncer instbnce is invoked.
type VCSSyncerRemoteShowCommbndFunc struct {
	defbultHook func(context.Context, *vcs.URL) (*exec.Cmd, error)
	hooks       []func(context.Context, *vcs.URL) (*exec.Cmd, error)
	history     []VCSSyncerRemoteShowCommbndFuncCbll
	mutex       sync.Mutex
}

// RemoteShowCommbnd delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockVCSSyncer) RemoteShowCommbnd(v0 context.Context, v1 *vcs.URL) (*exec.Cmd, error) {
	r0, r1 := m.RemoteShowCommbndFunc.nextHook()(v0, v1)
	m.RemoteShowCommbndFunc.bppendCbll(VCSSyncerRemoteShowCommbndFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RemoteShowCommbnd
// method of the pbrent MockVCSSyncer instbnce is invoked bnd the hook queue
// is empty.
func (f *VCSSyncerRemoteShowCommbndFunc) SetDefbultHook(hook func(context.Context, *vcs.URL) (*exec.Cmd, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RemoteShowCommbnd method of the pbrent MockVCSSyncer instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *VCSSyncerRemoteShowCommbndFunc) PushHook(hook func(context.Context, *vcs.URL) (*exec.Cmd, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *VCSSyncerRemoteShowCommbndFunc) SetDefbultReturn(r0 *exec.Cmd, r1 error) {
	f.SetDefbultHook(func(context.Context, *vcs.URL) (*exec.Cmd, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *VCSSyncerRemoteShowCommbndFunc) PushReturn(r0 *exec.Cmd, r1 error) {
	f.PushHook(func(context.Context, *vcs.URL) (*exec.Cmd, error) {
		return r0, r1
	})
}

func (f *VCSSyncerRemoteShowCommbndFunc) nextHook() func(context.Context, *vcs.URL) (*exec.Cmd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerRemoteShowCommbndFunc) bppendCbll(r0 VCSSyncerRemoteShowCommbndFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of VCSSyncerRemoteShowCommbndFuncCbll objects
// describing the invocbtions of this function.
func (f *VCSSyncerRemoteShowCommbndFunc) History() []VCSSyncerRemoteShowCommbndFuncCbll {
	f.mutex.Lock()
	history := mbke([]VCSSyncerRemoteShowCommbndFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerRemoteShowCommbndFuncCbll is bn object thbt describes bn
// invocbtion of method RemoteShowCommbnd on bn instbnce of MockVCSSyncer.
type VCSSyncerRemoteShowCommbndFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *vcs.URL
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *exec.Cmd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c VCSSyncerRemoteShowCommbndFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c VCSSyncerRemoteShowCommbndFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// VCSSyncerTypeFunc describes the behbvior when the Type method of the
// pbrent MockVCSSyncer instbnce is invoked.
type VCSSyncerTypeFunc struct {
	defbultHook func() string
	hooks       []func() string
	history     []VCSSyncerTypeFuncCbll
	mutex       sync.Mutex
}

// Type delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockVCSSyncer) Type() string {
	r0 := m.TypeFunc.nextHook()()
	m.TypeFunc.bppendCbll(VCSSyncerTypeFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Type method of the
// pbrent MockVCSSyncer instbnce is invoked bnd the hook queue is empty.
func (f *VCSSyncerTypeFunc) SetDefbultHook(hook func() string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Type method of the pbrent MockVCSSyncer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *VCSSyncerTypeFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *VCSSyncerTypeFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func() string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *VCSSyncerTypeFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *VCSSyncerTypeFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerTypeFunc) bppendCbll(r0 VCSSyncerTypeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of VCSSyncerTypeFuncCbll objects describing
// the invocbtions of this function.
func (f *VCSSyncerTypeFunc) History() []VCSSyncerTypeFuncCbll {
	f.mutex.Lock()
	history := mbke([]VCSSyncerTypeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerTypeFuncCbll is bn object thbt describes bn invocbtion of method
// Type on bn instbnce of MockVCSSyncer.
type VCSSyncerTypeFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c VCSSyncerTypeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c VCSSyncerTypeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
