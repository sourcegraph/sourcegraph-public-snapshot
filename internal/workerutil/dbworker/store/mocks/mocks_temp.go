// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge mocks

import (
	"context"
	"sync"
	"time"

	sqlf "github.com/keegbncsmith/sqlf"
	bbsestore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	executor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	workerutil "github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	store "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store)
// used for unit testing.
type MockStore[T workerutil.Record] struct {
	// AddExecutionLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AddExecutionLogEntry.
	AddExecutionLogEntryFunc *StoreAddExecutionLogEntryFunc[T]
	// DequeueFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Dequeue.
	DequeueFunc *StoreDequeueFunc[T]
	// HbndleFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hbndle.
	HbndleFunc *StoreHbndleFunc[T]
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
	// MbxDurbtionInQueueFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method MbxDurbtionInQueue.
	MbxDurbtionInQueueFunc *StoreMbxDurbtionInQueueFunc[T]
	// QueuedCountFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method QueuedCount.
	QueuedCountFunc *StoreQueuedCountFunc[T]
	// RequeueFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Requeue.
	RequeueFunc *StoreRequeueFunc[T]
	// ResetStblledFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ResetStblled.
	ResetStblledFunc *StoreResetStblledFunc[T]
	// UpdbteExecutionLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteExecutionLogEntry.
	UpdbteExecutionLogEntryFunc *StoreUpdbteExecutionLogEntryFunc[T]
	// WithFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method With.
	WithFunc *StoreWithFunc[T]
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore[T workerutil.Record]() *MockStore[T] {
	return &MockStore[T]{
		AddExecutionLogEntryFunc: &StoreAddExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (r0 int, r1 error) {
				return
			},
		},
		DequeueFunc: &StoreDequeueFunc[T]{
			defbultHook: func(context.Context, string, []*sqlf.Query) (r0 T, r1 bool, r2 error) {
				return
			},
		},
		HbndleFunc: &StoreHbndleFunc[T]{
			defbultHook: func() (r0 bbsestore.TrbnsbctbbleHbndle) {
				return
			},
		},
		HebrtbebtFunc: &StoreHebrtbebtFunc[T]{
			defbultHook: func(context.Context, []string, store.HebrtbebtOptions) (r0 []string, r1 []string, r2 error) {
				return
			},
		},
		MbrkCompleteFunc: &StoreMbrkCompleteFunc[T]{
			defbultHook: func(context.Context, int, store.MbrkFinblOptions) (r0 bool, r1 error) {
				return
			},
		},
		MbrkErroredFunc: &StoreMbrkErroredFunc[T]{
			defbultHook: func(context.Context, int, string, store.MbrkFinblOptions) (r0 bool, r1 error) {
				return
			},
		},
		MbrkFbiledFunc: &StoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, int, string, store.MbrkFinblOptions) (r0 bool, r1 error) {
				return
			},
		},
		MbxDurbtionInQueueFunc: &StoreMbxDurbtionInQueueFunc[T]{
			defbultHook: func(context.Context) (r0 time.Durbtion, r1 error) {
				return
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc[T]{
			defbultHook: func(context.Context, bool) (r0 int, r1 error) {
				return
			},
		},
		RequeueFunc: &StoreRequeueFunc[T]{
			defbultHook: func(context.Context, int, time.Time) (r0 error) {
				return
			},
		},
		ResetStblledFunc: &StoreResetStblledFunc[T]{
			defbultHook: func(context.Context) (r0 mbp[int]time.Durbtion, r1 mbp[int]time.Durbtion, r2 error) {
				return
			},
		},
		UpdbteExecutionLogEntryFunc: &StoreUpdbteExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (r0 error) {
				return
			},
		},
		WithFunc: &StoreWithFunc[T]{
			defbultHook: func(bbsestore.ShbrebbleStore) (r0 store.Store[T]) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore[T workerutil.Record]() *MockStore[T] {
	return &MockStore[T]{
		AddExecutionLogEntryFunc: &StoreAddExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error) {
				pbnic("unexpected invocbtion of MockStore.AddExecutionLogEntry")
			},
		},
		DequeueFunc: &StoreDequeueFunc[T]{
			defbultHook: func(context.Context, string, []*sqlf.Query) (T, bool, error) {
				pbnic("unexpected invocbtion of MockStore.Dequeue")
			},
		},
		HbndleFunc: &StoreHbndleFunc[T]{
			defbultHook: func() bbsestore.TrbnsbctbbleHbndle {
				pbnic("unexpected invocbtion of MockStore.Hbndle")
			},
		},
		HebrtbebtFunc: &StoreHebrtbebtFunc[T]{
			defbultHook: func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error) {
				pbnic("unexpected invocbtion of MockStore.Hebrtbebt")
			},
		},
		MbrkCompleteFunc: &StoreMbrkCompleteFunc[T]{
			defbultHook: func(context.Context, int, store.MbrkFinblOptions) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.MbrkComplete")
			},
		},
		MbrkErroredFunc: &StoreMbrkErroredFunc[T]{
			defbultHook: func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.MbrkErrored")
			},
		},
		MbrkFbiledFunc: &StoreMbrkFbiledFunc[T]{
			defbultHook: func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.MbrkFbiled")
			},
		},
		MbxDurbtionInQueueFunc: &StoreMbxDurbtionInQueueFunc[T]{
			defbultHook: func(context.Context) (time.Durbtion, error) {
				pbnic("unexpected invocbtion of MockStore.MbxDurbtionInQueue")
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc[T]{
			defbultHook: func(context.Context, bool) (int, error) {
				pbnic("unexpected invocbtion of MockStore.QueuedCount")
			},
		},
		RequeueFunc: &StoreRequeueFunc[T]{
			defbultHook: func(context.Context, int, time.Time) error {
				pbnic("unexpected invocbtion of MockStore.Requeue")
			},
		},
		ResetStblledFunc: &StoreResetStblledFunc[T]{
			defbultHook: func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
				pbnic("unexpected invocbtion of MockStore.ResetStblled")
			},
		},
		UpdbteExecutionLogEntryFunc: &StoreUpdbteExecutionLogEntryFunc[T]{
			defbultHook: func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error {
				pbnic("unexpected invocbtion of MockStore.UpdbteExecutionLogEntry")
			},
		},
		WithFunc: &StoreWithFunc[T]{
			defbultHook: func(bbsestore.ShbrebbleStore) store.Store[T] {
				pbnic("unexpected invocbtion of MockStore.With")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom[T workerutil.Record](i store.Store[T]) *MockStore[T] {
	return &MockStore[T]{
		AddExecutionLogEntryFunc: &StoreAddExecutionLogEntryFunc[T]{
			defbultHook: i.AddExecutionLogEntry,
		},
		DequeueFunc: &StoreDequeueFunc[T]{
			defbultHook: i.Dequeue,
		},
		HbndleFunc: &StoreHbndleFunc[T]{
			defbultHook: i.Hbndle,
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
		MbxDurbtionInQueueFunc: &StoreMbxDurbtionInQueueFunc[T]{
			defbultHook: i.MbxDurbtionInQueue,
		},
		QueuedCountFunc: &StoreQueuedCountFunc[T]{
			defbultHook: i.QueuedCount,
		},
		RequeueFunc: &StoreRequeueFunc[T]{
			defbultHook: i.Requeue,
		},
		ResetStblledFunc: &StoreResetStblledFunc[T]{
			defbultHook: i.ResetStblled,
		},
		UpdbteExecutionLogEntryFunc: &StoreUpdbteExecutionLogEntryFunc[T]{
			defbultHook: i.UpdbteExecutionLogEntry,
		},
		WithFunc: &StoreWithFunc[T]{
			defbultHook: i.With,
		},
	}
}

// StoreAddExecutionLogEntryFunc describes the behbvior when the
// AddExecutionLogEntry method of the pbrent MockStore instbnce is invoked.
type StoreAddExecutionLogEntryFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error)
	hooks       []func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error)
	history     []StoreAddExecutionLogEntryFuncCbll[T]
	mutex       sync.Mutex
}

// AddExecutionLogEntry delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) AddExecutionLogEntry(v0 context.Context, v1 int, v2 executor.ExecutionLogEntry, v3 store.ExecutionLogEntryOptions) (int, error) {
	r0, r1 := m.AddExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3)
	m.AddExecutionLogEntryFunc.bppendCbll(StoreAddExecutionLogEntryFuncCbll[T]{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AddExecutionLogEntry
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreAddExecutionLogEntryFunc[T]) SetDefbultHook(hook func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddExecutionLogEntry method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreAddExecutionLogEntryFunc[T]) PushHook(hook func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreAddExecutionLogEntryFunc[T]) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreAddExecutionLogEntryFunc[T]) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error) {
		return r0, r1
	})
}

func (f *StoreAddExecutionLogEntryFunc[T]) nextHook() func(context.Context, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreAddExecutionLogEntryFunc[T]) bppendCbll(r0 StoreAddExecutionLogEntryFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreAddExecutionLogEntryFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreAddExecutionLogEntryFunc[T]) History() []StoreAddExecutionLogEntryFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreAddExecutionLogEntryFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreAddExecutionLogEntryFuncCbll is bn object thbt describes bn
// invocbtion of method AddExecutionLogEntry on bn instbnce of MockStore.
type StoreAddExecutionLogEntryFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 executor.ExecutionLogEntry
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 store.ExecutionLogEntryOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreAddExecutionLogEntryFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreAddExecutionLogEntryFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreDequeueFunc describes the behbvior when the Dequeue method of the
// pbrent MockStore instbnce is invoked.
type StoreDequeueFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, string, []*sqlf.Query) (T, bool, error)
	hooks       []func(context.Context, string, []*sqlf.Query) (T, bool, error)
	history     []StoreDequeueFuncCbll[T]
	mutex       sync.Mutex
}

// Dequeue delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) Dequeue(v0 context.Context, v1 string, v2 []*sqlf.Query) (T, bool, error) {
	r0, r1, r2 := m.DequeueFunc.nextHook()(v0, v1, v2)
	m.DequeueFunc.bppendCbll(StoreDequeueFuncCbll[T]{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Dequeue method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDequeueFunc[T]) SetDefbultHook(hook func(context.Context, string, []*sqlf.Query) (T, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Dequeue method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDequeueFunc[T]) PushHook(hook func(context.Context, string, []*sqlf.Query) (T, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDequeueFunc[T]) SetDefbultReturn(r0 T, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, []*sqlf.Query) (T, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDequeueFunc[T]) PushReturn(r0 T, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, []*sqlf.Query) (T, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreDequeueFunc[T]) nextHook() func(context.Context, string, []*sqlf.Query) (T, bool, error) {
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
type StoreDequeueFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []*sqlf.Query
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

// StoreHbndleFunc describes the behbvior when the Hbndle method of the
// pbrent MockStore instbnce is invoked.
type StoreHbndleFunc[T workerutil.Record] struct {
	defbultHook func() bbsestore.TrbnsbctbbleHbndle
	hooks       []func() bbsestore.TrbnsbctbbleHbndle
	history     []StoreHbndleFuncCbll[T]
	mutex       sync.Mutex
}

// Hbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) Hbndle() bbsestore.TrbnsbctbbleHbndle {
	r0 := m.HbndleFunc.nextHook()()
	m.HbndleFunc.bppendCbll(StoreHbndleFuncCbll[T]{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Hbndle method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreHbndleFunc[T]) SetDefbultHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hbndle method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreHbndleFunc[T]) PushHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreHbndleFunc[T]) SetDefbultReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.SetDefbultHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreHbndleFunc[T]) PushReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.PushHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

func (f *StoreHbndleFunc[T]) nextHook() func() bbsestore.TrbnsbctbbleHbndle {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreHbndleFunc[T]) bppendCbll(r0 StoreHbndleFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreHbndleFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreHbndleFunc[T]) History() []StoreHbndleFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreHbndleFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreHbndleFuncCbll is bn object thbt describes bn invocbtion of method
// Hbndle on bn instbnce of MockStore.
type StoreHbndleFuncCbll[T workerutil.Record] struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bbsestore.TrbnsbctbbleHbndle
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreHbndleFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreHbndleFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreHebrtbebtFunc describes the behbvior when the Hebrtbebt method of
// the pbrent MockStore instbnce is invoked.
type StoreHebrtbebtFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error)
	hooks       []func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error)
	history     []StoreHebrtbebtFuncCbll[T]
	mutex       sync.Mutex
}

// Hebrtbebt delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) Hebrtbebt(v0 context.Context, v1 []string, v2 store.HebrtbebtOptions) ([]string, []string, error) {
	r0, r1, r2 := m.HebrtbebtFunc.nextHook()(v0, v1, v2)
	m.HebrtbebtFunc.bppendCbll(StoreHebrtbebtFuncCbll[T]{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Hebrtbebt method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreHebrtbebtFunc[T]) SetDefbultHook(hook func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hebrtbebt method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreHebrtbebtFunc[T]) PushHook(hook func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreHebrtbebtFunc[T]) SetDefbultReturn(r0 []string, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreHebrtbebtFunc[T]) PushReturn(r0 []string, r1 []string, r2 error) {
	f.PushHook(func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error) {
		return r0, r1, r2
	})
}

func (f *StoreHebrtbebtFunc[T]) nextHook() func(context.Context, []string, store.HebrtbebtOptions) ([]string, []string, error) {
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
type StoreHebrtbebtFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 store.HebrtbebtOptions
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
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreHebrtbebtFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreMbrkCompleteFunc describes the behbvior when the MbrkComplete method
// of the pbrent MockStore instbnce is invoked.
type StoreMbrkCompleteFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, store.MbrkFinblOptions) (bool, error)
	hooks       []func(context.Context, int, store.MbrkFinblOptions) (bool, error)
	history     []StoreMbrkCompleteFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkComplete delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbrkComplete(v0 context.Context, v1 int, v2 store.MbrkFinblOptions) (bool, error) {
	r0, r1 := m.MbrkCompleteFunc.nextHook()(v0, v1, v2)
	m.MbrkCompleteFunc.bppendCbll(StoreMbrkCompleteFuncCbll[T]{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkComplete method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreMbrkCompleteFunc[T]) SetDefbultHook(hook func(context.Context, int, store.MbrkFinblOptions) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkComplete method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreMbrkCompleteFunc[T]) PushHook(hook func(context.Context, int, store.MbrkFinblOptions) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkCompleteFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, store.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkCompleteFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, store.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMbrkCompleteFunc[T]) nextHook() func(context.Context, int, store.MbrkFinblOptions) (bool, error) {
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
type StoreMbrkCompleteFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 store.MbrkFinblOptions
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
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkCompleteFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreMbrkErroredFunc describes the behbvior when the MbrkErrored method
// of the pbrent MockStore instbnce is invoked.
type StoreMbrkErroredFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)
	hooks       []func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)
	history     []StoreMbrkErroredFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkErrored delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbrkErrored(v0 context.Context, v1 int, v2 string, v3 store.MbrkFinblOptions) (bool, error) {
	r0, r1 := m.MbrkErroredFunc.nextHook()(v0, v1, v2, v3)
	m.MbrkErroredFunc.bppendCbll(StoreMbrkErroredFuncCbll[T]{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkErrored method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreMbrkErroredFunc[T]) SetDefbultHook(hook func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkErrored method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreMbrkErroredFunc[T]) PushHook(hook func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkErroredFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkErroredFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMbrkErroredFunc[T]) nextHook() func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
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
type StoreMbrkErroredFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 store.MbrkFinblOptions
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
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkErroredFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreMbrkFbiledFunc describes the behbvior when the MbrkFbiled method of
// the pbrent MockStore instbnce is invoked.
type StoreMbrkFbiledFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)
	hooks       []func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)
	history     []StoreMbrkFbiledFuncCbll[T]
	mutex       sync.Mutex
}

// MbrkFbiled delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbrkFbiled(v0 context.Context, v1 int, v2 string, v3 store.MbrkFinblOptions) (bool, error) {
	r0, r1 := m.MbrkFbiledFunc.nextHook()(v0, v1, v2, v3)
	m.MbrkFbiledFunc.bppendCbll(StoreMbrkFbiledFuncCbll[T]{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbrkFbiled method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreMbrkFbiledFunc[T]) SetDefbultHook(hook func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbrkFbiled method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreMbrkFbiledFunc[T]) PushHook(hook func(context.Context, int, string, store.MbrkFinblOptions) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbrkFbiledFunc[T]) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbrkFbiledFunc[T]) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMbrkFbiledFunc[T]) nextHook() func(context.Context, int, string, store.MbrkFinblOptions) (bool, error) {
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
type StoreMbrkFbiledFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 store.MbrkFinblOptions
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
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbrkFbiledFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreMbxDurbtionInQueueFunc describes the behbvior when the
// MbxDurbtionInQueue method of the pbrent MockStore instbnce is invoked.
type StoreMbxDurbtionInQueueFunc[T workerutil.Record] struct {
	defbultHook func(context.Context) (time.Durbtion, error)
	hooks       []func(context.Context) (time.Durbtion, error)
	history     []StoreMbxDurbtionInQueueFuncCbll[T]
	mutex       sync.Mutex
}

// MbxDurbtionInQueue delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) MbxDurbtionInQueue(v0 context.Context) (time.Durbtion, error) {
	r0, r1 := m.MbxDurbtionInQueueFunc.nextHook()(v0)
	m.MbxDurbtionInQueueFunc.bppendCbll(StoreMbxDurbtionInQueueFuncCbll[T]{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MbxDurbtionInQueue
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreMbxDurbtionInQueueFunc[T]) SetDefbultHook(hook func(context.Context) (time.Durbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbxDurbtionInQueue method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreMbxDurbtionInQueueFunc[T]) PushHook(hook func(context.Context) (time.Durbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreMbxDurbtionInQueueFunc[T]) SetDefbultReturn(r0 time.Durbtion, r1 error) {
	f.SetDefbultHook(func(context.Context) (time.Durbtion, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreMbxDurbtionInQueueFunc[T]) PushReturn(r0 time.Durbtion, r1 error) {
	f.PushHook(func(context.Context) (time.Durbtion, error) {
		return r0, r1
	})
}

func (f *StoreMbxDurbtionInQueueFunc[T]) nextHook() func(context.Context) (time.Durbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMbxDurbtionInQueueFunc[T]) bppendCbll(r0 StoreMbxDurbtionInQueueFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreMbxDurbtionInQueueFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreMbxDurbtionInQueueFunc[T]) History() []StoreMbxDurbtionInQueueFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreMbxDurbtionInQueueFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMbxDurbtionInQueueFuncCbll is bn object thbt describes bn invocbtion
// of method MbxDurbtionInQueue on bn instbnce of MockStore.
type StoreMbxDurbtionInQueueFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 time.Durbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreMbxDurbtionInQueueFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreMbxDurbtionInQueueFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreQueuedCountFunc describes the behbvior when the QueuedCount method
// of the pbrent MockStore instbnce is invoked.
type StoreQueuedCountFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, bool) (int, error)
	hooks       []func(context.Context, bool) (int, error)
	history     []StoreQueuedCountFuncCbll[T]
	mutex       sync.Mutex
}

// QueuedCount delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) QueuedCount(v0 context.Context, v1 bool) (int, error) {
	r0, r1 := m.QueuedCountFunc.nextHook()(v0, v1)
	m.QueuedCountFunc.bppendCbll(StoreQueuedCountFuncCbll[T]{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the QueuedCount method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreQueuedCountFunc[T]) SetDefbultHook(hook func(context.Context, bool) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueuedCount method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreQueuedCountFunc[T]) PushHook(hook func(context.Context, bool) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreQueuedCountFunc[T]) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, bool) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreQueuedCountFunc[T]) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, bool) (int, error) {
		return r0, r1
	})
}

func (f *StoreQueuedCountFunc[T]) nextHook() func(context.Context, bool) (int, error) {
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
type StoreQueuedCountFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bool
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
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreQueuedCountFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreRequeueFunc describes the behbvior when the Requeue method of the
// pbrent MockStore instbnce is invoked.
type StoreRequeueFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, time.Time) error
	hooks       []func(context.Context, int, time.Time) error
	history     []StoreRequeueFuncCbll[T]
	mutex       sync.Mutex
}

// Requeue delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) Requeue(v0 context.Context, v1 int, v2 time.Time) error {
	r0 := m.RequeueFunc.nextHook()(v0, v1, v2)
	m.RequeueFunc.bppendCbll(StoreRequeueFuncCbll[T]{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Requeue method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreRequeueFunc[T]) SetDefbultHook(hook func(context.Context, int, time.Time) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Requeue method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreRequeueFunc[T]) PushHook(hook func(context.Context, int, time.Time) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreRequeueFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, time.Time) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreRequeueFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, time.Time) error {
		return r0
	})
}

func (f *StoreRequeueFunc[T]) nextHook() func(context.Context, int, time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreRequeueFunc[T]) bppendCbll(r0 StoreRequeueFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreRequeueFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreRequeueFunc[T]) History() []StoreRequeueFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreRequeueFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreRequeueFuncCbll is bn object thbt describes bn invocbtion of method
// Requeue on bn instbnce of MockStore.
type StoreRequeueFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreRequeueFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreRequeueFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreResetStblledFunc describes the behbvior when the ResetStblled method
// of the pbrent MockStore instbnce is invoked.
type StoreResetStblledFunc[T workerutil.Record] struct {
	defbultHook func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)
	hooks       []func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)
	history     []StoreResetStblledFuncCbll[T]
	mutex       sync.Mutex
}

// ResetStblled delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) ResetStblled(v0 context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
	r0, r1, r2 := m.ResetStblledFunc.nextHook()(v0)
	m.ResetStblledFunc.bppendCbll(StoreResetStblledFuncCbll[T]{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ResetStblled method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreResetStblledFunc[T]) SetDefbultHook(hook func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResetStblled method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreResetStblledFunc[T]) PushHook(hook func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreResetStblledFunc[T]) SetDefbultReturn(r0 mbp[int]time.Durbtion, r1 mbp[int]time.Durbtion, r2 error) {
	f.SetDefbultHook(func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreResetStblledFunc[T]) PushReturn(r0 mbp[int]time.Durbtion, r1 mbp[int]time.Durbtion, r2 error) {
	f.PushHook(func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
		return r0, r1, r2
	})
}

func (f *StoreResetStblledFunc[T]) nextHook() func(context.Context) (mbp[int]time.Durbtion, mbp[int]time.Durbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreResetStblledFunc[T]) bppendCbll(r0 StoreResetStblledFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreResetStblledFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreResetStblledFunc[T]) History() []StoreResetStblledFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreResetStblledFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreResetStblledFuncCbll is bn object thbt describes bn invocbtion of
// method ResetStblled on bn instbnce of MockStore.
type StoreResetStblledFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[int]time.Durbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 mbp[int]time.Durbtion
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreResetStblledFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreResetStblledFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreUpdbteExecutionLogEntryFunc describes the behbvior when the
// UpdbteExecutionLogEntry method of the pbrent MockStore instbnce is
// invoked.
type StoreUpdbteExecutionLogEntryFunc[T workerutil.Record] struct {
	defbultHook func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error
	hooks       []func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error
	history     []StoreUpdbteExecutionLogEntryFuncCbll[T]
	mutex       sync.Mutex
}

// UpdbteExecutionLogEntry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) UpdbteExecutionLogEntry(v0 context.Context, v1 int, v2 int, v3 executor.ExecutionLogEntry, v4 store.ExecutionLogEntryOptions) error {
	r0 := m.UpdbteExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3, v4)
	m.UpdbteExecutionLogEntryFunc.bppendCbll(StoreUpdbteExecutionLogEntryFuncCbll[T]{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteExecutionLogEntry method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreUpdbteExecutionLogEntryFunc[T]) SetDefbultHook(hook func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteExecutionLogEntry method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreUpdbteExecutionLogEntryFunc[T]) PushHook(hook func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpdbteExecutionLogEntryFunc[T]) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpdbteExecutionLogEntryFunc[T]) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error {
		return r0
	})
}

func (f *StoreUpdbteExecutionLogEntryFunc[T]) nextHook() func(context.Context, int, int, executor.ExecutionLogEntry, store.ExecutionLogEntryOptions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdbteExecutionLogEntryFunc[T]) bppendCbll(r0 StoreUpdbteExecutionLogEntryFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUpdbteExecutionLogEntryFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreUpdbteExecutionLogEntryFunc[T]) History() []StoreUpdbteExecutionLogEntryFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreUpdbteExecutionLogEntryFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdbteExecutionLogEntryFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteExecutionLogEntry on bn instbnce of MockStore.
type StoreUpdbteExecutionLogEntryFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 executor.ExecutionLogEntry
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 store.ExecutionLogEntryOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpdbteExecutionLogEntryFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpdbteExecutionLogEntryFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreWithFunc describes the behbvior when the With method of the pbrent
// MockStore instbnce is invoked.
type StoreWithFunc[T workerutil.Record] struct {
	defbultHook func(bbsestore.ShbrebbleStore) store.Store[T]
	hooks       []func(bbsestore.ShbrebbleStore) store.Store[T]
	history     []StoreWithFuncCbll[T]
	mutex       sync.Mutex
}

// With delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore[T]) With(v0 bbsestore.ShbrebbleStore) store.Store[T] {
	r0 := m.WithFunc.nextHook()(v0)
	m.WithFunc.bppendCbll(StoreWithFuncCbll[T]{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the With method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreWithFunc[T]) SetDefbultHook(hook func(bbsestore.ShbrebbleStore) store.Store[T]) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// With method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreWithFunc[T]) PushHook(hook func(bbsestore.ShbrebbleStore) store.Store[T]) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreWithFunc[T]) SetDefbultReturn(r0 store.Store[T]) {
	f.SetDefbultHook(func(bbsestore.ShbrebbleStore) store.Store[T] {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreWithFunc[T]) PushReturn(r0 store.Store[T]) {
	f.PushHook(func(bbsestore.ShbrebbleStore) store.Store[T] {
		return r0
	})
}

func (f *StoreWithFunc[T]) nextHook() func(bbsestore.ShbrebbleStore) store.Store[T] {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreWithFunc[T]) bppendCbll(r0 StoreWithFuncCbll[T]) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreWithFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreWithFunc[T]) History() []StoreWithFuncCbll[T] {
	f.mutex.Lock()
	history := mbke([]StoreWithFuncCbll[T], len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreWithFuncCbll is bn object thbt describes bn invocbtion of method
// With on bn instbnce of MockStore.
type StoreWithFuncCbll[T workerutil.Record] struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 bbsestore.ShbrebbleStore
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 store.Store[T]
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreWithFuncCbll[T]) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreWithFuncCbll[T]) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
