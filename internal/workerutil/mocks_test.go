// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge workerutil

import (
	"context"
	"sync"

	log "github.com/sourcegrbph/log"
)

// MockHbndler is b mock implementbtion of the Hbndler interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/workerutil) used for
// unit testing.
type MockHbndler[T Record] struct {
	// HbndleFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hbndle.
	HbndleFunc *HbndlerHbndleFunc[T]
}

// NewMockHbndler crebtes b new mock of the Hbndler interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockHbndler[T Record]() *MockHbndler[T] {
	return &MockHbndler[T]{
		HbndleFunc: &HbndlerHbndleFunc[T]{
			defbultHook: func(context.Context, log.Logger, T) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockHbndler crebtes b new mock of the Hbndler interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockHbndler[T Record]() *MockHbndler[T] {
	return &MockHbndler[T]{
		HbndleFunc: &HbndlerHbndleFunc[T]{
			defbultHook: func(context.Context, log.Logger, T) error {
				pbnic("unexpected invocbtion of MockHbndler.Hbndle")
			},
		},
	}
}

// NewMockHbndlerFrom crebtes b new mock of the MockHbndler interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockHbndlerFrom[T Record](i Hbndler[T]) *MockHbndler[T] {
	return &MockHbndler[T]{
		HbndleFunc: &HbndlerHbndleFunc[T]{
			defbultHook: i.Hbndle,
		},
	}
}

// HbndlerHbndleFunc describes the behbvior when the Hbndle method of the
// pbrent MockHbndler instbnce is invoked.
type HbndlerHbndleFunc[T Record] struct {
	defbultHook func(context.Context, log.Logger, T) error
	hooks       []func(context.Context, log.Logger, T) error
	history     []HbndlerHbndleFuncCbll[T]
	mutex       sync.Mutex
}

// Hbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockHbndler[T]) Hbndle(v0 context.Context, v1 log.Logger, v2 T) error {
	r0 := m.HbndleFunc.nextHook()(v0, v1, v2)
	m.HbndleFunc.bppendCbll(HbndlerHbndleFuncCbll[T]{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Hbndle method of the
// pbrent MockHbndler instbnce is invoked bnd the hook queue is empty.
func (f *HbndlerHbndleFunc[T]) SetDefbultHook(hook func(context.Context, log.Logger, T) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hbndle method of the pbrent MockHbndler instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *HbndlerHbndleFunc[T]) PushHook(hook func(context.Context, log.Logger, T) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *HbndlerHbndleFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, log.Logger, T) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *HbndlerHbndleFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, log.Logger, T) error {
		return r0
	})
}

func (f *HbndlerHbndleFunc[T]) nextHook() func(context.Context, log.Logger, T) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *HbndlerHbndleFunc[T]) bppendCbll(r0 HbndlerHbndleFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of HbndlerHbndleFuncCbll objects describing
// the invocbtions of this function.
func (f *HbndlerHbndleFunc[T]) History() []HbndlerHbndleFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]HbndlerHbndleFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// HbndlerHbndleFuncCbll is bn object thbt describes bn invocbtion of method
// Hbndle on bn instbnce of MockHbndler.
type HbndlerHbndleFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 log.Logger
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 T
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c HbndlerHbndleFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c HbndlerHbndleFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/workerutil) used for
// unit testing.
type MockStore[T Record] struct {
	// DequeueFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Dequeue.
	DequeueFunc *StoreDequeueFunc[T]
	// HebrtbebtFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method Hebrtbebt.
	HebrtbebtFunc *StoreHebrtbebtFunc[T]
	// MbrkCompleteFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkComplete.
	MbrkCompleteFunc *StoreMbrkCompleteFunc[T]
	// MbrkErroredFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkErrored.
	MbrkErroredFunc *StoreMbrkErroredFunc[T]
	// MbrkFbiledFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbrkFbiled.
	MbrkFbiledFunc *StoreMbrkFbiledFunc[T]
	// QueuedCountFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method QueuedCount.
	QueuedCountFunc *StoreQueuedCountFunc[T]
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore[T Record]() *MockStore[T] {
	return &MockStore[T]{
		DequeueFunc: &StoreDequeueFunc[T]{
			defbultHook: func(context.Context, string, interfbce{}) (r0 T, r1 bool, r2 error) {
				return
			},
		},
		HebrtbebtFunc: &StoreHebrtbebtFunc[T]{
			defbultHook: func(context.Context, []string) (r0 []string, r1 []string, r2 error) {
				return
			},
		},
		MbrkCompleteFunc: &StoreMbrkCompleteFunc[T]{
			defbultHook: func(context.Context, T) (r0 bool, r1 error) {
				return
			},
		},
		MbrkErroredFunc: &StoreMbrkErroredFunc[T]{
			defbultHook: func(context.Context, T, string) (r0 bool, r1 error) {
				return
			},
		},
		MbrkFbiledFunc: &StoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, T, string) (r0 bool, r1 error) {
				return
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc[T]{
			defbultHook: func(context.Context) (r0 int, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore[T Record]() *MockStore[T] {
	return &MockStore[T]{
		DequeueFunc: &StoreDequeueFunc[T]{
			defbultHook: func(context.Context, string, interfbce{}) (T, bool, error) {
				pbnic("unexpected invocbtion of MockStore.Dequeue")
			},
		},
		HebrtbebtFunc: &StoreHebrtbebtFunc[T]{
			defbultHook: func(context.Context, []string) ([]string, []string, error) {
				pbnic("unexpected invocbtion of MockStore.Hebrtbebt")
			},
		},
		MbrkCompleteFunc: &StoreMbrkCompleteFunc[T]{
			defbultHook: func(context.Context, T) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.MbrkComplete")
			},
		},
		MbrkErroredFunc: &StoreMbrkErroredFunc[T]{
			defbultHook: func(context.Context, T, string) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.MbrkErrored")
			},
		},
		MbrkFbiledFunc: &StoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, T, string) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.MbrkFbiled")
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc[T]{
			defbultHook: func(context.Context) (int, error) {
				pbnic("unexpected invocbtion of MockStore.QueuedCount")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom[T Record](i Store[T]) *MockStore[T] {
	return &MockStore[T]{
		DequeueFunc: &StoreDequeueFunc[T]{
			defbultHook: i.Dequeue,
		},
		HebrtbebtFunc: &StoreHebrtbebtFunc[T]{
			defbultHook: i.Hebrtbebt,
		},
		MbrkCompleteFunc: &StoreMbrkCompleteFunc[T]{
			defbultHook: i.MbrkComplete,
		},
		MbrkErroredFunc: &StoreMbrkErroredFunc[T]{
			defbultHook: i.MbrkErrored,
		},
		MbrkFbiledFunc: &StoreMbrkFbiledFunc[T]{
			defbultHook: i.MbrkFbiled,
		},
		QueuedCountFunc: &StoreQueuedCountFunc[T]{
			defbultHook: i.QueuedCount,
		},
	}
}

// StoreDequeueFunc describes the behbvior when the Dequeue method of the
// pbrent MockStore instbnce is invoked.
type StoreDequeueFunc[T Record] struct {
	defbultHook func(context.Context, string, interfbce{}) (T, bool, error)
	hooks       []func(context.Context, string, interfbce{}) (T, bool, error)
	history     []StoreDequeueFuncCbll[T]
	mutex       sync.Mutex
}

// Dequeue delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) Dequeue(v0 context.Context, v1 string, v2 interfbce{}) (T, bool, error) {
	r0, r1, r2 := m.DequeueFunc.nextHook()(v0, v1, v2)
	m.DequeueFunc.bppendCbll(StoreDequeueFuncCbll[T]{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Dequeue method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDequeueFunc[T]) SetDefbultHook(hook func(context.Context, string, interfbce{}) (T, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Dequeue method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDequeueFunc[T]) PushHook(hook func(context.Context, string, interfbce{}) (T, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDequeueFunc[T]) SetDefbultReturn(r0 T, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, interfbce{}) (T, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDequeueFunc[T]) PushReturn(r0 T, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, interfbce{}) (T, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreDequeueFunc[T]) nextHook() func(context.Context, string, interfbce{}) (T, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDequeueFunc[T]) bppendCbll(r0 StoreDequeueFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDequeueFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreDequeueFunc[T]) History() []StoreDequeueFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreDequeueFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDequeueFuncCbll is bn object thbt describes bn invocbtion of method
// Dequeue on bn instbnce of MockStore.
type StoreDequeueFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 T
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDequeueFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDequeueFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreHebrtbebtFunc describes the behbvior when the Hebrtbebt method of
// the pbrent MockStore instbnce is invoked.
type StoreHebrtbebtFunc[T Record] struct {
	defbultHook func(context.Context, []string) ([]string, []string, error)
	hooks       []func(context.Context, []string) ([]string, []string, error)
	history     []StoreHebrtbebtFuncCbll[T]
	mutex       sync.Mutex
}

// Hebrtbebt delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) Hebrtbebt(v0 context.Context, v1 []string) ([]string, []string, error) {
	r0, r1, r2 := m.HebrtbebtFunc.nextHook()(v0, v1)
	m.HebrtbebtFunc.bppendCbll(StoreHebrtbebtFuncCbll[T]{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Hebrtbebt method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreHebrtbebtFunc[T]) SetDefbultHook(hook func(context.Context, []string) ([]string, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hebrtbebt method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreHebrtbebtFunc[T]) PushHook(hook func(context.Context, []string) ([]string, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreHebrtbebtFunc[T]) SetDefbultReturn(r0 []string, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, []string) ([]string, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreHebrtbebtFunc[T]) PushReturn(r0 []string, r1 []string, r2 error) {
	f.PushHook(func(context.Context, []string) ([]string, []string, error) {
		return r0, r1, r2
	})
}

func (f *StoreHebrtbebtFunc[T]) nextHook() func(context.Context, []string) ([]string, []string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreHebrtbebtFunc[T]) bppendCbll(r0 StoreHebrtbebtFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreHebrtbebtFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreHebrtbebtFunc[T]) History() []StoreHebrtbebtFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreHebrtbebtFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreHebrtbebtFuncCbll is bn object thbt describes bn invocbtion of
// method Hebrtbebt on bn instbnce of MockStore.
type StoreHebrtbebtFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreHebrtbebtFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreHebrtbebtFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreMbrkCompleteFunc describes the behbvior when the MbrkComplete method
// of the pbrent MockStore instbnce is invoked.
type StoreMbrkCompleteFunc[T Record] struct {
	defbultHook func(context.Context, T) (bool, error)
	hooks       []func(context.Context, T) (bool, error)
	history     []StoreMbrkCompleteFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkComplete delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbrkComplete(v0 context.Context, v1 T) (bool, error) {
	r0, r1 := m.MbrkCompleteFunc.nextHook()(v0, v1)
	m.MbrkCompleteFunc.bppendCbll(StoreMbrkCompleteFuncCbll[T]{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkComplete method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreMbrkCompleteFunc[T]) SetDefbultHook(hook func(context.Context, T) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkComplete method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreMbrkCompleteFunc[T]) PushHook(hook func(context.Context, T) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkCompleteFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, T) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkCompleteFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, T) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMbrkCompleteFunc[T]) nextHook() func(context.Context, T) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMbrkCompleteFunc[T]) bppendCbll(r0 StoreMbrkCompleteFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreMbrkCompleteFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreMbrkCompleteFunc[T]) History() []StoreMbrkCompleteFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreMbrkCompleteFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMbrkCompleteFuncCbll is bn object thbt describes bn invocbtion of
// method MbrkComplete on bn instbnce of MockStore.
type StoreMbrkCompleteFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 T
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreMbrkCompleteFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkCompleteFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreMbrkErroredFunc describes the behbvior when the MbrkErrored method
// of the pbrent MockStore instbnce is invoked.
type StoreMbrkErroredFunc[T Record] struct {
	defbultHook func(context.Context, T, string) (bool, error)
	hooks       []func(context.Context, T, string) (bool, error)
	history     []StoreMbrkErroredFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkErrored delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbrkErrored(v0 context.Context, v1 T, v2 string) (bool, error) {
	r0, r1 := m.MbrkErroredFunc.nextHook()(v0, v1, v2)
	m.MbrkErroredFunc.bppendCbll(StoreMbrkErroredFuncCbll[T]{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkErrored method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreMbrkErroredFunc[T]) SetDefbultHook(hook func(context.Context, T, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkErrored method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreMbrkErroredFunc[T]) PushHook(hook func(context.Context, T, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkErroredFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, T, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkErroredFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, T, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMbrkErroredFunc[T]) nextHook() func(context.Context, T, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMbrkErroredFunc[T]) bppendCbll(r0 StoreMbrkErroredFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreMbrkErroredFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreMbrkErroredFunc[T]) History() []StoreMbrkErroredFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreMbrkErroredFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMbrkErroredFuncCbll is bn object thbt describes bn invocbtion of
// method MbrkErrored on bn instbnce of MockStore.
type StoreMbrkErroredFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 T
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreMbrkErroredFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkErroredFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreMbrkFbiledFunc describes the behbvior when the MbrkFbiled method of
// the pbrent MockStore instbnce is invoked.
type StoreMbrkFbiledFunc[T Record] struct {
	defbultHook func(context.Context, T, string) (bool, error)
	hooks       []func(context.Context, T, string) (bool, error)
	history     []StoreMbrkFbiledFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkFbiled delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbrkFbiled(v0 context.Context, v1 T, v2 string) (bool, error) {
	r0, r1 := m.MbrkFbiledFunc.nextHook()(v0, v1, v2)
	m.MbrkFbiledFunc.bppendCbll(StoreMbrkFbiledFuncCbll[T]{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkFbiled method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreMbrkFbiledFunc[T]) SetDefbultHook(hook func(context.Context, T, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkFbiled method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreMbrkFbiledFunc[T]) PushHook(hook func(context.Context, T, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkFbiledFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, T, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkFbiledFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, T, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMbrkFbiledFunc[T]) nextHook() func(context.Context, T, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMbrkFbiledFunc[T]) bppendCbll(r0 StoreMbrkFbiledFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreMbrkFbiledFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreMbrkFbiledFunc[T]) History() []StoreMbrkFbiledFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreMbrkFbiledFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMbrkFbiledFuncCbll is bn object thbt describes bn invocbtion of
// method MbrkFbiled on bn instbnce of MockStore.
type StoreMbrkFbiledFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 T
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreMbrkFbiledFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkFbiledFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreQueuedCountFunc describes the behbvior when the QueuedCount method
// of the pbrent MockStore instbnce is invoked.
type StoreQueuedCountFunc[T Record] struct {
	defbultHook func(context.Context) (int, error)
	hooks       []func(context.Context) (int, error)
	history     []StoreQueuedCountFuncCbll[T]
	mutex       sync.Mutex
}

// QueuedCount delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) QueuedCount(v0 context.Context) (int, error) {
	r0, r1 := m.QueuedCountFunc.nextHook()(v0)
	m.QueuedCountFunc.bppendCbll(StoreQueuedCountFuncCbll[T]{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the QueuedCount method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreQueuedCountFunc[T]) SetDefbultHook(hook func(context.Context) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueuedCount method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreQueuedCountFunc[T]) PushHook(hook func(context.Context) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreQueuedCountFunc[T]) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreQueuedCountFunc[T]) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

func (f *StoreQueuedCountFunc[T]) nextHook() func(context.Context) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreQueuedCountFunc[T]) bppendCbll(r0 StoreQueuedCountFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreQueuedCountFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreQueuedCountFunc[T]) History() []StoreQueuedCountFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreQueuedCountFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreQueuedCountFuncCbll is bn object thbt describes bn invocbtion of
// method QueuedCount on bn instbnce of MockStore.
type StoreQueuedCountFuncCbll[T Record] struct {
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
func (c StoreQueuedCountFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreQueuedCountFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockWithHooks is b mock implementbtion of the WithHooks interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/workerutil) used
// for unit testing.
type MockWithHooks[T Record] struct {
	// PostHbndleFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method PostHbndle.
	PostHbndleFunc *WithHooksPostHbndleFunc[T]
	// PreHbndleFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method PreHbndle.
	PreHbndleFunc *WithHooksPreHbndleFunc[T]
}

// NewMockWithHooks crebtes b new mock of the WithHooks interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockWithHooks[T Record]() *MockWithHooks[T] {
	return &MockWithHooks[T]{
		PostHbndleFunc: &WithHooksPostHbndleFunc[T]{
			defbultHook: func(context.Context, log.Logger, T) {
				return
			},
		},
		PreHbndleFunc: &WithHooksPreHbndleFunc[T]{
			defbultHook: func(context.Context, log.Logger, T) {
				return
			},
		},
	}
}

// NewStrictMockWithHooks crebtes b new mock of the WithHooks interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockWithHooks[T Record]() *MockWithHooks[T] {
	return &MockWithHooks[T]{
		PostHbndleFunc: &WithHooksPostHbndleFunc[T]{
			defbultHook: func(context.Context, log.Logger, T) {
				pbnic("unexpected invocbtion of MockWithHooks.PostHbndle")
			},
		},
		PreHbndleFunc: &WithHooksPreHbndleFunc[T]{
			defbultHook: func(context.Context, log.Logger, T) {
				pbnic("unexpected invocbtion of MockWithHooks.PreHbndle")
			},
		},
	}
}

// NewMockWithHooksFrom crebtes b new mock of the MockWithHooks interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockWithHooksFrom[T Record](i WithHooks[T]) *MockWithHooks[T] {
	return &MockWithHooks[T]{
		PostHbndleFunc: &WithHooksPostHbndleFunc[T]{
			defbultHook: i.PostHbndle,
		},
		PreHbndleFunc: &WithHooksPreHbndleFunc[T]{
			defbultHook: i.PreHbndle,
		},
	}
}

// WithHooksPostHbndleFunc describes the behbvior when the PostHbndle method
// of the pbrent MockWithHooks instbnce is invoked.
type WithHooksPostHbndleFunc[T Record] struct {
	defbultHook func(context.Context, log.Logger, T)
	hooks       []func(context.Context, log.Logger, T)
	history     []WithHooksPostHbndleFuncCbll[T]
	mutex       sync.Mutex
}

// PostHbndle delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWithHooks[T]) PostHbndle(v0 context.Context, v1 log.Logger, v2 T) {
	m.PostHbndleFunc.nextHook()(v0, v1, v2)
	m.PostHbndleFunc.bppendCbll(WithHooksPostHbndleFuncCbll[T]{v0, v1, v2})
	return
}

// SetDefbultHook sets function thbt is cblled when the PostHbndle method of
// the pbrent MockWithHooks instbnce is invoked bnd the hook queue is empty.
func (f *WithHooksPostHbndleFunc[T]) SetDefbultHook(hook func(context.Context, log.Logger, T)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// PostHbndle method of the pbrent MockWithHooks instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WithHooksPostHbndleFunc[T]) PushHook(hook func(context.Context, log.Logger, T)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WithHooksPostHbndleFunc[T]) SetDefbultReturn() {
	f.SetDefbultHook(func(context.Context, log.Logger, T) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WithHooksPostHbndleFunc[T]) PushReturn() {
	f.PushHook(func(context.Context, log.Logger, T) {
		return
	})
}

func (f *WithHooksPostHbndleFunc[T]) nextHook() func(context.Context, log.Logger, T) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WithHooksPostHbndleFunc[T]) bppendCbll(r0 WithHooksPostHbndleFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WithHooksPostHbndleFuncCbll objects
// describing the invocbtions of this function.
func (f *WithHooksPostHbndleFunc[T]) History() []WithHooksPostHbndleFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WithHooksPostHbndleFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WithHooksPostHbndleFuncCbll is bn object thbt describes bn invocbtion of
// method PostHbndle on bn instbnce of MockWithHooks.
type WithHooksPostHbndleFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 log.Logger
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 T
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WithHooksPostHbndleFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WithHooksPostHbndleFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{}
}

// WithHooksPreHbndleFunc describes the behbvior when the PreHbndle method
// of the pbrent MockWithHooks instbnce is invoked.
type WithHooksPreHbndleFunc[T Record] struct {
	defbultHook func(context.Context, log.Logger, T)
	hooks       []func(context.Context, log.Logger, T)
	history     []WithHooksPreHbndleFuncCbll[T]
	mutex       sync.Mutex
}

// PreHbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWithHooks[T]) PreHbndle(v0 context.Context, v1 log.Logger, v2 T) {
	m.PreHbndleFunc.nextHook()(v0, v1, v2)
	m.PreHbndleFunc.bppendCbll(WithHooksPreHbndleFuncCbll[T]{v0, v1, v2})
	return
}

// SetDefbultHook sets function thbt is cblled when the PreHbndle method of
// the pbrent MockWithHooks instbnce is invoked bnd the hook queue is empty.
func (f *WithHooksPreHbndleFunc[T]) SetDefbultHook(hook func(context.Context, log.Logger, T)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// PreHbndle method of the pbrent MockWithHooks instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WithHooksPreHbndleFunc[T]) PushHook(hook func(context.Context, log.Logger, T)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WithHooksPreHbndleFunc[T]) SetDefbultReturn() {
	f.SetDefbultHook(func(context.Context, log.Logger, T) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WithHooksPreHbndleFunc[T]) PushReturn() {
	f.PushHook(func(context.Context, log.Logger, T) {
		return
	})
}

func (f *WithHooksPreHbndleFunc[T]) nextHook() func(context.Context, log.Logger, T) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WithHooksPreHbndleFunc[T]) bppendCbll(r0 WithHooksPreHbndleFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WithHooksPreHbndleFuncCbll objects
// describing the invocbtions of this function.
func (f *WithHooksPreHbndleFunc[T]) History() []WithHooksPreHbndleFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]WithHooksPreHbndleFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WithHooksPreHbndleFuncCbll is bn object thbt describes bn invocbtion of
// method PreHbndle on bn instbnce of MockWithHooks.
type WithHooksPreHbndleFuncCbll[T Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 log.Logger
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 T
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WithHooksPreHbndleFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WithHooksPreHbndleFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{}
}

// MockWithPreDequeue is b mock implementbtion of the WithPreDequeue
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/workerutil) used for unit
// testing.
type MockWithPreDequeue struct {
	// PreDequeueFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method PreDequeue.
	PreDequeueFunc *WithPreDequeuePreDequeueFunc
}

// NewMockWithPreDequeue crebtes b new mock of the WithPreDequeue interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockWithPreDequeue() *MockWithPreDequeue {
	return &MockWithPreDequeue{
		PreDequeueFunc: &WithPreDequeuePreDequeueFunc{
			defbultHook: func(context.Context, log.Logger) (r0 bool, r1 interfbce{}, r2 error) {
				return
			},
		},
	}
}

// NewStrictMockWithPreDequeue crebtes b new mock of the WithPreDequeue
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockWithPreDequeue() *MockWithPreDequeue {
	return &MockWithPreDequeue{
		PreDequeueFunc: &WithPreDequeuePreDequeueFunc{
			defbultHook: func(context.Context, log.Logger) (bool, interfbce{}, error) {
				pbnic("unexpected invocbtion of MockWithPreDequeue.PreDequeue")
			},
		},
	}
}

// NewMockWithPreDequeueFrom crebtes b new mock of the MockWithPreDequeue
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockWithPreDequeueFrom(i WithPreDequeue) *MockWithPreDequeue {
	return &MockWithPreDequeue{
		PreDequeueFunc: &WithPreDequeuePreDequeueFunc{
			defbultHook: i.PreDequeue,
		},
	}
}

// WithPreDequeuePreDequeueFunc describes the behbvior when the PreDequeue
// method of the pbrent MockWithPreDequeue instbnce is invoked.
type WithPreDequeuePreDequeueFunc struct {
	defbultHook func(context.Context, log.Logger) (bool, interfbce{}, error)
	hooks       []func(context.Context, log.Logger) (bool, interfbce{}, error)
	history     []WithPreDequeuePreDequeueFuncCbll
	mutex       sync.Mutex
}

// PreDequeue delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWithPreDequeue) PreDequeue(v0 context.Context, v1 log.Logger) (bool, interfbce{}, error) {
	r0, r1, r2 := m.PreDequeueFunc.nextHook()(v0, v1)
	m.PreDequeueFunc.bppendCbll(WithPreDequeuePreDequeueFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the PreDequeue method of
// the pbrent MockWithPreDequeue instbnce is invoked bnd the hook queue is
// empty.
func (f *WithPreDequeuePreDequeueFunc) SetDefbultHook(hook func(context.Context, log.Logger) (bool, interfbce{}, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// PreDequeue method of the pbrent MockWithPreDequeue instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WithPreDequeuePreDequeueFunc) PushHook(hook func(context.Context, log.Logger) (bool, interfbce{}, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WithPreDequeuePreDequeueFunc) SetDefbultReturn(r0 bool, r1 interfbce{}, r2 error) {
	f.SetDefbultHook(func(context.Context, log.Logger) (bool, interfbce{}, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WithPreDequeuePreDequeueFunc) PushReturn(r0 bool, r1 interfbce{}, r2 error) {
	f.PushHook(func(context.Context, log.Logger) (bool, interfbce{}, error) {
		return r0, r1, r2
	})
}

func (f *WithPreDequeuePreDequeueFunc) nextHook() func(context.Context, log.Logger) (bool, interfbce{}, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WithPreDequeuePreDequeueFunc) bppendCbll(r0 WithPreDequeuePreDequeueFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WithPreDequeuePreDequeueFuncCbll objects
// describing the invocbtions of this function.
func (f *WithPreDequeuePreDequeueFunc) History() []WithPreDequeuePreDequeueFuncCbll {
	f.mutex.Lock()
	history := mbke([]WithPreDequeuePreDequeueFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WithPreDequeuePreDequeueFuncCbll is bn object thbt describes bn
// invocbtion of method PreDequeue on bn instbnce of MockWithPreDequeue.
type WithPreDequeuePreDequeueFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 log.Logger
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 interfbce{}
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WithPreDequeuePreDequeueFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WithPreDequeuePreDequeueFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}
