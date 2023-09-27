// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge http

import (
	"context"
	"sync"

	uplobdhbndler "github.com/sourcegrbph/sourcegrbph/internbl/uplobdhbndler"
)

// MockDBStore is b mock implementbtion of the DBStore interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdhbndler) used
// for unit testing.
type MockDBStore[T interfbce{}] struct {
	// AddUplobdPbrtFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AddUplobdPbrt.
	AddUplobdPbrtFunc *DBStoreAddUplobdPbrtFunc[T]
	// GetUplobdByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetUplobdByID.
	GetUplobdByIDFunc *DBStoreGetUplobdByIDFunc[T]
	// InsertUplobdFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method InsertUplobd.
	InsertUplobdFunc *DBStoreInsertUplobdFunc[T]
	// MbrkFbiledFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkFbiled.
	MbrkFbiledFunc *DBStoreMbrkFbiledFunc[T]
	// MbrkQueuedFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkQueued.
	MbrkQueuedFunc *DBStoreMbrkQueuedFunc[T]
	// WithTrbnsbctionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithTrbnsbction.
	WithTrbnsbctionFunc *DBStoreWithTrbnsbctionFunc[T]
}

// NewMockDBStore crebtes b new mock of the DBStore interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockDBStore[T interfbce{}]() *MockDBStore[T] {
	return &MockDBStore[T]{
		AddUplobdPbrtFunc: &DBStoreAddUplobdPbrtFunc[T]{
			defbultHook: func(context.Context, int, int) (r0 error) {
				return
			},
		},
		GetUplobdByIDFunc: &DBStoreGetUplobdByIDFunc[T]{
			defbultHook: func(context.Context, int) (r0 uplobdhbndler.Uplobd[T], r1 bool, r2 error) {
				return
			},
		},
		InsertUplobdFunc: &DBStoreInsertUplobdFunc[T]{
			defbultHook: func(context.Context, uplobdhbndler.Uplobd[T]) (r0 int, r1 error) {
				return
			},
		},
		MbrkFbiledFunc: &DBStoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, int, string) (r0 error) {
				return
			},
		},
		MbrkQueuedFunc: &DBStoreMbrkQueuedFunc[T]{
			defbultHook: func(context.Context, int, *int64) (r0 error) {
				return
			},
		},
		WithTrbnsbctionFunc: &DBStoreWithTrbnsbctionFunc[T]{
			defbultHook: func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockDBStore crebtes b new mock of the DBStore interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockDBStore[T interfbce{}]() *MockDBStore[T] {
	return &MockDBStore[T]{
		AddUplobdPbrtFunc: &DBStoreAddUplobdPbrtFunc[T]{
			defbultHook: func(context.Context, int, int) error {
				pbnic("unexpected invocbtion of MockDBStore.AddUplobdPbrt")
			},
		},
		GetUplobdByIDFunc: &DBStoreGetUplobdByIDFunc[T]{
			defbultHook: func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error) {
				pbnic("unexpected invocbtion of MockDBStore.GetUplobdByID")
			},
		},
		InsertUplobdFunc: &DBStoreInsertUplobdFunc[T]{
			defbultHook: func(context.Context, uplobdhbndler.Uplobd[T]) (int, error) {
				pbnic("unexpected invocbtion of MockDBStore.InsertUplobd")
			},
		},
		MbrkFbiledFunc: &DBStoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, int, string) error {
				pbnic("unexpected invocbtion of MockDBStore.MbrkFbiled")
			},
		},
		MbrkQueuedFunc: &DBStoreMbrkQueuedFunc[T]{
			defbultHook: func(context.Context, int, *int64) error {
				pbnic("unexpected invocbtion of MockDBStore.MbrkQueued")
			},
		},
		WithTrbnsbctionFunc: &DBStoreWithTrbnsbctionFunc[T]{
			defbultHook: func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error {
				pbnic("unexpected invocbtion of MockDBStore.WithTrbnsbction")
			},
		},
	}
}

// NewMockDBStoreFrom crebtes b new mock of the MockDBStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockDBStoreFrom[T interfbce{}](i uplobdhbndler.DBStore[T]) *MockDBStore[T] {
	return &MockDBStore[T]{
		AddUplobdPbrtFunc: &DBStoreAddUplobdPbrtFunc[T]{
			defbultHook: i.AddUplobdPbrt,
		},
		GetUplobdByIDFunc: &DBStoreGetUplobdByIDFunc[T]{
			defbultHook: i.GetUplobdByID,
		},
		InsertUplobdFunc: &DBStoreInsertUplobdFunc[T]{
			defbultHook: i.InsertUplobd,
		},
		MbrkFbiledFunc: &DBStoreMbrkFbiledFunc[T]{
			defbultHook: i.MbrkFbiled,
		},
		MbrkQueuedFunc: &DBStoreMbrkQueuedFunc[T]{
			defbultHook: i.MbrkQueued,
		},
		WithTrbnsbctionFunc: &DBStoreWithTrbnsbctionFunc[T]{
			defbultHook: i.WithTrbnsbction,
		},
	}
}

// DBStoreAddUplobdPbrtFunc describes the behbvior when the AddUplobdPbrt
// method of the pbrent MockDBStore instbnce is invoked.
type DBStoreAddUplobdPbrtFunc[T interfbce{}] struct {
	defbultHook func(context.Context, int, int) error
	hooks       []func(context.Context, int, int) error
	history     []DBStoreAddUplobdPbrtFuncCbll[T]
	mutex       sync.Mutex
}

// AddUplobdPbrt delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDBStore[T]) AddUplobdPbrt(v0 context.Context, v1 int, v2 int) error {
	r0 := m.AddUplobdPbrtFunc.nextHook()(v0, v1, v2)
	m.AddUplobdPbrtFunc.bppendCbll(DBStoreAddUplobdPbrtFuncCbll[T]{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the AddUplobdPbrt method
// of the pbrent MockDBStore instbnce is invoked bnd the hook queue is
// empty.
func (f *DBStoreAddUplobdPbrtFunc[T]) SetDefbultHook(hook func(context.Context, int, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddUplobdPbrt method of the pbrent MockDBStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *DBStoreAddUplobdPbrtFunc[T]) PushHook(hook func(context.Context, int, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DBStoreAddUplobdPbrtFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DBStoreAddUplobdPbrtFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int) error {
		return r0
	})
}

func (f *DBStoreAddUplobdPbrtFunc[T]) nextHook() func(context.Context, int, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreAddUplobdPbrtFunc[T]) bppendCbll(r0 DBStoreAddUplobdPbrtFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DBStoreAddUplobdPbrtFuncCbll objects
// describing the invocbtions of this function.
func (f *DBStoreAddUplobdPbrtFunc[T]) History() []DBStoreAddUplobdPbrtFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]DBStoreAddUplobdPbrtFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreAddUplobdPbrtFuncCbll is bn object thbt describes bn invocbtion of
// method AddUplobdPbrt on bn instbnce of MockDBStore.
type DBStoreAddUplobdPbrtFuncCbll[T interfbce{}] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DBStoreAddUplobdPbrtFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DBStoreAddUplobdPbrtFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// DBStoreGetUplobdByIDFunc describes the behbvior when the GetUplobdByID
// method of the pbrent MockDBStore instbnce is invoked.
type DBStoreGetUplobdByIDFunc[T interfbce{}] struct {
	defbultHook func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error)
	hooks       []func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error)
	history     []DBStoreGetUplobdByIDFuncCbll[T]
	mutex       sync.Mutex
}

// GetUplobdByID delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDBStore[T]) GetUplobdByID(v0 context.Context, v1 int) (uplobdhbndler.Uplobd[T], bool, error) {
	r0, r1, r2 := m.GetUplobdByIDFunc.nextHook()(v0, v1)
	m.GetUplobdByIDFunc.bppendCbll(DBStoreGetUplobdByIDFuncCbll[T]{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetUplobdByID method
// of the pbrent MockDBStore instbnce is invoked bnd the hook queue is
// empty.
func (f *DBStoreGetUplobdByIDFunc[T]) SetDefbultHook(hook func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobdByID method of the pbrent MockDBStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *DBStoreGetUplobdByIDFunc[T]) PushHook(hook func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DBStoreGetUplobdByIDFunc[T]) SetDefbultReturn(r0 uplobdhbndler.Uplobd[T], r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DBStoreGetUplobdByIDFunc[T]) PushReturn(r0 uplobdhbndler.Uplobd[T], r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error) {
		return r0, r1, r2
	})
}

func (f *DBStoreGetUplobdByIDFunc[T]) nextHook() func(context.Context, int) (uplobdhbndler.Uplobd[T], bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreGetUplobdByIDFunc[T]) bppendCbll(r0 DBStoreGetUplobdByIDFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DBStoreGetUplobdByIDFuncCbll objects
// describing the invocbtions of this function.
func (f *DBStoreGetUplobdByIDFunc[T]) History() []DBStoreGetUplobdByIDFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]DBStoreGetUplobdByIDFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreGetUplobdByIDFuncCbll is bn object thbt describes bn invocbtion of
// method GetUplobdByID on bn instbnce of MockDBStore.
type DBStoreGetUplobdByIDFuncCbll[T interfbce{}] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 uplobdhbndler.Uplobd[T]
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DBStoreGetUplobdByIDFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DBStoreGetUplobdByIDFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// DBStoreInsertUplobdFunc describes the behbvior when the InsertUplobd
// method of the pbrent MockDBStore instbnce is invoked.
type DBStoreInsertUplobdFunc[T interfbce{}] struct {
	defbultHook func(context.Context, uplobdhbndler.Uplobd[T]) (int, error)
	hooks       []func(context.Context, uplobdhbndler.Uplobd[T]) (int, error)
	history     []DBStoreInsertUplobdFuncCbll[T]
	mutex       sync.Mutex
}

// InsertUplobd delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDBStore[T]) InsertUplobd(v0 context.Context, v1 uplobdhbndler.Uplobd[T]) (int, error) {
	r0, r1 := m.InsertUplobdFunc.nextHook()(v0, v1)
	m.InsertUplobdFunc.bppendCbll(DBStoreInsertUplobdFuncCbll[T]{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the InsertUplobd method
// of the pbrent MockDBStore instbnce is invoked bnd the hook queue is
// empty.
func (f *DBStoreInsertUplobdFunc[T]) SetDefbultHook(hook func(context.Context, uplobdhbndler.Uplobd[T]) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertUplobd method of the pbrent MockDBStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *DBStoreInsertUplobdFunc[T]) PushHook(hook func(context.Context, uplobdhbndler.Uplobd[T]) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DBStoreInsertUplobdFunc[T]) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, uplobdhbndler.Uplobd[T]) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DBStoreInsertUplobdFunc[T]) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, uplobdhbndler.Uplobd[T]) (int, error) {
		return r0, r1
	})
}

func (f *DBStoreInsertUplobdFunc[T]) nextHook() func(context.Context, uplobdhbndler.Uplobd[T]) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreInsertUplobdFunc[T]) bppendCbll(r0 DBStoreInsertUplobdFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DBStoreInsertUplobdFuncCbll objects
// describing the invocbtions of this function.
func (f *DBStoreInsertUplobdFunc[T]) History() []DBStoreInsertUplobdFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]DBStoreInsertUplobdFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreInsertUplobdFuncCbll is bn object thbt describes bn invocbtion of
// method InsertUplobd on bn instbnce of MockDBStore.
type DBStoreInsertUplobdFuncCbll[T interfbce{}] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 uplobdhbndler.Uplobd[T]
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DBStoreInsertUplobdFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DBStoreInsertUplobdFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DBStoreMbrkFbiledFunc describes the behbvior when the MbrkFbiled method
// of the pbrent MockDBStore instbnce is invoked.
type DBStoreMbrkFbiledFunc[T interfbce{}] struct {
	defbultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []DBStoreMbrkFbiledFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkFbiled delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDBStore[T]) MbrkFbiled(v0 context.Context, v1 int, v2 string) error {
	r0 := m.MbrkFbiledFunc.nextHook()(v0, v1, v2)
	m.MbrkFbiledFunc.bppendCbll(DBStoreMbrkFbiledFuncCbll[T]{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the MbrkFbiled method of
// the pbrent MockDBStore instbnce is invoked bnd the hook queue is empty.
func (f *DBStoreMbrkFbiledFunc[T]) SetDefbultHook(hook func(context.Context, int, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkFbiled method of the pbrent MockDBStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *DBStoreMbrkFbiledFunc[T]) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DBStoreMbrkFbiledFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DBStoreMbrkFbiledFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *DBStoreMbrkFbiledFunc[T]) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreMbrkFbiledFunc[T]) bppendCbll(r0 DBStoreMbrkFbiledFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DBStoreMbrkFbiledFuncCbll objects
// describing the invocbtions of this function.
func (f *DBStoreMbrkFbiledFunc[T]) History() []DBStoreMbrkFbiledFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]DBStoreMbrkFbiledFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreMbrkFbiledFuncCbll is bn object thbt describes bn invocbtion of
// method MbrkFbiled on bn instbnce of MockDBStore.
type DBStoreMbrkFbiledFuncCbll[T interfbce{}] struct {
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
func (c DBStoreMbrkFbiledFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DBStoreMbrkFbiledFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// DBStoreMbrkQueuedFunc describes the behbvior when the MbrkQueued method
// of the pbrent MockDBStore instbnce is invoked.
type DBStoreMbrkQueuedFunc[T interfbce{}] struct {
	defbultHook func(context.Context, int, *int64) error
	hooks       []func(context.Context, int, *int64) error
	history     []DBStoreMbrkQueuedFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkQueued delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDBStore[T]) MbrkQueued(v0 context.Context, v1 int, v2 *int64) error {
	r0 := m.MbrkQueuedFunc.nextHook()(v0, v1, v2)
	m.MbrkQueuedFunc.bppendCbll(DBStoreMbrkQueuedFuncCbll[T]{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the MbrkQueued method of
// the pbrent MockDBStore instbnce is invoked bnd the hook queue is empty.
func (f *DBStoreMbrkQueuedFunc[T]) SetDefbultHook(hook func(context.Context, int, *int64) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkQueued method of the pbrent MockDBStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *DBStoreMbrkQueuedFunc[T]) PushHook(hook func(context.Context, int, *int64) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DBStoreMbrkQueuedFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, *int64) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DBStoreMbrkQueuedFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, *int64) error {
		return r0
	})
}

func (f *DBStoreMbrkQueuedFunc[T]) nextHook() func(context.Context, int, *int64) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreMbrkQueuedFunc[T]) bppendCbll(r0 DBStoreMbrkQueuedFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DBStoreMbrkQueuedFuncCbll objects
// describing the invocbtions of this function.
func (f *DBStoreMbrkQueuedFunc[T]) History() []DBStoreMbrkQueuedFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]DBStoreMbrkQueuedFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreMbrkQueuedFuncCbll is bn object thbt describes bn invocbtion of
// method MbrkQueued on bn instbnce of MockDBStore.
type DBStoreMbrkQueuedFuncCbll[T interfbce{}] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DBStoreMbrkQueuedFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DBStoreMbrkQueuedFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// DBStoreWithTrbnsbctionFunc describes the behbvior when the
// WithTrbnsbction method of the pbrent MockDBStore instbnce is invoked.
type DBStoreWithTrbnsbctionFunc[T interfbce{}] struct {
	defbultHook func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error
	hooks       []func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error
	history     []DBStoreWithTrbnsbctionFuncCbll[T]
	mutex       sync.Mutex
}

// WithTrbnsbction delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDBStore[T]) WithTrbnsbction(v0 context.Context, v1 func(tx uplobdhbndler.DBStore[T]) error) error {
	r0 := m.WithTrbnsbctionFunc.nextHook()(v0, v1)
	m.WithTrbnsbctionFunc.bppendCbll(DBStoreWithTrbnsbctionFuncCbll[T]{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithTrbnsbction
// method of the pbrent MockDBStore instbnce is invoked bnd the hook queue
// is empty.
func (f *DBStoreWithTrbnsbctionFunc[T]) SetDefbultHook(hook func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithTrbnsbction method of the pbrent MockDBStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *DBStoreWithTrbnsbctionFunc[T]) PushHook(hook func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DBStoreWithTrbnsbctionFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DBStoreWithTrbnsbctionFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error {
		return r0
	})
}

func (f *DBStoreWithTrbnsbctionFunc[T]) nextHook() func(context.Context, func(tx uplobdhbndler.DBStore[T]) error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DBStoreWithTrbnsbctionFunc[T]) bppendCbll(r0 DBStoreWithTrbnsbctionFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DBStoreWithTrbnsbctionFuncCbll objects
// describing the invocbtions of this function.
func (f *DBStoreWithTrbnsbctionFunc[T]) History() []DBStoreWithTrbnsbctionFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]DBStoreWithTrbnsbctionFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DBStoreWithTrbnsbctionFuncCbll is bn object thbt describes bn invocbtion
// of method WithTrbnsbction on bn instbnce of MockDBStore.
type DBStoreWithTrbnsbctionFuncCbll[T interfbce{}] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 func(tx uplobdhbndler.DBStore[T]) error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DBStoreWithTrbnsbctionFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DBStoreWithTrbnsbctionFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
