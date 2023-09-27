// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge febtureflbg

import (
	"context"
	"sync"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg) used for
// unit testing.
type MockStore struct {
	// GetAnonymousUserFlbgsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetAnonymousUserFlbgs.
	GetAnonymousUserFlbgsFunc *StoreGetAnonymousUserFlbgsFunc
	// GetGlobblFebtureFlbgsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetGlobblFebtureFlbgs.
	GetGlobblFebtureFlbgsFunc *StoreGetGlobblFebtureFlbgsFunc
	// GetUserFlbgsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetUserFlbgs.
	GetUserFlbgsFunc *StoreGetUserFlbgsFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		GetAnonymousUserFlbgsFunc: &StoreGetAnonymousUserFlbgsFunc{
			defbultHook: func(context.Context, string) (r0 mbp[string]bool, r1 error) {
				return
			},
		},
		GetGlobblFebtureFlbgsFunc: &StoreGetGlobblFebtureFlbgsFunc{
			defbultHook: func(context.Context) (r0 mbp[string]bool, r1 error) {
				return
			},
		},
		GetUserFlbgsFunc: &StoreGetUserFlbgsFunc{
			defbultHook: func(context.Context, int32) (r0 mbp[string]bool, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		GetAnonymousUserFlbgsFunc: &StoreGetAnonymousUserFlbgsFunc{
			defbultHook: func(context.Context, string) (mbp[string]bool, error) {
				pbnic("unexpected invocbtion of MockStore.GetAnonymousUserFlbgs")
			},
		},
		GetGlobblFebtureFlbgsFunc: &StoreGetGlobblFebtureFlbgsFunc{
			defbultHook: func(context.Context) (mbp[string]bool, error) {
				pbnic("unexpected invocbtion of MockStore.GetGlobblFebtureFlbgs")
			},
		},
		GetUserFlbgsFunc: &StoreGetUserFlbgsFunc{
			defbultHook: func(context.Context, int32) (mbp[string]bool, error) {
				pbnic("unexpected invocbtion of MockStore.GetUserFlbgs")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		GetAnonymousUserFlbgsFunc: &StoreGetAnonymousUserFlbgsFunc{
			defbultHook: i.GetAnonymousUserFlbgs,
		},
		GetGlobblFebtureFlbgsFunc: &StoreGetGlobblFebtureFlbgsFunc{
			defbultHook: i.GetGlobblFebtureFlbgs,
		},
		GetUserFlbgsFunc: &StoreGetUserFlbgsFunc{
			defbultHook: i.GetUserFlbgs,
		},
	}
}

// StoreGetAnonymousUserFlbgsFunc describes the behbvior when the
// GetAnonymousUserFlbgs method of the pbrent MockStore instbnce is invoked.
type StoreGetAnonymousUserFlbgsFunc struct {
	defbultHook func(context.Context, string) (mbp[string]bool, error)
	hooks       []func(context.Context, string) (mbp[string]bool, error)
	history     []StoreGetAnonymousUserFlbgsFuncCbll
	mutex       sync.Mutex
}

// GetAnonymousUserFlbgs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetAnonymousUserFlbgs(v0 context.Context, v1 string) (mbp[string]bool, error) {
	r0, r1 := m.GetAnonymousUserFlbgsFunc.nextHook()(v0, v1)
	m.GetAnonymousUserFlbgsFunc.bppendCbll(StoreGetAnonymousUserFlbgsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetAnonymousUserFlbgs method of the pbrent MockStore instbnce is invoked
// bnd the hook queue is empty.
func (f *StoreGetAnonymousUserFlbgsFunc) SetDefbultHook(hook func(context.Context, string) (mbp[string]bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAnonymousUserFlbgs method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreGetAnonymousUserFlbgsFunc) PushHook(hook func(context.Context, string) (mbp[string]bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetAnonymousUserFlbgsFunc) SetDefbultReturn(r0 mbp[string]bool, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (mbp[string]bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetAnonymousUserFlbgsFunc) PushReturn(r0 mbp[string]bool, r1 error) {
	f.PushHook(func(context.Context, string) (mbp[string]bool, error) {
		return r0, r1
	})
}

func (f *StoreGetAnonymousUserFlbgsFunc) nextHook() func(context.Context, string) (mbp[string]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetAnonymousUserFlbgsFunc) bppendCbll(r0 StoreGetAnonymousUserFlbgsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetAnonymousUserFlbgsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetAnonymousUserFlbgsFunc) History() []StoreGetAnonymousUserFlbgsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetAnonymousUserFlbgsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetAnonymousUserFlbgsFuncCbll is bn object thbt describes bn
// invocbtion of method GetAnonymousUserFlbgs on bn instbnce of MockStore.
type StoreGetAnonymousUserFlbgsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetAnonymousUserFlbgsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetAnonymousUserFlbgsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetGlobblFebtureFlbgsFunc describes the behbvior when the
// GetGlobblFebtureFlbgs method of the pbrent MockStore instbnce is invoked.
type StoreGetGlobblFebtureFlbgsFunc struct {
	defbultHook func(context.Context) (mbp[string]bool, error)
	hooks       []func(context.Context) (mbp[string]bool, error)
	history     []StoreGetGlobblFebtureFlbgsFuncCbll
	mutex       sync.Mutex
}

// GetGlobblFebtureFlbgs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetGlobblFebtureFlbgs(v0 context.Context) (mbp[string]bool, error) {
	r0, r1 := m.GetGlobblFebtureFlbgsFunc.nextHook()(v0)
	m.GetGlobblFebtureFlbgsFunc.bppendCbll(StoreGetGlobblFebtureFlbgsFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetGlobblFebtureFlbgs method of the pbrent MockStore instbnce is invoked
// bnd the hook queue is empty.
func (f *StoreGetGlobblFebtureFlbgsFunc) SetDefbultHook(hook func(context.Context) (mbp[string]bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetGlobblFebtureFlbgs method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreGetGlobblFebtureFlbgsFunc) PushHook(hook func(context.Context) (mbp[string]bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetGlobblFebtureFlbgsFunc) SetDefbultReturn(r0 mbp[string]bool, r1 error) {
	f.SetDefbultHook(func(context.Context) (mbp[string]bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetGlobblFebtureFlbgsFunc) PushReturn(r0 mbp[string]bool, r1 error) {
	f.PushHook(func(context.Context) (mbp[string]bool, error) {
		return r0, r1
	})
}

func (f *StoreGetGlobblFebtureFlbgsFunc) nextHook() func(context.Context) (mbp[string]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetGlobblFebtureFlbgsFunc) bppendCbll(r0 StoreGetGlobblFebtureFlbgsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetGlobblFebtureFlbgsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetGlobblFebtureFlbgsFunc) History() []StoreGetGlobblFebtureFlbgsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetGlobblFebtureFlbgsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetGlobblFebtureFlbgsFuncCbll is bn object thbt describes bn
// invocbtion of method GetGlobblFebtureFlbgs on bn instbnce of MockStore.
type StoreGetGlobblFebtureFlbgsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetGlobblFebtureFlbgsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetGlobblFebtureFlbgsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetUserFlbgsFunc describes the behbvior when the GetUserFlbgs method
// of the pbrent MockStore instbnce is invoked.
type StoreGetUserFlbgsFunc struct {
	defbultHook func(context.Context, int32) (mbp[string]bool, error)
	hooks       []func(context.Context, int32) (mbp[string]bool, error)
	history     []StoreGetUserFlbgsFuncCbll
	mutex       sync.Mutex
}

// GetUserFlbgs delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetUserFlbgs(v0 context.Context, v1 int32) (mbp[string]bool, error) {
	r0, r1 := m.GetUserFlbgsFunc.nextHook()(v0, v1)
	m.GetUserFlbgsFunc.bppendCbll(StoreGetUserFlbgsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetUserFlbgs method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreGetUserFlbgsFunc) SetDefbultHook(hook func(context.Context, int32) (mbp[string]bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUserFlbgs method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreGetUserFlbgsFunc) PushHook(hook func(context.Context, int32) (mbp[string]bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetUserFlbgsFunc) SetDefbultReturn(r0 mbp[string]bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int32) (mbp[string]bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetUserFlbgsFunc) PushReturn(r0 mbp[string]bool, r1 error) {
	f.PushHook(func(context.Context, int32) (mbp[string]bool, error) {
		return r0, r1
	})
}

func (f *StoreGetUserFlbgsFunc) nextHook() func(context.Context, int32) (mbp[string]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetUserFlbgsFunc) bppendCbll(r0 StoreGetUserFlbgsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetUserFlbgsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetUserFlbgsFunc) History() []StoreGetUserFlbgsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetUserFlbgsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetUserFlbgsFuncCbll is bn object thbt describes bn invocbtion of
// method GetUserFlbgs on bn instbnce of MockStore.
type StoreGetUserFlbgsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int32
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetUserFlbgsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetUserFlbgsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
