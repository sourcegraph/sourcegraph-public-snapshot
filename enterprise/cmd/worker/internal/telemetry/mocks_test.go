// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge telemetry

import (
	"context"
	"sync"
)

// MockBookmbrkStore is b mock implementbtion of the bookmbrkStore interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/telemetry)
// used for unit testing.
type MockBookmbrkStore struct {
	// GetBookmbrkFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetBookmbrk.
	GetBookmbrkFunc *BookmbrkStoreGetBookmbrkFunc
	// UpdbteBookmbrkFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteBookmbrk.
	UpdbteBookmbrkFunc *BookmbrkStoreUpdbteBookmbrkFunc
}

// NewMockBookmbrkStore crebtes b new mock of the bookmbrkStore interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockBookmbrkStore() *MockBookmbrkStore {
	return &MockBookmbrkStore{
		GetBookmbrkFunc: &BookmbrkStoreGetBookmbrkFunc{
			defbultHook: func(context.Context) (r0 int, r1 error) {
				return
			},
		},
		UpdbteBookmbrkFunc: &BookmbrkStoreUpdbteBookmbrkFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockBookmbrkStore crebtes b new mock of the bookmbrkStore
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockBookmbrkStore() *MockBookmbrkStore {
	return &MockBookmbrkStore{
		GetBookmbrkFunc: &BookmbrkStoreGetBookmbrkFunc{
			defbultHook: func(context.Context) (int, error) {
				pbnic("unexpected invocbtion of MockBookmbrkStore.GetBookmbrk")
			},
		},
		UpdbteBookmbrkFunc: &BookmbrkStoreUpdbteBookmbrkFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockBookmbrkStore.UpdbteBookmbrk")
			},
		},
	}
}

// surrogbteMockBookmbrkStore is b copy of the bookmbrkStore interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/telemetry).
// It is redefined here bs it is unexported in the source pbckbge.
type surrogbteMockBookmbrkStore interfbce {
	GetBookmbrk(context.Context) (int, error)
	UpdbteBookmbrk(context.Context, int) error
}

// NewMockBookmbrkStoreFrom crebtes b new mock of the MockBookmbrkStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockBookmbrkStoreFrom(i surrogbteMockBookmbrkStore) *MockBookmbrkStore {
	return &MockBookmbrkStore{
		GetBookmbrkFunc: &BookmbrkStoreGetBookmbrkFunc{
			defbultHook: i.GetBookmbrk,
		},
		UpdbteBookmbrkFunc: &BookmbrkStoreUpdbteBookmbrkFunc{
			defbultHook: i.UpdbteBookmbrk,
		},
	}
}

// BookmbrkStoreGetBookmbrkFunc describes the behbvior when the GetBookmbrk
// method of the pbrent MockBookmbrkStore instbnce is invoked.
type BookmbrkStoreGetBookmbrkFunc struct {
	defbultHook func(context.Context) (int, error)
	hooks       []func(context.Context) (int, error)
	history     []BookmbrkStoreGetBookmbrkFuncCbll
	mutex       sync.Mutex
}

// GetBookmbrk delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBookmbrkStore) GetBookmbrk(v0 context.Context) (int, error) {
	r0, r1 := m.GetBookmbrkFunc.nextHook()(v0)
	m.GetBookmbrkFunc.bppendCbll(BookmbrkStoreGetBookmbrkFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBookmbrk method
// of the pbrent MockBookmbrkStore instbnce is invoked bnd the hook queue is
// empty.
func (f *BookmbrkStoreGetBookmbrkFunc) SetDefbultHook(hook func(context.Context) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBookmbrk method of the pbrent MockBookmbrkStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *BookmbrkStoreGetBookmbrkFunc) PushHook(hook func(context.Context) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BookmbrkStoreGetBookmbrkFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BookmbrkStoreGetBookmbrkFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

func (f *BookmbrkStoreGetBookmbrkFunc) nextHook() func(context.Context) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BookmbrkStoreGetBookmbrkFunc) bppendCbll(r0 BookmbrkStoreGetBookmbrkFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BookmbrkStoreGetBookmbrkFuncCbll objects
// describing the invocbtions of this function.
func (f *BookmbrkStoreGetBookmbrkFunc) History() []BookmbrkStoreGetBookmbrkFuncCbll {
	f.mutex.Lock()
	history := mbke([]BookmbrkStoreGetBookmbrkFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BookmbrkStoreGetBookmbrkFuncCbll is bn object thbt describes bn
// invocbtion of method GetBookmbrk on bn instbnce of MockBookmbrkStore.
type BookmbrkStoreGetBookmbrkFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c BookmbrkStoreGetBookmbrkFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BookmbrkStoreGetBookmbrkFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// BookmbrkStoreUpdbteBookmbrkFunc describes the behbvior when the
// UpdbteBookmbrk method of the pbrent MockBookmbrkStore instbnce is
// invoked.
type BookmbrkStoreUpdbteBookmbrkFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []BookmbrkStoreUpdbteBookmbrkFuncCbll
	mutex       sync.Mutex
}

// UpdbteBookmbrk delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockBookmbrkStore) UpdbteBookmbrk(v0 context.Context, v1 int) error {
	r0 := m.UpdbteBookmbrkFunc.nextHook()(v0, v1)
	m.UpdbteBookmbrkFunc.bppendCbll(BookmbrkStoreUpdbteBookmbrkFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UpdbteBookmbrk
// method of the pbrent MockBookmbrkStore instbnce is invoked bnd the hook
// queue is empty.
func (f *BookmbrkStoreUpdbteBookmbrkFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteBookmbrk method of the pbrent MockBookmbrkStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *BookmbrkStoreUpdbteBookmbrkFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *BookmbrkStoreUpdbteBookmbrkFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *BookmbrkStoreUpdbteBookmbrkFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *BookmbrkStoreUpdbteBookmbrkFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BookmbrkStoreUpdbteBookmbrkFunc) bppendCbll(r0 BookmbrkStoreUpdbteBookmbrkFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of BookmbrkStoreUpdbteBookmbrkFuncCbll objects
// describing the invocbtions of this function.
func (f *BookmbrkStoreUpdbteBookmbrkFunc) History() []BookmbrkStoreUpdbteBookmbrkFuncCbll {
	f.mutex.Lock()
	history := mbke([]BookmbrkStoreUpdbteBookmbrkFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BookmbrkStoreUpdbteBookmbrkFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteBookmbrk on bn instbnce of MockBookmbrkStore.
type BookmbrkStoreUpdbteBookmbrkFuncCbll struct {
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
func (c BookmbrkStoreUpdbteBookmbrkFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c BookmbrkStoreUpdbteBookmbrkFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
