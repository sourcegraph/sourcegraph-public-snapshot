// Code generated by go-mockgen 1.1.4; DO NOT EDIT.

package batches

import (
	"context"
	"sync"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	types "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	database "github.com/sourcegraph/sourcegraph/internal/database"
)

// MockBatchesStore is a mock implementation of the BatchesStore interface
// (from the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches)
// used for unit testing.
type MockBatchesStore struct {
	// DatabaseDBFunc is an instance of a mock function object controlling
	// the behavior of the method DatabaseDB.
	DatabaseDBFunc *BatchesStoreDatabaseDBFunc
	// GetBatchSpecFunc is an instance of a mock function object controlling
	// the behavior of the method GetBatchSpec.
	GetBatchSpecFunc *BatchesStoreGetBatchSpecFunc
	// GetBatchSpecWorkspaceFunc is an instance of a mock function object
	// controlling the behavior of the method GetBatchSpecWorkspace.
	GetBatchSpecWorkspaceFunc *BatchesStoreGetBatchSpecWorkspaceFunc
	// SetBatchSpecWorkspaceExecutionJobAccessTokenFunc is an instance of a
	// mock function object controlling the behavior of the method
	// SetBatchSpecWorkspaceExecutionJobAccessToken.
	SetBatchSpecWorkspaceExecutionJobAccessTokenFunc *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc
}

// NewMockBatchesStore creates a new mock of the BatchesStore interface. All
// methods return zero values for all results, unless overwritten.
func NewMockBatchesStore() *MockBatchesStore {
	return &MockBatchesStore{
		DatabaseDBFunc: &BatchesStoreDatabaseDBFunc{
			defaultHook: func() database.DB {
				return nil
			},
		},
		GetBatchSpecFunc: &BatchesStoreGetBatchSpecFunc{
			defaultHook: func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error) {
				return nil, nil
			},
		},
		GetBatchSpecWorkspaceFunc: &BatchesStoreGetBatchSpecWorkspaceFunc{
			defaultHook: func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error) {
				return nil, nil
			},
		},
		SetBatchSpecWorkspaceExecutionJobAccessTokenFunc: &BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc{
			defaultHook: func(context.Context, int64, int64) error {
				return nil
			},
		},
	}
}

// NewStrictMockBatchesStore creates a new mock of the BatchesStore
// interface. All methods panic on invocation, unless overwritten.
func NewStrictMockBatchesStore() *MockBatchesStore {
	return &MockBatchesStore{
		DatabaseDBFunc: &BatchesStoreDatabaseDBFunc{
			defaultHook: func() database.DB {
				panic("unexpected invocation of MockBatchesStore.DatabaseDB")
			},
		},
		GetBatchSpecFunc: &BatchesStoreGetBatchSpecFunc{
			defaultHook: func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error) {
				panic("unexpected invocation of MockBatchesStore.GetBatchSpec")
			},
		},
		GetBatchSpecWorkspaceFunc: &BatchesStoreGetBatchSpecWorkspaceFunc{
			defaultHook: func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error) {
				panic("unexpected invocation of MockBatchesStore.GetBatchSpecWorkspace")
			},
		},
		SetBatchSpecWorkspaceExecutionJobAccessTokenFunc: &BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc{
			defaultHook: func(context.Context, int64, int64) error {
				panic("unexpected invocation of MockBatchesStore.SetBatchSpecWorkspaceExecutionJobAccessToken")
			},
		},
	}
}

// NewMockBatchesStoreFrom creates a new mock of the MockBatchesStore
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockBatchesStoreFrom(i BatchesStore) *MockBatchesStore {
	return &MockBatchesStore{
		DatabaseDBFunc: &BatchesStoreDatabaseDBFunc{
			defaultHook: i.DatabaseDB,
		},
		GetBatchSpecFunc: &BatchesStoreGetBatchSpecFunc{
			defaultHook: i.GetBatchSpec,
		},
		GetBatchSpecWorkspaceFunc: &BatchesStoreGetBatchSpecWorkspaceFunc{
			defaultHook: i.GetBatchSpecWorkspace,
		},
		SetBatchSpecWorkspaceExecutionJobAccessTokenFunc: &BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc{
			defaultHook: i.SetBatchSpecWorkspaceExecutionJobAccessToken,
		},
	}
}

// BatchesStoreDatabaseDBFunc describes the behavior when the DatabaseDB
// method of the parent MockBatchesStore instance is invoked.
type BatchesStoreDatabaseDBFunc struct {
	defaultHook func() database.DB
	hooks       []func() database.DB
	history     []BatchesStoreDatabaseDBFuncCall
	mutex       sync.Mutex
}

// DatabaseDB delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockBatchesStore) DatabaseDB() database.DB {
	r0 := m.DatabaseDBFunc.nextHook()()
	m.DatabaseDBFunc.appendCall(BatchesStoreDatabaseDBFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the DatabaseDB method of
// the parent MockBatchesStore instance is invoked and the hook queue is
// empty.
func (f *BatchesStoreDatabaseDBFunc) SetDefaultHook(hook func() database.DB) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// DatabaseDB method of the parent MockBatchesStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *BatchesStoreDatabaseDBFunc) PushHook(hook func() database.DB) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *BatchesStoreDatabaseDBFunc) SetDefaultReturn(r0 database.DB) {
	f.SetDefaultHook(func() database.DB {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *BatchesStoreDatabaseDBFunc) PushReturn(r0 database.DB) {
	f.PushHook(func() database.DB {
		return r0
	})
}

func (f *BatchesStoreDatabaseDBFunc) nextHook() func() database.DB {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BatchesStoreDatabaseDBFunc) appendCall(r0 BatchesStoreDatabaseDBFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BatchesStoreDatabaseDBFuncCall objects
// describing the invocations of this function.
func (f *BatchesStoreDatabaseDBFunc) History() []BatchesStoreDatabaseDBFuncCall {
	f.mutex.Lock()
	history := make([]BatchesStoreDatabaseDBFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BatchesStoreDatabaseDBFuncCall is an object that describes an invocation
// of method DatabaseDB on an instance of MockBatchesStore.
type BatchesStoreDatabaseDBFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 database.DB
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BatchesStoreDatabaseDBFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BatchesStoreDatabaseDBFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// BatchesStoreGetBatchSpecFunc describes the behavior when the GetBatchSpec
// method of the parent MockBatchesStore instance is invoked.
type BatchesStoreGetBatchSpecFunc struct {
	defaultHook func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error)
	hooks       []func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error)
	history     []BatchesStoreGetBatchSpecFuncCall
	mutex       sync.Mutex
}

// GetBatchSpec delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockBatchesStore) GetBatchSpec(v0 context.Context, v1 store.GetBatchSpecOpts) (*types.BatchSpec, error) {
	r0, r1 := m.GetBatchSpecFunc.nextHook()(v0, v1)
	m.GetBatchSpecFunc.appendCall(BatchesStoreGetBatchSpecFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetBatchSpec method
// of the parent MockBatchesStore instance is invoked and the hook queue is
// empty.
func (f *BatchesStoreGetBatchSpecFunc) SetDefaultHook(hook func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetBatchSpec method of the parent MockBatchesStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *BatchesStoreGetBatchSpecFunc) PushHook(hook func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *BatchesStoreGetBatchSpecFunc) SetDefaultReturn(r0 *types.BatchSpec, r1 error) {
	f.SetDefaultHook(func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *BatchesStoreGetBatchSpecFunc) PushReturn(r0 *types.BatchSpec, r1 error) {
	f.PushHook(func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error) {
		return r0, r1
	})
}

func (f *BatchesStoreGetBatchSpecFunc) nextHook() func(context.Context, store.GetBatchSpecOpts) (*types.BatchSpec, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BatchesStoreGetBatchSpecFunc) appendCall(r0 BatchesStoreGetBatchSpecFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BatchesStoreGetBatchSpecFuncCall objects
// describing the invocations of this function.
func (f *BatchesStoreGetBatchSpecFunc) History() []BatchesStoreGetBatchSpecFuncCall {
	f.mutex.Lock()
	history := make([]BatchesStoreGetBatchSpecFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BatchesStoreGetBatchSpecFuncCall is an object that describes an
// invocation of method GetBatchSpec on an instance of MockBatchesStore.
type BatchesStoreGetBatchSpecFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.GetBatchSpecOpts
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *types.BatchSpec
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BatchesStoreGetBatchSpecFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BatchesStoreGetBatchSpecFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// BatchesStoreGetBatchSpecWorkspaceFunc describes the behavior when the
// GetBatchSpecWorkspace method of the parent MockBatchesStore instance is
// invoked.
type BatchesStoreGetBatchSpecWorkspaceFunc struct {
	defaultHook func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error)
	hooks       []func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error)
	history     []BatchesStoreGetBatchSpecWorkspaceFuncCall
	mutex       sync.Mutex
}

// GetBatchSpecWorkspace delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockBatchesStore) GetBatchSpecWorkspace(v0 context.Context, v1 store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error) {
	r0, r1 := m.GetBatchSpecWorkspaceFunc.nextHook()(v0, v1)
	m.GetBatchSpecWorkspaceFunc.appendCall(BatchesStoreGetBatchSpecWorkspaceFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// GetBatchSpecWorkspace method of the parent MockBatchesStore instance is
// invoked and the hook queue is empty.
func (f *BatchesStoreGetBatchSpecWorkspaceFunc) SetDefaultHook(hook func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetBatchSpecWorkspace method of the parent MockBatchesStore instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *BatchesStoreGetBatchSpecWorkspaceFunc) PushHook(hook func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *BatchesStoreGetBatchSpecWorkspaceFunc) SetDefaultReturn(r0 *types.BatchSpecWorkspace, r1 error) {
	f.SetDefaultHook(func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *BatchesStoreGetBatchSpecWorkspaceFunc) PushReturn(r0 *types.BatchSpecWorkspace, r1 error) {
	f.PushHook(func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error) {
		return r0, r1
	})
}

func (f *BatchesStoreGetBatchSpecWorkspaceFunc) nextHook() func(context.Context, store.GetBatchSpecWorkspaceOpts) (*types.BatchSpecWorkspace, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BatchesStoreGetBatchSpecWorkspaceFunc) appendCall(r0 BatchesStoreGetBatchSpecWorkspaceFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BatchesStoreGetBatchSpecWorkspaceFuncCall
// objects describing the invocations of this function.
func (f *BatchesStoreGetBatchSpecWorkspaceFunc) History() []BatchesStoreGetBatchSpecWorkspaceFuncCall {
	f.mutex.Lock()
	history := make([]BatchesStoreGetBatchSpecWorkspaceFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BatchesStoreGetBatchSpecWorkspaceFuncCall is an object that describes an
// invocation of method GetBatchSpecWorkspace on an instance of
// MockBatchesStore.
type BatchesStoreGetBatchSpecWorkspaceFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.GetBatchSpecWorkspaceOpts
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *types.BatchSpecWorkspace
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BatchesStoreGetBatchSpecWorkspaceFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BatchesStoreGetBatchSpecWorkspaceFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc describes
// the behavior when the SetBatchSpecWorkspaceExecutionJobAccessToken method
// of the parent MockBatchesStore instance is invoked.
type BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc struct {
	defaultHook func(context.Context, int64, int64) error
	hooks       []func(context.Context, int64, int64) error
	history     []BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall
	mutex       sync.Mutex
}

// SetBatchSpecWorkspaceExecutionJobAccessToken delegates to the next hook
// function in the queue and stores the parameter and result values of this
// invocation.
func (m *MockBatchesStore) SetBatchSpecWorkspaceExecutionJobAccessToken(v0 context.Context, v1 int64, v2 int64) error {
	r0 := m.SetBatchSpecWorkspaceExecutionJobAccessTokenFunc.nextHook()(v0, v1, v2)
	m.SetBatchSpecWorkspaceExecutionJobAccessTokenFunc.appendCall(BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the
// SetBatchSpecWorkspaceExecutionJobAccessToken method of the parent
// MockBatchesStore instance is invoked and the hook queue is empty.
func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) SetDefaultHook(hook func(context.Context, int64, int64) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SetBatchSpecWorkspaceExecutionJobAccessToken method of the parent
// MockBatchesStore instance invokes the hook at the front of the queue and
// discards it. After the queue is empty, the default hook function is
// invoked for any future action.
func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) PushHook(hook func(context.Context, int64, int64) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int64, int64) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int64, int64) error {
		return r0
	})
}

func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) nextHook() func(context.Context, int64, int64) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) appendCall(r0 BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall objects
// describing the invocations of this function.
func (f *BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFunc) History() []BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall {
	f.mutex.Lock()
	history := make([]BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall is an
// object that describes an invocation of method
// SetBatchSpecWorkspaceExecutionJobAccessToken on an instance of
// MockBatchesStore.
type BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int64
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int64
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BatchesStoreSetBatchSpecWorkspaceExecutionJobAccessTokenFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
