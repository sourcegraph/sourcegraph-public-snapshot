// Code generated by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package mocks

import (
	"context"
	"io"
	"sync"
	"time"

	uploadstore "github.com/sourcegraph/sourcegraph/internal/uploadstore"
	iterator "github.com/sourcegraph/sourcegraph/lib/iterator"
)

// MockStore is a mock implementation of the Store interface (from the
// package github.com/sourcegraph/sourcegraph/internal/uploadstore) used for
// unit testing.
type MockStore struct {
	// ComposeFunc is an instance of a mock function object controlling the
	// behavior of the method Compose.
	ComposeFunc *StoreComposeFunc
	// DeleteFunc is an instance of a mock function object controlling the
	// behavior of the method Delete.
	DeleteFunc *StoreDeleteFunc
	// ExpireObjectsFunc is an instance of a mock function object
	// controlling the behavior of the method ExpireObjects.
	ExpireObjectsFunc *StoreExpireObjectsFunc
	// GetFunc is an instance of a mock function object controlling the
	// behavior of the method Get.
	GetFunc *StoreGetFunc
	// InitFunc is an instance of a mock function object controlling the
	// behavior of the method Init.
	InitFunc *StoreInitFunc
	// ListFunc is an instance of a mock function object controlling the
	// behavior of the method List.
	ListFunc *StoreListFunc
	// UploadFunc is an instance of a mock function object controlling the
	// behavior of the method Upload.
	UploadFunc *StoreUploadFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		ComposeFunc: &StoreComposeFunc{
			defaultHook: func(context.Context, string, ...string) (r0 int64, r1 error) {
				return
			},
		},
		DeleteFunc: &StoreDeleteFunc{
			defaultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		ExpireObjectsFunc: &StoreExpireObjectsFunc{
			defaultHook: func(context.Context, string, time.Duration) (r0 error) {
				return
			},
		},
		GetFunc: &StoreGetFunc{
			defaultHook: func(context.Context, string) (r0 io.ReadCloser, r1 error) {
				return
			},
		},
		InitFunc: &StoreInitFunc{
			defaultHook: func(context.Context) (r0 error) {
				return
			},
		},
		ListFunc: &StoreListFunc{
			defaultHook: func(context.Context, string) (r0 *iterator.Iterator[string], r1 error) {
				return
			},
		},
		UploadFunc: &StoreUploadFunc{
			defaultHook: func(context.Context, string, io.Reader) (r0 int64, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore creates a new mock of the Store interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		ComposeFunc: &StoreComposeFunc{
			defaultHook: func(context.Context, string, ...string) (int64, error) {
				panic("unexpected invocation of MockStore.Compose")
			},
		},
		DeleteFunc: &StoreDeleteFunc{
			defaultHook: func(context.Context, string) error {
				panic("unexpected invocation of MockStore.Delete")
			},
		},
		ExpireObjectsFunc: &StoreExpireObjectsFunc{
			defaultHook: func(context.Context, string, time.Duration) error {
				panic("unexpected invocation of MockStore.ExpireObjects")
			},
		},
		GetFunc: &StoreGetFunc{
			defaultHook: func(context.Context, string) (io.ReadCloser, error) {
				panic("unexpected invocation of MockStore.Get")
			},
		},
		InitFunc: &StoreInitFunc{
			defaultHook: func(context.Context) error {
				panic("unexpected invocation of MockStore.Init")
			},
		},
		ListFunc: &StoreListFunc{
			defaultHook: func(context.Context, string) (*iterator.Iterator[string], error) {
				panic("unexpected invocation of MockStore.List")
			},
		},
		UploadFunc: &StoreUploadFunc{
			defaultHook: func(context.Context, string, io.Reader) (int64, error) {
				panic("unexpected invocation of MockStore.Upload")
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i uploadstore.Store) *MockStore {
	return &MockStore{
		ComposeFunc: &StoreComposeFunc{
			defaultHook: i.Compose,
		},
		DeleteFunc: &StoreDeleteFunc{
			defaultHook: i.Delete,
		},
		ExpireObjectsFunc: &StoreExpireObjectsFunc{
			defaultHook: i.ExpireObjects,
		},
		GetFunc: &StoreGetFunc{
			defaultHook: i.Get,
		},
		InitFunc: &StoreInitFunc{
			defaultHook: i.Init,
		},
		ListFunc: &StoreListFunc{
			defaultHook: i.List,
		},
		UploadFunc: &StoreUploadFunc{
			defaultHook: i.Upload,
		},
	}
}

// StoreComposeFunc describes the behavior when the Compose method of the
// parent MockStore instance is invoked.
type StoreComposeFunc struct {
	defaultHook func(context.Context, string, ...string) (int64, error)
	hooks       []func(context.Context, string, ...string) (int64, error)
	history     []StoreComposeFuncCall
	mutex       sync.Mutex
}

// Compose delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Compose(v0 context.Context, v1 string, v2 ...string) (int64, error) {
	r0, r1 := m.ComposeFunc.nextHook()(v0, v1, v2...)
	m.ComposeFunc.appendCall(StoreComposeFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Compose method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreComposeFunc) SetDefaultHook(hook func(context.Context, string, ...string) (int64, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Compose method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreComposeFunc) PushHook(hook func(context.Context, string, ...string) (int64, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreComposeFunc) SetDefaultReturn(r0 int64, r1 error) {
	f.SetDefaultHook(func(context.Context, string, ...string) (int64, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreComposeFunc) PushReturn(r0 int64, r1 error) {
	f.PushHook(func(context.Context, string, ...string) (int64, error) {
		return r0, r1
	})
}

func (f *StoreComposeFunc) nextHook() func(context.Context, string, ...string) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreComposeFunc) appendCall(r0 StoreComposeFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreComposeFuncCall objects describing the
// invocations of this function.
func (f *StoreComposeFunc) History() []StoreComposeFuncCall {
	f.mutex.Lock()
	history := make([]StoreComposeFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreComposeFuncCall is an object that describes an invocation of method
// Compose on an instance of MockStore.
type StoreComposeFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is a slice containing the values of the variadic arguments
	// passed to this method invocation.
	Arg2 []string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int64
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation. The variadic slice argument is flattened in this array such
// that one positional argument and three variadic arguments would result in
// a slice of four, not two.
func (c StoreComposeFuncCall) Args() []interface{} {
	trailing := []interface{}{}
	for _, val := range c.Arg2 {
		trailing = append(trailing, val)
	}

	return append([]interface{}{c.Arg0, c.Arg1}, trailing...)
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreComposeFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreDeleteFunc describes the behavior when the Delete method of the
// parent MockStore instance is invoked.
type StoreDeleteFunc struct {
	defaultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []StoreDeleteFuncCall
	mutex       sync.Mutex
}

// Delete delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Delete(v0 context.Context, v1 string) error {
	r0 := m.DeleteFunc.nextHook()(v0, v1)
	m.DeleteFunc.appendCall(StoreDeleteFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Delete method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreDeleteFunc) SetDefaultHook(hook func(context.Context, string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Delete method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreDeleteFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreDeleteFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreDeleteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *StoreDeleteFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteFunc) appendCall(r0 StoreDeleteFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDeleteFuncCall objects describing the
// invocations of this function.
func (f *StoreDeleteFunc) History() []StoreDeleteFuncCall {
	f.mutex.Lock()
	history := make([]StoreDeleteFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteFuncCall is an object that describes an invocation of method
// Delete on an instance of MockStore.
type StoreDeleteFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDeleteFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDeleteFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreExpireObjectsFunc describes the behavior when the ExpireObjects
// method of the parent MockStore instance is invoked.
type StoreExpireObjectsFunc struct {
	defaultHook func(context.Context, string, time.Duration) error
	hooks       []func(context.Context, string, time.Duration) error
	history     []StoreExpireObjectsFuncCall
	mutex       sync.Mutex
}

// ExpireObjects delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) ExpireObjects(v0 context.Context, v1 string, v2 time.Duration) error {
	r0 := m.ExpireObjectsFunc.nextHook()(v0, v1, v2)
	m.ExpireObjectsFunc.appendCall(StoreExpireObjectsFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the ExpireObjects method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreExpireObjectsFunc) SetDefaultHook(hook func(context.Context, string, time.Duration) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// ExpireObjects method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreExpireObjectsFunc) PushHook(hook func(context.Context, string, time.Duration) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreExpireObjectsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, string, time.Duration) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreExpireObjectsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, time.Duration) error {
		return r0
	})
}

func (f *StoreExpireObjectsFunc) nextHook() func(context.Context, string, time.Duration) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreExpireObjectsFunc) appendCall(r0 StoreExpireObjectsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreExpireObjectsFuncCall objects
// describing the invocations of this function.
func (f *StoreExpireObjectsFunc) History() []StoreExpireObjectsFuncCall {
	f.mutex.Lock()
	history := make([]StoreExpireObjectsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreExpireObjectsFuncCall is an object that describes an invocation of
// method ExpireObjects on an instance of MockStore.
type StoreExpireObjectsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 time.Duration
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreExpireObjectsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreExpireObjectsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreGetFunc describes the behavior when the Get method of the parent
// MockStore instance is invoked.
type StoreGetFunc struct {
	defaultHook func(context.Context, string) (io.ReadCloser, error)
	hooks       []func(context.Context, string) (io.ReadCloser, error)
	history     []StoreGetFuncCall
	mutex       sync.Mutex
}

// Get delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Get(v0 context.Context, v1 string) (io.ReadCloser, error) {
	r0, r1 := m.GetFunc.nextHook()(v0, v1)
	m.GetFunc.appendCall(StoreGetFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Get method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreGetFunc) SetDefaultHook(hook func(context.Context, string) (io.ReadCloser, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Get method of the parent MockStore instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *StoreGetFunc) PushHook(hook func(context.Context, string) (io.ReadCloser, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetFunc) SetDefaultReturn(r0 io.ReadCloser, r1 error) {
	f.SetDefaultHook(func(context.Context, string) (io.ReadCloser, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetFunc) PushReturn(r0 io.ReadCloser, r1 error) {
	f.PushHook(func(context.Context, string) (io.ReadCloser, error) {
		return r0, r1
	})
}

func (f *StoreGetFunc) nextHook() func(context.Context, string) (io.ReadCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetFunc) appendCall(r0 StoreGetFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetFuncCall objects describing the
// invocations of this function.
func (f *StoreGetFunc) History() []StoreGetFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetFuncCall is an object that describes an invocation of method Get
// on an instance of MockStore.
type StoreGetFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 io.ReadCloser
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreInitFunc describes the behavior when the Init method of the parent
// MockStore instance is invoked.
type StoreInitFunc struct {
	defaultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []StoreInitFuncCall
	mutex       sync.Mutex
}

// Init delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Init(v0 context.Context) error {
	r0 := m.InitFunc.nextHook()(v0)
	m.InitFunc.appendCall(StoreInitFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Init method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreInitFunc) SetDefaultHook(hook func(context.Context) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Init method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreInitFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreInitFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreInitFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *StoreInitFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInitFunc) appendCall(r0 StoreInitFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreInitFuncCall objects describing the
// invocations of this function.
func (f *StoreInitFunc) History() []StoreInitFuncCall {
	f.mutex.Lock()
	history := make([]StoreInitFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInitFuncCall is an object that describes an invocation of method
// Init on an instance of MockStore.
type StoreInitFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreInitFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreInitFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreListFunc describes the behavior when the List method of the parent
// MockStore instance is invoked.
type StoreListFunc struct {
	defaultHook func(context.Context, string) (*iterator.Iterator[string], error)
	hooks       []func(context.Context, string) (*iterator.Iterator[string], error)
	history     []StoreListFuncCall
	mutex       sync.Mutex
}

// List delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) List(v0 context.Context, v1 string) (*iterator.Iterator[string], error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.appendCall(StoreListFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the List method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreListFunc) SetDefaultHook(hook func(context.Context, string) (*iterator.Iterator[string], error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// List method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreListFunc) PushHook(hook func(context.Context, string) (*iterator.Iterator[string], error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreListFunc) SetDefaultReturn(r0 *iterator.Iterator[string], r1 error) {
	f.SetDefaultHook(func(context.Context, string) (*iterator.Iterator[string], error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreListFunc) PushReturn(r0 *iterator.Iterator[string], r1 error) {
	f.PushHook(func(context.Context, string) (*iterator.Iterator[string], error) {
		return r0, r1
	})
}

func (f *StoreListFunc) nextHook() func(context.Context, string) (*iterator.Iterator[string], error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreListFunc) appendCall(r0 StoreListFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreListFuncCall objects describing the
// invocations of this function.
func (f *StoreListFunc) History() []StoreListFuncCall {
	f.mutex.Lock()
	history := make([]StoreListFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreListFuncCall is an object that describes an invocation of method
// List on an instance of MockStore.
type StoreListFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *iterator.Iterator[string]
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreListFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreListFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreUploadFunc describes the behavior when the Upload method of the
// parent MockStore instance is invoked.
type StoreUploadFunc struct {
	defaultHook func(context.Context, string, io.Reader) (int64, error)
	hooks       []func(context.Context, string, io.Reader) (int64, error)
	history     []StoreUploadFuncCall
	mutex       sync.Mutex
}

// Upload delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Upload(v0 context.Context, v1 string, v2 io.Reader) (int64, error) {
	r0, r1 := m.UploadFunc.nextHook()(v0, v1, v2)
	m.UploadFunc.appendCall(StoreUploadFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Upload method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreUploadFunc) SetDefaultHook(hook func(context.Context, string, io.Reader) (int64, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Upload method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreUploadFunc) PushHook(hook func(context.Context, string, io.Reader) (int64, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreUploadFunc) SetDefaultReturn(r0 int64, r1 error) {
	f.SetDefaultHook(func(context.Context, string, io.Reader) (int64, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreUploadFunc) PushReturn(r0 int64, r1 error) {
	f.PushHook(func(context.Context, string, io.Reader) (int64, error) {
		return r0, r1
	})
}

func (f *StoreUploadFunc) nextHook() func(context.Context, string, io.Reader) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUploadFunc) appendCall(r0 StoreUploadFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreUploadFuncCall objects describing the
// invocations of this function.
func (f *StoreUploadFunc) History() []StoreUploadFuncCall {
	f.mutex.Lock()
	history := make([]StoreUploadFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUploadFuncCall is an object that describes an invocation of method
// Upload on an instance of MockStore.
type StoreUploadFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 io.Reader
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int64
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreUploadFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreUploadFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
