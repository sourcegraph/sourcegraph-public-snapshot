// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package policies

import (
	"context"
	"sync"
	"time"

	gitserver "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	api "github.com/sourcegraph/sourcegraph/internal/api"
)

// MockGitserverClient is a mock implementation of the GitserverClient
// interface (from the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies)
// used for unit testing.
type MockGitserverClient struct {
	// CommitDateFunc is an instance of a mock function object controlling
	// the behavior of the method CommitDate.
	CommitDateFunc *GitserverClientCommitDateFunc
	// CommitsUniqueToBranchFunc is an instance of a mock function object
	// controlling the behavior of the method CommitsUniqueToBranch.
	CommitsUniqueToBranchFunc *GitserverClientCommitsUniqueToBranchFunc
	// RefDescriptionsFunc is an instance of a mock function object
	// controlling the behavior of the method RefDescriptions.
	RefDescriptionsFunc *GitserverClientRefDescriptionsFunc
	// ResolveRevisionFunc is an instance of a mock function object
	// controlling the behavior of the method ResolveRevision.
	ResolveRevisionFunc *GitserverClientResolveRevisionFunc
}

// NewMockGitserverClient creates a new mock of the GitserverClient
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockGitserverClient() *MockGitserverClient {
	return &MockGitserverClient{
		CommitDateFunc: &GitserverClientCommitDateFunc{
			defaultHook: func(context.Context, int, string) (string, time.Time, bool, error) {
				return "", time.Time{}, false, nil
			},
		},
		CommitsUniqueToBranchFunc: &GitserverClientCommitsUniqueToBranchFunc{
			defaultHook: func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error) {
				return nil, nil
			},
		},
		RefDescriptionsFunc: &GitserverClientRefDescriptionsFunc{
			defaultHook: func(context.Context, int) (map[string][]gitserver.RefDescription, error) {
				return nil, nil
			},
		},
		ResolveRevisionFunc: &GitserverClientResolveRevisionFunc{
			defaultHook: func(context.Context, int, string) (api.CommitID, error) {
				return "", nil
			},
		},
	}
}

// NewMockGitserverClientFrom creates a new mock of the MockGitserverClient
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockGitserverClientFrom(i GitserverClient) *MockGitserverClient {
	return &MockGitserverClient{
		CommitDateFunc: &GitserverClientCommitDateFunc{
			defaultHook: i.CommitDate,
		},
		CommitsUniqueToBranchFunc: &GitserverClientCommitsUniqueToBranchFunc{
			defaultHook: i.CommitsUniqueToBranch,
		},
		RefDescriptionsFunc: &GitserverClientRefDescriptionsFunc{
			defaultHook: i.RefDescriptions,
		},
		ResolveRevisionFunc: &GitserverClientResolveRevisionFunc{
			defaultHook: i.ResolveRevision,
		},
	}
}

// GitserverClientCommitDateFunc describes the behavior when the CommitDate
// method of the parent MockGitserverClient instance is invoked.
type GitserverClientCommitDateFunc struct {
	defaultHook func(context.Context, int, string) (string, time.Time, bool, error)
	hooks       []func(context.Context, int, string) (string, time.Time, bool, error)
	history     []GitserverClientCommitDateFuncCall
	mutex       sync.Mutex
}

// CommitDate delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockGitserverClient) CommitDate(v0 context.Context, v1 int, v2 string) (string, time.Time, bool, error) {
	r0, r1, r2, r3 := m.CommitDateFunc.nextHook()(v0, v1, v2)
	m.CommitDateFunc.appendCall(GitserverClientCommitDateFuncCall{v0, v1, v2, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefaultHook sets function that is called when the CommitDate method of
// the parent MockGitserverClient instance is invoked and the hook queue is
// empty.
func (f *GitserverClientCommitDateFunc) SetDefaultHook(hook func(context.Context, int, string) (string, time.Time, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CommitDate method of the parent MockGitserverClient instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *GitserverClientCommitDateFunc) PushHook(hook func(context.Context, int, string) (string, time.Time, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientCommitDateFunc) SetDefaultReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.SetDefaultHook(func(context.Context, int, string) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientCommitDateFunc) PushReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.PushHook(func(context.Context, int, string) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *GitserverClientCommitDateFunc) nextHook() func(context.Context, int, string) (string, time.Time, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitDateFunc) appendCall(r0 GitserverClientCommitDateFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientCommitDateFuncCall objects
// describing the invocations of this function.
func (f *GitserverClientCommitDateFunc) History() []GitserverClientCommitDateFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientCommitDateFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitDateFuncCall is an object that describes an
// invocation of method CommitDate on an instance of MockGitserverClient.
type GitserverClientCommitDateFuncCall struct {
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
	Result0 string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 time.Time
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 bool
	// Result3 is the value of the 4th result returned from this method
	// invocation.
	Result3 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientCommitDateFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientCommitDateFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// GitserverClientCommitsUniqueToBranchFunc describes the behavior when the
// CommitsUniqueToBranch method of the parent MockGitserverClient instance
// is invoked.
type GitserverClientCommitsUniqueToBranchFunc struct {
	defaultHook func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error)
	hooks       []func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error)
	history     []GitserverClientCommitsUniqueToBranchFuncCall
	mutex       sync.Mutex
}

// CommitsUniqueToBranch delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockGitserverClient) CommitsUniqueToBranch(v0 context.Context, v1 int, v2 string, v3 bool, v4 *time.Time) (map[string]time.Time, error) {
	r0, r1 := m.CommitsUniqueToBranchFunc.nextHook()(v0, v1, v2, v3, v4)
	m.CommitsUniqueToBranchFunc.appendCall(GitserverClientCommitsUniqueToBranchFuncCall{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// CommitsUniqueToBranch method of the parent MockGitserverClient instance
// is invoked and the hook queue is empty.
func (f *GitserverClientCommitsUniqueToBranchFunc) SetDefaultHook(hook func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CommitsUniqueToBranch method of the parent MockGitserverClient instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *GitserverClientCommitsUniqueToBranchFunc) PushHook(hook func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientCommitsUniqueToBranchFunc) SetDefaultReturn(r0 map[string]time.Time, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientCommitsUniqueToBranchFunc) PushReturn(r0 map[string]time.Time, r1 error) {
	f.PushHook(func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error) {
		return r0, r1
	})
}

func (f *GitserverClientCommitsUniqueToBranchFunc) nextHook() func(context.Context, int, string, bool, *time.Time) (map[string]time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientCommitsUniqueToBranchFunc) appendCall(r0 GitserverClientCommitsUniqueToBranchFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// GitserverClientCommitsUniqueToBranchFuncCall objects describing the
// invocations of this function.
func (f *GitserverClientCommitsUniqueToBranchFunc) History() []GitserverClientCommitsUniqueToBranchFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientCommitsUniqueToBranchFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientCommitsUniqueToBranchFuncCall is an object that describes
// an invocation of method CommitsUniqueToBranch on an instance of
// MockGitserverClient.
type GitserverClientCommitsUniqueToBranchFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 bool
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 *time.Time
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string]time.Time
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientCommitsUniqueToBranchFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientCommitsUniqueToBranchFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// GitserverClientRefDescriptionsFunc describes the behavior when the
// RefDescriptions method of the parent MockGitserverClient instance is
// invoked.
type GitserverClientRefDescriptionsFunc struct {
	defaultHook func(context.Context, int) (map[string][]gitserver.RefDescription, error)
	hooks       []func(context.Context, int) (map[string][]gitserver.RefDescription, error)
	history     []GitserverClientRefDescriptionsFuncCall
	mutex       sync.Mutex
}

// RefDescriptions delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockGitserverClient) RefDescriptions(v0 context.Context, v1 int) (map[string][]gitserver.RefDescription, error) {
	r0, r1 := m.RefDescriptionsFunc.nextHook()(v0, v1)
	m.RefDescriptionsFunc.appendCall(GitserverClientRefDescriptionsFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the RefDescriptions
// method of the parent MockGitserverClient instance is invoked and the hook
// queue is empty.
func (f *GitserverClientRefDescriptionsFunc) SetDefaultHook(hook func(context.Context, int) (map[string][]gitserver.RefDescription, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RefDescriptions method of the parent MockGitserverClient instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *GitserverClientRefDescriptionsFunc) PushHook(hook func(context.Context, int) (map[string][]gitserver.RefDescription, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientRefDescriptionsFunc) SetDefaultReturn(r0 map[string][]gitserver.RefDescription, r1 error) {
	f.SetDefaultHook(func(context.Context, int) (map[string][]gitserver.RefDescription, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientRefDescriptionsFunc) PushReturn(r0 map[string][]gitserver.RefDescription, r1 error) {
	f.PushHook(func(context.Context, int) (map[string][]gitserver.RefDescription, error) {
		return r0, r1
	})
}

func (f *GitserverClientRefDescriptionsFunc) nextHook() func(context.Context, int) (map[string][]gitserver.RefDescription, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientRefDescriptionsFunc) appendCall(r0 GitserverClientRefDescriptionsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientRefDescriptionsFuncCall
// objects describing the invocations of this function.
func (f *GitserverClientRefDescriptionsFunc) History() []GitserverClientRefDescriptionsFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientRefDescriptionsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientRefDescriptionsFuncCall is an object that describes an
// invocation of method RefDescriptions on an instance of
// MockGitserverClient.
type GitserverClientRefDescriptionsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string][]gitserver.RefDescription
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientRefDescriptionsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientRefDescriptionsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// GitserverClientResolveRevisionFunc describes the behavior when the
// ResolveRevision method of the parent MockGitserverClient instance is
// invoked.
type GitserverClientResolveRevisionFunc struct {
	defaultHook func(context.Context, int, string) (api.CommitID, error)
	hooks       []func(context.Context, int, string) (api.CommitID, error)
	history     []GitserverClientResolveRevisionFuncCall
	mutex       sync.Mutex
}

// ResolveRevision delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockGitserverClient) ResolveRevision(v0 context.Context, v1 int, v2 string) (api.CommitID, error) {
	r0, r1 := m.ResolveRevisionFunc.nextHook()(v0, v1, v2)
	m.ResolveRevisionFunc.appendCall(GitserverClientResolveRevisionFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the ResolveRevision
// method of the parent MockGitserverClient instance is invoked and the hook
// queue is empty.
func (f *GitserverClientResolveRevisionFunc) SetDefaultHook(hook func(context.Context, int, string) (api.CommitID, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// ResolveRevision method of the parent MockGitserverClient instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *GitserverClientResolveRevisionFunc) PushHook(hook func(context.Context, int, string) (api.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *GitserverClientResolveRevisionFunc) SetDefaultReturn(r0 api.CommitID, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string) (api.CommitID, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *GitserverClientResolveRevisionFunc) PushReturn(r0 api.CommitID, r1 error) {
	f.PushHook(func(context.Context, int, string) (api.CommitID, error) {
		return r0, r1
	})
}

func (f *GitserverClientResolveRevisionFunc) nextHook() func(context.Context, int, string) (api.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitserverClientResolveRevisionFunc) appendCall(r0 GitserverClientResolveRevisionFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of GitserverClientResolveRevisionFuncCall
// objects describing the invocations of this function.
func (f *GitserverClientResolveRevisionFunc) History() []GitserverClientResolveRevisionFuncCall {
	f.mutex.Lock()
	history := make([]GitserverClientResolveRevisionFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitserverClientResolveRevisionFuncCall is an object that describes an
// invocation of method ResolveRevision on an instance of
// MockGitserverClient.
type GitserverClientResolveRevisionFuncCall struct {
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
	Result0 api.CommitID
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c GitserverClientResolveRevisionFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c GitserverClientResolveRevisionFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
