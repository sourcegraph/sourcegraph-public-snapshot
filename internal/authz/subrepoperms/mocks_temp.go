// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge subrepoperms

import (
	"context"
	"sync"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	buthz "github.com/sourcegrbph/sourcegrbph/internbl/buthz"
)

// MockSubRepoPermissionsGetter is b mock implementbtion of the
// SubRepoPermissionsGetter interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms) used for
// unit testing.
type MockSubRepoPermissionsGetter struct {
	// GetByUserFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetByUser.
	GetByUserFunc *SubRepoPermissionsGetterGetByUserFunc
	// RepoIDSupportedFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RepoIDSupported.
	RepoIDSupportedFunc *SubRepoPermissionsGetterRepoIDSupportedFunc
	// RepoSupportedFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RepoSupported.
	RepoSupportedFunc *SubRepoPermissionsGetterRepoSupportedFunc
}

// NewMockSubRepoPermissionsGetter crebtes b new mock of the
// SubRepoPermissionsGetter interfbce. All methods return zero vblues for
// bll results, unless overwritten.
func NewMockSubRepoPermissionsGetter() *MockSubRepoPermissionsGetter {
	return &MockSubRepoPermissionsGetter{
		GetByUserFunc: &SubRepoPermissionsGetterGetByUserFunc{
			defbultHook: func(context.Context, int32) (r0 mbp[bpi.RepoNbme]buthz.SubRepoPermissions, r1 error) {
				return
			},
		},
		RepoIDSupportedFunc: &SubRepoPermissionsGetterRepoIDSupportedFunc{
			defbultHook: func(context.Context, bpi.RepoID) (r0 bool, r1 error) {
				return
			},
		},
		RepoSupportedFunc: &SubRepoPermissionsGetterRepoSupportedFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 bool, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockSubRepoPermissionsGetter crebtes b new mock of the
// SubRepoPermissionsGetter interfbce. All methods pbnic on invocbtion,
// unless overwritten.
func NewStrictMockSubRepoPermissionsGetter() *MockSubRepoPermissionsGetter {
	return &MockSubRepoPermissionsGetter{
		GetByUserFunc: &SubRepoPermissionsGetterGetByUserFunc{
			defbultHook: func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionsGetter.GetByUser")
			},
		},
		RepoIDSupportedFunc: &SubRepoPermissionsGetterRepoIDSupportedFunc{
			defbultHook: func(context.Context, bpi.RepoID) (bool, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionsGetter.RepoIDSupported")
			},
		},
		RepoSupportedFunc: &SubRepoPermissionsGetterRepoSupportedFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (bool, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionsGetter.RepoSupported")
			},
		},
	}
}

// NewMockSubRepoPermissionsGetterFrom crebtes b new mock of the
// MockSubRepoPermissionsGetter interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockSubRepoPermissionsGetterFrom(i SubRepoPermissionsGetter) *MockSubRepoPermissionsGetter {
	return &MockSubRepoPermissionsGetter{
		GetByUserFunc: &SubRepoPermissionsGetterGetByUserFunc{
			defbultHook: i.GetByUser,
		},
		RepoIDSupportedFunc: &SubRepoPermissionsGetterRepoIDSupportedFunc{
			defbultHook: i.RepoIDSupported,
		},
		RepoSupportedFunc: &SubRepoPermissionsGetterRepoSupportedFunc{
			defbultHook: i.RepoSupported,
		},
	}
}

// SubRepoPermissionsGetterGetByUserFunc describes the behbvior when the
// GetByUser method of the pbrent MockSubRepoPermissionsGetter instbnce is
// invoked.
type SubRepoPermissionsGetterGetByUserFunc struct {
	defbultHook func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error)
	hooks       []func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error)
	history     []SubRepoPermissionsGetterGetByUserFuncCbll
	mutex       sync.Mutex
}

// GetByUser delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionsGetter) GetByUser(v0 context.Context, v1 int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
	r0, r1 := m.GetByUserFunc.nextHook()(v0, v1)
	m.GetByUserFunc.bppendCbll(SubRepoPermissionsGetterGetByUserFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetByUser method of
// the pbrent MockSubRepoPermissionsGetter instbnce is invoked bnd the hook
// queue is empty.
func (f *SubRepoPermissionsGetterGetByUserFunc) SetDefbultHook(hook func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetByUser method of the pbrent MockSubRepoPermissionsGetter instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SubRepoPermissionsGetterGetByUserFunc) PushHook(hook func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionsGetterGetByUserFunc) SetDefbultReturn(r0 mbp[bpi.RepoNbme]buthz.SubRepoPermissions, r1 error) {
	f.SetDefbultHook(func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionsGetterGetByUserFunc) PushReturn(r0 mbp[bpi.RepoNbme]buthz.SubRepoPermissions, r1 error) {
	f.PushHook(func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionsGetterGetByUserFunc) nextHook() func(context.Context, int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionsGetterGetByUserFunc) bppendCbll(r0 SubRepoPermissionsGetterGetByUserFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SubRepoPermissionsGetterGetByUserFuncCbll
// objects describing the invocbtions of this function.
func (f *SubRepoPermissionsGetterGetByUserFunc) History() []SubRepoPermissionsGetterGetByUserFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionsGetterGetByUserFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionsGetterGetByUserFuncCbll is bn object thbt describes bn
// invocbtion of method GetByUser on bn instbnce of
// MockSubRepoPermissionsGetter.
type SubRepoPermissionsGetterGetByUserFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[bpi.RepoNbme]buthz.SubRepoPermissions
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SubRepoPermissionsGetterGetByUserFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionsGetterGetByUserFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SubRepoPermissionsGetterRepoIDSupportedFunc describes the behbvior when
// the RepoIDSupported method of the pbrent MockSubRepoPermissionsGetter
// instbnce is invoked.
type SubRepoPermissionsGetterRepoIDSupportedFunc struct {
	defbultHook func(context.Context, bpi.RepoID) (bool, error)
	hooks       []func(context.Context, bpi.RepoID) (bool, error)
	history     []SubRepoPermissionsGetterRepoIDSupportedFuncCbll
	mutex       sync.Mutex
}

// RepoIDSupported delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionsGetter) RepoIDSupported(v0 context.Context, v1 bpi.RepoID) (bool, error) {
	r0, r1 := m.RepoIDSupportedFunc.nextHook()(v0, v1)
	m.RepoIDSupportedFunc.bppendCbll(SubRepoPermissionsGetterRepoIDSupportedFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RepoIDSupported
// method of the pbrent MockSubRepoPermissionsGetter instbnce is invoked bnd
// the hook queue is empty.
func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoIDSupported method of the pbrent MockSubRepoPermissionsGetter
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) PushHook(hook func(context.Context, bpi.RepoID) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoID) (bool, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) nextHook() func(context.Context, bpi.RepoID) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) bppendCbll(r0 SubRepoPermissionsGetterRepoIDSupportedFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// SubRepoPermissionsGetterRepoIDSupportedFuncCbll objects describing the
// invocbtions of this function.
func (f *SubRepoPermissionsGetterRepoIDSupportedFunc) History() []SubRepoPermissionsGetterRepoIDSupportedFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionsGetterRepoIDSupportedFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionsGetterRepoIDSupportedFuncCbll is bn object thbt
// describes bn invocbtion of method RepoIDSupported on bn instbnce of
// MockSubRepoPermissionsGetter.
type SubRepoPermissionsGetterRepoIDSupportedFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SubRepoPermissionsGetterRepoIDSupportedFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionsGetterRepoIDSupportedFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SubRepoPermissionsGetterRepoSupportedFunc describes the behbvior when the
// RepoSupported method of the pbrent MockSubRepoPermissionsGetter instbnce
// is invoked.
type SubRepoPermissionsGetterRepoSupportedFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (bool, error)
	hooks       []func(context.Context, bpi.RepoNbme) (bool, error)
	history     []SubRepoPermissionsGetterRepoSupportedFuncCbll
	mutex       sync.Mutex
}

// RepoSupported delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionsGetter) RepoSupported(v0 context.Context, v1 bpi.RepoNbme) (bool, error) {
	r0, r1 := m.RepoSupportedFunc.nextHook()(v0, v1)
	m.RepoSupportedFunc.bppendCbll(SubRepoPermissionsGetterRepoSupportedFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RepoSupported method
// of the pbrent MockSubRepoPermissionsGetter instbnce is invoked bnd the
// hook queue is empty.
func (f *SubRepoPermissionsGetterRepoSupportedFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoSupported method of the pbrent MockSubRepoPermissionsGetter instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SubRepoPermissionsGetterRepoSupportedFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionsGetterRepoSupportedFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionsGetterRepoSupportedFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (bool, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionsGetterRepoSupportedFunc) nextHook() func(context.Context, bpi.RepoNbme) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionsGetterRepoSupportedFunc) bppendCbll(r0 SubRepoPermissionsGetterRepoSupportedFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// SubRepoPermissionsGetterRepoSupportedFuncCbll objects describing the
// invocbtions of this function.
func (f *SubRepoPermissionsGetterRepoSupportedFunc) History() []SubRepoPermissionsGetterRepoSupportedFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionsGetterRepoSupportedFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionsGetterRepoSupportedFuncCbll is bn object thbt describes
// bn invocbtion of method RepoSupported on bn instbnce of
// MockSubRepoPermissionsGetter.
type SubRepoPermissionsGetterRepoSupportedFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SubRepoPermissionsGetterRepoSupportedFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionsGetterRepoSupportedFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
