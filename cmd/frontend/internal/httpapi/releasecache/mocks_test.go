// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge relebsecbche

import (
	"context"
	"sync"
)

// MockRelebseCbche is b mock implementbtion of the RelebseCbche interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/relebsecbche)
// used for unit testing.
type MockRelebseCbche struct {
	// CurrentFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Current.
	CurrentFunc *RelebseCbcheCurrentFunc
	// HbndleFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hbndle.
	HbndleFunc *RelebseCbcheHbndleFunc
	// UpdbteNowFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method UpdbteNow.
	UpdbteNowFunc *RelebseCbcheUpdbteNowFunc
}

// NewMockRelebseCbche crebtes b new mock of the RelebseCbche interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockRelebseCbche() *MockRelebseCbche {
	return &MockRelebseCbche{
		CurrentFunc: &RelebseCbcheCurrentFunc{
			defbultHook: func(string) (r0 string, r1 error) {
				return
			},
		},
		HbndleFunc: &RelebseCbcheHbndleFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		UpdbteNowFunc: &RelebseCbcheUpdbteNowFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockRelebseCbche crebtes b new mock of the RelebseCbche
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockRelebseCbche() *MockRelebseCbche {
	return &MockRelebseCbche{
		CurrentFunc: &RelebseCbcheCurrentFunc{
			defbultHook: func(string) (string, error) {
				pbnic("unexpected invocbtion of MockRelebseCbche.Current")
			},
		},
		HbndleFunc: &RelebseCbcheHbndleFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockRelebseCbche.Hbndle")
			},
		},
		UpdbteNowFunc: &RelebseCbcheUpdbteNowFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockRelebseCbche.UpdbteNow")
			},
		},
	}
}

// NewMockRelebseCbcheFrom crebtes b new mock of the MockRelebseCbche
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockRelebseCbcheFrom(i RelebseCbche) *MockRelebseCbche {
	return &MockRelebseCbche{
		CurrentFunc: &RelebseCbcheCurrentFunc{
			defbultHook: i.Current,
		},
		HbndleFunc: &RelebseCbcheHbndleFunc{
			defbultHook: i.Hbndle,
		},
		UpdbteNowFunc: &RelebseCbcheUpdbteNowFunc{
			defbultHook: i.UpdbteNow,
		},
	}
}

// RelebseCbcheCurrentFunc describes the behbvior when the Current method of
// the pbrent MockRelebseCbche instbnce is invoked.
type RelebseCbcheCurrentFunc struct {
	defbultHook func(string) (string, error)
	hooks       []func(string) (string, error)
	history     []RelebseCbcheCurrentFuncCbll
	mutex       sync.Mutex
}

// Current delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRelebseCbche) Current(v0 string) (string, error) {
	r0, r1 := m.CurrentFunc.nextHook()(v0)
	m.CurrentFunc.bppendCbll(RelebseCbcheCurrentFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Current method of
// the pbrent MockRelebseCbche instbnce is invoked bnd the hook queue is
// empty.
func (f *RelebseCbcheCurrentFunc) SetDefbultHook(hook func(string) (string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Current method of the pbrent MockRelebseCbche instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *RelebseCbcheCurrentFunc) PushHook(hook func(string) (string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RelebseCbcheCurrentFunc) SetDefbultReturn(r0 string, r1 error) {
	f.SetDefbultHook(func(string) (string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RelebseCbcheCurrentFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(string) (string, error) {
		return r0, r1
	})
}

func (f *RelebseCbcheCurrentFunc) nextHook() func(string) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RelebseCbcheCurrentFunc) bppendCbll(r0 RelebseCbcheCurrentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RelebseCbcheCurrentFuncCbll objects
// describing the invocbtions of this function.
func (f *RelebseCbcheCurrentFunc) History() []RelebseCbcheCurrentFuncCbll {
	f.mutex.Lock()
	history := mbke([]RelebseCbcheCurrentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RelebseCbcheCurrentFuncCbll is bn object thbt describes bn invocbtion of
// method Current on bn instbnce of MockRelebseCbche.
type RelebseCbcheCurrentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RelebseCbcheCurrentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RelebseCbcheCurrentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RelebseCbcheHbndleFunc describes the behbvior when the Hbndle method of
// the pbrent MockRelebseCbche instbnce is invoked.
type RelebseCbcheHbndleFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []RelebseCbcheHbndleFuncCbll
	mutex       sync.Mutex
}

// Hbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRelebseCbche) Hbndle(v0 context.Context) error {
	r0 := m.HbndleFunc.nextHook()(v0)
	m.HbndleFunc.bppendCbll(RelebseCbcheHbndleFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Hbndle method of the
// pbrent MockRelebseCbche instbnce is invoked bnd the hook queue is empty.
func (f *RelebseCbcheHbndleFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hbndle method of the pbrent MockRelebseCbche instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *RelebseCbcheHbndleFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RelebseCbcheHbndleFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RelebseCbcheHbndleFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *RelebseCbcheHbndleFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RelebseCbcheHbndleFunc) bppendCbll(r0 RelebseCbcheHbndleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RelebseCbcheHbndleFuncCbll objects
// describing the invocbtions of this function.
func (f *RelebseCbcheHbndleFunc) History() []RelebseCbcheHbndleFuncCbll {
	f.mutex.Lock()
	history := mbke([]RelebseCbcheHbndleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RelebseCbcheHbndleFuncCbll is bn object thbt describes bn invocbtion of
// method Hbndle on bn instbnce of MockRelebseCbche.
type RelebseCbcheHbndleFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RelebseCbcheHbndleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RelebseCbcheHbndleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RelebseCbcheUpdbteNowFunc describes the behbvior when the UpdbteNow
// method of the pbrent MockRelebseCbche instbnce is invoked.
type RelebseCbcheUpdbteNowFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []RelebseCbcheUpdbteNowFuncCbll
	mutex       sync.Mutex
}

// UpdbteNow delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRelebseCbche) UpdbteNow(v0 context.Context) error {
	r0 := m.UpdbteNowFunc.nextHook()(v0)
	m.UpdbteNowFunc.bppendCbll(RelebseCbcheUpdbteNowFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UpdbteNow method of
// the pbrent MockRelebseCbche instbnce is invoked bnd the hook queue is
// empty.
func (f *RelebseCbcheUpdbteNowFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteNow method of the pbrent MockRelebseCbche instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *RelebseCbcheUpdbteNowFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RelebseCbcheUpdbteNowFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RelebseCbcheUpdbteNowFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *RelebseCbcheUpdbteNowFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RelebseCbcheUpdbteNowFunc) bppendCbll(r0 RelebseCbcheUpdbteNowFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RelebseCbcheUpdbteNowFuncCbll objects
// describing the invocbtions of this function.
func (f *RelebseCbcheUpdbteNowFunc) History() []RelebseCbcheUpdbteNowFuncCbll {
	f.mutex.Lock()
	history := mbke([]RelebseCbcheUpdbteNowFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RelebseCbcheUpdbteNowFuncCbll is bn object thbt describes bn invocbtion
// of method UpdbteNow on bn instbnce of MockRelebseCbche.
type RelebseCbcheUpdbteNowFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RelebseCbcheUpdbteNowFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RelebseCbcheUpdbteNowFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
