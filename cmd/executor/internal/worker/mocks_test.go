// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge worker

import (
	"context"
	"io"
	"io/fs"
	"os/exec"
	"sync"

	util "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	cmdlogger "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	commbnd "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	files "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	runner "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	runtime "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runtime"
	workspbce "github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	executor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	types "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

// MockCommbnd is b mock implementbtion of the Commbnd interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd)
// used for unit testing.
type MockCommbnd struct {
	// RunFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Run.
	RunFunc *CommbndRunFunc
}

// NewMockCommbnd crebtes b new mock of the Commbnd interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockCommbnd() *MockCommbnd {
	return &MockCommbnd{
		RunFunc: &CommbndRunFunc{
			defbultHook: func(context.Context, cmdlogger.Logger, commbnd.Spec) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockCommbnd crebtes b new mock of the Commbnd interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockCommbnd() *MockCommbnd {
	return &MockCommbnd{
		RunFunc: &CommbndRunFunc{
			defbultHook: func(context.Context, cmdlogger.Logger, commbnd.Spec) error {
				pbnic("unexpected invocbtion of MockCommbnd.Run")
			},
		},
	}
}

// NewMockCommbndFrom crebtes b new mock of the MockCommbnd interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockCommbndFrom(i commbnd.Commbnd) *MockCommbnd {
	return &MockCommbnd{
		RunFunc: &CommbndRunFunc{
			defbultHook: i.Run,
		},
	}
}

// CommbndRunFunc describes the behbvior when the Run method of the pbrent
// MockCommbnd instbnce is invoked.
type CommbndRunFunc struct {
	defbultHook func(context.Context, cmdlogger.Logger, commbnd.Spec) error
	hooks       []func(context.Context, cmdlogger.Logger, commbnd.Spec) error
	history     []CommbndRunFuncCbll
	mutex       sync.Mutex
}

// Run delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCommbnd) Run(v0 context.Context, v1 cmdlogger.Logger, v2 commbnd.Spec) error {
	r0 := m.RunFunc.nextHook()(v0, v1, v2)
	m.RunFunc.bppendCbll(CommbndRunFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Run method of the
// pbrent MockCommbnd instbnce is invoked bnd the hook queue is empty.
func (f *CommbndRunFunc) SetDefbultHook(hook func(context.Context, cmdlogger.Logger, commbnd.Spec) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Run method of the pbrent MockCommbnd instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *CommbndRunFunc) PushHook(hook func(context.Context, cmdlogger.Logger, commbnd.Spec) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CommbndRunFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, cmdlogger.Logger, commbnd.Spec) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CommbndRunFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, cmdlogger.Logger, commbnd.Spec) error {
		return r0
	})
}

func (f *CommbndRunFunc) nextHook() func(context.Context, cmdlogger.Logger, commbnd.Spec) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommbndRunFunc) bppendCbll(r0 CommbndRunFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CommbndRunFuncCbll objects describing the
// invocbtions of this function.
func (f *CommbndRunFunc) History() []CommbndRunFuncCbll {
	f.mutex.Lock()
	history := mbke([]CommbndRunFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommbndRunFuncCbll is bn object thbt describes bn invocbtion of method
// Run on bn instbnce of MockCommbnd.
type CommbndRunFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 cmdlogger.Logger
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 commbnd.Spec
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CommbndRunFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CommbndRunFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

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
func NewMockExecutionLogEntryStoreFrom(i cmdlogger.ExecutionLogEntryStore) *MockExecutionLogEntryStore {
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
func NewMockLogEntryFrom(i cmdlogger.LogEntry) *MockLogEntry {
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
			defbultHook: func(string, []string) (r0 cmdlogger.LogEntry) {
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
			defbultHook: func(string, []string) cmdlogger.LogEntry {
				pbnic("unexpected invocbtion of MockLogger.LogEntry")
			},
		},
	}
}

// NewMockLoggerFrom crebtes b new mock of the MockLogger interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockLoggerFrom(i cmdlogger.Logger) *MockLogger {
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
	defbultHook func(string, []string) cmdlogger.LogEntry
	hooks       []func(string, []string) cmdlogger.LogEntry
	history     []LoggerLogEntryFuncCbll
	mutex       sync.Mutex
}

// LogEntry delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLogger) LogEntry(v0 string, v1 []string) cmdlogger.LogEntry {
	r0 := m.LogEntryFunc.nextHook()(v0, v1)
	m.LogEntryFunc.bppendCbll(LoggerLogEntryFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LogEntry method of
// the pbrent MockLogger instbnce is invoked bnd the hook queue is empty.
func (f *LoggerLogEntryFunc) SetDefbultHook(hook func(string, []string) cmdlogger.LogEntry) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LogEntry method of the pbrent MockLogger instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *LoggerLogEntryFunc) PushHook(hook func(string, []string) cmdlogger.LogEntry) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LoggerLogEntryFunc) SetDefbultReturn(r0 cmdlogger.LogEntry) {
	f.SetDefbultHook(func(string, []string) cmdlogger.LogEntry {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LoggerLogEntryFunc) PushReturn(r0 cmdlogger.LogEntry) {
	f.PushHook(func(string, []string) cmdlogger.LogEntry {
		return r0
	})
}

func (f *LoggerLogEntryFunc) nextHook() func(string, []string) cmdlogger.LogEntry {
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
	Result0 cmdlogger.LogEntry
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

// MockCmdRunner is b mock implementbtion of the CmdRunner interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util) used for
// unit testing.
type MockCmdRunner struct {
	// CombinedOutputFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CombinedOutput.
	CombinedOutputFunc *CmdRunnerCombinedOutputFunc
	// CommbndContextFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CommbndContext.
	CommbndContextFunc *CmdRunnerCommbndContextFunc
	// LookPbthFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LookPbth.
	LookPbthFunc *CmdRunnerLookPbthFunc
	// StbtFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Stbt.
	StbtFunc *CmdRunnerStbtFunc
}

// NewMockCmdRunner crebtes b new mock of the CmdRunner interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockCmdRunner() *MockCmdRunner {
	return &MockCmdRunner{
		CombinedOutputFunc: &CmdRunnerCombinedOutputFunc{
			defbultHook: func(context.Context, string, ...string) (r0 []byte, r1 error) {
				return
			},
		},
		CommbndContextFunc: &CmdRunnerCommbndContextFunc{
			defbultHook: func(context.Context, string, ...string) (r0 *exec.Cmd) {
				return
			},
		},
		LookPbthFunc: &CmdRunnerLookPbthFunc{
			defbultHook: func(string) (r0 string, r1 error) {
				return
			},
		},
		StbtFunc: &CmdRunnerStbtFunc{
			defbultHook: func(string) (r0 fs.FileInfo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockCmdRunner crebtes b new mock of the CmdRunner interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockCmdRunner() *MockCmdRunner {
	return &MockCmdRunner{
		CombinedOutputFunc: &CmdRunnerCombinedOutputFunc{
			defbultHook: func(context.Context, string, ...string) ([]byte, error) {
				pbnic("unexpected invocbtion of MockCmdRunner.CombinedOutput")
			},
		},
		CommbndContextFunc: &CmdRunnerCommbndContextFunc{
			defbultHook: func(context.Context, string, ...string) *exec.Cmd {
				pbnic("unexpected invocbtion of MockCmdRunner.CommbndContext")
			},
		},
		LookPbthFunc: &CmdRunnerLookPbthFunc{
			defbultHook: func(string) (string, error) {
				pbnic("unexpected invocbtion of MockCmdRunner.LookPbth")
			},
		},
		StbtFunc: &CmdRunnerStbtFunc{
			defbultHook: func(string) (fs.FileInfo, error) {
				pbnic("unexpected invocbtion of MockCmdRunner.Stbt")
			},
		},
	}
}

// NewMockCmdRunnerFrom crebtes b new mock of the MockCmdRunner interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockCmdRunnerFrom(i util.CmdRunner) *MockCmdRunner {
	return &MockCmdRunner{
		CombinedOutputFunc: &CmdRunnerCombinedOutputFunc{
			defbultHook: i.CombinedOutput,
		},
		CommbndContextFunc: &CmdRunnerCommbndContextFunc{
			defbultHook: i.CommbndContext,
		},
		LookPbthFunc: &CmdRunnerLookPbthFunc{
			defbultHook: i.LookPbth,
		},
		StbtFunc: &CmdRunnerStbtFunc{
			defbultHook: i.Stbt,
		},
	}
}

// CmdRunnerCombinedOutputFunc describes the behbvior when the
// CombinedOutput method of the pbrent MockCmdRunner instbnce is invoked.
type CmdRunnerCombinedOutputFunc struct {
	defbultHook func(context.Context, string, ...string) ([]byte, error)
	hooks       []func(context.Context, string, ...string) ([]byte, error)
	history     []CmdRunnerCombinedOutputFuncCbll
	mutex       sync.Mutex
}

// CombinedOutput delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCmdRunner) CombinedOutput(v0 context.Context, v1 string, v2 ...string) ([]byte, error) {
	r0, r1 := m.CombinedOutputFunc.nextHook()(v0, v1, v2...)
	m.CombinedOutputFunc.bppendCbll(CmdRunnerCombinedOutputFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CombinedOutput
// method of the pbrent MockCmdRunner instbnce is invoked bnd the hook queue
// is empty.
func (f *CmdRunnerCombinedOutputFunc) SetDefbultHook(hook func(context.Context, string, ...string) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CombinedOutput method of the pbrent MockCmdRunner instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *CmdRunnerCombinedOutputFunc) PushHook(hook func(context.Context, string, ...string) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CmdRunnerCombinedOutputFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, string, ...string) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CmdRunnerCombinedOutputFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, string, ...string) ([]byte, error) {
		return r0, r1
	})
}

func (f *CmdRunnerCombinedOutputFunc) nextHook() func(context.Context, string, ...string) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CmdRunnerCombinedOutputFunc) bppendCbll(r0 CmdRunnerCombinedOutputFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CmdRunnerCombinedOutputFuncCbll objects
// describing the invocbtions of this function.
func (f *CmdRunnerCombinedOutputFunc) History() []CmdRunnerCombinedOutputFuncCbll {
	f.mutex.Lock()
	history := mbke([]CmdRunnerCombinedOutputFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CmdRunnerCombinedOutputFuncCbll is bn object thbt describes bn invocbtion
// of method CombinedOutput on bn instbnce of MockCmdRunner.
type CmdRunnerCombinedOutputFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg2 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []byte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c CmdRunnerCombinedOutputFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg2 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CmdRunnerCombinedOutputFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CmdRunnerCommbndContextFunc describes the behbvior when the
// CommbndContext method of the pbrent MockCmdRunner instbnce is invoked.
type CmdRunnerCommbndContextFunc struct {
	defbultHook func(context.Context, string, ...string) *exec.Cmd
	hooks       []func(context.Context, string, ...string) *exec.Cmd
	history     []CmdRunnerCommbndContextFuncCbll
	mutex       sync.Mutex
}

// CommbndContext delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCmdRunner) CommbndContext(v0 context.Context, v1 string, v2 ...string) *exec.Cmd {
	r0 := m.CommbndContextFunc.nextHook()(v0, v1, v2...)
	m.CommbndContextFunc.bppendCbll(CmdRunnerCommbndContextFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the CommbndContext
// method of the pbrent MockCmdRunner instbnce is invoked bnd the hook queue
// is empty.
func (f *CmdRunnerCommbndContextFunc) SetDefbultHook(hook func(context.Context, string, ...string) *exec.Cmd) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommbndContext method of the pbrent MockCmdRunner instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *CmdRunnerCommbndContextFunc) PushHook(hook func(context.Context, string, ...string) *exec.Cmd) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CmdRunnerCommbndContextFunc) SetDefbultReturn(r0 *exec.Cmd) {
	f.SetDefbultHook(func(context.Context, string, ...string) *exec.Cmd {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CmdRunnerCommbndContextFunc) PushReturn(r0 *exec.Cmd) {
	f.PushHook(func(context.Context, string, ...string) *exec.Cmd {
		return r0
	})
}

func (f *CmdRunnerCommbndContextFunc) nextHook() func(context.Context, string, ...string) *exec.Cmd {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CmdRunnerCommbndContextFunc) bppendCbll(r0 CmdRunnerCommbndContextFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CmdRunnerCommbndContextFuncCbll objects
// describing the invocbtions of this function.
func (f *CmdRunnerCommbndContextFunc) History() []CmdRunnerCommbndContextFuncCbll {
	f.mutex.Lock()
	history := mbke([]CmdRunnerCommbndContextFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CmdRunnerCommbndContextFuncCbll is bn object thbt describes bn invocbtion
// of method CommbndContext on bn instbnce of MockCmdRunner.
type CmdRunnerCommbndContextFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg2 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *exec.Cmd
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c CmdRunnerCommbndContextFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg2 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CmdRunnerCommbndContextFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// CmdRunnerLookPbthFunc describes the behbvior when the LookPbth method of
// the pbrent MockCmdRunner instbnce is invoked.
type CmdRunnerLookPbthFunc struct {
	defbultHook func(string) (string, error)
	hooks       []func(string) (string, error)
	history     []CmdRunnerLookPbthFuncCbll
	mutex       sync.Mutex
}

// LookPbth delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCmdRunner) LookPbth(v0 string) (string, error) {
	r0, r1 := m.LookPbthFunc.nextHook()(v0)
	m.LookPbthFunc.bppendCbll(CmdRunnerLookPbthFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the LookPbth method of
// the pbrent MockCmdRunner instbnce is invoked bnd the hook queue is empty.
func (f *CmdRunnerLookPbthFunc) SetDefbultHook(hook func(string) (string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LookPbth method of the pbrent MockCmdRunner instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *CmdRunnerLookPbthFunc) PushHook(hook func(string) (string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CmdRunnerLookPbthFunc) SetDefbultReturn(r0 string, r1 error) {
	f.SetDefbultHook(func(string) (string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CmdRunnerLookPbthFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(string) (string, error) {
		return r0, r1
	})
}

func (f *CmdRunnerLookPbthFunc) nextHook() func(string) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CmdRunnerLookPbthFunc) bppendCbll(r0 CmdRunnerLookPbthFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CmdRunnerLookPbthFuncCbll objects
// describing the invocbtions of this function.
func (f *CmdRunnerLookPbthFunc) History() []CmdRunnerLookPbthFuncCbll {
	f.mutex.Lock()
	history := mbke([]CmdRunnerLookPbthFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CmdRunnerLookPbthFuncCbll is bn object thbt describes bn invocbtion of
// method LookPbth on bn instbnce of MockCmdRunner.
type CmdRunnerLookPbthFuncCbll struct {
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
func (c CmdRunnerLookPbthFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CmdRunnerLookPbthFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CmdRunnerStbtFunc describes the behbvior when the Stbt method of the
// pbrent MockCmdRunner instbnce is invoked.
type CmdRunnerStbtFunc struct {
	defbultHook func(string) (fs.FileInfo, error)
	hooks       []func(string) (fs.FileInfo, error)
	history     []CmdRunnerStbtFuncCbll
	mutex       sync.Mutex
}

// Stbt delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCmdRunner) Stbt(v0 string) (fs.FileInfo, error) {
	r0, r1 := m.StbtFunc.nextHook()(v0)
	m.StbtFunc.bppendCbll(CmdRunnerStbtFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Stbt method of the
// pbrent MockCmdRunner instbnce is invoked bnd the hook queue is empty.
func (f *CmdRunnerStbtFunc) SetDefbultHook(hook func(string) (fs.FileInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Stbt method of the pbrent MockCmdRunner instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *CmdRunnerStbtFunc) PushHook(hook func(string) (fs.FileInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CmdRunnerStbtFunc) SetDefbultReturn(r0 fs.FileInfo, r1 error) {
	f.SetDefbultHook(func(string) (fs.FileInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CmdRunnerStbtFunc) PushReturn(r0 fs.FileInfo, r1 error) {
	f.PushHook(func(string) (fs.FileInfo, error) {
		return r0, r1
	})
}

func (f *CmdRunnerStbtFunc) nextHook() func(string) (fs.FileInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CmdRunnerStbtFunc) bppendCbll(r0 CmdRunnerStbtFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CmdRunnerStbtFuncCbll objects describing
// the invocbtions of this function.
func (f *CmdRunnerStbtFunc) History() []CmdRunnerStbtFuncCbll {
	f.mutex.Lock()
	history := mbke([]CmdRunnerStbtFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CmdRunnerStbtFuncCbll is bn object thbt describes bn invocbtion of method
// Stbt on bn instbnce of MockCmdRunner.
type CmdRunnerStbtFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 fs.FileInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CmdRunnerStbtFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CmdRunnerStbtFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files)
// used for unit testing.
type MockStore struct {
	// ExistsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Exists.
	ExistsFunc *StoreExistsFunc
	// GetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Get.
	GetFunc *StoreGetFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		ExistsFunc: &StoreExistsFunc{
			defbultHook: func(context.Context, types.Job, string, string) (r0 bool, r1 error) {
				return
			},
		},
		GetFunc: &StoreGetFunc{
			defbultHook: func(context.Context, types.Job, string, string) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		ExistsFunc: &StoreExistsFunc{
			defbultHook: func(context.Context, types.Job, string, string) (bool, error) {
				pbnic("unexpected invocbtion of MockStore.Exists")
			},
		},
		GetFunc: &StoreGetFunc{
			defbultHook: func(context.Context, types.Job, string, string) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockStore.Get")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i files.Store) *MockStore {
	return &MockStore{
		ExistsFunc: &StoreExistsFunc{
			defbultHook: i.Exists,
		},
		GetFunc: &StoreGetFunc{
			defbultHook: i.Get,
		},
	}
}

// StoreExistsFunc describes the behbvior when the Exists method of the
// pbrent MockStore instbnce is invoked.
type StoreExistsFunc struct {
	defbultHook func(context.Context, types.Job, string, string) (bool, error)
	hooks       []func(context.Context, types.Job, string, string) (bool, error)
	history     []StoreExistsFuncCbll
	mutex       sync.Mutex
}

// Exists delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Exists(v0 context.Context, v1 types.Job, v2 string, v3 string) (bool, error) {
	r0, r1 := m.ExistsFunc.nextHook()(v0, v1, v2, v3)
	m.ExistsFunc.bppendCbll(StoreExistsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Exists method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreExistsFunc) SetDefbultHook(hook func(context.Context, types.Job, string, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Exists method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreExistsFunc) PushHook(hook func(context.Context, types.Job, string, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreExistsFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, types.Job, string, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreExistsFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, types.Job, string, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreExistsFunc) nextHook() func(context.Context, types.Job, string, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreExistsFunc) bppendCbll(r0 StoreExistsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreExistsFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreExistsFunc) History() []StoreExistsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreExistsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreExistsFuncCbll is bn object thbt describes bn invocbtion of method
// Exists on bn instbnce of MockStore.
type StoreExistsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.Job
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreExistsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreExistsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetFunc describes the behbvior when the Get method of the pbrent
// MockStore instbnce is invoked.
type StoreGetFunc struct {
	defbultHook func(context.Context, types.Job, string, string) (io.RebdCloser, error)
	hooks       []func(context.Context, types.Job, string, string) (io.RebdCloser, error)
	history     []StoreGetFuncCbll
	mutex       sync.Mutex
}

// Get delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Get(v0 context.Context, v1 types.Job, v2 string, v3 string) (io.RebdCloser, error) {
	r0, r1 := m.GetFunc.nextHook()(v0, v1, v2, v3)
	m.GetFunc.bppendCbll(StoreGetFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Get method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreGetFunc) SetDefbultHook(hook func(context.Context, types.Job, string, string) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Get method of the pbrent MockStore instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *StoreGetFunc) PushHook(hook func(context.Context, types.Job, string, string) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, types.Job, string, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, types.Job, string, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *StoreGetFunc) nextHook() func(context.Context, types.Job, string, string) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetFunc) bppendCbll(r0 StoreGetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreGetFunc) History() []StoreGetFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetFuncCbll is bn object thbt describes bn invocbtion of method Get
// on bn instbnce of MockStore.
type StoreGetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.Job
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockRuntime is b mock implementbtion of the Runtime interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runtime)
// used for unit testing.
type MockRuntime struct {
	// NbmeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Nbme.
	NbmeFunc *RuntimeNbmeFunc
	// NewRunnerFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method NewRunner.
	NewRunnerFunc *RuntimeNewRunnerFunc
	// NewRunnerSpecsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewRunnerSpecs.
	NewRunnerSpecsFunc *RuntimeNewRunnerSpecsFunc
	// PrepbreWorkspbceFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method PrepbreWorkspbce.
	PrepbreWorkspbceFunc *RuntimePrepbreWorkspbceFunc
}

// NewMockRuntime crebtes b new mock of the Runtime interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{
		NbmeFunc: &RuntimeNbmeFunc{
			defbultHook: func() (r0 runtime.Nbme) {
				return
			},
		},
		NewRunnerFunc: &RuntimeNewRunnerFunc{
			defbultHook: func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (r0 runner.Runner, r1 error) {
				return
			},
		},
		NewRunnerSpecsFunc: &RuntimeNewRunnerSpecsFunc{
			defbultHook: func(workspbce.Workspbce, types.Job) (r0 []runner.Spec, r1 error) {
				return
			},
		},
		PrepbreWorkspbceFunc: &RuntimePrepbreWorkspbceFunc{
			defbultHook: func(context.Context, cmdlogger.Logger, types.Job) (r0 workspbce.Workspbce, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockRuntime crebtes b new mock of the Runtime interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockRuntime() *MockRuntime {
	return &MockRuntime{
		NbmeFunc: &RuntimeNbmeFunc{
			defbultHook: func() runtime.Nbme {
				pbnic("unexpected invocbtion of MockRuntime.Nbme")
			},
		},
		NewRunnerFunc: &RuntimeNewRunnerFunc{
			defbultHook: func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error) {
				pbnic("unexpected invocbtion of MockRuntime.NewRunner")
			},
		},
		NewRunnerSpecsFunc: &RuntimeNewRunnerSpecsFunc{
			defbultHook: func(workspbce.Workspbce, types.Job) ([]runner.Spec, error) {
				pbnic("unexpected invocbtion of MockRuntime.NewRunnerSpecs")
			},
		},
		PrepbreWorkspbceFunc: &RuntimePrepbreWorkspbceFunc{
			defbultHook: func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error) {
				pbnic("unexpected invocbtion of MockRuntime.PrepbreWorkspbce")
			},
		},
	}
}

// NewMockRuntimeFrom crebtes b new mock of the MockRuntime interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockRuntimeFrom(i runtime.Runtime) *MockRuntime {
	return &MockRuntime{
		NbmeFunc: &RuntimeNbmeFunc{
			defbultHook: i.Nbme,
		},
		NewRunnerFunc: &RuntimeNewRunnerFunc{
			defbultHook: i.NewRunner,
		},
		NewRunnerSpecsFunc: &RuntimeNewRunnerSpecsFunc{
			defbultHook: i.NewRunnerSpecs,
		},
		PrepbreWorkspbceFunc: &RuntimePrepbreWorkspbceFunc{
			defbultHook: i.PrepbreWorkspbce,
		},
	}
}

// RuntimeNbmeFunc describes the behbvior when the Nbme method of the pbrent
// MockRuntime instbnce is invoked.
type RuntimeNbmeFunc struct {
	defbultHook func() runtime.Nbme
	hooks       []func() runtime.Nbme
	history     []RuntimeNbmeFuncCbll
	mutex       sync.Mutex
}

// Nbme delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRuntime) Nbme() runtime.Nbme {
	r0 := m.NbmeFunc.nextHook()()
	m.NbmeFunc.bppendCbll(RuntimeNbmeFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Nbme method of the
// pbrent MockRuntime instbnce is invoked bnd the hook queue is empty.
func (f *RuntimeNbmeFunc) SetDefbultHook(hook func() runtime.Nbme) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Nbme method of the pbrent MockRuntime instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *RuntimeNbmeFunc) PushHook(hook func() runtime.Nbme) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RuntimeNbmeFunc) SetDefbultReturn(r0 runtime.Nbme) {
	f.SetDefbultHook(func() runtime.Nbme {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RuntimeNbmeFunc) PushReturn(r0 runtime.Nbme) {
	f.PushHook(func() runtime.Nbme {
		return r0
	})
}

func (f *RuntimeNbmeFunc) nextHook() func() runtime.Nbme {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RuntimeNbmeFunc) bppendCbll(r0 RuntimeNbmeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RuntimeNbmeFuncCbll objects describing the
// invocbtions of this function.
func (f *RuntimeNbmeFunc) History() []RuntimeNbmeFuncCbll {
	f.mutex.Lock()
	history := mbke([]RuntimeNbmeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RuntimeNbmeFuncCbll is bn object thbt describes bn invocbtion of method
// Nbme on bn instbnce of MockRuntime.
type RuntimeNbmeFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 runtime.Nbme
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RuntimeNbmeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RuntimeNbmeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RuntimeNewRunnerFunc describes the behbvior when the NewRunner method of
// the pbrent MockRuntime instbnce is invoked.
type RuntimeNewRunnerFunc struct {
	defbultHook func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error)
	hooks       []func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error)
	history     []RuntimeNewRunnerFuncCbll
	mutex       sync.Mutex
}

// NewRunner delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRuntime) NewRunner(v0 context.Context, v1 cmdlogger.Logger, v2 files.Store, v3 runtime.RunnerOptions) (runner.Runner, error) {
	r0, r1 := m.NewRunnerFunc.nextHook()(v0, v1, v2, v3)
	m.NewRunnerFunc.bppendCbll(RuntimeNewRunnerFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the NewRunner method of
// the pbrent MockRuntime instbnce is invoked bnd the hook queue is empty.
func (f *RuntimeNewRunnerFunc) SetDefbultHook(hook func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewRunner method of the pbrent MockRuntime instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *RuntimeNewRunnerFunc) PushHook(hook func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RuntimeNewRunnerFunc) SetDefbultReturn(r0 runner.Runner, r1 error) {
	f.SetDefbultHook(func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RuntimeNewRunnerFunc) PushReturn(r0 runner.Runner, r1 error) {
	f.PushHook(func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error) {
		return r0, r1
	})
}

func (f *RuntimeNewRunnerFunc) nextHook() func(context.Context, cmdlogger.Logger, files.Store, runtime.RunnerOptions) (runner.Runner, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RuntimeNewRunnerFunc) bppendCbll(r0 RuntimeNewRunnerFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RuntimeNewRunnerFuncCbll objects describing
// the invocbtions of this function.
func (f *RuntimeNewRunnerFunc) History() []RuntimeNewRunnerFuncCbll {
	f.mutex.Lock()
	history := mbke([]RuntimeNewRunnerFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RuntimeNewRunnerFuncCbll is bn object thbt describes bn invocbtion of
// method NewRunner on bn instbnce of MockRuntime.
type RuntimeNewRunnerFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 cmdlogger.Logger
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 files.Store
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 runtime.RunnerOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 runner.Runner
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RuntimeNewRunnerFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RuntimeNewRunnerFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RuntimeNewRunnerSpecsFunc describes the behbvior when the NewRunnerSpecs
// method of the pbrent MockRuntime instbnce is invoked.
type RuntimeNewRunnerSpecsFunc struct {
	defbultHook func(workspbce.Workspbce, types.Job) ([]runner.Spec, error)
	hooks       []func(workspbce.Workspbce, types.Job) ([]runner.Spec, error)
	history     []RuntimeNewRunnerSpecsFuncCbll
	mutex       sync.Mutex
}

// NewRunnerSpecs delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRuntime) NewRunnerSpecs(v0 workspbce.Workspbce, v1 types.Job) ([]runner.Spec, error) {
	r0, r1 := m.NewRunnerSpecsFunc.nextHook()(v0, v1)
	m.NewRunnerSpecsFunc.bppendCbll(RuntimeNewRunnerSpecsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the NewRunnerSpecs
// method of the pbrent MockRuntime instbnce is invoked bnd the hook queue
// is empty.
func (f *RuntimeNewRunnerSpecsFunc) SetDefbultHook(hook func(workspbce.Workspbce, types.Job) ([]runner.Spec, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewRunnerSpecs method of the pbrent MockRuntime instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *RuntimeNewRunnerSpecsFunc) PushHook(hook func(workspbce.Workspbce, types.Job) ([]runner.Spec, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RuntimeNewRunnerSpecsFunc) SetDefbultReturn(r0 []runner.Spec, r1 error) {
	f.SetDefbultHook(func(workspbce.Workspbce, types.Job) ([]runner.Spec, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RuntimeNewRunnerSpecsFunc) PushReturn(r0 []runner.Spec, r1 error) {
	f.PushHook(func(workspbce.Workspbce, types.Job) ([]runner.Spec, error) {
		return r0, r1
	})
}

func (f *RuntimeNewRunnerSpecsFunc) nextHook() func(workspbce.Workspbce, types.Job) ([]runner.Spec, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RuntimeNewRunnerSpecsFunc) bppendCbll(r0 RuntimeNewRunnerSpecsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RuntimeNewRunnerSpecsFuncCbll objects
// describing the invocbtions of this function.
func (f *RuntimeNewRunnerSpecsFunc) History() []RuntimeNewRunnerSpecsFuncCbll {
	f.mutex.Lock()
	history := mbke([]RuntimeNewRunnerSpecsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RuntimeNewRunnerSpecsFuncCbll is bn object thbt describes bn invocbtion
// of method NewRunnerSpecs on bn instbnce of MockRuntime.
type RuntimeNewRunnerSpecsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 workspbce.Workspbce
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.Job
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []runner.Spec
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RuntimeNewRunnerSpecsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RuntimeNewRunnerSpecsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RuntimePrepbreWorkspbceFunc describes the behbvior when the
// PrepbreWorkspbce method of the pbrent MockRuntime instbnce is invoked.
type RuntimePrepbreWorkspbceFunc struct {
	defbultHook func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error)
	hooks       []func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error)
	history     []RuntimePrepbreWorkspbceFuncCbll
	mutex       sync.Mutex
}

// PrepbreWorkspbce delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRuntime) PrepbreWorkspbce(v0 context.Context, v1 cmdlogger.Logger, v2 types.Job) (workspbce.Workspbce, error) {
	r0, r1 := m.PrepbreWorkspbceFunc.nextHook()(v0, v1, v2)
	m.PrepbreWorkspbceFunc.bppendCbll(RuntimePrepbreWorkspbceFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the PrepbreWorkspbce
// method of the pbrent MockRuntime instbnce is invoked bnd the hook queue
// is empty.
func (f *RuntimePrepbreWorkspbceFunc) SetDefbultHook(hook func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// PrepbreWorkspbce method of the pbrent MockRuntime instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *RuntimePrepbreWorkspbceFunc) PushHook(hook func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RuntimePrepbreWorkspbceFunc) SetDefbultReturn(r0 workspbce.Workspbce, r1 error) {
	f.SetDefbultHook(func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RuntimePrepbreWorkspbceFunc) PushReturn(r0 workspbce.Workspbce, r1 error) {
	f.PushHook(func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error) {
		return r0, r1
	})
}

func (f *RuntimePrepbreWorkspbceFunc) nextHook() func(context.Context, cmdlogger.Logger, types.Job) (workspbce.Workspbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RuntimePrepbreWorkspbceFunc) bppendCbll(r0 RuntimePrepbreWorkspbceFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RuntimePrepbreWorkspbceFuncCbll objects
// describing the invocbtions of this function.
func (f *RuntimePrepbreWorkspbceFunc) History() []RuntimePrepbreWorkspbceFuncCbll {
	f.mutex.Lock()
	history := mbke([]RuntimePrepbreWorkspbceFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RuntimePrepbreWorkspbceFuncCbll is bn object thbt describes bn invocbtion
// of method PrepbreWorkspbce on bn instbnce of MockRuntime.
type RuntimePrepbreWorkspbceFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 cmdlogger.Logger
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 types.Job
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 workspbce.Workspbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RuntimePrepbreWorkspbceFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RuntimePrepbreWorkspbceFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockRunner is b mock implementbtion of the Runner interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner)
// used for unit testing.
type MockRunner struct {
	// RunFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Run.
	RunFunc *RunnerRunFunc
	// SetupFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Setup.
	SetupFunc *RunnerSetupFunc
	// TebrdownFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Tebrdown.
	TebrdownFunc *RunnerTebrdownFunc
	// TempDirFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method TempDir.
	TempDirFunc *RunnerTempDirFunc
}

// NewMockRunner crebtes b new mock of the Runner interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		RunFunc: &RunnerRunFunc{
			defbultHook: func(context.Context, runner.Spec) (r0 error) {
				return
			},
		},
		SetupFunc: &RunnerSetupFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		TebrdownFunc: &RunnerTebrdownFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		TempDirFunc: &RunnerTempDirFunc{
			defbultHook: func() (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockRunner crebtes b new mock of the Runner interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockRunner() *MockRunner {
	return &MockRunner{
		RunFunc: &RunnerRunFunc{
			defbultHook: func(context.Context, runner.Spec) error {
				pbnic("unexpected invocbtion of MockRunner.Run")
			},
		},
		SetupFunc: &RunnerSetupFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockRunner.Setup")
			},
		},
		TebrdownFunc: &RunnerTebrdownFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockRunner.Tebrdown")
			},
		},
		TempDirFunc: &RunnerTempDirFunc{
			defbultHook: func() string {
				pbnic("unexpected invocbtion of MockRunner.TempDir")
			},
		},
	}
}

// NewMockRunnerFrom crebtes b new mock of the MockRunner interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockRunnerFrom(i runner.Runner) *MockRunner {
	return &MockRunner{
		RunFunc: &RunnerRunFunc{
			defbultHook: i.Run,
		},
		SetupFunc: &RunnerSetupFunc{
			defbultHook: i.Setup,
		},
		TebrdownFunc: &RunnerTebrdownFunc{
			defbultHook: i.Tebrdown,
		},
		TempDirFunc: &RunnerTempDirFunc{
			defbultHook: i.TempDir,
		},
	}
}

// RunnerRunFunc describes the behbvior when the Run method of the pbrent
// MockRunner instbnce is invoked.
type RunnerRunFunc struct {
	defbultHook func(context.Context, runner.Spec) error
	hooks       []func(context.Context, runner.Spec) error
	history     []RunnerRunFuncCbll
	mutex       sync.Mutex
}

// Run delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRunner) Run(v0 context.Context, v1 runner.Spec) error {
	r0 := m.RunFunc.nextHook()(v0, v1)
	m.RunFunc.bppendCbll(RunnerRunFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Run method of the
// pbrent MockRunner instbnce is invoked bnd the hook queue is empty.
func (f *RunnerRunFunc) SetDefbultHook(hook func(context.Context, runner.Spec) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Run method of the pbrent MockRunner instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *RunnerRunFunc) PushHook(hook func(context.Context, runner.Spec) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RunnerRunFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, runner.Spec) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RunnerRunFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, runner.Spec) error {
		return r0
	})
}

func (f *RunnerRunFunc) nextHook() func(context.Context, runner.Spec) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RunnerRunFunc) bppendCbll(r0 RunnerRunFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RunnerRunFuncCbll objects describing the
// invocbtions of this function.
func (f *RunnerRunFunc) History() []RunnerRunFuncCbll {
	f.mutex.Lock()
	history := mbke([]RunnerRunFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RunnerRunFuncCbll is bn object thbt describes bn invocbtion of method Run
// on bn instbnce of MockRunner.
type RunnerRunFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 runner.Spec
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RunnerRunFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RunnerRunFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RunnerSetupFunc describes the behbvior when the Setup method of the
// pbrent MockRunner instbnce is invoked.
type RunnerSetupFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []RunnerSetupFuncCbll
	mutex       sync.Mutex
}

// Setup delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRunner) Setup(v0 context.Context) error {
	r0 := m.SetupFunc.nextHook()(v0)
	m.SetupFunc.bppendCbll(RunnerSetupFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Setup method of the
// pbrent MockRunner instbnce is invoked bnd the hook queue is empty.
func (f *RunnerSetupFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Setup method of the pbrent MockRunner instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *RunnerSetupFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RunnerSetupFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RunnerSetupFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *RunnerSetupFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RunnerSetupFunc) bppendCbll(r0 RunnerSetupFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RunnerSetupFuncCbll objects describing the
// invocbtions of this function.
func (f *RunnerSetupFunc) History() []RunnerSetupFuncCbll {
	f.mutex.Lock()
	history := mbke([]RunnerSetupFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RunnerSetupFuncCbll is bn object thbt describes bn invocbtion of method
// Setup on bn instbnce of MockRunner.
type RunnerSetupFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RunnerSetupFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RunnerSetupFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RunnerTebrdownFunc describes the behbvior when the Tebrdown method of the
// pbrent MockRunner instbnce is invoked.
type RunnerTebrdownFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []RunnerTebrdownFuncCbll
	mutex       sync.Mutex
}

// Tebrdown delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRunner) Tebrdown(v0 context.Context) error {
	r0 := m.TebrdownFunc.nextHook()(v0)
	m.TebrdownFunc.bppendCbll(RunnerTebrdownFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Tebrdown method of
// the pbrent MockRunner instbnce is invoked bnd the hook queue is empty.
func (f *RunnerTebrdownFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Tebrdown method of the pbrent MockRunner instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *RunnerTebrdownFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RunnerTebrdownFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RunnerTebrdownFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *RunnerTebrdownFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RunnerTebrdownFunc) bppendCbll(r0 RunnerTebrdownFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RunnerTebrdownFuncCbll objects describing
// the invocbtions of this function.
func (f *RunnerTebrdownFunc) History() []RunnerTebrdownFuncCbll {
	f.mutex.Lock()
	history := mbke([]RunnerTebrdownFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RunnerTebrdownFuncCbll is bn object thbt describes bn invocbtion of
// method Tebrdown on bn instbnce of MockRunner.
type RunnerTebrdownFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RunnerTebrdownFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RunnerTebrdownFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RunnerTempDirFunc describes the behbvior when the TempDir method of the
// pbrent MockRunner instbnce is invoked.
type RunnerTempDirFunc struct {
	defbultHook func() string
	hooks       []func() string
	history     []RunnerTempDirFuncCbll
	mutex       sync.Mutex
}

// TempDir delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRunner) TempDir() string {
	r0 := m.TempDirFunc.nextHook()()
	m.TempDirFunc.bppendCbll(RunnerTempDirFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the TempDir method of
// the pbrent MockRunner instbnce is invoked bnd the hook queue is empty.
func (f *RunnerTempDirFunc) SetDefbultHook(hook func() string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// TempDir method of the pbrent MockRunner instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *RunnerTempDirFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RunnerTempDirFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func() string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RunnerTempDirFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *RunnerTempDirFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RunnerTempDirFunc) bppendCbll(r0 RunnerTempDirFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RunnerTempDirFuncCbll objects describing
// the invocbtions of this function.
func (f *RunnerTempDirFunc) History() []RunnerTempDirFuncCbll {
	f.mutex.Lock()
	history := mbke([]RunnerTempDirFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RunnerTempDirFuncCbll is bn object thbt describes bn invocbtion of method
// TempDir on bn instbnce of MockRunner.
type RunnerTempDirFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RunnerTempDirFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RunnerTempDirFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockWorkspbce is b mock implementbtion of the Workspbce interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce)
// used for unit testing.
type MockWorkspbce struct {
	// PbthFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Pbth.
	PbthFunc *WorkspbcePbthFunc
	// RemoveFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Remove.
	RemoveFunc *WorkspbceRemoveFunc
	// ScriptFilenbmesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ScriptFilenbmes.
	ScriptFilenbmesFunc *WorkspbceScriptFilenbmesFunc
	// WorkingDirectoryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WorkingDirectory.
	WorkingDirectoryFunc *WorkspbceWorkingDirectoryFunc
}

// NewMockWorkspbce crebtes b new mock of the Workspbce interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockWorkspbce() *MockWorkspbce {
	return &MockWorkspbce{
		PbthFunc: &WorkspbcePbthFunc{
			defbultHook: func() (r0 string) {
				return
			},
		},
		RemoveFunc: &WorkspbceRemoveFunc{
			defbultHook: func(context.Context, bool) {
				return
			},
		},
		ScriptFilenbmesFunc: &WorkspbceScriptFilenbmesFunc{
			defbultHook: func() (r0 []string) {
				return
			},
		},
		WorkingDirectoryFunc: &WorkspbceWorkingDirectoryFunc{
			defbultHook: func() (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockWorkspbce crebtes b new mock of the Workspbce interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockWorkspbce() *MockWorkspbce {
	return &MockWorkspbce{
		PbthFunc: &WorkspbcePbthFunc{
			defbultHook: func() string {
				pbnic("unexpected invocbtion of MockWorkspbce.Pbth")
			},
		},
		RemoveFunc: &WorkspbceRemoveFunc{
			defbultHook: func(context.Context, bool) {
				pbnic("unexpected invocbtion of MockWorkspbce.Remove")
			},
		},
		ScriptFilenbmesFunc: &WorkspbceScriptFilenbmesFunc{
			defbultHook: func() []string {
				pbnic("unexpected invocbtion of MockWorkspbce.ScriptFilenbmes")
			},
		},
		WorkingDirectoryFunc: &WorkspbceWorkingDirectoryFunc{
			defbultHook: func() string {
				pbnic("unexpected invocbtion of MockWorkspbce.WorkingDirectory")
			},
		},
	}
}

// NewMockWorkspbceFrom crebtes b new mock of the MockWorkspbce interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockWorkspbceFrom(i workspbce.Workspbce) *MockWorkspbce {
	return &MockWorkspbce{
		PbthFunc: &WorkspbcePbthFunc{
			defbultHook: i.Pbth,
		},
		RemoveFunc: &WorkspbceRemoveFunc{
			defbultHook: i.Remove,
		},
		ScriptFilenbmesFunc: &WorkspbceScriptFilenbmesFunc{
			defbultHook: i.ScriptFilenbmes,
		},
		WorkingDirectoryFunc: &WorkspbceWorkingDirectoryFunc{
			defbultHook: i.WorkingDirectory,
		},
	}
}

// WorkspbcePbthFunc describes the behbvior when the Pbth method of the
// pbrent MockWorkspbce instbnce is invoked.
type WorkspbcePbthFunc struct {
	defbultHook func() string
	hooks       []func() string
	history     []WorkspbcePbthFuncCbll
	mutex       sync.Mutex
}

// Pbth delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkspbce) Pbth() string {
	r0 := m.PbthFunc.nextHook()()
	m.PbthFunc.bppendCbll(WorkspbcePbthFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Pbth method of the
// pbrent MockWorkspbce instbnce is invoked bnd the hook queue is empty.
func (f *WorkspbcePbthFunc) SetDefbultHook(hook func() string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Pbth method of the pbrent MockWorkspbce instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *WorkspbcePbthFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkspbcePbthFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func() string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkspbcePbthFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *WorkspbcePbthFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkspbcePbthFunc) bppendCbll(r0 WorkspbcePbthFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkspbcePbthFuncCbll objects describing
// the invocbtions of this function.
func (f *WorkspbcePbthFunc) History() []WorkspbcePbthFuncCbll {
	f.mutex.Lock()
	history := mbke([]WorkspbcePbthFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkspbcePbthFuncCbll is bn object thbt describes bn invocbtion of method
// Pbth on bn instbnce of MockWorkspbce.
type WorkspbcePbthFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkspbcePbthFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkspbcePbthFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// WorkspbceRemoveFunc describes the behbvior when the Remove method of the
// pbrent MockWorkspbce instbnce is invoked.
type WorkspbceRemoveFunc struct {
	defbultHook func(context.Context, bool)
	hooks       []func(context.Context, bool)
	history     []WorkspbceRemoveFuncCbll
	mutex       sync.Mutex
}

// Remove delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkspbce) Remove(v0 context.Context, v1 bool) {
	m.RemoveFunc.nextHook()(v0, v1)
	m.RemoveFunc.bppendCbll(WorkspbceRemoveFuncCbll{v0, v1})
	return
}

// SetDefbultHook sets function thbt is cblled when the Remove method of the
// pbrent MockWorkspbce instbnce is invoked bnd the hook queue is empty.
func (f *WorkspbceRemoveFunc) SetDefbultHook(hook func(context.Context, bool)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Remove method of the pbrent MockWorkspbce instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *WorkspbceRemoveFunc) PushHook(hook func(context.Context, bool)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkspbceRemoveFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(context.Context, bool) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkspbceRemoveFunc) PushReturn() {
	f.PushHook(func(context.Context, bool) {
		return
	})
}

func (f *WorkspbceRemoveFunc) nextHook() func(context.Context, bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkspbceRemoveFunc) bppendCbll(r0 WorkspbceRemoveFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkspbceRemoveFuncCbll objects describing
// the invocbtions of this function.
func (f *WorkspbceRemoveFunc) History() []WorkspbceRemoveFuncCbll {
	f.mutex.Lock()
	history := mbke([]WorkspbceRemoveFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkspbceRemoveFuncCbll is bn object thbt describes bn invocbtion of
// method Remove on bn instbnce of MockWorkspbce.
type WorkspbceRemoveFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkspbceRemoveFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkspbceRemoveFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// WorkspbceScriptFilenbmesFunc describes the behbvior when the
// ScriptFilenbmes method of the pbrent MockWorkspbce instbnce is invoked.
type WorkspbceScriptFilenbmesFunc struct {
	defbultHook func() []string
	hooks       []func() []string
	history     []WorkspbceScriptFilenbmesFuncCbll
	mutex       sync.Mutex
}

// ScriptFilenbmes delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkspbce) ScriptFilenbmes() []string {
	r0 := m.ScriptFilenbmesFunc.nextHook()()
	m.ScriptFilenbmesFunc.bppendCbll(WorkspbceScriptFilenbmesFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ScriptFilenbmes
// method of the pbrent MockWorkspbce instbnce is invoked bnd the hook queue
// is empty.
func (f *WorkspbceScriptFilenbmesFunc) SetDefbultHook(hook func() []string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ScriptFilenbmes method of the pbrent MockWorkspbce instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WorkspbceScriptFilenbmesFunc) PushHook(hook func() []string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkspbceScriptFilenbmesFunc) SetDefbultReturn(r0 []string) {
	f.SetDefbultHook(func() []string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkspbceScriptFilenbmesFunc) PushReturn(r0 []string) {
	f.PushHook(func() []string {
		return r0
	})
}

func (f *WorkspbceScriptFilenbmesFunc) nextHook() func() []string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkspbceScriptFilenbmesFunc) bppendCbll(r0 WorkspbceScriptFilenbmesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkspbceScriptFilenbmesFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkspbceScriptFilenbmesFunc) History() []WorkspbceScriptFilenbmesFuncCbll {
	f.mutex.Lock()
	history := mbke([]WorkspbceScriptFilenbmesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkspbceScriptFilenbmesFuncCbll is bn object thbt describes bn
// invocbtion of method ScriptFilenbmes on bn instbnce of MockWorkspbce.
type WorkspbceScriptFilenbmesFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkspbceScriptFilenbmesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkspbceScriptFilenbmesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// WorkspbceWorkingDirectoryFunc describes the behbvior when the
// WorkingDirectory method of the pbrent MockWorkspbce instbnce is invoked.
type WorkspbceWorkingDirectoryFunc struct {
	defbultHook func() string
	hooks       []func() string
	history     []WorkspbceWorkingDirectoryFuncCbll
	mutex       sync.Mutex
}

// WorkingDirectory delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockWorkspbce) WorkingDirectory() string {
	r0 := m.WorkingDirectoryFunc.nextHook()()
	m.WorkingDirectoryFunc.bppendCbll(WorkspbceWorkingDirectoryFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WorkingDirectory
// method of the pbrent MockWorkspbce instbnce is invoked bnd the hook queue
// is empty.
func (f *WorkspbceWorkingDirectoryFunc) SetDefbultHook(hook func() string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WorkingDirectory method of the pbrent MockWorkspbce instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *WorkspbceWorkingDirectoryFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *WorkspbceWorkingDirectoryFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func() string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *WorkspbceWorkingDirectoryFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *WorkspbceWorkingDirectoryFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WorkspbceWorkingDirectoryFunc) bppendCbll(r0 WorkspbceWorkingDirectoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of WorkspbceWorkingDirectoryFuncCbll objects
// describing the invocbtions of this function.
func (f *WorkspbceWorkingDirectoryFunc) History() []WorkspbceWorkingDirectoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]WorkspbceWorkingDirectoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WorkspbceWorkingDirectoryFuncCbll is bn object thbt describes bn
// invocbtion of method WorkingDirectory on bn instbnce of MockWorkspbce.
type WorkspbceWorkingDirectoryFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c WorkspbceWorkingDirectoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c WorkspbceWorkingDirectoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
