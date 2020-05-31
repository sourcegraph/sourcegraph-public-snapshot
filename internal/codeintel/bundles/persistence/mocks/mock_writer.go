// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package mocks

import (
	"context"
	persistence "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	types "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"sync"
)

// MockWriter is a mock impelementation of the Writer interface (from the
// package
// github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence)
// used for unit testing.
type MockWriter struct {
	// CloseFunc is an instance of a mock function object controlling the
	// behavior of the method Close.
	CloseFunc *WriterCloseFunc
	// WriteDefinitionsFunc is an instance of a mock function object
	// controlling the behavior of the method WriteDefinitions.
	WriteDefinitionsFunc *WriterWriteDefinitionsFunc
	// WriteDocumentsFunc is an instance of a mock function object
	// controlling the behavior of the method WriteDocuments.
	WriteDocumentsFunc *WriterWriteDocumentsFunc
	// WriteMetaFunc is an instance of a mock function object controlling
	// the behavior of the method WriteMeta.
	WriteMetaFunc *WriterWriteMetaFunc
	// WriteReferencesFunc is an instance of a mock function object
	// controlling the behavior of the method WriteReferences.
	WriteReferencesFunc *WriterWriteReferencesFunc
	// WriteResultChunksFunc is an instance of a mock function object
	// controlling the behavior of the method WriteResultChunks.
	WriteResultChunksFunc *WriterWriteResultChunksFunc
}

// NewMockWriter creates a new mock of the Writer interface. All methods
// return zero values for all results, unless overwritten.
func NewMockWriter() *MockWriter {
	return &MockWriter{
		CloseFunc: &WriterCloseFunc{
			defaultHook: func() error {
				return nil
			},
		},
		WriteDefinitionsFunc: &WriterWriteDefinitionsFunc{
			defaultHook: func(context.Context, []types.MonikerLocations) error {
				return nil
			},
		},
		WriteDocumentsFunc: &WriterWriteDocumentsFunc{
			defaultHook: func(context.Context, map[string]types.DocumentData) error {
				return nil
			},
		},
		WriteMetaFunc: &WriterWriteMetaFunc{
			defaultHook: func(context.Context, types.MetaData) error {
				return nil
			},
		},
		WriteReferencesFunc: &WriterWriteReferencesFunc{
			defaultHook: func(context.Context, []types.MonikerLocations) error {
				return nil
			},
		},
		WriteResultChunksFunc: &WriterWriteResultChunksFunc{
			defaultHook: func(context.Context, map[int]types.ResultChunkData) error {
				return nil
			},
		},
	}
}

// NewMockWriterFrom creates a new mock of the MockWriter interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockWriterFrom(i persistence.Writer) *MockWriter {
	return &MockWriter{
		CloseFunc: &WriterCloseFunc{
			defaultHook: i.Close,
		},
		WriteDefinitionsFunc: &WriterWriteDefinitionsFunc{
			defaultHook: i.WriteDefinitions,
		},
		WriteDocumentsFunc: &WriterWriteDocumentsFunc{
			defaultHook: i.WriteDocuments,
		},
		WriteMetaFunc: &WriterWriteMetaFunc{
			defaultHook: i.WriteMeta,
		},
		WriteReferencesFunc: &WriterWriteReferencesFunc{
			defaultHook: i.WriteReferences,
		},
		WriteResultChunksFunc: &WriterWriteResultChunksFunc{
			defaultHook: i.WriteResultChunks,
		},
	}
}

// WriterCloseFunc describes the behavior when the Close method of the
// parent MockWriter instance is invoked.
type WriterCloseFunc struct {
	defaultHook func() error
	hooks       []func() error
	history     []WriterCloseFuncCall
	mutex       sync.Mutex
}

// Close delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockWriter) Close() error {
	r0 := m.CloseFunc.nextHook()()
	m.CloseFunc.appendCall(WriterCloseFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Close method of the
// parent MockWriter instance is invoked and the hook queue is empty.
func (f *WriterCloseFunc) SetDefaultHook(hook func() error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Close method of the parent MockWriter instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *WriterCloseFunc) PushHook(hook func() error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *WriterCloseFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func() error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *WriterCloseFunc) PushReturn(r0 error) {
	f.PushHook(func() error {
		return r0
	})
}

func (f *WriterCloseFunc) nextHook() func() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WriterCloseFunc) appendCall(r0 WriterCloseFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WriterCloseFuncCall objects describing the
// invocations of this function.
func (f *WriterCloseFunc) History() []WriterCloseFuncCall {
	f.mutex.Lock()
	history := make([]WriterCloseFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WriterCloseFuncCall is an object that describes an invocation of method
// Close on an instance of MockWriter.
type WriterCloseFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WriterCloseFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WriterCloseFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// WriterWriteDefinitionsFunc describes the behavior when the
// WriteDefinitions method of the parent MockWriter instance is invoked.
type WriterWriteDefinitionsFunc struct {
	defaultHook func(context.Context, []types.MonikerLocations) error
	hooks       []func(context.Context, []types.MonikerLocations) error
	history     []WriterWriteDefinitionsFuncCall
	mutex       sync.Mutex
}

// WriteDefinitions delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockWriter) WriteDefinitions(v0 context.Context, v1 []types.MonikerLocations) error {
	r0 := m.WriteDefinitionsFunc.nextHook()(v0, v1)
	m.WriteDefinitionsFunc.appendCall(WriterWriteDefinitionsFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the WriteDefinitions
// method of the parent MockWriter instance is invoked and the hook queue is
// empty.
func (f *WriterWriteDefinitionsFunc) SetDefaultHook(hook func(context.Context, []types.MonikerLocations) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// WriteDefinitions method of the parent MockWriter instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *WriterWriteDefinitionsFunc) PushHook(hook func(context.Context, []types.MonikerLocations) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *WriterWriteDefinitionsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, []types.MonikerLocations) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *WriterWriteDefinitionsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []types.MonikerLocations) error {
		return r0
	})
}

func (f *WriterWriteDefinitionsFunc) nextHook() func(context.Context, []types.MonikerLocations) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WriterWriteDefinitionsFunc) appendCall(r0 WriterWriteDefinitionsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WriterWriteDefinitionsFuncCall objects
// describing the invocations of this function.
func (f *WriterWriteDefinitionsFunc) History() []WriterWriteDefinitionsFuncCall {
	f.mutex.Lock()
	history := make([]WriterWriteDefinitionsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WriterWriteDefinitionsFuncCall is an object that describes an invocation
// of method WriteDefinitions on an instance of MockWriter.
type WriterWriteDefinitionsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 []types.MonikerLocations
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WriterWriteDefinitionsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WriterWriteDefinitionsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// WriterWriteDocumentsFunc describes the behavior when the WriteDocuments
// method of the parent MockWriter instance is invoked.
type WriterWriteDocumentsFunc struct {
	defaultHook func(context.Context, map[string]types.DocumentData) error
	hooks       []func(context.Context, map[string]types.DocumentData) error
	history     []WriterWriteDocumentsFuncCall
	mutex       sync.Mutex
}

// WriteDocuments delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockWriter) WriteDocuments(v0 context.Context, v1 map[string]types.DocumentData) error {
	r0 := m.WriteDocumentsFunc.nextHook()(v0, v1)
	m.WriteDocumentsFunc.appendCall(WriterWriteDocumentsFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the WriteDocuments
// method of the parent MockWriter instance is invoked and the hook queue is
// empty.
func (f *WriterWriteDocumentsFunc) SetDefaultHook(hook func(context.Context, map[string]types.DocumentData) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// WriteDocuments method of the parent MockWriter instance inovkes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *WriterWriteDocumentsFunc) PushHook(hook func(context.Context, map[string]types.DocumentData) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *WriterWriteDocumentsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, map[string]types.DocumentData) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *WriterWriteDocumentsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, map[string]types.DocumentData) error {
		return r0
	})
}

func (f *WriterWriteDocumentsFunc) nextHook() func(context.Context, map[string]types.DocumentData) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WriterWriteDocumentsFunc) appendCall(r0 WriterWriteDocumentsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WriterWriteDocumentsFuncCall objects
// describing the invocations of this function.
func (f *WriterWriteDocumentsFunc) History() []WriterWriteDocumentsFuncCall {
	f.mutex.Lock()
	history := make([]WriterWriteDocumentsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WriterWriteDocumentsFuncCall is an object that describes an invocation of
// method WriteDocuments on an instance of MockWriter.
type WriterWriteDocumentsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 map[string]types.DocumentData
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WriterWriteDocumentsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WriterWriteDocumentsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// WriterWriteMetaFunc describes the behavior when the WriteMeta method of
// the parent MockWriter instance is invoked.
type WriterWriteMetaFunc struct {
	defaultHook func(context.Context, types.MetaData) error
	hooks       []func(context.Context, types.MetaData) error
	history     []WriterWriteMetaFuncCall
	mutex       sync.Mutex
}

// WriteMeta delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockWriter) WriteMeta(v0 context.Context, v1 types.MetaData) error {
	r0 := m.WriteMetaFunc.nextHook()(v0, v1)
	m.WriteMetaFunc.appendCall(WriterWriteMetaFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the WriteMeta method of
// the parent MockWriter instance is invoked and the hook queue is empty.
func (f *WriterWriteMetaFunc) SetDefaultHook(hook func(context.Context, types.MetaData) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// WriteMeta method of the parent MockWriter instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *WriterWriteMetaFunc) PushHook(hook func(context.Context, types.MetaData) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *WriterWriteMetaFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, types.MetaData) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *WriterWriteMetaFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, types.MetaData) error {
		return r0
	})
}

func (f *WriterWriteMetaFunc) nextHook() func(context.Context, types.MetaData) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WriterWriteMetaFunc) appendCall(r0 WriterWriteMetaFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WriterWriteMetaFuncCall objects describing
// the invocations of this function.
func (f *WriterWriteMetaFunc) History() []WriterWriteMetaFuncCall {
	f.mutex.Lock()
	history := make([]WriterWriteMetaFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WriterWriteMetaFuncCall is an object that describes an invocation of
// method WriteMeta on an instance of MockWriter.
type WriterWriteMetaFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 types.MetaData
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WriterWriteMetaFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WriterWriteMetaFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// WriterWriteReferencesFunc describes the behavior when the WriteReferences
// method of the parent MockWriter instance is invoked.
type WriterWriteReferencesFunc struct {
	defaultHook func(context.Context, []types.MonikerLocations) error
	hooks       []func(context.Context, []types.MonikerLocations) error
	history     []WriterWriteReferencesFuncCall
	mutex       sync.Mutex
}

// WriteReferences delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockWriter) WriteReferences(v0 context.Context, v1 []types.MonikerLocations) error {
	r0 := m.WriteReferencesFunc.nextHook()(v0, v1)
	m.WriteReferencesFunc.appendCall(WriterWriteReferencesFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the WriteReferences
// method of the parent MockWriter instance is invoked and the hook queue is
// empty.
func (f *WriterWriteReferencesFunc) SetDefaultHook(hook func(context.Context, []types.MonikerLocations) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// WriteReferences method of the parent MockWriter instance inovkes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *WriterWriteReferencesFunc) PushHook(hook func(context.Context, []types.MonikerLocations) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *WriterWriteReferencesFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, []types.MonikerLocations) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *WriterWriteReferencesFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []types.MonikerLocations) error {
		return r0
	})
}

func (f *WriterWriteReferencesFunc) nextHook() func(context.Context, []types.MonikerLocations) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WriterWriteReferencesFunc) appendCall(r0 WriterWriteReferencesFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WriterWriteReferencesFuncCall objects
// describing the invocations of this function.
func (f *WriterWriteReferencesFunc) History() []WriterWriteReferencesFuncCall {
	f.mutex.Lock()
	history := make([]WriterWriteReferencesFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WriterWriteReferencesFuncCall is an object that describes an invocation
// of method WriteReferences on an instance of MockWriter.
type WriterWriteReferencesFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 []types.MonikerLocations
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WriterWriteReferencesFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WriterWriteReferencesFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// WriterWriteResultChunksFunc describes the behavior when the
// WriteResultChunks method of the parent MockWriter instance is invoked.
type WriterWriteResultChunksFunc struct {
	defaultHook func(context.Context, map[int]types.ResultChunkData) error
	hooks       []func(context.Context, map[int]types.ResultChunkData) error
	history     []WriterWriteResultChunksFuncCall
	mutex       sync.Mutex
}

// WriteResultChunks delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockWriter) WriteResultChunks(v0 context.Context, v1 map[int]types.ResultChunkData) error {
	r0 := m.WriteResultChunksFunc.nextHook()(v0, v1)
	m.WriteResultChunksFunc.appendCall(WriterWriteResultChunksFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the WriteResultChunks
// method of the parent MockWriter instance is invoked and the hook queue is
// empty.
func (f *WriterWriteResultChunksFunc) SetDefaultHook(hook func(context.Context, map[int]types.ResultChunkData) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// WriteResultChunks method of the parent MockWriter instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *WriterWriteResultChunksFunc) PushHook(hook func(context.Context, map[int]types.ResultChunkData) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *WriterWriteResultChunksFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, map[int]types.ResultChunkData) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *WriterWriteResultChunksFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, map[int]types.ResultChunkData) error {
		return r0
	})
}

func (f *WriterWriteResultChunksFunc) nextHook() func(context.Context, map[int]types.ResultChunkData) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *WriterWriteResultChunksFunc) appendCall(r0 WriterWriteResultChunksFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of WriterWriteResultChunksFuncCall objects
// describing the invocations of this function.
func (f *WriterWriteResultChunksFunc) History() []WriterWriteResultChunksFuncCall {
	f.mutex.Lock()
	history := make([]WriterWriteResultChunksFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// WriterWriteResultChunksFuncCall is an object that describes an invocation
// of method WriteResultChunks on an instance of MockWriter.
type WriterWriteResultChunksFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 map[int]types.ResultChunkData
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c WriterWriteResultChunksFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c WriterWriteResultChunksFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
