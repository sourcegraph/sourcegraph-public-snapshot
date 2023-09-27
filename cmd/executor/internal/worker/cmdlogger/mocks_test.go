// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge cmdlogger

import (
	"context"
	"sync"

	executor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	types "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

// MockExecutionLogEntryStore is b mock implementbtion of the
// ExecutionLogEntryStore interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger)
// used for unit testing.
type MockExecutionLogEntryStore struct {
	// AddExecutionLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AddExecutionLogEntry.
	AddExecutionLogEntryFunc *ExecutionLogEntryStoreAddExecutionLogEntryFunc
	// UpdbteExecutionLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpdbteExecutionLogEntry.
	UpdbteExecutionLogEntryFunc *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc
}

// NewMockExecutionLogEntryStore crebtes b new mock of the
// ExecutionLogEntryStore interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockExecutionLogEntryStore() *MockExecutionLogEntryStore {
	return &MockExecutionLogEntryStore{
		AddExecutionLogEntryFunc: &ExecutionLogEntryStoreAddExecutionLogEntryFunc{
			defbultHook: func(context.Context, types.Job, executor.ExecutionLogEntry) (r0 int, r1 error) {
				return
			},
		},
		UpdbteExecutionLogEntryFunc: &ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc{
			defbultHook: func(context.Context, types.Job, int, executor.ExecutionLogEntry) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockExecutionLogEntryStore crebtes b new mock of the
// ExecutionLogEntryStore interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockExecutionLogEntryStore() *MockExecutionLogEntryStore {
	return &MockExecutionLogEntryStore{
		AddExecutionLogEntryFunc: &ExecutionLogEntryStoreAddExecutionLogEntryFunc{
			defbultHook: func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error) {
				pbnic("unexpected invocbtion of MockExecutionLogEntryStore.AddExecutionLogEntry")
			},
		},
		UpdbteExecutionLogEntryFunc: &ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc{
			defbultHook: func(context.Context, types.Job, int, executor.ExecutionLogEntry) error {
				pbnic("unexpected invocbtion of MockExecutionLogEntryStore.UpdbteExecutionLogEntry")
			},
		},
	}
}

// NewMockExecutionLogEntryStoreFrom crebtes b new mock of the
// MockExecutionLogEntryStore interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockExecutionLogEntryStoreFrom(i ExecutionLogEntryStore) *MockExecutionLogEntryStore {
	return &MockExecutionLogEntryStore{
		AddExecutionLogEntryFunc: &ExecutionLogEntryStoreAddExecutionLogEntryFunc{
			defbultHook: i.AddExecutionLogEntry,
		},
		UpdbteExecutionLogEntryFunc: &ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc{
			defbultHook: i.UpdbteExecutionLogEntry,
		},
	}
}

// ExecutionLogEntryStoreAddExecutionLogEntryFunc describes the behbvior
// when the AddExecutionLogEntry method of the pbrent
// MockExecutionLogEntryStore instbnce is invoked.
type ExecutionLogEntryStoreAddExecutionLogEntryFunc struct {
	defbultHook func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error)
	hooks       []func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error)
	history     []ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll
	mutex       sync.Mutex
}

// AddExecutionLogEntry delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockExecutionLogEntryStore) AddExecutionLogEntry(v0 context.Context, v1 types.Job, v2 executor.ExecutionLogEntry) (int, error) {
	r0, r1 := m.AddExecutionLogEntryFunc.nextHook()(v0, v1, v2)
	m.AddExecutionLogEntryFunc.bppendCbll(ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AddExecutionLogEntry
// method of the pbrent MockExecutionLogEntryStore instbnce is invoked bnd
// the hook queue is empty.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) SetDefbultHook(hook func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddExecutionLogEntry method of the pbrent MockExecutionLogEntryStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) PushHook(hook func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error) {
		return r0, r1
	})
}

func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) nextHook() func(context.Context, types.Job, executor.ExecutionLogEntry) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) bppendCbll(r0 ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll objects describing the
// invocbtions of this function.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) History() []ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll {
	f.mutex.Lock()
	history := mbke([]ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll is bn object thbt
// describes bn invocbtion of method AddExecutionLogEntry on bn instbnce of
// MockExecutionLogEntryStore.
type ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.Job
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 executor.ExecutionLogEntry
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ExecutionLogEntryStoreAddExecutionLogEntryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc describes the behbvior
// when the UpdbteExecutionLogEntry method of the pbrent
// MockExecutionLogEntryStore instbnce is invoked.
type ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc struct {
	defbultHook func(context.Context, types.Job, int, executor.ExecutionLogEntry) error
	hooks       []func(context.Context, types.Job, int, executor.ExecutionLogEntry) error
	history     []ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll
	mutex       sync.Mutex
}

// UpdbteExecutionLogEntry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockExecutionLogEntryStore) UpdbteExecutionLogEntry(v0 context.Context, v1 types.Job, v2 int, v3 executor.ExecutionLogEntry) error {
	r0 := m.UpdbteExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3)
	m.UpdbteExecutionLogEntryFunc.bppendCbll(ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteExecutionLogEntry method of the pbrent MockExecutionLogEntryStore
// instbnce is invoked bnd the hook queue is empty.
func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) SetDefbultHook(hook func(context.Context, types.Job, int, executor.ExecutionLogEntry) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteExecutionLogEntry method of the pbrent MockExecutionLogEntryStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) PushHook(hook func(context.Context, types.Job, int, executor.ExecutionLogEntry) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, types.Job, int, executor.ExecutionLogEntry) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, types.Job, int, executor.ExecutionLogEntry) error {
		return r0
	})
}

func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) nextHook() func(context.Context, types.Job, int, executor.ExecutionLogEntry) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) bppendCbll(r0 ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll objects describing
// the invocbtions of this function.
func (f *ExecutionLogEntryStoreUpdbteExecutionLogEntryFunc) History() []ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll {
	f.mutex.Lock()
	history := mbke([]ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll is bn object thbt
// describes bn invocbtion of method UpdbteExecutionLogEntry on bn instbnce
// of MockExecutionLogEntryStore.
type ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.Job
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 executor.ExecutionLogEntry
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ExecutionLogEntryStoreUpdbteExecutionLogEntryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockLogEntry is b mock implementbtion of the LogEntry interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger)
// used for unit testing.
type MockLogEntry struct {
	// CloseFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Close.
	CloseFunc *LogEntryCloseFunc
	// CurrentLogEntryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CurrentLogEntry.
	CurrentLogEntryFunc *LogEntryCurrentLogEntryFunc
	// FinblizeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Finblize.
	FinblizeFunc *LogEntryFinblizeFunc
	// WriteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Write.
	WriteFunc *LogEntryWriteFunc
}

// NewMockLogEntry crebtes b new mock of the LogEntry interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockLogEntry() *MockLogEntry {
	return &MockLogEntry{
		CloseFunc: &LogEntryCloseFunc{
			defbultHook: func() (r0 error) {
				return
			},
		},
		CurrentLogEntryFunc: &LogEntryCurrentLogEntryFunc{
			defbultHook: func() (r0 executor.ExecutionLogEntry) {
				return
			},
		},
		FinblizeFunc: &LogEntryFinblizeFunc{
			defbultHook: func(int) {
				return
			},
		},
		WriteFunc: &LogEntryWriteFunc{
			defbultHook: func([]byte) (r0 int, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockLogEntry crebtes b new mock of the LogEntry interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockLogEntry() *MockLogEntry {
	return &MockLogEntry{
		CloseFunc: &LogEntryCloseFunc{
			defbultHook: func() error {
				pbnic("unexpected invocbtion of MockLogEntry.Close")
			},
		},
		CurrentLogEntryFunc: &LogEntryCurrentLogEntryFunc{
			defbultHook: func() executor.ExecutionLogEntry {
				pbnic("unexpected invocbtion of MockLogEntry.CurrentLogEntry")
			},
		},
		FinblizeFunc: &LogEntryFinblizeFunc{
			defbultHook: func(int) {
				pbnic("unexpected invocbtion of MockLogEntry.Finblize")
			},
		},
		WriteFunc: &LogEntryWriteFunc{
			defbultHook: func([]byte) (int, error) {
				pbnic("unexpected invocbtion of MockLogEntry.Write")
			},
		},
	}
}

// NewMockLogEntryFrom crebtes b new mock of the MockLogEntry interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockLogEntryFrom(i LogEntry) *MockLogEntry {
	return &MockLogEntry{
		CloseFunc: &LogEntryCloseFunc{
			defbultHook: i.Close,
		},
		CurrentLogEntryFunc: &LogEntryCurrentLogEntryFunc{
			defbultHook: i.CurrentLogEntry,
		},
		FinblizeFunc: &LogEntryFinblizeFunc{
			defbultHook: i.Finblize,
		},
		WriteFunc: &LogEntryWriteFunc{
			defbultHook: i.Write,
		},
	}
}

// LogEntryCloseFunc describes the behbvior when the Close method of the
// pbrent MockLogEntry instbnce is invoked.
type LogEntryCloseFunc struct {
	defbultHook func() error
	hooks       []func() error
	history     []LogEntryCloseFuncCbll
	mutex       sync.Mutex
}

// Close delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogEntry) Close() error {
	r0 := m.CloseFunc.nextHook()()
	m.CloseFunc.bppendCbll(LogEntryCloseFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Close method of the
// pbrent MockLogEntry instbnce is invoked bnd the hook queue is empty.
func (f *LogEntryCloseFunc) SetDefbultHook(hook func() error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Close method of the pbrent MockLogEntry instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *LogEntryCloseFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LogEntryCloseFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func() error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LogEntryCloseFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *LogEntryCloseFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LogEntryCloseFunc) bppendCbll(r0 LogEntryCloseFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LogEntryCloseFuncCbll objects describing
// the invocbtions of this function.
func (f *LogEntryCloseFunc) History() []LogEntryCloseFuncCbll {
	f.mutex.Lock()
	history := mbke([]LogEntryCloseFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LogEntryCloseFuncCbll is bn object thbt describes bn invocbtion of method
// Close on bn instbnce of MockLogEntry.
type LogEntryCloseFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LogEntryCloseFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LogEntryCloseFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// LogEntryCurrentLogEntryFunc describes the behbvior when the
// CurrentLogEntry method of the pbrent MockLogEntry instbnce is invoked.
type LogEntryCurrentLogEntryFunc struct {
	defbultHook func() executor.ExecutionLogEntry
	hooks       []func() executor.ExecutionLogEntry
	history     []LogEntryCurrentLogEntryFuncCbll
	mutex       sync.Mutex
}

// CurrentLogEntry delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogEntry) CurrentLogEntry() executor.ExecutionLogEntry {
	r0 := m.CurrentLogEntryFunc.nextHook()()
	m.CurrentLogEntryFunc.bppendCbll(LogEntryCurrentLogEntryFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the CurrentLogEntry
// method of the pbrent MockLogEntry instbnce is invoked bnd the hook queue
// is empty.
func (f *LogEntryCurrentLogEntryFunc) SetDefbultHook(hook func() executor.ExecutionLogEntry) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CurrentLogEntry method of the pbrent MockLogEntry instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *LogEntryCurrentLogEntryFunc) PushHook(hook func() executor.ExecutionLogEntry) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LogEntryCurrentLogEntryFunc) SetDefbultReturn(r0 executor.ExecutionLogEntry) {
	f.SetDefbultHook(func() executor.ExecutionLogEntry {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LogEntryCurrentLogEntryFunc) PushReturn(r0 executor.ExecutionLogEntry) {
	f.PushHook(func() executor.ExecutionLogEntry {
		return r0
	})
}

func (f *LogEntryCurrentLogEntryFunc) nextHook() func() executor.ExecutionLogEntry {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LogEntryCurrentLogEntryFunc) bppendCbll(r0 LogEntryCurrentLogEntryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LogEntryCurrentLogEntryFuncCbll objects
// describing the invocbtions of this function.
func (f *LogEntryCurrentLogEntryFunc) History() []LogEntryCurrentLogEntryFuncCbll {
	f.mutex.Lock()
	history := mbke([]LogEntryCurrentLogEntryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LogEntryCurrentLogEntryFuncCbll is bn object thbt describes bn invocbtion
// of method CurrentLogEntry on bn instbnce of MockLogEntry.
type LogEntryCurrentLogEntryFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 executor.ExecutionLogEntry
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LogEntryCurrentLogEntryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LogEntryCurrentLogEntryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// LogEntryFinblizeFunc describes the behbvior when the Finblize method of
// the pbrent MockLogEntry instbnce is invoked.
type LogEntryFinblizeFunc struct {
	defbultHook func(int)
	hooks       []func(int)
	history     []LogEntryFinblizeFuncCbll
	mutex       sync.Mutex
}

// Finblize delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogEntry) Finblize(v0 int) {
	m.FinblizeFunc.nextHook()(v0)
	m.FinblizeFunc.bppendCbll(LogEntryFinblizeFuncCbll{v0})
	return
}

// SetDefbultHook sets function thbt is cblled when the Finblize method of
// the pbrent MockLogEntry instbnce is invoked bnd the hook queue is empty.
func (f *LogEntryFinblizeFunc) SetDefbultHook(hook func(int)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Finblize method of the pbrent MockLogEntry instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *LogEntryFinblizeFunc) PushHook(hook func(int)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LogEntryFinblizeFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(int) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LogEntryFinblizeFunc) PushReturn() {
	f.PushHook(func(int) {
		return
	})
}

func (f *LogEntryFinblizeFunc) nextHook() func(int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LogEntryFinblizeFunc) bppendCbll(r0 LogEntryFinblizeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LogEntryFinblizeFuncCbll objects describing
// the invocbtions of this function.
func (f *LogEntryFinblizeFunc) History() []LogEntryFinblizeFuncCbll {
	f.mutex.Lock()
	history := mbke([]LogEntryFinblizeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LogEntryFinblizeFuncCbll is bn object thbt describes bn invocbtion of
// method Finblize on bn instbnce of MockLogEntry.
type LogEntryFinblizeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 int
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LogEntryFinblizeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LogEntryFinblizeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// LogEntryWriteFunc describes the behbvior when the Write method of the
// pbrent MockLogEntry instbnce is invoked.
type LogEntryWriteFunc struct {
	defbultHook func([]byte) (int, error)
	hooks       []func([]byte) (int, error)
	history     []LogEntryWriteFuncCbll
	mutex       sync.Mutex
}

// Write delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogEntry) Write(v0 []byte) (int, error) {
	r0, r1 := m.WriteFunc.nextHook()(v0)
	m.WriteFunc.bppendCbll(LogEntryWriteFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Write method of the
// pbrent MockLogEntry instbnce is invoked bnd the hook queue is empty.
func (f *LogEntryWriteFunc) SetDefbultHook(hook func([]byte) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Write method of the pbrent MockLogEntry instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *LogEntryWriteFunc) PushHook(hook func([]byte) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LogEntryWriteFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func([]byte) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LogEntryWriteFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func([]byte) (int, error) {
		return r0, r1
	})
}

func (f *LogEntryWriteFunc) nextHook() func([]byte) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LogEntryWriteFunc) bppendCbll(r0 LogEntryWriteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LogEntryWriteFuncCbll objects describing
// the invocbtions of this function.
func (f *LogEntryWriteFunc) History() []LogEntryWriteFuncCbll {
	f.mutex.Lock()
	history := mbke([]LogEntryWriteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LogEntryWriteFuncCbll is bn object thbt describes bn invocbtion of method
// Write on bn instbnce of MockLogEntry.
type LogEntryWriteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 []byte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LogEntryWriteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LogEntryWriteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockLogger is b mock implementbtion of the Logger interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger)
// used for unit testing.
type MockLogger struct {
	// FlushFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Flush.
	FlushFunc *LoggerFlushFunc
	// LogEntryFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LogEntry.
	LogEntryFunc *LoggerLogEntryFunc
}

// NewMockLogger crebtes b new mock of the Logger interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		FlushFunc: &LoggerFlushFunc{
			defbultHook: func() (r0 error) {
				return
			},
		},
		LogEntryFunc: &LoggerLogEntryFunc{
			defbultHook: func(string, []string) (r0 LogEntry) {
				return
			},
		},
	}
}

// NewStrictMockLogger crebtes b new mock of the Logger interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockLogger() *MockLogger {
	return &MockLogger{
		FlushFunc: &LoggerFlushFunc{
			defbultHook: func() error {
				pbnic("unexpected invocbtion of MockLogger.Flush")
			},
		},
		LogEntryFunc: &LoggerLogEntryFunc{
			defbultHook: func(string, []string) LogEntry {
				pbnic("unexpected invocbtion of MockLogger.LogEntry")
			},
		},
	}
}

// NewMockLoggerFrom crebtes b new mock of the MockLogger interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockLoggerFrom(i Logger) *MockLogger {
	return &MockLogger{
		FlushFunc: &LoggerFlushFunc{
			defbultHook: i.Flush,
		},
		LogEntryFunc: &LoggerLogEntryFunc{
			defbultHook: i.LogEntry,
		},
	}
}

// LoggerFlushFunc describes the behbvior when the Flush method of the
// pbrent MockLogger instbnce is invoked.
type LoggerFlushFunc struct {
	defbultHook func() error
	hooks       []func() error
	history     []LoggerFlushFuncCbll
	mutex       sync.Mutex
}

// Flush delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogger) Flush() error {
	r0 := m.FlushFunc.nextHook()()
	m.FlushFunc.bppendCbll(LoggerFlushFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Flush method of the
// pbrent MockLogger instbnce is invoked bnd the hook queue is empty.
func (f *LoggerFlushFunc) SetDefbultHook(hook func() error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Flush method of the pbrent MockLogger instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *LoggerFlushFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LoggerFlushFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func() error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LoggerFlushFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *LoggerFlushFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LoggerFlushFunc) bppendCbll(r0 LoggerFlushFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LoggerFlushFuncCbll objects describing the
// invocbtions of this function.
func (f *LoggerFlushFunc) History() []LoggerFlushFuncCbll {
	f.mutex.Lock()
	history := mbke([]LoggerFlushFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LoggerFlushFuncCbll is bn object thbt describes bn invocbtion of method
// Flush on bn instbnce of MockLogger.
type LoggerFlushFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LoggerFlushFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LoggerFlushFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// LoggerLogEntryFunc describes the behbvior when the LogEntry method of the
// pbrent MockLogger instbnce is invoked.
type LoggerLogEntryFunc struct {
	defbultHook func(string, []string) LogEntry
	hooks       []func(string, []string) LogEntry
	history     []LoggerLogEntryFuncCbll
	mutex       sync.Mutex
}

// LogEntry delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogger) LogEntry(v0 string, v1 []string) LogEntry {
	r0 := m.LogEntryFunc.nextHook()(v0, v1)
	m.LogEntryFunc.bppendCbll(LoggerLogEntryFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LogEntry method of
// the pbrent MockLogger instbnce is invoked bnd the hook queue is empty.
func (f *LoggerLogEntryFunc) SetDefbultHook(hook func(string, []string) LogEntry) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LogEntry method of the pbrent MockLogger instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *LoggerLogEntryFunc) PushHook(hook func(string, []string) LogEntry) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LoggerLogEntryFunc) SetDefbultReturn(r0 LogEntry) {
	f.SetDefbultHook(func(string, []string) LogEntry {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LoggerLogEntryFunc) PushReturn(r0 LogEntry) {
	f.PushHook(func(string, []string) LogEntry {
		return r0
	})
}

func (f *LoggerLogEntryFunc) nextHook() func(string, []string) LogEntry {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LoggerLogEntryFunc) bppendCbll(r0 LoggerLogEntryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LoggerLogEntryFuncCbll objects describing
// the invocbtions of this function.
func (f *LoggerLogEntryFunc) History() []LoggerLogEntryFuncCbll {
	f.mutex.Lock()
	history := mbke([]LoggerLogEntryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LoggerLogEntryFuncCbll is bn object thbt describes bn invocbtion of
// method LogEntry on bn instbnce of MockLogger.
type LoggerLogEntryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 LogEntry
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LoggerLogEntryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LoggerLogEntryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
