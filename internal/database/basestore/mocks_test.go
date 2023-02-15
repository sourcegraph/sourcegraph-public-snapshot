// Code generated by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package basestore

import "sync"

// MockRows is a mock implementation of the Rows interface (from the package
// github.com/sourcegraph/sourcegraph/internal/database/basestore) used for
// unit testing.
type MockRows struct {
	// CloseFunc is an instance of a mock function object controlling the
	// behavior of the method Close.
	CloseFunc *RowsCloseFunc
	// ErrFunc is an instance of a mock function object controlling the
	// behavior of the method Err.
	ErrFunc *RowsErrFunc
	// NextFunc is an instance of a mock function object controlling the
	// behavior of the method Next.
	NextFunc *RowsNextFunc
	// ScanFunc is an instance of a mock function object controlling the
	// behavior of the method Scan.
	ScanFunc *RowsScanFunc
}

// NewMockRows creates a new mock of the Rows interface. All methods return
// zero values for all results, unless overwritten.
func NewMockRows() *MockRows {
	return &MockRows{
		CloseFunc: &RowsCloseFunc{
			defaultHook: func() (r0 error) {
				return
			},
		},
		ErrFunc: &RowsErrFunc{
			defaultHook: func() (r0 error) {
				return
			},
		},
		NextFunc: &RowsNextFunc{
			defaultHook: func() (r0 bool) {
				return
			},
		},
		ScanFunc: &RowsScanFunc{
			defaultHook: func(...interface{}) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockRows creates a new mock of the Rows interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockRows() *MockRows {
	return &MockRows{
		CloseFunc: &RowsCloseFunc{
			defaultHook: func() error {
				panic("unexpected invocation of MockRows.Close")
			},
		},
		ErrFunc: &RowsErrFunc{
			defaultHook: func() error {
				panic("unexpected invocation of MockRows.Err")
			},
		},
		NextFunc: &RowsNextFunc{
			defaultHook: func() bool {
				panic("unexpected invocation of MockRows.Next")
			},
		},
		ScanFunc: &RowsScanFunc{
			defaultHook: func(...interface{}) error {
				panic("unexpected invocation of MockRows.Scan")
			},
		},
	}
}

// NewMockRowsFrom creates a new mock of the MockRows interface. All methods
// delegate to the given implementation, unless overwritten.
func NewMockRowsFrom(i Rows) *MockRows {
	return &MockRows{
		CloseFunc: &RowsCloseFunc{
			defaultHook: i.Close,
		},
		ErrFunc: &RowsErrFunc{
			defaultHook: i.Err,
		},
		NextFunc: &RowsNextFunc{
			defaultHook: i.Next,
		},
		ScanFunc: &RowsScanFunc{
			defaultHook: i.Scan,
		},
	}
}

// RowsCloseFunc describes the behavior when the Close method of the parent
// MockRows instance is invoked.
type RowsCloseFunc struct {
	defaultHook func() error
	hooks       []func() error
	history     []RowsCloseFuncCall
	mutex       sync.Mutex
}

// Close delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockRows) Close() error {
	r0 := m.CloseFunc.nextHook()()
	m.CloseFunc.appendCall(RowsCloseFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Close method of the
// parent MockRows instance is invoked and the hook queue is empty.
func (f *RowsCloseFunc) SetDefaultHook(hook func() error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Close method of the parent MockRows instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *RowsCloseFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *RowsCloseFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func() error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *RowsCloseFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *RowsCloseFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RowsCloseFunc) appendCall(r0 RowsCloseFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of RowsCloseFuncCall objects describing the
// invocations of this function.
func (f *RowsCloseFunc) History() []RowsCloseFuncCall {
	f.mutex.Lock()
	history := make([]RowsCloseFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RowsCloseFuncCall is an object that describes an invocation of method
// Close on an instance of MockRows.
type RowsCloseFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c RowsCloseFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c RowsCloseFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// RowsErrFunc describes the behavior when the Err method of the parent
// MockRows instance is invoked.
type RowsErrFunc struct {
	defaultHook func() error
	hooks       []func() error
	history     []RowsErrFuncCall
	mutex       sync.Mutex
}

// Err delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockRows) Err() error {
	r0 := m.ErrFunc.nextHook()()
	m.ErrFunc.appendCall(RowsErrFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Err method of the
// parent MockRows instance is invoked and the hook queue is empty.
func (f *RowsErrFunc) SetDefaultHook(hook func() error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Err method of the parent MockRows instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *RowsErrFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *RowsErrFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func() error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *RowsErrFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *RowsErrFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RowsErrFunc) appendCall(r0 RowsErrFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of RowsErrFuncCall objects describing the
// invocations of this function.
func (f *RowsErrFunc) History() []RowsErrFuncCall {
	f.mutex.Lock()
	history := make([]RowsErrFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RowsErrFuncCall is an object that describes an invocation of method Err
// on an instance of MockRows.
type RowsErrFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c RowsErrFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c RowsErrFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// RowsNextFunc describes the behavior when the Next method of the parent
// MockRows instance is invoked.
type RowsNextFunc struct {
	defaultHook func() bool
	hooks       []func() bool
	history     []RowsNextFuncCall
	mutex       sync.Mutex
}

// Next delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockRows) Next() bool {
	r0 := m.NextFunc.nextHook()()
	m.NextFunc.appendCall(RowsNextFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Next method of the
// parent MockRows instance is invoked and the hook queue is empty.
func (f *RowsNextFunc) SetDefaultHook(hook func() bool) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Next method of the parent MockRows instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *RowsNextFunc) PushHook(hook func() bool) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *RowsNextFunc) SetDefaultReturn(r0 bool) {
	f.SetDefaultHook(func() bool {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *RowsNextFunc) PushReturn(r0 bool) {
	f.PushHook(func() bool {
		return r0
	})
}

func (f *RowsNextFunc) nextHook() func() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RowsNextFunc) appendCall(r0 RowsNextFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of RowsNextFuncCall objects describing the
// invocations of this function.
func (f *RowsNextFunc) History() []RowsNextFuncCall {
	f.mutex.Lock()
	history := make([]RowsNextFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RowsNextFuncCall is an object that describes an invocation of method Next
// on an instance of MockRows.
type RowsNextFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c RowsNextFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c RowsNextFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// RowsScanFunc describes the behavior when the Scan method of the parent
// MockRows instance is invoked.
type RowsScanFunc struct {
	defaultHook func(...interface{}) error
	hooks       []func(...interface{}) error
	history     []RowsScanFuncCall
	mutex       sync.Mutex
}

// Scan delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockRows) Scan(v0 ...interface{}) error {
	r0 := m.ScanFunc.nextHook()(v0...)
	m.ScanFunc.appendCall(RowsScanFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Scan method of the
// parent MockRows instance is invoked and the hook queue is empty.
func (f *RowsScanFunc) SetDefaultHook(hook func(...interface{}) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Scan method of the parent MockRows instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *RowsScanFunc) PushHook(hook func(...interface{}) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *RowsScanFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(...interface{}) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *RowsScanFunc) PushReturn(r0 error) {
	f.PushHook(func(...interface{}) error {
		return r0
	})
}

func (f *RowsScanFunc) nextHook() func(...interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RowsScanFunc) appendCall(r0 RowsScanFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of RowsScanFuncCall objects describing the
// invocations of this function.
func (f *RowsScanFunc) History() []RowsScanFuncCall {
	f.mutex.Lock()
	history := make([]RowsScanFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RowsScanFuncCall is an object that describes an invocation of method Scan
// on an instance of MockRows.
type RowsScanFuncCall struct {
	// Arg0 is a slice containing the values of the variadic arguments
	// passed to this method invocation.
	Arg0 []interface{}
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation. The variadic slice argument is flattened in this array such
// that one positional argument and three variadic arguments would result in
// a slice of four, not two.
func (c RowsScanFuncCall) Args() []interface{} {
	trailing := []interface{}{}
	for _, val := range c.Arg0 {
		trailing = append(trailing, val)
	}

	return append([]interface{}{}, trailing...)
}

// Results returns an interface slice containing the results of this
// invocation.
func (c RowsScanFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
