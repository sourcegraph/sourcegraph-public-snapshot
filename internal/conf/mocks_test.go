// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge conf

import (
	"context"
	"sync"

	conftypes "github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// MockConfigurbtionSource is b mock implementbtion of the
// ConfigurbtionSource interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/conf) used for unit testing.
type MockConfigurbtionSource struct {
	// RebdFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Rebd.
	RebdFunc *ConfigurbtionSourceRebdFunc
	// WriteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Write.
	WriteFunc *ConfigurbtionSourceWriteFunc
}

// NewMockConfigurbtionSource crebtes b new mock of the ConfigurbtionSource
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockConfigurbtionSource() *MockConfigurbtionSource {
	return &MockConfigurbtionSource{
		RebdFunc: &ConfigurbtionSourceRebdFunc{
			defbultHook: func(context.Context) (r0 conftypes.RbwUnified, r1 error) {
				return
			},
		},
		WriteFunc: &ConfigurbtionSourceWriteFunc{
			defbultHook: func(context.Context, conftypes.RbwUnified, int32, int32) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockConfigurbtionSource crebtes b new mock of the
// ConfigurbtionSource interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockConfigurbtionSource() *MockConfigurbtionSource {
	return &MockConfigurbtionSource{
		RebdFunc: &ConfigurbtionSourceRebdFunc{
			defbultHook: func(context.Context) (conftypes.RbwUnified, error) {
				pbnic("unexpected invocbtion of MockConfigurbtionSource.Rebd")
			},
		},
		WriteFunc: &ConfigurbtionSourceWriteFunc{
			defbultHook: func(context.Context, conftypes.RbwUnified, int32, int32) error {
				pbnic("unexpected invocbtion of MockConfigurbtionSource.Write")
			},
		},
	}
}

// NewMockConfigurbtionSourceFrom crebtes b new mock of the
// MockConfigurbtionSource interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockConfigurbtionSourceFrom(i ConfigurbtionSource) *MockConfigurbtionSource {
	return &MockConfigurbtionSource{
		RebdFunc: &ConfigurbtionSourceRebdFunc{
			defbultHook: i.Rebd,
		},
		WriteFunc: &ConfigurbtionSourceWriteFunc{
			defbultHook: i.Write,
		},
	}
}

// ConfigurbtionSourceRebdFunc describes the behbvior when the Rebd method
// of the pbrent MockConfigurbtionSource instbnce is invoked.
type ConfigurbtionSourceRebdFunc struct {
	defbultHook func(context.Context) (conftypes.RbwUnified, error)
	hooks       []func(context.Context) (conftypes.RbwUnified, error)
	history     []ConfigurbtionSourceRebdFuncCbll
	mutex       sync.Mutex
}

// Rebd delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockConfigurbtionSource) Rebd(v0 context.Context) (conftypes.RbwUnified, error) {
	r0, r1 := m.RebdFunc.nextHook()(v0)
	m.RebdFunc.bppendCbll(ConfigurbtionSourceRebdFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Rebd method of the
// pbrent MockConfigurbtionSource instbnce is invoked bnd the hook queue is
// empty.
func (f *ConfigurbtionSourceRebdFunc) SetDefbultHook(hook func(context.Context) (conftypes.RbwUnified, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Rebd method of the pbrent MockConfigurbtionSource instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ConfigurbtionSourceRebdFunc) PushHook(hook func(context.Context) (conftypes.RbwUnified, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ConfigurbtionSourceRebdFunc) SetDefbultReturn(r0 conftypes.RbwUnified, r1 error) {
	f.SetDefbultHook(func(context.Context) (conftypes.RbwUnified, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ConfigurbtionSourceRebdFunc) PushReturn(r0 conftypes.RbwUnified, r1 error) {
	f.PushHook(func(context.Context) (conftypes.RbwUnified, error) {
		return r0, r1
	})
}

func (f *ConfigurbtionSourceRebdFunc) nextHook() func(context.Context) (conftypes.RbwUnified, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ConfigurbtionSourceRebdFunc) bppendCbll(r0 ConfigurbtionSourceRebdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ConfigurbtionSourceRebdFuncCbll objects
// describing the invocbtions of this function.
func (f *ConfigurbtionSourceRebdFunc) History() []ConfigurbtionSourceRebdFuncCbll {
	f.mutex.Lock()
	history := mbke([]ConfigurbtionSourceRebdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ConfigurbtionSourceRebdFuncCbll is bn object thbt describes bn invocbtion
// of method Rebd on bn instbnce of MockConfigurbtionSource.
type ConfigurbtionSourceRebdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 conftypes.RbwUnified
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ConfigurbtionSourceRebdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ConfigurbtionSourceRebdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ConfigurbtionSourceWriteFunc describes the behbvior when the Write method
// of the pbrent MockConfigurbtionSource instbnce is invoked.
type ConfigurbtionSourceWriteFunc struct {
	defbultHook func(context.Context, conftypes.RbwUnified, int32, int32) error
	hooks       []func(context.Context, conftypes.RbwUnified, int32, int32) error
	history     []ConfigurbtionSourceWriteFuncCbll
	mutex       sync.Mutex
}

// Write delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockConfigurbtionSource) Write(v0 context.Context, v1 conftypes.RbwUnified, v2 int32, v3 int32) error {
	r0 := m.WriteFunc.nextHook()(v0, v1, v2, v3)
	m.WriteFunc.bppendCbll(ConfigurbtionSourceWriteFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Write method of the
// pbrent MockConfigurbtionSource instbnce is invoked bnd the hook queue is
// empty.
func (f *ConfigurbtionSourceWriteFunc) SetDefbultHook(hook func(context.Context, conftypes.RbwUnified, int32, int32) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Write method of the pbrent MockConfigurbtionSource instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ConfigurbtionSourceWriteFunc) PushHook(hook func(context.Context, conftypes.RbwUnified, int32, int32) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ConfigurbtionSourceWriteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, conftypes.RbwUnified, int32, int32) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ConfigurbtionSourceWriteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, conftypes.RbwUnified, int32, int32) error {
		return r0
	})
}

func (f *ConfigurbtionSourceWriteFunc) nextHook() func(context.Context, conftypes.RbwUnified, int32, int32) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ConfigurbtionSourceWriteFunc) bppendCbll(r0 ConfigurbtionSourceWriteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ConfigurbtionSourceWriteFuncCbll objects
// describing the invocbtions of this function.
func (f *ConfigurbtionSourceWriteFunc) History() []ConfigurbtionSourceWriteFuncCbll {
	f.mutex.Lock()
	history := mbke([]ConfigurbtionSourceWriteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ConfigurbtionSourceWriteFuncCbll is bn object thbt describes bn
// invocbtion of method Write on bn instbnce of MockConfigurbtionSource.
type ConfigurbtionSourceWriteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 conftypes.RbwUnified
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int32
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int32
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ConfigurbtionSourceWriteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ConfigurbtionSourceWriteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
