// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge repos

import (
	"context"
	"sync"

	zoekt "github.com/sourcegrbph/zoekt"
	query "github.com/sourcegrbph/zoekt/query"
)

// MockStrebmer is b mock implementbtion of the Strebmer interfbce (from the
// pbckbge github.com/sourcegrbph/zoekt) used for unit testing.
type MockStrebmer struct {
	// CloseFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Close.
	CloseFunc *StrebmerCloseFunc
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *StrebmerListFunc
	// SebrchFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Sebrch.
	SebrchFunc *StrebmerSebrchFunc
	// StrebmSebrchFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method StrebmSebrch.
	StrebmSebrchFunc *StrebmerStrebmSebrchFunc
	// StringFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method String.
	StringFunc *StrebmerStringFunc
}

// NewMockStrebmer crebtes b new mock of the Strebmer interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStrebmer() *MockStrebmer {
	return &MockStrebmer{
		CloseFunc: &StrebmerCloseFunc{
			defbultHook: func() {
				return
			},
		},
		ListFunc: &StrebmerListFunc{
			defbultHook: func(context.Context, query.Q, *zoekt.ListOptions) (r0 *zoekt.RepoList, r1 error) {
				return
			},
		},
		SebrchFunc: &StrebmerSebrchFunc{
			defbultHook: func(context.Context, query.Q, *zoekt.SebrchOptions) (r0 *zoekt.SebrchResult, r1 error) {
				return
			},
		},
		StrebmSebrchFunc: &StrebmerStrebmSebrchFunc{
			defbultHook: func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) (r0 error) {
				return
			},
		},
		StringFunc: &StrebmerStringFunc{
			defbultHook: func() (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockStrebmer crebtes b new mock of the Strebmer interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockStrebmer() *MockStrebmer {
	return &MockStrebmer{
		CloseFunc: &StrebmerCloseFunc{
			defbultHook: func() {
				pbnic("unexpected invocbtion of MockStrebmer.Close")
			},
		},
		ListFunc: &StrebmerListFunc{
			defbultHook: func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error) {
				pbnic("unexpected invocbtion of MockStrebmer.List")
			},
		},
		SebrchFunc: &StrebmerSebrchFunc{
			defbultHook: func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
				pbnic("unexpected invocbtion of MockStrebmer.Sebrch")
			},
		},
		StrebmSebrchFunc: &StrebmerStrebmSebrchFunc{
			defbultHook: func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error {
				pbnic("unexpected invocbtion of MockStrebmer.StrebmSebrch")
			},
		},
		StringFunc: &StrebmerStringFunc{
			defbultHook: func() string {
				pbnic("unexpected invocbtion of MockStrebmer.String")
			},
		},
	}
}

// NewMockStrebmerFrom crebtes b new mock of the MockStrebmer interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStrebmerFrom(i zoekt.Strebmer) *MockStrebmer {
	return &MockStrebmer{
		CloseFunc: &StrebmerCloseFunc{
			defbultHook: i.Close,
		},
		ListFunc: &StrebmerListFunc{
			defbultHook: i.List,
		},
		SebrchFunc: &StrebmerSebrchFunc{
			defbultHook: i.Sebrch,
		},
		StrebmSebrchFunc: &StrebmerStrebmSebrchFunc{
			defbultHook: i.StrebmSebrch,
		},
		StringFunc: &StrebmerStringFunc{
			defbultHook: i.String,
		},
	}
}

// StrebmerCloseFunc describes the behbvior when the Close method of the
// pbrent MockStrebmer instbnce is invoked.
type StrebmerCloseFunc struct {
	defbultHook func()
	hooks       []func()
	history     []StrebmerCloseFuncCbll
	mutex       sync.Mutex
}

// Close delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStrebmer) Close() {
	m.CloseFunc.nextHook()()
	m.CloseFunc.bppendCbll(StrebmerCloseFuncCbll{})
	return
}

// SetDefbultHook sets function thbt is cblled when the Close method of the
// pbrent MockStrebmer instbnce is invoked bnd the hook queue is empty.
func (f *StrebmerCloseFunc) SetDefbultHook(hook func()) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Close method of the pbrent MockStrebmer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StrebmerCloseFunc) PushHook(hook func()) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StrebmerCloseFunc) SetDefbultReturn() {
	f.SetDefbultHook(func() {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StrebmerCloseFunc) PushReturn() {
	f.PushHook(func() {
		return
	})
}

func (f *StrebmerCloseFunc) nextHook() func() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StrebmerCloseFunc) bppendCbll(r0 StrebmerCloseFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StrebmerCloseFuncCbll objects describing
// the invocbtions of this function.
func (f *StrebmerCloseFunc) History() []StrebmerCloseFuncCbll {
	f.mutex.Lock()
	history := mbke([]StrebmerCloseFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StrebmerCloseFuncCbll is bn object thbt describes bn invocbtion of method
// Close on bn instbnce of MockStrebmer.
type StrebmerCloseFuncCbll struct{}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StrebmerCloseFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StrebmerCloseFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// StrebmerListFunc describes the behbvior when the List method of the
// pbrent MockStrebmer instbnce is invoked.
type StrebmerListFunc struct {
	defbultHook func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error)
	hooks       []func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error)
	history     []StrebmerListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStrebmer) List(v0 context.Context, v1 query.Q, v2 *zoekt.ListOptions) (*zoekt.RepoList, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1, v2)
	m.ListFunc.bppendCbll(StrebmerListFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockStrebmer instbnce is invoked bnd the hook queue is empty.
func (f *StrebmerListFunc) SetDefbultHook(hook func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockStrebmer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StrebmerListFunc) PushHook(hook func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StrebmerListFunc) SetDefbultReturn(r0 *zoekt.RepoList, r1 error) {
	f.SetDefbultHook(func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StrebmerListFunc) PushReturn(r0 *zoekt.RepoList, r1 error) {
	f.PushHook(func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error) {
		return r0, r1
	})
}

func (f *StrebmerListFunc) nextHook() func(context.Context, query.Q, *zoekt.ListOptions) (*zoekt.RepoList, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StrebmerListFunc) bppendCbll(r0 StrebmerListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StrebmerListFuncCbll objects describing the
// invocbtions of this function.
func (f *StrebmerListFunc) History() []StrebmerListFuncCbll {
	f.mutex.Lock()
	history := mbke([]StrebmerListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StrebmerListFuncCbll is bn object thbt describes bn invocbtion of method
// List on bn instbnce of MockStrebmer.
type StrebmerListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 query.Q
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *zoekt.ListOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *zoekt.RepoList
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StrebmerListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StrebmerListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StrebmerSebrchFunc describes the behbvior when the Sebrch method of the
// pbrent MockStrebmer instbnce is invoked.
type StrebmerSebrchFunc struct {
	defbultHook func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error)
	hooks       []func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error)
	history     []StrebmerSebrchFuncCbll
	mutex       sync.Mutex
}

// Sebrch delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStrebmer) Sebrch(v0 context.Context, v1 query.Q, v2 *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	r0, r1 := m.SebrchFunc.nextHook()(v0, v1, v2)
	m.SebrchFunc.bppendCbll(StrebmerSebrchFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Sebrch method of the
// pbrent MockStrebmer instbnce is invoked bnd the hook queue is empty.
func (f *StrebmerSebrchFunc) SetDefbultHook(hook func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Sebrch method of the pbrent MockStrebmer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StrebmerSebrchFunc) PushHook(hook func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StrebmerSebrchFunc) SetDefbultReturn(r0 *zoekt.SebrchResult, r1 error) {
	f.SetDefbultHook(func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StrebmerSebrchFunc) PushReturn(r0 *zoekt.SebrchResult, r1 error) {
	f.PushHook(func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
		return r0, r1
	})
}

func (f *StrebmerSebrchFunc) nextHook() func(context.Context, query.Q, *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StrebmerSebrchFunc) bppendCbll(r0 StrebmerSebrchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StrebmerSebrchFuncCbll objects describing
// the invocbtions of this function.
func (f *StrebmerSebrchFunc) History() []StrebmerSebrchFuncCbll {
	f.mutex.Lock()
	history := mbke([]StrebmerSebrchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StrebmerSebrchFuncCbll is bn object thbt describes bn invocbtion of
// method Sebrch on bn instbnce of MockStrebmer.
type StrebmerSebrchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 query.Q
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *zoekt.SebrchOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *zoekt.SebrchResult
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StrebmerSebrchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StrebmerSebrchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StrebmerStrebmSebrchFunc describes the behbvior when the StrebmSebrch
// method of the pbrent MockStrebmer instbnce is invoked.
type StrebmerStrebmSebrchFunc struct {
	defbultHook func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error
	hooks       []func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error
	history     []StrebmerStrebmSebrchFuncCbll
	mutex       sync.Mutex
}

// StrebmSebrch delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStrebmer) StrebmSebrch(v0 context.Context, v1 query.Q, v2 *zoekt.SebrchOptions, v3 zoekt.Sender) error {
	r0 := m.StrebmSebrchFunc.nextHook()(v0, v1, v2, v3)
	m.StrebmSebrchFunc.bppendCbll(StrebmerStrebmSebrchFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the StrebmSebrch method
// of the pbrent MockStrebmer instbnce is invoked bnd the hook queue is
// empty.
func (f *StrebmerStrebmSebrchFunc) SetDefbultHook(hook func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StrebmSebrch method of the pbrent MockStrebmer instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StrebmerStrebmSebrchFunc) PushHook(hook func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StrebmerStrebmSebrchFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StrebmerStrebmSebrchFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error {
		return r0
	})
}

func (f *StrebmerStrebmSebrchFunc) nextHook() func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StrebmerStrebmSebrchFunc) bppendCbll(r0 StrebmerStrebmSebrchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StrebmerStrebmSebrchFuncCbll objects
// describing the invocbtions of this function.
func (f *StrebmerStrebmSebrchFunc) History() []StrebmerStrebmSebrchFuncCbll {
	f.mutex.Lock()
	history := mbke([]StrebmerStrebmSebrchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StrebmerStrebmSebrchFuncCbll is bn object thbt describes bn invocbtion of
// method StrebmSebrch on bn instbnce of MockStrebmer.
type StrebmerStrebmSebrchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 query.Q
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *zoekt.SebrchOptions
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 zoekt.Sender
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StrebmerStrebmSebrchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StrebmerStrebmSebrchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StrebmerStringFunc describes the behbvior when the String method of the
// pbrent MockStrebmer instbnce is invoked.
type StrebmerStringFunc struct {
	defbultHook func() string
	hooks       []func() string
	history     []StrebmerStringFuncCbll
	mutex       sync.Mutex
}

// String delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStrebmer) String() string {
	r0 := m.StringFunc.nextHook()()
	m.StringFunc.bppendCbll(StrebmerStringFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the String method of the
// pbrent MockStrebmer instbnce is invoked bnd the hook queue is empty.
func (f *StrebmerStringFunc) SetDefbultHook(hook func() string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// String method of the pbrent MockStrebmer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StrebmerStringFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StrebmerStringFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func() string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StrebmerStringFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *StrebmerStringFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StrebmerStringFunc) bppendCbll(r0 StrebmerStringFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StrebmerStringFuncCbll objects describing
// the invocbtions of this function.
func (f *StrebmerStringFunc) History() []StrebmerStringFuncCbll {
	f.mutex.Lock()
	history := mbke([]StrebmerStringFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StrebmerStringFuncCbll is bn object thbt describes bn invocbtion of
// method String on bn instbnce of MockStrebmer.
type StrebmerStringFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StrebmerStringFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StrebmerStringFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
