// Code generated by go-mockgen 1.1.5; DO NOT EDIT.

package workerutil

import (
	"context"
	"sync"
)

// MockWithPreDequeue is a mock implementation of the WithPreDequeue
// interface (from the package
// github.com/sourcegraph/sourcegraph/internal/workerutil) used for unit
// testing.
type MockWithPreDequeue struct {
	// PreDequeueFunc is an instance of a mock function object controlling
	// the behavior of the method PreDequeue.
	PreDequeueFunc *WithPreDequeuePreDequeueFunc
}

// NewMockWithPreDequeue creates a new mock of the WithPreDequeue interface.
// All methods return zero values for all results, unless overwritten.
func NewMockWithPreDequeue() *MockWithPreDequeue {
	return &MockWithPreDequeue{
		PreDequeueFunc: &WithPreDequeuePreDequeueFunc{
			defaultHook: func(context.Context) (bool, interface{}, error) {
				return false, nil, nil
			},
		},
	}
}

// NewStrictMockWithPreDequeue creates a new mock of the WithPreDequeue
// interface. All methods panic on invocation, unless overwritten.
func NewStrictMockWithPreDequeue() *MockWithPreDequeue {
	return &MockWithPreDequeue{
		PreDequeueFunc: &WithPreDequeuePreDequeueFunc{
			defaultHook: func(context.Context) (bool, interface{}, error) {
				panic("unexpected invocation of MockWithPreDequeue.PreDequeue")
			},
		},
	}
}

// NewMockWithPreDequeueFrom creates a new mock of the MockWithPreDequeue
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockWithPreDequeueFrom(i WithPreDequeue) *MockWithPreDequeue {
	return &MockWithPreDequeue{
		PreDequeueFunc: &WithPreDequeuePreDequeueFunc{
			defaultHook: i.PreDequeue,
		},
	}
}

// WithPreDequeuePreDequeueFunc describes the behavior when the PreDequeue
// method of the parent MockWithPreDequeue instance is invoked.
type WithPreDequeuePreDequeueFunc struct {
	defaultHook func(context.Context) (bool, interface{}, error)
	hooks       []func(context.Context) (bool, interface{}, error)
	history     []WithPreDequeuePreDequeueFuncCall
	mutex       sync.Mutex
}

// PreDequeue delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockWithPreDequeue) PreDequeue(v0 context.Context) (bool, interface{}, error) {
	r0, r1, r2 := m.PreDequeueFunc.nextHook()(v0)
	m.PreDequeueFunc.appendCall(WithPreDequeuePreDequeueFuncCall{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the PreDequeue method of
// the parent MockWithPreDequeue instance is invoked and the hook queue is
// empty.
func (f *WithPreDequeuePreDequeueFunc) SetDefaultHook(hook func(context.Context) (bool, interface{}, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// PreDequeue method of the parent MockWithPreDequeue instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *WithPreDequeuePreDequeueFunc) PushHook(hook func(context.Context) (bool, interface{}, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *WithPreDequeuePreDequeueFunc) SetDefaultReturn(r0 bool, r1 interface{}, r2 error) {
	f.SetDefaultHook(func(context.Context) (bool, interface{}, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *WithPreDequeuePreDequeueFunc) PushReturn(r0 bool, r1 interface{}, r2 error) {
	f.PushHook(func(context.Context) (bool, interface{}, error) {
		return r0, r1, r2
	})
}

func (f *WithPreDequeuePreDequeueFunc) nextHook() func(context.Context) (bool, interface{}, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WithPreDequeuePreDequeueFunc) appendCall(r0 WithPreDequeuePreDequeueFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WithPreDequeuePreDequeueFuncCall objects
// describing the invocations of this function.
func (f *WithPreDequeuePreDequeueFunc) History() []WithPreDequeuePreDequeueFuncCall {
	f.mutex.Lock()
	history := make([]WithPreDequeuePreDequeueFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WithPreDequeuePreDequeueFuncCall is an object that describes an
// invocation of method PreDequeue on an instance of MockWithPreDequeue.
type WithPreDequeuePreDequeueFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 interface{}
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WithPreDequeuePreDequeueFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WithPreDequeuePreDequeueFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}
