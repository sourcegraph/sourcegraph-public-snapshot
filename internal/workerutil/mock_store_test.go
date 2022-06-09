// Code generated by go-mockgen 1.3.1; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package workerutil

import (
	"context"
	"sync"
)

// MockStore is a mock implementation of the Store interface (from the
// package github.com/sourcegraph/sourcegraph/internal/workerutil) used for
// unit testing.
type MockStore struct {
	// AddExecutionLogEntryFunc is an instance of a mock function object
	// controlling the behavior of the method AddExecutionLogEntry.
	AddExecutionLogEntryFunc *StoreAddExecutionLogEntryFunc
	// DequeueFunc is an instance of a mock function object controlling the
	// behavior of the method Dequeue.
	DequeueFunc *StoreDequeueFunc
	// HeartbeatFunc is an instance of a mock function object controlling
	// the behavior of the method Heartbeat.
	HeartbeatFunc *StoreHeartbeatFunc
	// MarkCompleteFunc is an instance of a mock function object controlling
	// the behavior of the method MarkComplete.
	MarkCompleteFunc *StoreMarkCompleteFunc
	// MarkErroredFunc is an instance of a mock function object controlling
	// the behavior of the method MarkErrored.
	MarkErroredFunc *StoreMarkErroredFunc
	// MarkFailedFunc is an instance of a mock function object controlling
	// the behavior of the method MarkFailed.
	MarkFailedFunc *StoreMarkFailedFunc
	// QueuedCountFunc is an instance of a mock function object controlling
	// the behavior of the method QueuedCount.
	QueuedCountFunc *StoreQueuedCountFunc
	// UpdateExecutionLogEntryFunc is an instance of a mock function object
	// controlling the behavior of the method UpdateExecutionLogEntry.
	UpdateExecutionLogEntryFunc *StoreUpdateExecutionLogEntryFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		AddExecutionLogEntryFunc: &StoreAddExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, ExecutionLogEntry) (r0 int, r1 error) {
				return
			},
		},
		DequeueFunc: &StoreDequeueFunc{
			defaultHook: func(context.Context, string, interface{}) (r0 Record, r1 bool, r2 error) {
				return
			},
		},
		HeartbeatFunc: &StoreHeartbeatFunc{
			defaultHook: func(context.Context, []int) (r0 []int, r1 error) {
				return
			},
		},
		MarkCompleteFunc: &StoreMarkCompleteFunc{
			defaultHook: func(context.Context, int) (r0 bool, r1 error) {
				return
			},
		},
		MarkErroredFunc: &StoreMarkErroredFunc{
			defaultHook: func(context.Context, int, string) (r0 bool, r1 error) {
				return
			},
		},
		MarkFailedFunc: &StoreMarkFailedFunc{
			defaultHook: func(context.Context, int, string) (r0 bool, r1 error) {
				return
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc{
			defaultHook: func(context.Context, interface{}) (r0 int, r1 error) {
				return
			},
		},
		UpdateExecutionLogEntryFunc: &StoreUpdateExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, int, ExecutionLogEntry) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockStore creates a new mock of the Store interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		AddExecutionLogEntryFunc: &StoreAddExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, ExecutionLogEntry) (int, error) {
				panic("unexpected invocation of MockStore.AddExecutionLogEntry")
			},
		},
		DequeueFunc: &StoreDequeueFunc{
			defaultHook: func(context.Context, string, interface{}) (Record, bool, error) {
				panic("unexpected invocation of MockStore.Dequeue")
			},
		},
		HeartbeatFunc: &StoreHeartbeatFunc{
			defaultHook: func(context.Context, []int) ([]int, error) {
				panic("unexpected invocation of MockStore.Heartbeat")
			},
		},
		MarkCompleteFunc: &StoreMarkCompleteFunc{
			defaultHook: func(context.Context, int) (bool, error) {
				panic("unexpected invocation of MockStore.MarkComplete")
			},
		},
		MarkErroredFunc: &StoreMarkErroredFunc{
			defaultHook: func(context.Context, int, string) (bool, error) {
				panic("unexpected invocation of MockStore.MarkErrored")
			},
		},
		MarkFailedFunc: &StoreMarkFailedFunc{
			defaultHook: func(context.Context, int, string) (bool, error) {
				panic("unexpected invocation of MockStore.MarkFailed")
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc{
			defaultHook: func(context.Context, interface{}) (int, error) {
				panic("unexpected invocation of MockStore.QueuedCount")
			},
		},
		UpdateExecutionLogEntryFunc: &StoreUpdateExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, int, ExecutionLogEntry) error {
				panic("unexpected invocation of MockStore.UpdateExecutionLogEntry")
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		AddExecutionLogEntryFunc: &StoreAddExecutionLogEntryFunc{
			defaultHook: i.AddExecutionLogEntry,
		},
		DequeueFunc: &StoreDequeueFunc{
			defaultHook: i.Dequeue,
		},
		HeartbeatFunc: &StoreHeartbeatFunc{
			defaultHook: i.Heartbeat,
		},
		MarkCompleteFunc: &StoreMarkCompleteFunc{
			defaultHook: i.MarkComplete,
		},
		MarkErroredFunc: &StoreMarkErroredFunc{
			defaultHook: i.MarkErrored,
		},
		MarkFailedFunc: &StoreMarkFailedFunc{
			defaultHook: i.MarkFailed,
		},
		QueuedCountFunc: &StoreQueuedCountFunc{
			defaultHook: i.QueuedCount,
		},
		UpdateExecutionLogEntryFunc: &StoreUpdateExecutionLogEntryFunc{
			defaultHook: i.UpdateExecutionLogEntry,
		},
	}
}

// StoreAddExecutionLogEntryFunc describes the behavior when the
// AddExecutionLogEntry method of the parent MockStore instance is invoked.
type StoreAddExecutionLogEntryFunc struct {
	defaultHook func(context.Context, int, ExecutionLogEntry) (int, error)
	hooks       []func(context.Context, int, ExecutionLogEntry) (int, error)
	history     []StoreAddExecutionLogEntryFuncCall
	mutex       sync.Mutex
}

// AddExecutionLogEntry delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) AddExecutionLogEntry(v0 context.Context, v1 int, v2 ExecutionLogEntry) (int, error) {
	r0, r1 := m.AddExecutionLogEntryFunc.nextHook()(v0, v1, v2)
	m.AddExecutionLogEntryFunc.appendCall(StoreAddExecutionLogEntryFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the AddExecutionLogEntry
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreAddExecutionLogEntryFunc) SetDefaultHook(hook func(context.Context, int, ExecutionLogEntry) (int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// AddExecutionLogEntry method of the parent MockStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *StoreAddExecutionLogEntryFunc) PushHook(hook func(context.Context, int, ExecutionLogEntry) (int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreAddExecutionLogEntryFunc) SetDefaultReturn(r0 int, r1 error) {
	f.SetDefaultHook(func(context.Context, int, ExecutionLogEntry) (int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreAddExecutionLogEntryFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, int, ExecutionLogEntry) (int, error) {
		return r0, r1
	})
}

func (f *StoreAddExecutionLogEntryFunc) nextHook() func(context.Context, int, ExecutionLogEntry) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreAddExecutionLogEntryFunc) appendCall(r0 StoreAddExecutionLogEntryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreAddExecutionLogEntryFuncCall objects
// describing the invocations of this function.
func (f *StoreAddExecutionLogEntryFunc) History() []StoreAddExecutionLogEntryFuncCall {
	f.mutex.Lock()
	history := make([]StoreAddExecutionLogEntryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreAddExecutionLogEntryFuncCall is an object that describes an
// invocation of method AddExecutionLogEntry on an instance of MockStore.
type StoreAddExecutionLogEntryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 ExecutionLogEntry
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreAddExecutionLogEntryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreAddExecutionLogEntryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreDequeueFunc describes the behavior when the Dequeue method of the
// parent MockStore instance is invoked.
type StoreDequeueFunc struct {
	defaultHook func(context.Context, string, interface{}) (Record, bool, error)
	hooks       []func(context.Context, string, interface{}) (Record, bool, error)
	history     []StoreDequeueFuncCall
	mutex       sync.Mutex
}

// Dequeue delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Dequeue(v0 context.Context, v1 string, v2 interface{}) (Record, bool, error) {
	r0, r1, r2 := m.DequeueFunc.nextHook()(v0, v1, v2)
	m.DequeueFunc.appendCall(StoreDequeueFuncCall{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the Dequeue method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreDequeueFunc) SetDefaultHook(hook func(context.Context, string, interface{}) (Record, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Dequeue method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreDequeueFunc) PushHook(hook func(context.Context, string, interface{}) (Record, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreDequeueFunc) SetDefaultReturn(r0 Record, r1 bool, r2 error) {
	f.SetDefaultHook(func(context.Context, string, interface{}) (Record, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreDequeueFunc) PushReturn(r0 Record, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, interface{}) (Record, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreDequeueFunc) nextHook() func(context.Context, string, interface{}) (Record, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDequeueFunc) appendCall(r0 StoreDequeueFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDequeueFuncCall objects describing the
// invocations of this function.
func (f *StoreDequeueFunc) History() []StoreDequeueFuncCall {
	f.mutex.Lock()
	history := make([]StoreDequeueFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDequeueFuncCall is an object that describes an invocation of method
// Dequeue on an instance of MockStore.
type StoreDequeueFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 interface{}
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 Record
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 bool
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDequeueFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDequeueFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// StoreHeartbeatFunc describes the behavior when the Heartbeat method of
// the parent MockStore instance is invoked.
type StoreHeartbeatFunc struct {
	defaultHook func(context.Context, []int) ([]int, error)
	hooks       []func(context.Context, []int) ([]int, error)
	history     []StoreHeartbeatFuncCall
	mutex       sync.Mutex
}

// Heartbeat delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Heartbeat(v0 context.Context, v1 []int) ([]int, error) {
	r0, r1 := m.HeartbeatFunc.nextHook()(v0, v1)
	m.HeartbeatFunc.appendCall(StoreHeartbeatFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Heartbeat method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreHeartbeatFunc) SetDefaultHook(hook func(context.Context, []int) ([]int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Heartbeat method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreHeartbeatFunc) PushHook(hook func(context.Context, []int) ([]int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreHeartbeatFunc) SetDefaultReturn(r0 []int, r1 error) {
	f.SetDefaultHook(func(context.Context, []int) ([]int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreHeartbeatFunc) PushReturn(r0 []int, r1 error) {
	f.PushHook(func(context.Context, []int) ([]int, error) {
		return r0, r1
	})
}

func (f *StoreHeartbeatFunc) nextHook() func(context.Context, []int) ([]int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreHeartbeatFunc) appendCall(r0 StoreHeartbeatFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreHeartbeatFuncCall objects describing
// the invocations of this function.
func (f *StoreHeartbeatFunc) History() []StoreHeartbeatFuncCall {
	f.mutex.Lock()
	history := make([]StoreHeartbeatFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreHeartbeatFuncCall is an object that describes an invocation of
// method Heartbeat on an instance of MockStore.
type StoreHeartbeatFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 []int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreHeartbeatFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreHeartbeatFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreMarkCompleteFunc describes the behavior when the MarkComplete method
// of the parent MockStore instance is invoked.
type StoreMarkCompleteFunc struct {
	defaultHook func(context.Context, int) (bool, error)
	hooks       []func(context.Context, int) (bool, error)
	history     []StoreMarkCompleteFuncCall
	mutex       sync.Mutex
}

// MarkComplete delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) MarkComplete(v0 context.Context, v1 int) (bool, error) {
	r0, r1 := m.MarkCompleteFunc.nextHook()(v0, v1)
	m.MarkCompleteFunc.appendCall(StoreMarkCompleteFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the MarkComplete method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreMarkCompleteFunc) SetDefaultHook(hook func(context.Context, int) (bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// MarkComplete method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreMarkCompleteFunc) PushHook(hook func(context.Context, int) (bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreMarkCompleteFunc) SetDefaultReturn(r0 bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreMarkCompleteFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMarkCompleteFunc) nextHook() func(context.Context, int) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMarkCompleteFunc) appendCall(r0 StoreMarkCompleteFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreMarkCompleteFuncCall objects
// describing the invocations of this function.
func (f *StoreMarkCompleteFunc) History() []StoreMarkCompleteFuncCall {
	f.mutex.Lock()
	history := make([]StoreMarkCompleteFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMarkCompleteFuncCall is an object that describes an invocation of
// method MarkComplete on an instance of MockStore.
type StoreMarkCompleteFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreMarkCompleteFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreMarkCompleteFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreMarkErroredFunc describes the behavior when the MarkErrored method
// of the parent MockStore instance is invoked.
type StoreMarkErroredFunc struct {
	defaultHook func(context.Context, int, string) (bool, error)
	hooks       []func(context.Context, int, string) (bool, error)
	history     []StoreMarkErroredFuncCall
	mutex       sync.Mutex
}

// MarkErrored delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) MarkErrored(v0 context.Context, v1 int, v2 string) (bool, error) {
	r0, r1 := m.MarkErroredFunc.nextHook()(v0, v1, v2)
	m.MarkErroredFunc.appendCall(StoreMarkErroredFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the MarkErrored method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreMarkErroredFunc) SetDefaultHook(hook func(context.Context, int, string) (bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// MarkErrored method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreMarkErroredFunc) PushHook(hook func(context.Context, int, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreMarkErroredFunc) SetDefaultReturn(r0 bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreMarkErroredFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMarkErroredFunc) nextHook() func(context.Context, int, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMarkErroredFunc) appendCall(r0 StoreMarkErroredFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreMarkErroredFuncCall objects describing
// the invocations of this function.
func (f *StoreMarkErroredFunc) History() []StoreMarkErroredFuncCall {
	f.mutex.Lock()
	history := make([]StoreMarkErroredFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMarkErroredFuncCall is an object that describes an invocation of
// method MarkErrored on an instance of MockStore.
type StoreMarkErroredFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreMarkErroredFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreMarkErroredFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreMarkFailedFunc describes the behavior when the MarkFailed method of
// the parent MockStore instance is invoked.
type StoreMarkFailedFunc struct {
	defaultHook func(context.Context, int, string) (bool, error)
	hooks       []func(context.Context, int, string) (bool, error)
	history     []StoreMarkFailedFuncCall
	mutex       sync.Mutex
}

// MarkFailed delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) MarkFailed(v0 context.Context, v1 int, v2 string) (bool, error) {
	r0, r1 := m.MarkFailedFunc.nextHook()(v0, v1, v2)
	m.MarkFailedFunc.appendCall(StoreMarkFailedFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the MarkFailed method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreMarkFailedFunc) SetDefaultHook(hook func(context.Context, int, string) (bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// MarkFailed method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreMarkFailedFunc) PushHook(hook func(context.Context, int, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreMarkFailedFunc) SetDefaultReturn(r0 bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreMarkFailedFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMarkFailedFunc) nextHook() func(context.Context, int, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMarkFailedFunc) appendCall(r0 StoreMarkFailedFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreMarkFailedFuncCall objects describing
// the invocations of this function.
func (f *StoreMarkFailedFunc) History() []StoreMarkFailedFuncCall {
	f.mutex.Lock()
	history := make([]StoreMarkFailedFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMarkFailedFuncCall is an object that describes an invocation of
// method MarkFailed on an instance of MockStore.
type StoreMarkFailedFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreMarkFailedFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreMarkFailedFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreQueuedCountFunc describes the behavior when the QueuedCount method
// of the parent MockStore instance is invoked.
type StoreQueuedCountFunc struct {
	defaultHook func(context.Context, interface{}) (int, error)
	hooks       []func(context.Context, interface{}) (int, error)
	history     []StoreQueuedCountFuncCall
	mutex       sync.Mutex
}

// QueuedCount delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) QueuedCount(v0 context.Context, v1 interface{}) (int, error) {
	r0, r1 := m.QueuedCountFunc.nextHook()(v0, v1)
	m.QueuedCountFunc.appendCall(StoreQueuedCountFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the QueuedCount method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreQueuedCountFunc) SetDefaultHook(hook func(context.Context, interface{}) (int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// QueuedCount method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreQueuedCountFunc) PushHook(hook func(context.Context, interface{}) (int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreQueuedCountFunc) SetDefaultReturn(r0 int, r1 error) {
	f.SetDefaultHook(func(context.Context, interface{}) (int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreQueuedCountFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, interface{}) (int, error) {
		return r0, r1
	})
}

func (f *StoreQueuedCountFunc) nextHook() func(context.Context, interface{}) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreQueuedCountFunc) appendCall(r0 StoreQueuedCountFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreQueuedCountFuncCall objects describing
// the invocations of this function.
func (f *StoreQueuedCountFunc) History() []StoreQueuedCountFuncCall {
	f.mutex.Lock()
	history := make([]StoreQueuedCountFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreQueuedCountFuncCall is an object that describes an invocation of
// method QueuedCount on an instance of MockStore.
type StoreQueuedCountFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 interface{}
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreQueuedCountFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreQueuedCountFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreUpdateExecutionLogEntryFunc describes the behavior when the
// UpdateExecutionLogEntry method of the parent MockStore instance is
// invoked.
type StoreUpdateExecutionLogEntryFunc struct {
	defaultHook func(context.Context, int, int, ExecutionLogEntry) error
	hooks       []func(context.Context, int, int, ExecutionLogEntry) error
	history     []StoreUpdateExecutionLogEntryFuncCall
	mutex       sync.Mutex
}

// UpdateExecutionLogEntry delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockStore) UpdateExecutionLogEntry(v0 context.Context, v1 int, v2 int, v3 ExecutionLogEntry) error {
	r0 := m.UpdateExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3)
	m.UpdateExecutionLogEntryFunc.appendCall(StoreUpdateExecutionLogEntryFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the
// UpdateExecutionLogEntry method of the parent MockStore instance is
// invoked and the hook queue is empty.
func (f *StoreUpdateExecutionLogEntryFunc) SetDefaultHook(hook func(context.Context, int, int, ExecutionLogEntry) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// UpdateExecutionLogEntry method of the parent MockStore instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *StoreUpdateExecutionLogEntryFunc) PushHook(hook func(context.Context, int, int, ExecutionLogEntry) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreUpdateExecutionLogEntryFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, int, ExecutionLogEntry) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreUpdateExecutionLogEntryFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int, ExecutionLogEntry) error {
		return r0
	})
}

func (f *StoreUpdateExecutionLogEntryFunc) nextHook() func(context.Context, int, int, ExecutionLogEntry) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdateExecutionLogEntryFunc) appendCall(r0 StoreUpdateExecutionLogEntryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreUpdateExecutionLogEntryFuncCall
// objects describing the invocations of this function.
func (f *StoreUpdateExecutionLogEntryFunc) History() []StoreUpdateExecutionLogEntryFuncCall {
	f.mutex.Lock()
	history := make([]StoreUpdateExecutionLogEntryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdateExecutionLogEntryFuncCall is an object that describes an
// invocation of method UpdateExecutionLogEntry on an instance of MockStore.
type StoreUpdateExecutionLogEntryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 ExecutionLogEntry
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreUpdateExecutionLogEntryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreUpdateExecutionLogEntryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
