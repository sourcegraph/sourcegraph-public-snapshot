// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge buthz

import (
	"context"
	"sync"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// MockSubRepoPermissionChecker is b mock implementbtion of the
// SubRepoPermissionChecker interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/buthz) used for unit testing.
type MockSubRepoPermissionChecker struct {
	// EnbbledFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Enbbled.
	EnbbledFunc *SubRepoPermissionCheckerEnbbledFunc
	// EnbbledForRepoFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method EnbbledForRepo.
	EnbbledForRepoFunc *SubRepoPermissionCheckerEnbbledForRepoFunc
	// EnbbledForRepoIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method EnbbledForRepoID.
	EnbbledForRepoIDFunc *SubRepoPermissionCheckerEnbbledForRepoIDFunc
	// FilePermissionsFuncFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method FilePermissionsFunc.
	FilePermissionsFuncFunc *SubRepoPermissionCheckerFilePermissionsFuncFunc
	// PermissionsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method Permissions.
	PermissionsFunc *SubRepoPermissionCheckerPermissionsFunc
}

// NewMockSubRepoPermissionChecker crebtes b new mock of the
// SubRepoPermissionChecker interfbce. All methods return zero vblues for
// bll results, unless overwritten.
func NewMockSubRepoPermissionChecker() *MockSubRepoPermissionChecker {
	return &MockSubRepoPermissionChecker{
		EnbbledFunc: &SubRepoPermissionCheckerEnbbledFunc{
			defbultHook: func() (r0 bool) {
				return
			},
		},
		EnbbledForRepoFunc: &SubRepoPermissionCheckerEnbbledForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 bool, r1 error) {
				return
			},
		},
		EnbbledForRepoIDFunc: &SubRepoPermissionCheckerEnbbledForRepoIDFunc{
			defbultHook: func(context.Context, bpi.RepoID) (r0 bool, r1 error) {
				return
			},
		},
		FilePermissionsFuncFunc: &SubRepoPermissionCheckerFilePermissionsFuncFunc{
			defbultHook: func(context.Context, int32, bpi.RepoNbme) (r0 FilePermissionFunc, r1 error) {
				return
			},
		},
		PermissionsFunc: &SubRepoPermissionCheckerPermissionsFunc{
			defbultHook: func(context.Context, int32, RepoContent) (r0 Perms, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockSubRepoPermissionChecker crebtes b new mock of the
// SubRepoPermissionChecker interfbce. All methods pbnic on invocbtion,
// unless overwritten.
func NewStrictMockSubRepoPermissionChecker() *MockSubRepoPermissionChecker {
	return &MockSubRepoPermissionChecker{
		EnbbledFunc: &SubRepoPermissionCheckerEnbbledFunc{
			defbultHook: func() bool {
				pbnic("unexpected invocbtion of MockSubRepoPermissionChecker.Enbbled")
			},
		},
		EnbbledForRepoFunc: &SubRepoPermissionCheckerEnbbledForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (bool, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionChecker.EnbbledForRepo")
			},
		},
		EnbbledForRepoIDFunc: &SubRepoPermissionCheckerEnbbledForRepoIDFunc{
			defbultHook: func(context.Context, bpi.RepoID) (bool, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionChecker.EnbbledForRepoID")
			},
		},
		FilePermissionsFuncFunc: &SubRepoPermissionCheckerFilePermissionsFuncFunc{
			defbultHook: func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionChecker.FilePermissionsFunc")
			},
		},
		PermissionsFunc: &SubRepoPermissionCheckerPermissionsFunc{
			defbultHook: func(context.Context, int32, RepoContent) (Perms, error) {
				pbnic("unexpected invocbtion of MockSubRepoPermissionChecker.Permissions")
			},
		},
	}
}

// NewMockSubRepoPermissionCheckerFrom crebtes b new mock of the
// MockSubRepoPermissionChecker interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockSubRepoPermissionCheckerFrom(i SubRepoPermissionChecker) *MockSubRepoPermissionChecker {
	return &MockSubRepoPermissionChecker{
		EnbbledFunc: &SubRepoPermissionCheckerEnbbledFunc{
			defbultHook: i.Enbbled,
		},
		EnbbledForRepoFunc: &SubRepoPermissionCheckerEnbbledForRepoFunc{
			defbultHook: i.EnbbledForRepo,
		},
		EnbbledForRepoIDFunc: &SubRepoPermissionCheckerEnbbledForRepoIDFunc{
			defbultHook: i.EnbbledForRepoID,
		},
		FilePermissionsFuncFunc: &SubRepoPermissionCheckerFilePermissionsFuncFunc{
			defbultHook: i.FilePermissionsFunc,
		},
		PermissionsFunc: &SubRepoPermissionCheckerPermissionsFunc{
			defbultHook: i.Permissions,
		},
	}
}

// SubRepoPermissionCheckerEnbbledFunc describes the behbvior when the
// Enbbled method of the pbrent MockSubRepoPermissionChecker instbnce is
// invoked.
type SubRepoPermissionCheckerEnbbledFunc struct {
	defbultHook func() bool
	hooks       []func() bool
	history     []SubRepoPermissionCheckerEnbbledFuncCbll
	mutex       sync.Mutex
}

// Enbbled delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionChecker) Enbbled() bool {
	r0 := m.EnbbledFunc.nextHook()()
	m.EnbbledFunc.bppendCbll(SubRepoPermissionCheckerEnbbledFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Enbbled method of
// the pbrent MockSubRepoPermissionChecker instbnce is invoked bnd the hook
// queue is empty.
func (f *SubRepoPermissionCheckerEnbbledFunc) SetDefbultHook(hook func() bool) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Enbbled method of the pbrent MockSubRepoPermissionChecker instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SubRepoPermissionCheckerEnbbledFunc) PushHook(hook func() bool) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionCheckerEnbbledFunc) SetDefbultReturn(r0 bool) {
	f.SetDefbultHook(func() bool {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionCheckerEnbbledFunc) PushReturn(r0 bool) {
	f.PushHook(func() bool {
		return r0
	})
}

func (f *SubRepoPermissionCheckerEnbbledFunc) nextHook() func() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionCheckerEnbbledFunc) bppendCbll(r0 SubRepoPermissionCheckerEnbbledFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SubRepoPermissionCheckerEnbbledFuncCbll
// objects describing the invocbtions of this function.
func (f *SubRepoPermissionCheckerEnbbledFunc) History() []SubRepoPermissionCheckerEnbbledFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionCheckerEnbbledFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionCheckerEnbbledFuncCbll is bn object thbt describes bn
// invocbtion of method Enbbled on bn instbnce of
// MockSubRepoPermissionChecker.
type SubRepoPermissionCheckerEnbbledFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SubRepoPermissionCheckerEnbbledFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionCheckerEnbbledFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SubRepoPermissionCheckerEnbbledForRepoFunc describes the behbvior when
// the EnbbledForRepo method of the pbrent MockSubRepoPermissionChecker
// instbnce is invoked.
type SubRepoPermissionCheckerEnbbledForRepoFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (bool, error)
	hooks       []func(context.Context, bpi.RepoNbme) (bool, error)
	history     []SubRepoPermissionCheckerEnbbledForRepoFuncCbll
	mutex       sync.Mutex
}

// EnbbledForRepo delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionChecker) EnbbledForRepo(v0 context.Context, v1 bpi.RepoNbme) (bool, error) {
	r0, r1 := m.EnbbledForRepoFunc.nextHook()(v0, v1)
	m.EnbbledForRepoFunc.bppendCbll(SubRepoPermissionCheckerEnbbledForRepoFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the EnbbledForRepo
// method of the pbrent MockSubRepoPermissionChecker instbnce is invoked bnd
// the hook queue is empty.
func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// EnbbledForRepo method of the pbrent MockSubRepoPermissionChecker instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (bool, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) nextHook() func(context.Context, bpi.RepoNbme) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) bppendCbll(r0 SubRepoPermissionCheckerEnbbledForRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// SubRepoPermissionCheckerEnbbledForRepoFuncCbll objects describing the
// invocbtions of this function.
func (f *SubRepoPermissionCheckerEnbbledForRepoFunc) History() []SubRepoPermissionCheckerEnbbledForRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionCheckerEnbbledForRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionCheckerEnbbledForRepoFuncCbll is bn object thbt
// describes bn invocbtion of method EnbbledForRepo on bn instbnce of
// MockSubRepoPermissionChecker.
type SubRepoPermissionCheckerEnbbledForRepoFuncCbll struct {
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
func (c SubRepoPermissionCheckerEnbbledForRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionCheckerEnbbledForRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SubRepoPermissionCheckerEnbbledForRepoIDFunc describes the behbvior when
// the EnbbledForRepoID method of the pbrent MockSubRepoPermissionChecker
// instbnce is invoked.
type SubRepoPermissionCheckerEnbbledForRepoIDFunc struct {
	defbultHook func(context.Context, bpi.RepoID) (bool, error)
	hooks       []func(context.Context, bpi.RepoID) (bool, error)
	history     []SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll
	mutex       sync.Mutex
}

// EnbbledForRepoID delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionChecker) EnbbledForRepoID(v0 context.Context, v1 bpi.RepoID) (bool, error) {
	r0, r1 := m.EnbbledForRepoIDFunc.nextHook()(v0, v1)
	m.EnbbledForRepoIDFunc.bppendCbll(SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the EnbbledForRepoID
// method of the pbrent MockSubRepoPermissionChecker instbnce is invoked bnd
// the hook queue is empty.
func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// EnbbledForRepoID method of the pbrent MockSubRepoPermissionChecker
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) PushHook(hook func(context.Context, bpi.RepoID) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoID) (bool, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) nextHook() func(context.Context, bpi.RepoID) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) bppendCbll(r0 SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll objects describing the
// invocbtions of this function.
func (f *SubRepoPermissionCheckerEnbbledForRepoIDFunc) History() []SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll is bn object thbt
// describes bn invocbtion of method EnbbledForRepoID on bn instbnce of
// MockSubRepoPermissionChecker.
type SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll struct {
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
func (c SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionCheckerEnbbledForRepoIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SubRepoPermissionCheckerFilePermissionsFuncFunc describes the behbvior
// when the FilePermissionsFunc method of the pbrent
// MockSubRepoPermissionChecker instbnce is invoked.
type SubRepoPermissionCheckerFilePermissionsFuncFunc struct {
	defbultHook func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error)
	hooks       []func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error)
	history     []SubRepoPermissionCheckerFilePermissionsFuncFuncCbll
	mutex       sync.Mutex
}

// FilePermissionsFunc delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionChecker) FilePermissionsFunc(v0 context.Context, v1 int32, v2 bpi.RepoNbme) (FilePermissionFunc, error) {
	r0, r1 := m.FilePermissionsFuncFunc.nextHook()(v0, v1, v2)
	m.FilePermissionsFuncFunc.bppendCbll(SubRepoPermissionCheckerFilePermissionsFuncFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the FilePermissionsFunc
// method of the pbrent MockSubRepoPermissionChecker instbnce is invoked bnd
// the hook queue is empty.
func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) SetDefbultHook(hook func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// FilePermissionsFunc method of the pbrent MockSubRepoPermissionChecker
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) PushHook(hook func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) SetDefbultReturn(r0 FilePermissionFunc, r1 error) {
	f.SetDefbultHook(func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) PushReturn(r0 FilePermissionFunc, r1 error) {
	f.PushHook(func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) nextHook() func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) bppendCbll(r0 SubRepoPermissionCheckerFilePermissionsFuncFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// SubRepoPermissionCheckerFilePermissionsFuncFuncCbll objects describing
// the invocbtions of this function.
func (f *SubRepoPermissionCheckerFilePermissionsFuncFunc) History() []SubRepoPermissionCheckerFilePermissionsFuncFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionCheckerFilePermissionsFuncFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionCheckerFilePermissionsFuncFuncCbll is bn object thbt
// describes bn invocbtion of method FilePermissionsFunc on bn instbnce of
// MockSubRepoPermissionChecker.
type SubRepoPermissionCheckerFilePermissionsFuncFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 FilePermissionFunc
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SubRepoPermissionCheckerFilePermissionsFuncFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionCheckerFilePermissionsFuncFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SubRepoPermissionCheckerPermissionsFunc describes the behbvior when the
// Permissions method of the pbrent MockSubRepoPermissionChecker instbnce is
// invoked.
type SubRepoPermissionCheckerPermissionsFunc struct {
	defbultHook func(context.Context, int32, RepoContent) (Perms, error)
	hooks       []func(context.Context, int32, RepoContent) (Perms, error)
	history     []SubRepoPermissionCheckerPermissionsFuncCbll
	mutex       sync.Mutex
}

// Permissions delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSubRepoPermissionChecker) Permissions(v0 context.Context, v1 int32, v2 RepoContent) (Perms, error) {
	r0, r1 := m.PermissionsFunc.nextHook()(v0, v1, v2)
	m.PermissionsFunc.bppendCbll(SubRepoPermissionCheckerPermissionsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Permissions method
// of the pbrent MockSubRepoPermissionChecker instbnce is invoked bnd the
// hook queue is empty.
func (f *SubRepoPermissionCheckerPermissionsFunc) SetDefbultHook(hook func(context.Context, int32, RepoContent) (Perms, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Permissions method of the pbrent MockSubRepoPermissionChecker instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SubRepoPermissionCheckerPermissionsFunc) PushHook(hook func(context.Context, int32, RepoContent) (Perms, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SubRepoPermissionCheckerPermissionsFunc) SetDefbultReturn(r0 Perms, r1 error) {
	f.SetDefbultHook(func(context.Context, int32, RepoContent) (Perms, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SubRepoPermissionCheckerPermissionsFunc) PushReturn(r0 Perms, r1 error) {
	f.PushHook(func(context.Context, int32, RepoContent) (Perms, error) {
		return r0, r1
	})
}

func (f *SubRepoPermissionCheckerPermissionsFunc) nextHook() func(context.Context, int32, RepoContent) (Perms, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermissionCheckerPermissionsFunc) bppendCbll(r0 SubRepoPermissionCheckerPermissionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SubRepoPermissionCheckerPermissionsFuncCbll
// objects describing the invocbtions of this function.
func (f *SubRepoPermissionCheckerPermissionsFunc) History() []SubRepoPermissionCheckerPermissionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SubRepoPermissionCheckerPermissionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermissionCheckerPermissionsFuncCbll is bn object thbt describes
// bn invocbtion of method Permissions on bn instbnce of
// MockSubRepoPermissionChecker.
type SubRepoPermissionCheckerPermissionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 RepoContent
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Perms
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SubRepoPermissionCheckerPermissionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SubRepoPermissionCheckerPermissionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
