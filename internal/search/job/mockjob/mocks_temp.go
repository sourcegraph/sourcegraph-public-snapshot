// Code generated by go-mockgen 1.3.3; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package mockjob

import (
	"context"
	"sync"

	log "github.com/opentracing/opentracing-go/log"
	search "github.com/sourcegraph/sourcegraph/internal/search"
	job "github.com/sourcegraph/sourcegraph/internal/search/job"
	streaming "github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// MockJob is a mock implementation of the Job interface (from the package
// github.com/sourcegraph/sourcegraph/internal/search/job) used for unit
// testing.
type MockJob struct {
	// ChildrenFunc is an instance of a mock function object controlling the
	// behavior of the method Children.
	ChildrenFunc *JobChildrenFunc
	// FieldsFunc is an instance of a mock function object controlling the
	// behavior of the method Fields.
	FieldsFunc *JobFieldsFunc
	// MapChildrenFunc is an instance of a mock function object controlling
	// the behavior of the method MapChildren.
	MapChildrenFunc *JobMapChildrenFunc
	// NameFunc is an instance of a mock function object controlling the
	// behavior of the method Name.
	NameFunc *JobNameFunc
	// RunFunc is an instance of a mock function object controlling the
	// behavior of the method Run.
	RunFunc *JobRunFunc
}

// NewMockJob creates a new mock of the Job interface. All methods return
// zero values for all results, unless overwritten.
func NewMockJob() *MockJob {
	return &MockJob{
		ChildrenFunc: &JobChildrenFunc{
			defaultHook: func() (r0 []job.Describer) {
				return
			},
		},
		FieldsFunc: &JobFieldsFunc{
			defaultHook: func(job.Verbosity) (r0 []log.Field) {
				return
			},
		},
		MapChildrenFunc: &JobMapChildrenFunc{
			defaultHook: func(job.MapFunc) (r0 job.Job) {
				return
			},
		},
		NameFunc: &JobNameFunc{
			defaultHook: func() (r0 string) {
				return
			},
		},
		RunFunc: &JobRunFunc{
			defaultHook: func(context.Context, job.RuntimeClients, streaming.Sender) (r0 *search.Alert, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockJob creates a new mock of the Job interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockJob() *MockJob {
	return &MockJob{
		ChildrenFunc: &JobChildrenFunc{
			defaultHook: func() []job.Describer {
				panic("unexpected invocation of MockJob.Children")
			},
		},
		FieldsFunc: &JobFieldsFunc{
			defaultHook: func(job.Verbosity) []log.Field {
				panic("unexpected invocation of MockJob.Fields")
			},
		},
		MapChildrenFunc: &JobMapChildrenFunc{
			defaultHook: func(job.MapFunc) job.Job {
				panic("unexpected invocation of MockJob.MapChildren")
			},
		},
		NameFunc: &JobNameFunc{
			defaultHook: func() string {
				panic("unexpected invocation of MockJob.Name")
			},
		},
		RunFunc: &JobRunFunc{
			defaultHook: func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error) {
				panic("unexpected invocation of MockJob.Run")
			},
		},
	}
}

// NewMockJobFrom creates a new mock of the MockJob interface. All methods
// delegate to the given implementation, unless overwritten.
func NewMockJobFrom(i job.Job) *MockJob {
	return &MockJob{
		ChildrenFunc: &JobChildrenFunc{
			defaultHook: i.Children,
		},
		FieldsFunc: &JobFieldsFunc{
			defaultHook: i.Fields,
		},
		MapChildrenFunc: &JobMapChildrenFunc{
			defaultHook: i.MapChildren,
		},
		NameFunc: &JobNameFunc{
			defaultHook: i.Name,
		},
		RunFunc: &JobRunFunc{
			defaultHook: i.Run,
		},
	}
}

// JobChildrenFunc describes the behavior when the Children method of the
// parent MockJob instance is invoked.
type JobChildrenFunc struct {
	defaultHook func() []job.Describer
	hooks       []func() []job.Describer
	history     []JobChildrenFuncCall
	mutex       sync.Mutex
}

// Children delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockJob) Children() []job.Describer {
	r0 := m.ChildrenFunc.nextHook()()
	m.ChildrenFunc.appendCall(JobChildrenFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Children method of
// the parent MockJob instance is invoked and the hook queue is empty.
func (f *JobChildrenFunc) SetDefaultHook(hook func() []job.Describer) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Children method of the parent MockJob instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *JobChildrenFunc) PushHook(hook func() []job.Describer) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *JobChildrenFunc) SetDefaultReturn(r0 []job.Describer) {
	f.SetDefaultHook(func() []job.Describer {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *JobChildrenFunc) PushReturn(r0 []job.Describer) {
	f.PushHook(func() []job.Describer {
		return r0
	})
}

func (f *JobChildrenFunc) nextHook() func() []job.Describer {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobChildrenFunc) appendCall(r0 JobChildrenFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of JobChildrenFuncCall objects describing the
// invocations of this function.
func (f *JobChildrenFunc) History() []JobChildrenFuncCall {
	f.mutex.Lock()
	history := make([]JobChildrenFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobChildrenFuncCall is an object that describes an invocation of method
// Children on an instance of MockJob.
type JobChildrenFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []job.Describer
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c JobChildrenFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c JobChildrenFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// JobFieldsFunc describes the behavior when the Fields method of the parent
// MockJob instance is invoked.
type JobFieldsFunc struct {
	defaultHook func(job.Verbosity) []log.Field
	hooks       []func(job.Verbosity) []log.Field
	history     []JobFieldsFuncCall
	mutex       sync.Mutex
}

// Fields delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockJob) Fields(v0 job.Verbosity) []log.Field {
	r0 := m.FieldsFunc.nextHook()(v0)
	m.FieldsFunc.appendCall(JobFieldsFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Fields method of the
// parent MockJob instance is invoked and the hook queue is empty.
func (f *JobFieldsFunc) SetDefaultHook(hook func(job.Verbosity) []log.Field) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Fields method of the parent MockJob instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *JobFieldsFunc) PushHook(hook func(job.Verbosity) []log.Field) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *JobFieldsFunc) SetDefaultReturn(r0 []log.Field) {
	f.SetDefaultHook(func(job.Verbosity) []log.Field {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *JobFieldsFunc) PushReturn(r0 []log.Field) {
	f.PushHook(func(job.Verbosity) []log.Field {
		return r0
	})
}

func (f *JobFieldsFunc) nextHook() func(job.Verbosity) []log.Field {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobFieldsFunc) appendCall(r0 JobFieldsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of JobFieldsFuncCall objects describing the
// invocations of this function.
func (f *JobFieldsFunc) History() []JobFieldsFuncCall {
	f.mutex.Lock()
	history := make([]JobFieldsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobFieldsFuncCall is an object that describes an invocation of method
// Fields on an instance of MockJob.
type JobFieldsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 job.Verbosity
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []log.Field
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c JobFieldsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c JobFieldsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// JobMapChildrenFunc describes the behavior when the MapChildren method of
// the parent MockJob instance is invoked.
type JobMapChildrenFunc struct {
	defaultHook func(job.MapFunc) job.Job
	hooks       []func(job.MapFunc) job.Job
	history     []JobMapChildrenFuncCall
	mutex       sync.Mutex
}

// MapChildren delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockJob) MapChildren(v0 job.MapFunc) job.Job {
	r0 := m.MapChildrenFunc.nextHook()(v0)
	m.MapChildrenFunc.appendCall(JobMapChildrenFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the MapChildren method
// of the parent MockJob instance is invoked and the hook queue is empty.
func (f *JobMapChildrenFunc) SetDefaultHook(hook func(job.MapFunc) job.Job) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// MapChildren method of the parent MockJob instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *JobMapChildrenFunc) PushHook(hook func(job.MapFunc) job.Job) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *JobMapChildrenFunc) SetDefaultReturn(r0 job.Job) {
	f.SetDefaultHook(func(job.MapFunc) job.Job {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *JobMapChildrenFunc) PushReturn(r0 job.Job) {
	f.PushHook(func(job.MapFunc) job.Job {
		return r0
	})
}

func (f *JobMapChildrenFunc) nextHook() func(job.MapFunc) job.Job {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobMapChildrenFunc) appendCall(r0 JobMapChildrenFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of JobMapChildrenFuncCall objects describing
// the invocations of this function.
func (f *JobMapChildrenFunc) History() []JobMapChildrenFuncCall {
	f.mutex.Lock()
	history := make([]JobMapChildrenFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobMapChildrenFuncCall is an object that describes an invocation of
// method MapChildren on an instance of MockJob.
type JobMapChildrenFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 job.MapFunc
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 job.Job
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c JobMapChildrenFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c JobMapChildrenFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// JobNameFunc describes the behavior when the Name method of the parent
// MockJob instance is invoked.
type JobNameFunc struct {
	defaultHook func() string
	hooks       []func() string
	history     []JobNameFuncCall
	mutex       sync.Mutex
}

// Name delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockJob) Name() string {
	r0 := m.NameFunc.nextHook()()
	m.NameFunc.appendCall(JobNameFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Name method of the
// parent MockJob instance is invoked and the hook queue is empty.
func (f *JobNameFunc) SetDefaultHook(hook func() string) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Name method of the parent MockJob instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *JobNameFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *JobNameFunc) SetDefaultReturn(r0 string) {
	f.SetDefaultHook(func() string {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *JobNameFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *JobNameFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobNameFunc) appendCall(r0 JobNameFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of JobNameFuncCall objects describing the
// invocations of this function.
func (f *JobNameFunc) History() []JobNameFuncCall {
	f.mutex.Lock()
	history := make([]JobNameFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobNameFuncCall is an object that describes an invocation of method Name
// on an instance of MockJob.
type JobNameFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c JobNameFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c JobNameFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// JobRunFunc describes the behavior when the Run method of the parent
// MockJob instance is invoked.
type JobRunFunc struct {
	defaultHook func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error)
	hooks       []func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error)
	history     []JobRunFuncCall
	mutex       sync.Mutex
}

// Run delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockJob) Run(v0 context.Context, v1 job.RuntimeClients, v2 streaming.Sender) (*search.Alert, error) {
	r0, r1 := m.RunFunc.nextHook()(v0, v1, v2)
	m.RunFunc.appendCall(JobRunFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Run method of the
// parent MockJob instance is invoked and the hook queue is empty.
func (f *JobRunFunc) SetDefaultHook(hook func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Run method of the parent MockJob instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *JobRunFunc) PushHook(hook func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *JobRunFunc) SetDefaultReturn(r0 *search.Alert, r1 error) {
	f.SetDefaultHook(func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *JobRunFunc) PushReturn(r0 *search.Alert, r1 error) {
	f.PushHook(func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error) {
		return r0, r1
	})
}

func (f *JobRunFunc) nextHook() func(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobRunFunc) appendCall(r0 JobRunFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of JobRunFuncCall objects describing the
// invocations of this function.
func (f *JobRunFunc) History() []JobRunFuncCall {
	f.mutex.Lock()
	history := make([]JobRunFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobRunFuncCall is an object that describes an invocation of method Run on
// an instance of MockJob.
type JobRunFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 job.RuntimeClients
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 streaming.Sender
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *search.Alert
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c JobRunFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c JobRunFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
