// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge oobmigrbtion

import (
	"context"
	"sync"
)

// MockMigrbtor is b mock implementbtion of the Migrbtor interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion) used
// for unit testing.
type MockMigrbtor struct {
	// DownFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Down.
	DownFunc *MigrbtorDownFunc
	// ProgressFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Progress.
	ProgressFunc *MigrbtorProgressFunc
	// UpFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Up.
	UpFunc *MigrbtorUpFunc
}

// NewMockMigrbtor crebtes b new mock of the Migrbtor interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockMigrbtor() *MockMigrbtor {
	return &MockMigrbtor{
		DownFunc: &MigrbtorDownFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		ProgressFunc: &MigrbtorProgressFunc{
			defbultHook: func(context.Context, bool) (r0 flobt64, r1 error) {
				return
			},
		},
		UpFunc: &MigrbtorUpFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockMigrbtor crebtes b new mock of the Migrbtor interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockMigrbtor() *MockMigrbtor {
	return &MockMigrbtor{
		DownFunc: &MigrbtorDownFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockMigrbtor.Down")
			},
		},
		ProgressFunc: &MigrbtorProgressFunc{
			defbultHook: func(context.Context, bool) (flobt64, error) {
				pbnic("unexpected invocbtion of MockMigrbtor.Progress")
			},
		},
		UpFunc: &MigrbtorUpFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockMigrbtor.Up")
			},
		},
	}
}

// NewMockMigrbtorFrom crebtes b new mock of the MockMigrbtor interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockMigrbtorFrom(i Migrbtor) *MockMigrbtor {
	return &MockMigrbtor{
		DownFunc: &MigrbtorDownFunc{
			defbultHook: i.Down,
		},
		ProgressFunc: &MigrbtorProgressFunc{
			defbultHook: i.Progress,
		},
		UpFunc: &MigrbtorUpFunc{
			defbultHook: i.Up,
		},
	}
}

// MigrbtorDownFunc describes the behbvior when the Down method of the
// pbrent MockMigrbtor instbnce is invoked.
type MigrbtorDownFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []MigrbtorDownFuncCbll
	mutex       sync.Mutex
}

// Down delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockMigrbtor) Down(v0 context.Context) error {
	r0 := m.DownFunc.nextHook()(v0)
	m.DownFunc.bppendCbll(MigrbtorDownFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Down method of the
// pbrent MockMigrbtor instbnce is invoked bnd the hook queue is empty.
func (f *MigrbtorDownFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Down method of the pbrent MockMigrbtor instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *MigrbtorDownFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *MigrbtorDownFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *MigrbtorDownFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *MigrbtorDownFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MigrbtorDownFunc) bppendCbll(r0 MigrbtorDownFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of MigrbtorDownFuncCbll objects describing the
// invocbtions of this function.
func (f *MigrbtorDownFunc) History() []MigrbtorDownFuncCbll {
	f.mutex.Lock()
	history := mbke([]MigrbtorDownFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MigrbtorDownFuncCbll is bn object thbt describes bn invocbtion of method
// Down on bn instbnce of MockMigrbtor.
type MigrbtorDownFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c MigrbtorDownFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c MigrbtorDownFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MigrbtorProgressFunc describes the behbvior when the Progress method of
// the pbrent MockMigrbtor instbnce is invoked.
type MigrbtorProgressFunc struct {
	defbultHook func(context.Context, bool) (flobt64, error)
	hooks       []func(context.Context, bool) (flobt64, error)
	history     []MigrbtorProgressFuncCbll
	mutex       sync.Mutex
}

// Progress delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockMigrbtor) Progress(v0 context.Context, v1 bool) (flobt64, error) {
	r0, r1 := m.ProgressFunc.nextHook()(v0, v1)
	m.ProgressFunc.bppendCbll(MigrbtorProgressFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Progress method of
// the pbrent MockMigrbtor instbnce is invoked bnd the hook queue is empty.
func (f *MigrbtorProgressFunc) SetDefbultHook(hook func(context.Context, bool) (flobt64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Progress method of the pbrent MockMigrbtor instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *MigrbtorProgressFunc) PushHook(hook func(context.Context, bool) (flobt64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *MigrbtorProgressFunc) SetDefbultReturn(r0 flobt64, r1 error) {
	f.SetDefbultHook(func(context.Context, bool) (flobt64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *MigrbtorProgressFunc) PushReturn(r0 flobt64, r1 error) {
	f.PushHook(func(context.Context, bool) (flobt64, error) {
		return r0, r1
	})
}

func (f *MigrbtorProgressFunc) nextHook() func(context.Context, bool) (flobt64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MigrbtorProgressFunc) bppendCbll(r0 MigrbtorProgressFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of MigrbtorProgressFuncCbll objects describing
// the invocbtions of this function.
func (f *MigrbtorProgressFunc) History() []MigrbtorProgressFuncCbll {
	f.mutex.Lock()
	history := mbke([]MigrbtorProgressFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MigrbtorProgressFuncCbll is bn object thbt describes bn invocbtion of
// method Progress on bn instbnce of MockMigrbtor.
type MigrbtorProgressFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 flobt64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c MigrbtorProgressFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c MigrbtorProgressFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MigrbtorUpFunc describes the behbvior when the Up method of the pbrent
// MockMigrbtor instbnce is invoked.
type MigrbtorUpFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []MigrbtorUpFuncCbll
	mutex       sync.Mutex
}

// Up delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockMigrbtor) Up(v0 context.Context) error {
	r0 := m.UpFunc.nextHook()(v0)
	m.UpFunc.bppendCbll(MigrbtorUpFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Up method of the
// pbrent MockMigrbtor instbnce is invoked bnd the hook queue is empty.
func (f *MigrbtorUpFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Up method of the pbrent MockMigrbtor instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *MigrbtorUpFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *MigrbtorUpFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *MigrbtorUpFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *MigrbtorUpFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *MigrbtorUpFunc) bppendCbll(r0 MigrbtorUpFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of MigrbtorUpFuncCbll objects describing the
// invocbtions of this function.
func (f *MigrbtorUpFunc) History() []MigrbtorUpFuncCbll {
	f.mutex.Lock()
	history := mbke([]MigrbtorUpFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// MigrbtorUpFuncCbll is bn object thbt describes bn invocbtion of method Up
// on bn instbnce of MockMigrbtor.
type MigrbtorUpFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c MigrbtorUpFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c MigrbtorUpFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockStoreIfbce is b mock implementbtion of the storeIfbce interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion)
// used for unit testing.
type MockStoreIfbce struct {
	// AddErrorFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method AddError.
	AddErrorFunc *StoreIfbceAddErrorFunc
	// DoneFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Done.
	DoneFunc *StoreIfbceDoneFunc
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *StoreIfbceListFunc
	// SynchronizeMetbdbtbFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SynchronizeMetbdbtb.
	SynchronizeMetbdbtbFunc *StoreIfbceSynchronizeMetbdbtbFunc
	// TrbnsbctFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Trbnsbct.
	TrbnsbctFunc *StoreIfbceTrbnsbctFunc
	// UpdbteDirectionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteDirection.
	UpdbteDirectionFunc *StoreIfbceUpdbteDirectionFunc
	// UpdbteProgressFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteProgress.
	UpdbteProgressFunc *StoreIfbceUpdbteProgressFunc
}

// NewMockStoreIfbce crebtes b new mock of the storeIfbce interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockStoreIfbce() *MockStoreIfbce {
	return &MockStoreIfbce{
		AddErrorFunc: &StoreIfbceAddErrorFunc{
			defbultHook: func(context.Context, int, string) (r0 error) {
				return
			},
		},
		DoneFunc: &StoreIfbceDoneFunc{
			defbultHook: func(error) (r0 error) {
				return
			},
		},
		ListFunc: &StoreIfbceListFunc{
			defbultHook: func(context.Context) (r0 []Migrbtion, r1 error) {
				return
			},
		},
		SynchronizeMetbdbtbFunc: &StoreIfbceSynchronizeMetbdbtbFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		TrbnsbctFunc: &StoreIfbceTrbnsbctFunc{
			defbultHook: func(context.Context) (r0 storeIfbce, r1 error) {
				return
			},
		},
		UpdbteDirectionFunc: &StoreIfbceUpdbteDirectionFunc{
			defbultHook: func(context.Context, int, bool) (r0 error) {
				return
			},
		},
		UpdbteProgressFunc: &StoreIfbceUpdbteProgressFunc{
			defbultHook: func(context.Context, int, flobt64) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockStoreIfbce crebtes b new mock of the storeIfbce interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockStoreIfbce() *MockStoreIfbce {
	return &MockStoreIfbce{
		AddErrorFunc: &StoreIfbceAddErrorFunc{
			defbultHook: func(context.Context, int, string) error {
				pbnic("unexpected invocbtion of MockStoreIfbce.AddError")
			},
		},
		DoneFunc: &StoreIfbceDoneFunc{
			defbultHook: func(error) error {
				pbnic("unexpected invocbtion of MockStoreIfbce.Done")
			},
		},
		ListFunc: &StoreIfbceListFunc{
			defbultHook: func(context.Context) ([]Migrbtion, error) {
				pbnic("unexpected invocbtion of MockStoreIfbce.List")
			},
		},
		SynchronizeMetbdbtbFunc: &StoreIfbceSynchronizeMetbdbtbFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockStoreIfbce.SynchronizeMetbdbtb")
			},
		},
		TrbnsbctFunc: &StoreIfbceTrbnsbctFunc{
			defbultHook: func(context.Context) (storeIfbce, error) {
				pbnic("unexpected invocbtion of MockStoreIfbce.Trbnsbct")
			},
		},
		UpdbteDirectionFunc: &StoreIfbceUpdbteDirectionFunc{
			defbultHook: func(context.Context, int, bool) error {
				pbnic("unexpected invocbtion of MockStoreIfbce.UpdbteDirection")
			},
		},
		UpdbteProgressFunc: &StoreIfbceUpdbteProgressFunc{
			defbultHook: func(context.Context, int, flobt64) error {
				pbnic("unexpected invocbtion of MockStoreIfbce.UpdbteProgress")
			},
		},
	}
}

// surrogbteMockStoreIfbce is b copy of the storeIfbce interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion). It is
// redefined here bs it is unexported in the source pbckbge.
type surrogbteMockStoreIfbce interfbce {
	AddError(context.Context, int, string) error
	Done(error) error
	List(context.Context) ([]Migrbtion, error)
	SynchronizeMetbdbtb(context.Context) error
	Trbnsbct(context.Context) (storeIfbce, error)
	UpdbteDirection(context.Context, int, bool) error
	UpdbteProgress(context.Context, int, flobt64) error
}

// NewMockStoreIfbceFrom crebtes b new mock of the MockStoreIfbce interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreIfbceFrom(i surrogbteMockStoreIfbce) *MockStoreIfbce {
	return &MockStoreIfbce{
		AddErrorFunc: &StoreIfbceAddErrorFunc{
			defbultHook: i.AddError,
		},
		DoneFunc: &StoreIfbceDoneFunc{
			defbultHook: i.Done,
		},
		ListFunc: &StoreIfbceListFunc{
			defbultHook: i.List,
		},
		SynchronizeMetbdbtbFunc: &StoreIfbceSynchronizeMetbdbtbFunc{
			defbultHook: i.SynchronizeMetbdbtb,
		},
		TrbnsbctFunc: &StoreIfbceTrbnsbctFunc{
			defbultHook: i.Trbnsbct,
		},
		UpdbteDirectionFunc: &StoreIfbceUpdbteDirectionFunc{
			defbultHook: i.UpdbteDirection,
		},
		UpdbteProgressFunc: &StoreIfbceUpdbteProgressFunc{
			defbultHook: i.UpdbteProgress,
		},
	}
}

// StoreIfbceAddErrorFunc describes the behbvior when the AddError method of
// the pbrent MockStoreIfbce instbnce is invoked.
type StoreIfbceAddErrorFunc struct {
	defbultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []StoreIfbceAddErrorFuncCbll
	mutex       sync.Mutex
}

// AddError delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) AddError(v0 context.Context, v1 int, v2 string) error {
	r0 := m.AddErrorFunc.nextHook()(v0, v1, v2)
	m.AddErrorFunc.bppendCbll(StoreIfbceAddErrorFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the AddError method of
// the pbrent MockStoreIfbce instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreIfbceAddErrorFunc) SetDefbultHook(hook func(context.Context, int, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddError method of the pbrent MockStoreIfbce instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreIfbceAddErrorFunc) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceAddErrorFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceAddErrorFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *StoreIfbceAddErrorFunc) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceAddErrorFunc) bppendCbll(r0 StoreIfbceAddErrorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceAddErrorFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreIfbceAddErrorFunc) History() []StoreIfbceAddErrorFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceAddErrorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceAddErrorFuncCbll is bn object thbt describes bn invocbtion of
// method AddError on bn instbnce of MockStoreIfbce.
type StoreIfbceAddErrorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceAddErrorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceAddErrorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreIfbceDoneFunc describes the behbvior when the Done method of the
// pbrent MockStoreIfbce instbnce is invoked.
type StoreIfbceDoneFunc struct {
	defbultHook func(error) error
	hooks       []func(error) error
	history     []StoreIfbceDoneFuncCbll
	mutex       sync.Mutex
}

// Done delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.bppendCbll(StoreIfbceDoneFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Done method of the
// pbrent MockStoreIfbce instbnce is invoked bnd the hook queue is empty.
func (f *StoreIfbceDoneFunc) SetDefbultHook(hook func(error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Done method of the pbrent MockStoreIfbce instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreIfbceDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceDoneFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *StoreIfbceDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceDoneFunc) bppendCbll(r0 StoreIfbceDoneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceDoneFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreIfbceDoneFunc) History() []StoreIfbceDoneFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceDoneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceDoneFuncCbll is bn object thbt describes bn invocbtion of
// method Done on bn instbnce of MockStoreIfbce.
type StoreIfbceDoneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceDoneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceDoneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreIfbceListFunc describes the behbvior when the List method of the
// pbrent MockStoreIfbce instbnce is invoked.
type StoreIfbceListFunc struct {
	defbultHook func(context.Context) ([]Migrbtion, error)
	hooks       []func(context.Context) ([]Migrbtion, error)
	history     []StoreIfbceListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) List(v0 context.Context) ([]Migrbtion, error) {
	r0, r1 := m.ListFunc.nextHook()(v0)
	m.ListFunc.bppendCbll(StoreIfbceListFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockStoreIfbce instbnce is invoked bnd the hook queue is empty.
func (f *StoreIfbceListFunc) SetDefbultHook(hook func(context.Context) ([]Migrbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockStoreIfbce instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreIfbceListFunc) PushHook(hook func(context.Context) ([]Migrbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceListFunc) SetDefbultReturn(r0 []Migrbtion, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]Migrbtion, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceListFunc) PushReturn(r0 []Migrbtion, r1 error) {
	f.PushHook(func(context.Context) ([]Migrbtion, error) {
		return r0, r1
	})
}

func (f *StoreIfbceListFunc) nextHook() func(context.Context) ([]Migrbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceListFunc) bppendCbll(r0 StoreIfbceListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceListFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreIfbceListFunc) History() []StoreIfbceListFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceListFuncCbll is bn object thbt describes bn invocbtion of
// method List on bn instbnce of MockStoreIfbce.
type StoreIfbceListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []Migrbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreIfbceSynchronizeMetbdbtbFunc describes the behbvior when the
// SynchronizeMetbdbtb method of the pbrent MockStoreIfbce instbnce is
// invoked.
type StoreIfbceSynchronizeMetbdbtbFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []StoreIfbceSynchronizeMetbdbtbFuncCbll
	mutex       sync.Mutex
}

// SynchronizeMetbdbtb delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) SynchronizeMetbdbtb(v0 context.Context) error {
	r0 := m.SynchronizeMetbdbtbFunc.nextHook()(v0)
	m.SynchronizeMetbdbtbFunc.bppendCbll(StoreIfbceSynchronizeMetbdbtbFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SynchronizeMetbdbtb
// method of the pbrent MockStoreIfbce instbnce is invoked bnd the hook
// queue is empty.
func (f *StoreIfbceSynchronizeMetbdbtbFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SynchronizeMetbdbtb method of the pbrent MockStoreIfbce instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreIfbceSynchronizeMetbdbtbFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceSynchronizeMetbdbtbFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceSynchronizeMetbdbtbFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *StoreIfbceSynchronizeMetbdbtbFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceSynchronizeMetbdbtbFunc) bppendCbll(r0 StoreIfbceSynchronizeMetbdbtbFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceSynchronizeMetbdbtbFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreIfbceSynchronizeMetbdbtbFunc) History() []StoreIfbceSynchronizeMetbdbtbFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceSynchronizeMetbdbtbFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceSynchronizeMetbdbtbFuncCbll is bn object thbt describes bn
// invocbtion of method SynchronizeMetbdbtb on bn instbnce of
// MockStoreIfbce.
type StoreIfbceSynchronizeMetbdbtbFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceSynchronizeMetbdbtbFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceSynchronizeMetbdbtbFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreIfbceTrbnsbctFunc describes the behbvior when the Trbnsbct method of
// the pbrent MockStoreIfbce instbnce is invoked.
type StoreIfbceTrbnsbctFunc struct {
	defbultHook func(context.Context) (storeIfbce, error)
	hooks       []func(context.Context) (storeIfbce, error)
	history     []StoreIfbceTrbnsbctFuncCbll
	mutex       sync.Mutex
}

// Trbnsbct delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) Trbnsbct(v0 context.Context) (storeIfbce, error) {
	r0, r1 := m.TrbnsbctFunc.nextHook()(v0)
	m.TrbnsbctFunc.bppendCbll(StoreIfbceTrbnsbctFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Trbnsbct method of
// the pbrent MockStoreIfbce instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreIfbceTrbnsbctFunc) SetDefbultHook(hook func(context.Context) (storeIfbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Trbnsbct method of the pbrent MockStoreIfbce instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreIfbceTrbnsbctFunc) PushHook(hook func(context.Context) (storeIfbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceTrbnsbctFunc) SetDefbultReturn(r0 storeIfbce, r1 error) {
	f.SetDefbultHook(func(context.Context) (storeIfbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceTrbnsbctFunc) PushReturn(r0 storeIfbce, r1 error) {
	f.PushHook(func(context.Context) (storeIfbce, error) {
		return r0, r1
	})
}

func (f *StoreIfbceTrbnsbctFunc) nextHook() func(context.Context) (storeIfbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceTrbnsbctFunc) bppendCbll(r0 StoreIfbceTrbnsbctFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceTrbnsbctFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreIfbceTrbnsbctFunc) History() []StoreIfbceTrbnsbctFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceTrbnsbctFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceTrbnsbctFuncCbll is bn object thbt describes bn invocbtion of
// method Trbnsbct on bn instbnce of MockStoreIfbce.
type StoreIfbceTrbnsbctFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 storeIfbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceTrbnsbctFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceTrbnsbctFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreIfbceUpdbteDirectionFunc describes the behbvior when the
// UpdbteDirection method of the pbrent MockStoreIfbce instbnce is invoked.
type StoreIfbceUpdbteDirectionFunc struct {
	defbultHook func(context.Context, int, bool) error
	hooks       []func(context.Context, int, bool) error
	history     []StoreIfbceUpdbteDirectionFuncCbll
	mutex       sync.Mutex
}

// UpdbteDirection delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) UpdbteDirection(v0 context.Context, v1 int, v2 bool) error {
	r0 := m.UpdbteDirectionFunc.nextHook()(v0, v1, v2)
	m.UpdbteDirectionFunc.bppendCbll(StoreIfbceUpdbteDirectionFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UpdbteDirection
// method of the pbrent MockStoreIfbce instbnce is invoked bnd the hook
// queue is empty.
func (f *StoreIfbceUpdbteDirectionFunc) SetDefbultHook(hook func(context.Context, int, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteDirection method of the pbrent MockStoreIfbce instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreIfbceUpdbteDirectionFunc) PushHook(hook func(context.Context, int, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceUpdbteDirectionFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceUpdbteDirectionFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, bool) error {
		return r0
	})
}

func (f *StoreIfbceUpdbteDirectionFunc) nextHook() func(context.Context, int, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceUpdbteDirectionFunc) bppendCbll(r0 StoreIfbceUpdbteDirectionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceUpdbteDirectionFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreIfbceUpdbteDirectionFunc) History() []StoreIfbceUpdbteDirectionFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceUpdbteDirectionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceUpdbteDirectionFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteDirection on bn instbnce of MockStoreIfbce.
type StoreIfbceUpdbteDirectionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceUpdbteDirectionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceUpdbteDirectionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreIfbceUpdbteProgressFunc describes the behbvior when the
// UpdbteProgress method of the pbrent MockStoreIfbce instbnce is invoked.
type StoreIfbceUpdbteProgressFunc struct {
	defbultHook func(context.Context, int, flobt64) error
	hooks       []func(context.Context, int, flobt64) error
	history     []StoreIfbceUpdbteProgressFuncCbll
	mutex       sync.Mutex
}

// UpdbteProgress delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStoreIfbce) UpdbteProgress(v0 context.Context, v1 int, v2 flobt64) error {
	r0 := m.UpdbteProgressFunc.nextHook()(v0, v1, v2)
	m.UpdbteProgressFunc.bppendCbll(StoreIfbceUpdbteProgressFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UpdbteProgress
// method of the pbrent MockStoreIfbce instbnce is invoked bnd the hook
// queue is empty.
func (f *StoreIfbceUpdbteProgressFunc) SetDefbultHook(hook func(context.Context, int, flobt64) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteProgress method of the pbrent MockStoreIfbce instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreIfbceUpdbteProgressFunc) PushHook(hook func(context.Context, int, flobt64) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIfbceUpdbteProgressFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, flobt64) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIfbceUpdbteProgressFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, flobt64) error {
		return r0
	})
}

func (f *StoreIfbceUpdbteProgressFunc) nextHook() func(context.Context, int, flobt64) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIfbceUpdbteProgressFunc) bppendCbll(r0 StoreIfbceUpdbteProgressFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIfbceUpdbteProgressFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreIfbceUpdbteProgressFunc) History() []StoreIfbceUpdbteProgressFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIfbceUpdbteProgressFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIfbceUpdbteProgressFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteProgress on bn instbnce of MockStoreIfbce.
type StoreIfbceUpdbteProgressFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 flobt64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIfbceUpdbteProgressFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIfbceUpdbteProgressFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
