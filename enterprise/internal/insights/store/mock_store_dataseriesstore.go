// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package store

import (
	"context"
	"sync"

	types "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

// MockDataSeriesStore is a mock implementation of the DataSeriesStore
// interface (from the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store)
// used for unit testing.
type MockDataSeriesStore struct {
	// GetDataSeriesFunc is an instance of a mock function object
	// controlling the behavior of the method GetDataSeries.
	GetDataSeriesFunc *DataSeriesStoreGetDataSeriesFunc
	// SetSeriesEnabledFunc is an instance of a mock function object
	// controlling the behavior of the method SetSeriesEnabled.
	SetSeriesEnabledFunc *DataSeriesStoreSetSeriesEnabledFunc
	// StampBackfillFunc is an instance of a mock function object
	// controlling the behavior of the method StampBackfill.
	StampBackfillFunc *DataSeriesStoreStampBackfillFunc
	// StampRecordingFunc is an instance of a mock function object
	// controlling the behavior of the method StampRecording.
	StampRecordingFunc *DataSeriesStoreStampRecordingFunc
	// StampSnapshotFunc is an instance of a mock function object
	// controlling the behavior of the method StampSnapshot.
	StampSnapshotFunc *DataSeriesStoreStampSnapshotFunc
}

// NewMockDataSeriesStore creates a new mock of the DataSeriesStore
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockDataSeriesStore() *MockDataSeriesStore {
	return &MockDataSeriesStore{
		GetDataSeriesFunc: &DataSeriesStoreGetDataSeriesFunc{
			defaultHook: func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error) {
				return nil, nil
			},
		},
		SetSeriesEnabledFunc: &DataSeriesStoreSetSeriesEnabledFunc{
			defaultHook: func(context.Context, string, bool) error {
				return nil
			},
		},
		StampBackfillFunc: &DataSeriesStoreStampBackfillFunc{
			defaultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				return types.InsightSeries{}, nil
			},
		},
		StampRecordingFunc: &DataSeriesStoreStampRecordingFunc{
			defaultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				return types.InsightSeries{}, nil
			},
		},
		StampSnapshotFunc: &DataSeriesStoreStampSnapshotFunc{
			defaultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				return types.InsightSeries{}, nil
			},
		},
	}
}

// NewStrictMockDataSeriesStore creates a new mock of the DataSeriesStore
// interface. All methods panic on invocation, unless overwritten.
func NewStrictMockDataSeriesStore() *MockDataSeriesStore {
	return &MockDataSeriesStore{
		GetDataSeriesFunc: &DataSeriesStoreGetDataSeriesFunc{
			defaultHook: func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error) {
				panic("unexpected invocation of MockDataSeriesStore.GetDataSeries")
			},
		},
		SetSeriesEnabledFunc: &DataSeriesStoreSetSeriesEnabledFunc{
			defaultHook: func(context.Context, string, bool) error {
				panic("unexpected invocation of MockDataSeriesStore.SetSeriesEnabled")
			},
		},
		StampBackfillFunc: &DataSeriesStoreStampBackfillFunc{
			defaultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				panic("unexpected invocation of MockDataSeriesStore.StampBackfill")
			},
		},
		StampRecordingFunc: &DataSeriesStoreStampRecordingFunc{
			defaultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				panic("unexpected invocation of MockDataSeriesStore.StampRecording")
			},
		},
		StampSnapshotFunc: &DataSeriesStoreStampSnapshotFunc{
			defaultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				panic("unexpected invocation of MockDataSeriesStore.StampSnapshot")
			},
		},
	}
}

// NewMockDataSeriesStoreFrom creates a new mock of the MockDataSeriesStore
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockDataSeriesStoreFrom(i DataSeriesStore) *MockDataSeriesStore {
	return &MockDataSeriesStore{
		GetDataSeriesFunc: &DataSeriesStoreGetDataSeriesFunc{
			defaultHook: i.GetDataSeries,
		},
		SetSeriesEnabledFunc: &DataSeriesStoreSetSeriesEnabledFunc{
			defaultHook: i.SetSeriesEnabled,
		},
		StampBackfillFunc: &DataSeriesStoreStampBackfillFunc{
			defaultHook: i.StampBackfill,
		},
		StampRecordingFunc: &DataSeriesStoreStampRecordingFunc{
			defaultHook: i.StampRecording,
		},
		StampSnapshotFunc: &DataSeriesStoreStampSnapshotFunc{
			defaultHook: i.StampSnapshot,
		},
	}
}

// DataSeriesStoreGetDataSeriesFunc describes the behavior when the
// GetDataSeries method of the parent MockDataSeriesStore instance is
// invoked.
type DataSeriesStoreGetDataSeriesFunc struct {
	defaultHook func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error)
	hooks       []func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error)
	history     []DataSeriesStoreGetDataSeriesFuncCall
	mutex       sync.Mutex
}

// GetDataSeries delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockDataSeriesStore) GetDataSeries(v0 context.Context, v1 GetDataSeriesArgs) ([]types.InsightSeries, error) {
	r0, r1 := m.GetDataSeriesFunc.nextHook()(v0, v1)
	m.GetDataSeriesFunc.appendCall(DataSeriesStoreGetDataSeriesFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetDataSeries method
// of the parent MockDataSeriesStore instance is invoked and the hook queue
// is empty.
func (f *DataSeriesStoreGetDataSeriesFunc) SetDefaultHook(hook func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetDataSeries method of the parent MockDataSeriesStore instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *DataSeriesStoreGetDataSeriesFunc) PushHook(hook func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DataSeriesStoreGetDataSeriesFunc) SetDefaultReturn(r0 []types.InsightSeries, r1 error) {
	f.SetDefaultHook(func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DataSeriesStoreGetDataSeriesFunc) PushReturn(r0 []types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DataSeriesStoreGetDataSeriesFunc) nextHook() func(context.Context, GetDataSeriesArgs) ([]types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DataSeriesStoreGetDataSeriesFunc) appendCall(r0 DataSeriesStoreGetDataSeriesFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DataSeriesStoreGetDataSeriesFuncCall
// objects describing the invocations of this function.
func (f *DataSeriesStoreGetDataSeriesFunc) History() []DataSeriesStoreGetDataSeriesFuncCall {
	f.mutex.Lock()
	history := make([]DataSeriesStoreGetDataSeriesFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DataSeriesStoreGetDataSeriesFuncCall is an object that describes an
// invocation of method GetDataSeries on an instance of MockDataSeriesStore.
type DataSeriesStoreGetDataSeriesFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 GetDataSeriesArgs
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []types.InsightSeries
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DataSeriesStoreGetDataSeriesFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DataSeriesStoreGetDataSeriesFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// DataSeriesStoreSetSeriesEnabledFunc describes the behavior when the
// SetSeriesEnabled method of the parent MockDataSeriesStore instance is
// invoked.
type DataSeriesStoreSetSeriesEnabledFunc struct {
	defaultHook func(context.Context, string, bool) error
	hooks       []func(context.Context, string, bool) error
	history     []DataSeriesStoreSetSeriesEnabledFuncCall
	mutex       sync.Mutex
}

// SetSeriesEnabled delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockDataSeriesStore) SetSeriesEnabled(v0 context.Context, v1 string, v2 bool) error {
	r0 := m.SetSeriesEnabledFunc.nextHook()(v0, v1, v2)
	m.SetSeriesEnabledFunc.appendCall(DataSeriesStoreSetSeriesEnabledFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the SetSeriesEnabled
// method of the parent MockDataSeriesStore instance is invoked and the hook
// queue is empty.
func (f *DataSeriesStoreSetSeriesEnabledFunc) SetDefaultHook(hook func(context.Context, string, bool) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SetSeriesEnabled method of the parent MockDataSeriesStore instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *DataSeriesStoreSetSeriesEnabledFunc) PushHook(hook func(context.Context, string, bool) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DataSeriesStoreSetSeriesEnabledFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, string, bool) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DataSeriesStoreSetSeriesEnabledFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, bool) error {
		return r0
	})
}

func (f *DataSeriesStoreSetSeriesEnabledFunc) nextHook() func(context.Context, string, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DataSeriesStoreSetSeriesEnabledFunc) appendCall(r0 DataSeriesStoreSetSeriesEnabledFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DataSeriesStoreSetSeriesEnabledFuncCall
// objects describing the invocations of this function.
func (f *DataSeriesStoreSetSeriesEnabledFunc) History() []DataSeriesStoreSetSeriesEnabledFuncCall {
	f.mutex.Lock()
	history := make([]DataSeriesStoreSetSeriesEnabledFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DataSeriesStoreSetSeriesEnabledFuncCall is an object that describes an
// invocation of method SetSeriesEnabled on an instance of
// MockDataSeriesStore.
type DataSeriesStoreSetSeriesEnabledFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 bool
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DataSeriesStoreSetSeriesEnabledFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DataSeriesStoreSetSeriesEnabledFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// DataSeriesStoreStampBackfillFunc describes the behavior when the
// StampBackfill method of the parent MockDataSeriesStore instance is
// invoked.
type DataSeriesStoreStampBackfillFunc struct {
	defaultHook func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	hooks       []func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	history     []DataSeriesStoreStampBackfillFuncCall
	mutex       sync.Mutex
}

// StampBackfill delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockDataSeriesStore) StampBackfill(v0 context.Context, v1 types.InsightSeries) (types.InsightSeries, error) {
	r0, r1 := m.StampBackfillFunc.nextHook()(v0, v1)
	m.StampBackfillFunc.appendCall(DataSeriesStoreStampBackfillFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the StampBackfill method
// of the parent MockDataSeriesStore instance is invoked and the hook queue
// is empty.
func (f *DataSeriesStoreStampBackfillFunc) SetDefaultHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// StampBackfill method of the parent MockDataSeriesStore instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *DataSeriesStoreStampBackfillFunc) PushHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DataSeriesStoreStampBackfillFunc) SetDefaultReturn(r0 types.InsightSeries, r1 error) {
	f.SetDefaultHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DataSeriesStoreStampBackfillFunc) PushReturn(r0 types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DataSeriesStoreStampBackfillFunc) nextHook() func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DataSeriesStoreStampBackfillFunc) appendCall(r0 DataSeriesStoreStampBackfillFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DataSeriesStoreStampBackfillFuncCall
// objects describing the invocations of this function.
func (f *DataSeriesStoreStampBackfillFunc) History() []DataSeriesStoreStampBackfillFuncCall {
	f.mutex.Lock()
	history := make([]DataSeriesStoreStampBackfillFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DataSeriesStoreStampBackfillFuncCall is an object that describes an
// invocation of method StampBackfill on an instance of MockDataSeriesStore.
type DataSeriesStoreStampBackfillFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 types.InsightSeries
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 types.InsightSeries
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DataSeriesStoreStampBackfillFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DataSeriesStoreStampBackfillFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// DataSeriesStoreStampRecordingFunc describes the behavior when the
// StampRecording method of the parent MockDataSeriesStore instance is
// invoked.
type DataSeriesStoreStampRecordingFunc struct {
	defaultHook func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	hooks       []func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	history     []DataSeriesStoreStampRecordingFuncCall
	mutex       sync.Mutex
}

// StampRecording delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockDataSeriesStore) StampRecording(v0 context.Context, v1 types.InsightSeries) (types.InsightSeries, error) {
	r0, r1 := m.StampRecordingFunc.nextHook()(v0, v1)
	m.StampRecordingFunc.appendCall(DataSeriesStoreStampRecordingFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the StampRecording
// method of the parent MockDataSeriesStore instance is invoked and the hook
// queue is empty.
func (f *DataSeriesStoreStampRecordingFunc) SetDefaultHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// StampRecording method of the parent MockDataSeriesStore instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *DataSeriesStoreStampRecordingFunc) PushHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DataSeriesStoreStampRecordingFunc) SetDefaultReturn(r0 types.InsightSeries, r1 error) {
	f.SetDefaultHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DataSeriesStoreStampRecordingFunc) PushReturn(r0 types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DataSeriesStoreStampRecordingFunc) nextHook() func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DataSeriesStoreStampRecordingFunc) appendCall(r0 DataSeriesStoreStampRecordingFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DataSeriesStoreStampRecordingFuncCall
// objects describing the invocations of this function.
func (f *DataSeriesStoreStampRecordingFunc) History() []DataSeriesStoreStampRecordingFuncCall {
	f.mutex.Lock()
	history := make([]DataSeriesStoreStampRecordingFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DataSeriesStoreStampRecordingFuncCall is an object that describes an
// invocation of method StampRecording on an instance of
// MockDataSeriesStore.
type DataSeriesStoreStampRecordingFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 types.InsightSeries
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 types.InsightSeries
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DataSeriesStoreStampRecordingFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DataSeriesStoreStampRecordingFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// DataSeriesStoreStampSnapshotFunc describes the behavior when the
// StampSnapshot method of the parent MockDataSeriesStore instance is
// invoked.
type DataSeriesStoreStampSnapshotFunc struct {
	defaultHook func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	hooks       []func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	history     []DataSeriesStoreStampSnapshotFuncCall
	mutex       sync.Mutex
}

// StampSnapshot delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockDataSeriesStore) StampSnapshot(v0 context.Context, v1 types.InsightSeries) (types.InsightSeries, error) {
	r0, r1 := m.StampSnapshotFunc.nextHook()(v0, v1)
	m.StampSnapshotFunc.appendCall(DataSeriesStoreStampSnapshotFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the StampSnapshot method
// of the parent MockDataSeriesStore instance is invoked and the hook queue
// is empty.
func (f *DataSeriesStoreStampSnapshotFunc) SetDefaultHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// StampSnapshot method of the parent MockDataSeriesStore instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *DataSeriesStoreStampSnapshotFunc) PushHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *DataSeriesStoreStampSnapshotFunc) SetDefaultReturn(r0 types.InsightSeries, r1 error) {
	f.SetDefaultHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *DataSeriesStoreStampSnapshotFunc) PushReturn(r0 types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DataSeriesStoreStampSnapshotFunc) nextHook() func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DataSeriesStoreStampSnapshotFunc) appendCall(r0 DataSeriesStoreStampSnapshotFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of DataSeriesStoreStampSnapshotFuncCall
// objects describing the invocations of this function.
func (f *DataSeriesStoreStampSnapshotFunc) History() []DataSeriesStoreStampSnapshotFuncCall {
	f.mutex.Lock()
	history := make([]DataSeriesStoreStampSnapshotFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DataSeriesStoreStampSnapshotFuncCall is an object that describes an
// invocation of method StampSnapshot on an instance of MockDataSeriesStore.
type DataSeriesStoreStampSnapshotFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 types.InsightSeries
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 types.InsightSeries
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c DataSeriesStoreStampSnapshotFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c DataSeriesStoreStampSnapshotFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
