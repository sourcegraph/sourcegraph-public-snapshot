// Code generated by go-mockgen 1.2.0; DO NOT EDIT.

package featureflag

import (
	"context"
	"sync"
)

// MockStore is a mock implementation of the Store interface (from the
// package github.com/sourcegraph/sourcegraph/internal/featureflag) used for
// unit testing.
type MockStore struct {
	// GetAnonymousUserFlagFunc is an instance of a mock function object
	// controlling the behavior of the method GetAnonymousUserFlag.
	GetAnonymousUserFlagFunc *StoreGetAnonymousUserFlagFunc
	// GetAnonymousUserFlagsFunc is an instance of a mock function object
	// controlling the behavior of the method GetAnonymousUserFlags.
	GetAnonymousUserFlagsFunc *StoreGetAnonymousUserFlagsFunc
	// GetGlobalFeatureFlagFunc is an instance of a mock function object
	// controlling the behavior of the method GetGlobalFeatureFlag.
	GetGlobalFeatureFlagFunc *StoreGetGlobalFeatureFlagFunc
	// GetGlobalFeatureFlagsFunc is an instance of a mock function object
	// controlling the behavior of the method GetGlobalFeatureFlags.
	GetGlobalFeatureFlagsFunc *StoreGetGlobalFeatureFlagsFunc
	// GetUserFlagFunc is an instance of a mock function object controlling
	// the behavior of the method GetUserFlag.
	GetUserFlagFunc *StoreGetUserFlagFunc
	// GetUserFlagsFunc is an instance of a mock function object controlling
	// the behavior of the method GetUserFlags.
	GetUserFlagsFunc *StoreGetUserFlagsFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		GetAnonymousUserFlagFunc: &StoreGetAnonymousUserFlagFunc{
			defaultHook: func(context.Context, string, string) (r0 *bool, r1 error) {
				return
			},
		},
		GetAnonymousUserFlagsFunc: &StoreGetAnonymousUserFlagsFunc{
			defaultHook: func(context.Context, string) (r0 map[string]bool, r1 error) {
				return
			},
		},
		GetGlobalFeatureFlagFunc: &StoreGetGlobalFeatureFlagFunc{
			defaultHook: func(context.Context, string) (r0 *bool, r1 error) {
				return
			},
		},
		GetGlobalFeatureFlagsFunc: &StoreGetGlobalFeatureFlagsFunc{
			defaultHook: func(context.Context) (r0 map[string]bool, r1 error) {
				return
			},
		},
		GetUserFlagFunc: &StoreGetUserFlagFunc{
			defaultHook: func(context.Context, int32, string) (r0 *bool, r1 error) {
				return
			},
		},
		GetUserFlagsFunc: &StoreGetUserFlagsFunc{
			defaultHook: func(context.Context, int32) (r0 map[string]bool, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore creates a new mock of the Store interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		GetAnonymousUserFlagFunc: &StoreGetAnonymousUserFlagFunc{
			defaultHook: func(context.Context, string, string) (*bool, error) {
				panic("unexpected invocation of MockStore.GetAnonymousUserFlag")
			},
		},
		GetAnonymousUserFlagsFunc: &StoreGetAnonymousUserFlagsFunc{
			defaultHook: func(context.Context, string) (map[string]bool, error) {
				panic("unexpected invocation of MockStore.GetAnonymousUserFlags")
			},
		},
		GetGlobalFeatureFlagFunc: &StoreGetGlobalFeatureFlagFunc{
			defaultHook: func(context.Context, string) (*bool, error) {
				panic("unexpected invocation of MockStore.GetGlobalFeatureFlag")
			},
		},
		GetGlobalFeatureFlagsFunc: &StoreGetGlobalFeatureFlagsFunc{
			defaultHook: func(context.Context) (map[string]bool, error) {
				panic("unexpected invocation of MockStore.GetGlobalFeatureFlags")
			},
		},
		GetUserFlagFunc: &StoreGetUserFlagFunc{
			defaultHook: func(context.Context, int32, string) (*bool, error) {
				panic("unexpected invocation of MockStore.GetUserFlag")
			},
		},
		GetUserFlagsFunc: &StoreGetUserFlagsFunc{
			defaultHook: func(context.Context, int32) (map[string]bool, error) {
				panic("unexpected invocation of MockStore.GetUserFlags")
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		GetAnonymousUserFlagFunc: &StoreGetAnonymousUserFlagFunc{
			defaultHook: i.GetAnonymousUserFlag,
		},
		GetAnonymousUserFlagsFunc: &StoreGetAnonymousUserFlagsFunc{
			defaultHook: i.GetAnonymousUserFlags,
		},
		GetGlobalFeatureFlagFunc: &StoreGetGlobalFeatureFlagFunc{
			defaultHook: i.GetGlobalFeatureFlag,
		},
		GetGlobalFeatureFlagsFunc: &StoreGetGlobalFeatureFlagsFunc{
			defaultHook: i.GetGlobalFeatureFlags,
		},
		GetUserFlagFunc: &StoreGetUserFlagFunc{
			defaultHook: i.GetUserFlag,
		},
		GetUserFlagsFunc: &StoreGetUserFlagsFunc{
			defaultHook: i.GetUserFlags,
		},
	}
}

// StoreGetAnonymousUserFlagFunc describes the behavior when the
// GetAnonymousUserFlag method of the parent MockStore instance is invoked.
type StoreGetAnonymousUserFlagFunc struct {
	defaultHook func(context.Context, string, string) (*bool, error)
	hooks       []func(context.Context, string, string) (*bool, error)
	history     []StoreGetAnonymousUserFlagFuncCall
	mutex       sync.Mutex
}

// GetAnonymousUserFlag delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) GetAnonymousUserFlag(v0 context.Context, v1 string, v2 string) (*bool, error) {
	r0, r1 := m.GetAnonymousUserFlagFunc.nextHook()(v0, v1, v2)
	m.GetAnonymousUserFlagFunc.appendCall(StoreGetAnonymousUserFlagFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetAnonymousUserFlag
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreGetAnonymousUserFlagFunc) SetDefaultHook(hook func(context.Context, string, string) (*bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetAnonymousUserFlag method of the parent MockStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *StoreGetAnonymousUserFlagFunc) PushHook(hook func(context.Context, string, string) (*bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetAnonymousUserFlagFunc) SetDefaultReturn(r0 *bool, r1 error) {
	f.SetDefaultHook(func(context.Context, string, string) (*bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetAnonymousUserFlagFunc) PushReturn(r0 *bool, r1 error) {
	f.PushHook(func(context.Context, string, string) (*bool, error) {
		return r0, r1
	})
}

func (f *StoreGetAnonymousUserFlagFunc) nextHook() func(context.Context, string, string) (*bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetAnonymousUserFlagFunc) appendCall(r0 StoreGetAnonymousUserFlagFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetAnonymousUserFlagFuncCall objects
// describing the invocations of this function.
func (f *StoreGetAnonymousUserFlagFunc) History() []StoreGetAnonymousUserFlagFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetAnonymousUserFlagFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetAnonymousUserFlagFuncCall is an object that describes an
// invocation of method GetAnonymousUserFlag on an instance of MockStore.
type StoreGetAnonymousUserFlagFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetAnonymousUserFlagFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetAnonymousUserFlagFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreGetAnonymousUserFlagsFunc describes the behavior when the
// GetAnonymousUserFlags method of the parent MockStore instance is invoked.
type StoreGetAnonymousUserFlagsFunc struct {
	defaultHook func(context.Context, string) (map[string]bool, error)
	hooks       []func(context.Context, string) (map[string]bool, error)
	history     []StoreGetAnonymousUserFlagsFuncCall
	mutex       sync.Mutex
}

// GetAnonymousUserFlags delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockStore) GetAnonymousUserFlags(v0 context.Context, v1 string) (map[string]bool, error) {
	r0, r1 := m.GetAnonymousUserFlagsFunc.nextHook()(v0, v1)
	m.GetAnonymousUserFlagsFunc.appendCall(StoreGetAnonymousUserFlagsFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// GetAnonymousUserFlags method of the parent MockStore instance is invoked
// and the hook queue is empty.
func (f *StoreGetAnonymousUserFlagsFunc) SetDefaultHook(hook func(context.Context, string) (map[string]bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetAnonymousUserFlags method of the parent MockStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *StoreGetAnonymousUserFlagsFunc) PushHook(hook func(context.Context, string) (map[string]bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetAnonymousUserFlagsFunc) SetDefaultReturn(r0 map[string]bool, r1 error) {
	f.SetDefaultHook(func(context.Context, string) (map[string]bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetAnonymousUserFlagsFunc) PushReturn(r0 map[string]bool, r1 error) {
	f.PushHook(func(context.Context, string) (map[string]bool, error) {
		return r0, r1
	})
}

func (f *StoreGetAnonymousUserFlagsFunc) nextHook() func(context.Context, string) (map[string]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetAnonymousUserFlagsFunc) appendCall(r0 StoreGetAnonymousUserFlagsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetAnonymousUserFlagsFuncCall objects
// describing the invocations of this function.
func (f *StoreGetAnonymousUserFlagsFunc) History() []StoreGetAnonymousUserFlagsFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetAnonymousUserFlagsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetAnonymousUserFlagsFuncCall is an object that describes an
// invocation of method GetAnonymousUserFlags on an instance of MockStore.
type StoreGetAnonymousUserFlagsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string]bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetAnonymousUserFlagsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetAnonymousUserFlagsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreGetGlobalFeatureFlagFunc describes the behavior when the
// GetGlobalFeatureFlag method of the parent MockStore instance is invoked.
type StoreGetGlobalFeatureFlagFunc struct {
	defaultHook func(context.Context, string) (*bool, error)
	hooks       []func(context.Context, string) (*bool, error)
	history     []StoreGetGlobalFeatureFlagFuncCall
	mutex       sync.Mutex
}

// GetGlobalFeatureFlag delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) GetGlobalFeatureFlag(v0 context.Context, v1 string) (*bool, error) {
	r0, r1 := m.GetGlobalFeatureFlagFunc.nextHook()(v0, v1)
	m.GetGlobalFeatureFlagFunc.appendCall(StoreGetGlobalFeatureFlagFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetGlobalFeatureFlag
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreGetGlobalFeatureFlagFunc) SetDefaultHook(hook func(context.Context, string) (*bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetGlobalFeatureFlag method of the parent MockStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *StoreGetGlobalFeatureFlagFunc) PushHook(hook func(context.Context, string) (*bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetGlobalFeatureFlagFunc) SetDefaultReturn(r0 *bool, r1 error) {
	f.SetDefaultHook(func(context.Context, string) (*bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetGlobalFeatureFlagFunc) PushReturn(r0 *bool, r1 error) {
	f.PushHook(func(context.Context, string) (*bool, error) {
		return r0, r1
	})
}

func (f *StoreGetGlobalFeatureFlagFunc) nextHook() func(context.Context, string) (*bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetGlobalFeatureFlagFunc) appendCall(r0 StoreGetGlobalFeatureFlagFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetGlobalFeatureFlagFuncCall objects
// describing the invocations of this function.
func (f *StoreGetGlobalFeatureFlagFunc) History() []StoreGetGlobalFeatureFlagFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetGlobalFeatureFlagFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetGlobalFeatureFlagFuncCall is an object that describes an
// invocation of method GetGlobalFeatureFlag on an instance of MockStore.
type StoreGetGlobalFeatureFlagFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetGlobalFeatureFlagFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetGlobalFeatureFlagFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreGetGlobalFeatureFlagsFunc describes the behavior when the
// GetGlobalFeatureFlags method of the parent MockStore instance is invoked.
type StoreGetGlobalFeatureFlagsFunc struct {
	defaultHook func(context.Context) (map[string]bool, error)
	hooks       []func(context.Context) (map[string]bool, error)
	history     []StoreGetGlobalFeatureFlagsFuncCall
	mutex       sync.Mutex
}

// GetGlobalFeatureFlags delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockStore) GetGlobalFeatureFlags(v0 context.Context) (map[string]bool, error) {
	r0, r1 := m.GetGlobalFeatureFlagsFunc.nextHook()(v0)
	m.GetGlobalFeatureFlagsFunc.appendCall(StoreGetGlobalFeatureFlagsFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// GetGlobalFeatureFlags method of the parent MockStore instance is invoked
// and the hook queue is empty.
func (f *StoreGetGlobalFeatureFlagsFunc) SetDefaultHook(hook func(context.Context) (map[string]bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetGlobalFeatureFlags method of the parent MockStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *StoreGetGlobalFeatureFlagsFunc) PushHook(hook func(context.Context) (map[string]bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetGlobalFeatureFlagsFunc) SetDefaultReturn(r0 map[string]bool, r1 error) {
	f.SetDefaultHook(func(context.Context) (map[string]bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetGlobalFeatureFlagsFunc) PushReturn(r0 map[string]bool, r1 error) {
	f.PushHook(func(context.Context) (map[string]bool, error) {
		return r0, r1
	})
}

func (f *StoreGetGlobalFeatureFlagsFunc) nextHook() func(context.Context) (map[string]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetGlobalFeatureFlagsFunc) appendCall(r0 StoreGetGlobalFeatureFlagsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetGlobalFeatureFlagsFuncCall objects
// describing the invocations of this function.
func (f *StoreGetGlobalFeatureFlagsFunc) History() []StoreGetGlobalFeatureFlagsFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetGlobalFeatureFlagsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetGlobalFeatureFlagsFuncCall is an object that describes an
// invocation of method GetGlobalFeatureFlags on an instance of MockStore.
type StoreGetGlobalFeatureFlagsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string]bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetGlobalFeatureFlagsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetGlobalFeatureFlagsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreGetUserFlagFunc describes the behavior when the GetUserFlag method
// of the parent MockStore instance is invoked.
type StoreGetUserFlagFunc struct {
	defaultHook func(context.Context, int32, string) (*bool, error)
	hooks       []func(context.Context, int32, string) (*bool, error)
	history     []StoreGetUserFlagFuncCall
	mutex       sync.Mutex
}

// GetUserFlag delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) GetUserFlag(v0 context.Context, v1 int32, v2 string) (*bool, error) {
	r0, r1 := m.GetUserFlagFunc.nextHook()(v0, v1, v2)
	m.GetUserFlagFunc.appendCall(StoreGetUserFlagFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetUserFlag method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreGetUserFlagFunc) SetDefaultHook(hook func(context.Context, int32, string) (*bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetUserFlag method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreGetUserFlagFunc) PushHook(hook func(context.Context, int32, string) (*bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetUserFlagFunc) SetDefaultReturn(r0 *bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int32, string) (*bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetUserFlagFunc) PushReturn(r0 *bool, r1 error) {
	f.PushHook(func(context.Context, int32, string) (*bool, error) {
		return r0, r1
	})
}

func (f *StoreGetUserFlagFunc) nextHook() func(context.Context, int32, string) (*bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetUserFlagFunc) appendCall(r0 StoreGetUserFlagFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetUserFlagFuncCall objects describing
// the invocations of this function.
func (f *StoreGetUserFlagFunc) History() []StoreGetUserFlagFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetUserFlagFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetUserFlagFuncCall is an object that describes an invocation of
// method GetUserFlag on an instance of MockStore.
type StoreGetUserFlagFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int32
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetUserFlagFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetUserFlagFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreGetUserFlagsFunc describes the behavior when the GetUserFlags method
// of the parent MockStore instance is invoked.
type StoreGetUserFlagsFunc struct {
	defaultHook func(context.Context, int32) (map[string]bool, error)
	hooks       []func(context.Context, int32) (map[string]bool, error)
	history     []StoreGetUserFlagsFuncCall
	mutex       sync.Mutex
}

// GetUserFlags delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) GetUserFlags(v0 context.Context, v1 int32) (map[string]bool, error) {
	r0, r1 := m.GetUserFlagsFunc.nextHook()(v0, v1)
	m.GetUserFlagsFunc.appendCall(StoreGetUserFlagsFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetUserFlags method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreGetUserFlagsFunc) SetDefaultHook(hook func(context.Context, int32) (map[string]bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetUserFlags method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreGetUserFlagsFunc) PushHook(hook func(context.Context, int32) (map[string]bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetUserFlagsFunc) SetDefaultReturn(r0 map[string]bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int32) (map[string]bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetUserFlagsFunc) PushReturn(r0 map[string]bool, r1 error) {
	f.PushHook(func(context.Context, int32) (map[string]bool, error) {
		return r0, r1
	})
}

func (f *StoreGetUserFlagsFunc) nextHook() func(context.Context, int32) (map[string]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetUserFlagsFunc) appendCall(r0 StoreGetUserFlagsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetUserFlagsFuncCall objects
// describing the invocations of this function.
func (f *StoreGetUserFlagsFunc) History() []StoreGetUserFlagsFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetUserFlagsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetUserFlagsFuncCall is an object that describes an invocation of
// method GetUserFlags on an instance of MockStore.
type StoreGetUserFlagsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int32
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string]bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetUserFlagsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetUserFlagsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
