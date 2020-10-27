// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package command

import (
	"context"
	"sync"
)

// MockCommandRunner is a mock implementation of the commandRunner interface
// (from the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/command)
// used for unit testing.
type MockCommandRunner struct {
	// RunCommandFunc is an instance of a mock function object controlling
	// the behavior of the method RunCommand.
	RunCommandFunc *CommandRunnerRunCommandFunc
}

// NewMockCommandRunner creates a new mock of the commandRunner interface.
// All methods return zero values for all results, unless overwritten.
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		RunCommandFunc: &CommandRunnerRunCommandFunc{
			defaultHook: func(context.Context, *Logger, command) error {
				return nil
			},
		},
	}
}

// surrogateMockCommandRunner is a copy of the commandRunner interface (from
// the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/command).
// It is redefined here as it is unexported in the source packge.
type surrogateMockCommandRunner interface {
	RunCommand(context.Context, *Logger, command) error
}

// NewMockCommandRunnerFrom creates a new mock of the MockCommandRunner
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockCommandRunnerFrom(i surrogateMockCommandRunner) *MockCommandRunner {
	return &MockCommandRunner{
		RunCommandFunc: &CommandRunnerRunCommandFunc{
			defaultHook: i.RunCommand,
		},
	}
}

// CommandRunnerRunCommandFunc describes the behavior when the RunCommand
// method of the parent MockCommandRunner instance is invoked.
type CommandRunnerRunCommandFunc struct {
	defaultHook func(context.Context, *Logger, command) error
	hooks       []func(context.Context, *Logger, command) error
	history     []CommandRunnerRunCommandFuncCall
	mutex       sync.Mutex
}

// RunCommand delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCommandRunner) RunCommand(v0 context.Context, v1 *Logger, v2 command) error {
	r0 := m.RunCommandFunc.nextHook()(v0, v1, v2)
	m.RunCommandFunc.appendCall(CommandRunnerRunCommandFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the RunCommand method of
// the parent MockCommandRunner instance is invoked and the hook queue is
// empty.
func (f *CommandRunnerRunCommandFunc) SetDefaultHook(hook func(context.Context, *Logger, command) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RunCommand method of the parent MockCommandRunner instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CommandRunnerRunCommandFunc) PushHook(hook func(context.Context, *Logger, command) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CommandRunnerRunCommandFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, *Logger, command) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CommandRunnerRunCommandFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *Logger, command) error {
		return r0
	})
}

func (f *CommandRunnerRunCommandFunc) nextHook() func(context.Context, *Logger, command) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommandRunnerRunCommandFunc) appendCall(r0 CommandRunnerRunCommandFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommandRunnerRunCommandFuncCall objects
// describing the invocations of this function.
func (f *CommandRunnerRunCommandFunc) History() []CommandRunnerRunCommandFuncCall {
	f.mutex.Lock()
	history := make([]CommandRunnerRunCommandFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommandRunnerRunCommandFuncCall is an object that describes an invocation
// of method RunCommand on an instance of MockCommandRunner.
type CommandRunnerRunCommandFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 *Logger
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 command
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommandRunnerRunCommandFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommandRunnerRunCommandFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
