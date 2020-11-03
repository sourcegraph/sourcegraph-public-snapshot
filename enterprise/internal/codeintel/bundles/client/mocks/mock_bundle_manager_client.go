// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package mocks

import (
	"context"
	client "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"io"
	"sync"
)

// MockBundleManagerClient is a mock implementation of the
// BundleManagerClient interface (from the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client)
// used for unit testing.
type MockBundleManagerClient struct {
	// BundleClientFunc is an instance of a mock function object controlling
	// the behavior of the method BundleClient.
	BundleClientFunc *BundleManagerClientBundleClientFunc
	// DeleteUploadFunc is an instance of a mock function object controlling
	// the behavior of the method DeleteUpload.
	DeleteUploadFunc *BundleManagerClientDeleteUploadFunc
	// ExistsFunc is an instance of a mock function object controlling the
	// behavior of the method Exists.
	ExistsFunc *BundleManagerClientExistsFunc
	// GetUploadFunc is an instance of a mock function object controlling
	// the behavior of the method GetUpload.
	GetUploadFunc *BundleManagerClientGetUploadFunc
	// SendDBFunc is an instance of a mock function object controlling the
	// behavior of the method SendDB.
	SendDBFunc *BundleManagerClientSendDBFunc
	// SendUploadFunc is an instance of a mock function object controlling
	// the behavior of the method SendUpload.
	SendUploadFunc *BundleManagerClientSendUploadFunc
	// SendUploadPartFunc is an instance of a mock function object
	// controlling the behavior of the method SendUploadPart.
	SendUploadPartFunc *BundleManagerClientSendUploadPartFunc
	// StitchPartsFunc is an instance of a mock function object controlling
	// the behavior of the method StitchParts.
	StitchPartsFunc *BundleManagerClientStitchPartsFunc
}

// NewMockBundleManagerClient creates a new mock of the BundleManagerClient
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockBundleManagerClient() *MockBundleManagerClient {
	return &MockBundleManagerClient{
		BundleClientFunc: &BundleManagerClientBundleClientFunc{
			defaultHook: func(int) client.BundleClient {
				return nil
			},
		},
		DeleteUploadFunc: &BundleManagerClientDeleteUploadFunc{
			defaultHook: func(context.Context, int) error {
				return nil
			},
		},
		ExistsFunc: &BundleManagerClientExistsFunc{
			defaultHook: func(context.Context, []int) (map[int]bool, error) {
				return nil, nil
			},
		},
		GetUploadFunc: &BundleManagerClientGetUploadFunc{
			defaultHook: func(context.Context, int) (io.ReadCloser, error) {
				return nil, nil
			},
		},
		SendDBFunc: &BundleManagerClientSendDBFunc{
			defaultHook: func(context.Context, int, string) error {
				return nil
			},
		},
		SendUploadFunc: &BundleManagerClientSendUploadFunc{
			defaultHook: func(context.Context, int, io.Reader) (int64, error) {
				return 0, nil
			},
		},
		SendUploadPartFunc: &BundleManagerClientSendUploadPartFunc{
			defaultHook: func(context.Context, int, int, io.Reader) error {
				return nil
			},
		},
		StitchPartsFunc: &BundleManagerClientStitchPartsFunc{
			defaultHook: func(context.Context, int, int) (int64, error) {
				return 0, nil
			},
		},
	}
}

// NewMockBundleManagerClientFrom creates a new mock of the
// MockBundleManagerClient interface. All methods delegate to the given
// implementation, unless overwritten.
func NewMockBundleManagerClientFrom(i client.BundleManagerClient) *MockBundleManagerClient {
	return &MockBundleManagerClient{
		BundleClientFunc: &BundleManagerClientBundleClientFunc{
			defaultHook: i.BundleClient,
		},
		DeleteUploadFunc: &BundleManagerClientDeleteUploadFunc{
			defaultHook: i.DeleteUpload,
		},
		ExistsFunc: &BundleManagerClientExistsFunc{
			defaultHook: i.Exists,
		},
		GetUploadFunc: &BundleManagerClientGetUploadFunc{
			defaultHook: i.GetUpload,
		},
		SendDBFunc: &BundleManagerClientSendDBFunc{
			defaultHook: i.SendDB,
		},
		SendUploadFunc: &BundleManagerClientSendUploadFunc{
			defaultHook: i.SendUpload,
		},
		SendUploadPartFunc: &BundleManagerClientSendUploadPartFunc{
			defaultHook: i.SendUploadPart,
		},
		StitchPartsFunc: &BundleManagerClientStitchPartsFunc{
			defaultHook: i.StitchParts,
		},
	}
}

// BundleManagerClientBundleClientFunc describes the behavior when the
// BundleClient method of the parent MockBundleManagerClient instance is
// invoked.
type BundleManagerClientBundleClientFunc struct {
	defaultHook func(int) client.BundleClient
	hooks       []func(int) client.BundleClient
	history     []BundleManagerClientBundleClientFuncCall
	mutex       sync.Mutex
}

// BundleClient delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockBundleManagerClient) BundleClient(v0 int) client.BundleClient {
	r0 := m.BundleClientFunc.nextHook()(v0)
	m.BundleClientFunc.appendCall(BundleManagerClientBundleClientFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the BundleClient method
// of the parent MockBundleManagerClient instance is invoked and the hook
// queue is empty.
func (f *BundleManagerClientBundleClientFunc) SetDefaultHook(hook func(int) client.BundleClient) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// BundleClient method of the parent MockBundleManagerClient instance
// inovkes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *BundleManagerClientBundleClientFunc) PushHook(hook func(int) client.BundleClient) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientBundleClientFunc) SetDefaultReturn(r0 client.BundleClient) {
	f.SetDefaultHook(func(int) client.BundleClient {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientBundleClientFunc) PushReturn(r0 client.BundleClient) {
	f.PushHook(func(int) client.BundleClient {
		return r0
	})
}

func (f *BundleManagerClientBundleClientFunc) nextHook() func(int) client.BundleClient {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientBundleClientFunc) appendCall(r0 BundleManagerClientBundleClientFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientBundleClientFuncCall
// objects describing the invocations of this function.
func (f *BundleManagerClientBundleClientFunc) History() []BundleManagerClientBundleClientFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientBundleClientFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientBundleClientFuncCall is an object that describes an
// invocation of method BundleClient on an instance of
// MockBundleManagerClient.
type BundleManagerClientBundleClientFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 client.BundleClient
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientBundleClientFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientBundleClientFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// BundleManagerClientDeleteUploadFunc describes the behavior when the
// DeleteUpload method of the parent MockBundleManagerClient instance is
// invoked.
type BundleManagerClientDeleteUploadFunc struct {
	defaultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []BundleManagerClientDeleteUploadFuncCall
	mutex       sync.Mutex
}

// DeleteUpload delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockBundleManagerClient) DeleteUpload(v0 context.Context, v1 int) error {
	r0 := m.DeleteUploadFunc.nextHook()(v0, v1)
	m.DeleteUploadFunc.appendCall(BundleManagerClientDeleteUploadFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the DeleteUpload method
// of the parent MockBundleManagerClient instance is invoked and the hook
// queue is empty.
func (f *BundleManagerClientDeleteUploadFunc) SetDefaultHook(hook func(context.Context, int) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// DeleteUpload method of the parent MockBundleManagerClient instance
// inovkes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *BundleManagerClientDeleteUploadFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientDeleteUploadFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientDeleteUploadFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *BundleManagerClientDeleteUploadFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientDeleteUploadFunc) appendCall(r0 BundleManagerClientDeleteUploadFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientDeleteUploadFuncCall
// objects describing the invocations of this function.
func (f *BundleManagerClientDeleteUploadFunc) History() []BundleManagerClientDeleteUploadFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientDeleteUploadFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientDeleteUploadFuncCall is an object that describes an
// invocation of method DeleteUpload on an instance of
// MockBundleManagerClient.
type BundleManagerClientDeleteUploadFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientDeleteUploadFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientDeleteUploadFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// BundleManagerClientExistsFunc describes the behavior when the Exists
// method of the parent MockBundleManagerClient instance is invoked.
type BundleManagerClientExistsFunc struct {
	defaultHook func(context.Context, []int) (map[int]bool, error)
	hooks       []func(context.Context, []int) (map[int]bool, error)
	history     []BundleManagerClientExistsFuncCall
	mutex       sync.Mutex
}

// Exists delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockBundleManagerClient) Exists(v0 context.Context, v1 []int) (map[int]bool, error) {
	r0, r1 := m.ExistsFunc.nextHook()(v0, v1)
	m.ExistsFunc.appendCall(BundleManagerClientExistsFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Exists method of the
// parent MockBundleManagerClient instance is invoked and the hook queue is
// empty.
func (f *BundleManagerClientExistsFunc) SetDefaultHook(hook func(context.Context, []int) (map[int]bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Exists method of the parent MockBundleManagerClient instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *BundleManagerClientExistsFunc) PushHook(hook func(context.Context, []int) (map[int]bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientExistsFunc) SetDefaultReturn(r0 map[int]bool, r1 error) {
	f.SetDefaultHook(func(context.Context, []int) (map[int]bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientExistsFunc) PushReturn(r0 map[int]bool, r1 error) {
	f.PushHook(func(context.Context, []int) (map[int]bool, error) {
		return r0, r1
	})
}

func (f *BundleManagerClientExistsFunc) nextHook() func(context.Context, []int) (map[int]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientExistsFunc) appendCall(r0 BundleManagerClientExistsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientExistsFuncCall objects
// describing the invocations of this function.
func (f *BundleManagerClientExistsFunc) History() []BundleManagerClientExistsFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientExistsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientExistsFuncCall is an object that describes an
// invocation of method Exists on an instance of MockBundleManagerClient.
type BundleManagerClientExistsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 []int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[int]bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientExistsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientExistsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// BundleManagerClientGetUploadFunc describes the behavior when the
// GetUpload method of the parent MockBundleManagerClient instance is
// invoked.
type BundleManagerClientGetUploadFunc struct {
	defaultHook func(context.Context, int) (io.ReadCloser, error)
	hooks       []func(context.Context, int) (io.ReadCloser, error)
	history     []BundleManagerClientGetUploadFuncCall
	mutex       sync.Mutex
}

// GetUpload delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockBundleManagerClient) GetUpload(v0 context.Context, v1 int) (io.ReadCloser, error) {
	r0, r1 := m.GetUploadFunc.nextHook()(v0, v1)
	m.GetUploadFunc.appendCall(BundleManagerClientGetUploadFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetUpload method of
// the parent MockBundleManagerClient instance is invoked and the hook queue
// is empty.
func (f *BundleManagerClientGetUploadFunc) SetDefaultHook(hook func(context.Context, int) (io.ReadCloser, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetUpload method of the parent MockBundleManagerClient instance inovkes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *BundleManagerClientGetUploadFunc) PushHook(hook func(context.Context, int) (io.ReadCloser, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientGetUploadFunc) SetDefaultReturn(r0 io.ReadCloser, r1 error) {
	f.SetDefaultHook(func(context.Context, int) (io.ReadCloser, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientGetUploadFunc) PushReturn(r0 io.ReadCloser, r1 error) {
	f.PushHook(func(context.Context, int) (io.ReadCloser, error) {
		return r0, r1
	})
}

func (f *BundleManagerClientGetUploadFunc) nextHook() func(context.Context, int) (io.ReadCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientGetUploadFunc) appendCall(r0 BundleManagerClientGetUploadFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientGetUploadFuncCall
// objects describing the invocations of this function.
func (f *BundleManagerClientGetUploadFunc) History() []BundleManagerClientGetUploadFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientGetUploadFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientGetUploadFuncCall is an object that describes an
// invocation of method GetUpload on an instance of MockBundleManagerClient.
type BundleManagerClientGetUploadFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 io.ReadCloser
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientGetUploadFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientGetUploadFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// BundleManagerClientSendDBFunc describes the behavior when the SendDB
// method of the parent MockBundleManagerClient instance is invoked.
type BundleManagerClientSendDBFunc struct {
	defaultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []BundleManagerClientSendDBFuncCall
	mutex       sync.Mutex
}

// SendDB delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockBundleManagerClient) SendDB(v0 context.Context, v1 int, v2 string) error {
	r0 := m.SendDBFunc.nextHook()(v0, v1, v2)
	m.SendDBFunc.appendCall(BundleManagerClientSendDBFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the SendDB method of the
// parent MockBundleManagerClient instance is invoked and the hook queue is
// empty.
func (f *BundleManagerClientSendDBFunc) SetDefaultHook(hook func(context.Context, int, string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SendDB method of the parent MockBundleManagerClient instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *BundleManagerClientSendDBFunc) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientSendDBFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientSendDBFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *BundleManagerClientSendDBFunc) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientSendDBFunc) appendCall(r0 BundleManagerClientSendDBFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientSendDBFuncCall objects
// describing the invocations of this function.
func (f *BundleManagerClientSendDBFunc) History() []BundleManagerClientSendDBFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientSendDBFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientSendDBFuncCall is an object that describes an
// invocation of method SendDB on an instance of MockBundleManagerClient.
type BundleManagerClientSendDBFuncCall struct {
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
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientSendDBFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientSendDBFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// BundleManagerClientSendUploadFunc describes the behavior when the
// SendUpload method of the parent MockBundleManagerClient instance is
// invoked.
type BundleManagerClientSendUploadFunc struct {
	defaultHook func(context.Context, int, io.Reader) (int64, error)
	hooks       []func(context.Context, int, io.Reader) (int64, error)
	history     []BundleManagerClientSendUploadFuncCall
	mutex       sync.Mutex
}

// SendUpload delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockBundleManagerClient) SendUpload(v0 context.Context, v1 int, v2 io.Reader) (int64, error) {
	r0, r1 := m.SendUploadFunc.nextHook()(v0, v1, v2)
	m.SendUploadFunc.appendCall(BundleManagerClientSendUploadFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the SendUpload method of
// the parent MockBundleManagerClient instance is invoked and the hook queue
// is empty.
func (f *BundleManagerClientSendUploadFunc) SetDefaultHook(hook func(context.Context, int, io.Reader) (int64, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SendUpload method of the parent MockBundleManagerClient instance inovkes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *BundleManagerClientSendUploadFunc) PushHook(hook func(context.Context, int, io.Reader) (int64, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientSendUploadFunc) SetDefaultReturn(r0 int64, r1 error) {
	f.SetDefaultHook(func(context.Context, int, io.Reader) (int64, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientSendUploadFunc) PushReturn(r0 int64, r1 error) {
	f.PushHook(func(context.Context, int, io.Reader) (int64, error) {
		return r0, r1
	})
}

func (f *BundleManagerClientSendUploadFunc) nextHook() func(context.Context, int, io.Reader) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientSendUploadFunc) appendCall(r0 BundleManagerClientSendUploadFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientSendUploadFuncCall
// objects describing the invocations of this function.
func (f *BundleManagerClientSendUploadFunc) History() []BundleManagerClientSendUploadFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientSendUploadFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientSendUploadFuncCall is an object that describes an
// invocation of method SendUpload on an instance of
// MockBundleManagerClient.
type BundleManagerClientSendUploadFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
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
func (c BundleManagerClientSendUploadFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientSendUploadFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// BundleManagerClientSendUploadPartFunc describes the behavior when the
// SendUploadPart method of the parent MockBundleManagerClient instance is
// invoked.
type BundleManagerClientSendUploadPartFunc struct {
	defaultHook func(context.Context, int, int, io.Reader) error
	hooks       []func(context.Context, int, int, io.Reader) error
	history     []BundleManagerClientSendUploadPartFuncCall
	mutex       sync.Mutex
}

// SendUploadPart delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockBundleManagerClient) SendUploadPart(v0 context.Context, v1 int, v2 int, v3 io.Reader) error {
	r0 := m.SendUploadPartFunc.nextHook()(v0, v1, v2, v3)
	m.SendUploadPartFunc.appendCall(BundleManagerClientSendUploadPartFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the SendUploadPart
// method of the parent MockBundleManagerClient instance is invoked and the
// hook queue is empty.
func (f *BundleManagerClientSendUploadPartFunc) SetDefaultHook(hook func(context.Context, int, int, io.Reader) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SendUploadPart method of the parent MockBundleManagerClient instance
// inovkes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *BundleManagerClientSendUploadPartFunc) PushHook(hook func(context.Context, int, int, io.Reader) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientSendUploadPartFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, int, io.Reader) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientSendUploadPartFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int, io.Reader) error {
		return r0
	})
}

func (f *BundleManagerClientSendUploadPartFunc) nextHook() func(context.Context, int, int, io.Reader) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientSendUploadPartFunc) appendCall(r0 BundleManagerClientSendUploadPartFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientSendUploadPartFuncCall
// objects describing the invocations of this function.
func (f *BundleManagerClientSendUploadPartFunc) History() []BundleManagerClientSendUploadPartFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientSendUploadPartFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientSendUploadPartFuncCall is an object that describes an
// invocation of method SendUploadPart on an instance of
// MockBundleManagerClient.
type BundleManagerClientSendUploadPartFuncCall struct {
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
	Arg3 io.Reader
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientSendUploadPartFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientSendUploadPartFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// BundleManagerClientStitchPartsFunc describes the behavior when the
// StitchParts method of the parent MockBundleManagerClient instance is
// invoked.
type BundleManagerClientStitchPartsFunc struct {
	defaultHook func(context.Context, int, int) (int64, error)
	hooks       []func(context.Context, int, int) (int64, error)
	history     []BundleManagerClientStitchPartsFuncCall
	mutex       sync.Mutex
}

// StitchParts delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockBundleManagerClient) StitchParts(v0 context.Context, v1 int, v2 int) (int64, error) {
	r0, r1 := m.StitchPartsFunc.nextHook()(v0, v1, v2)
	m.StitchPartsFunc.appendCall(BundleManagerClientStitchPartsFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the StitchParts method
// of the parent MockBundleManagerClient instance is invoked and the hook
// queue is empty.
func (f *BundleManagerClientStitchPartsFunc) SetDefaultHook(hook func(context.Context, int, int) (int64, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// StitchParts method of the parent MockBundleManagerClient instance inovkes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *BundleManagerClientStitchPartsFunc) PushHook(hook func(context.Context, int, int) (int64, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *BundleManagerClientStitchPartsFunc) SetDefaultReturn(r0 int64, r1 error) {
	f.SetDefaultHook(func(context.Context, int, int) (int64, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *BundleManagerClientStitchPartsFunc) PushReturn(r0 int64, r1 error) {
	f.PushHook(func(context.Context, int, int) (int64, error) {
		return r0, r1
	})
}

func (f *BundleManagerClientStitchPartsFunc) nextHook() func(context.Context, int, int) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *BundleManagerClientStitchPartsFunc) appendCall(r0 BundleManagerClientStitchPartsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of BundleManagerClientStitchPartsFuncCall
// objects describing the invocations of this function.
func (f *BundleManagerClientStitchPartsFunc) History() []BundleManagerClientStitchPartsFuncCall {
	f.mutex.Lock()
	history := make([]BundleManagerClientStitchPartsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// BundleManagerClientStitchPartsFuncCall is an object that describes an
// invocation of method StitchParts on an instance of
// MockBundleManagerClient.
type BundleManagerClientStitchPartsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int64
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c BundleManagerClientStitchPartsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c BundleManagerClientStitchPartsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
