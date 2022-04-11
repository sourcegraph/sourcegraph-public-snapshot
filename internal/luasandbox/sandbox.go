package luasandbox

import (
	"context"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Sandbox struct {
	// note: operations around vm state are kept isolated in
	// the run function, which ensures mutual access with the
	// mutex.
	state *lua.LState
	m     sync.Mutex

	operations *operations
}

// Close releases resources occupied by the underlying Lua VM.
// No calls to the sandbox should be made after closing it.
func (s *Sandbox) Close() {
	s.state.Close()
}

// RunScript runs the given Lua script text in the sandbox.
func (s *Sandbox) RunScript(ctx context.Context, opts RunOptions, script string) (retValue lua.LValue, err error) {
	ctx, endObservation := s.operations.runScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	f := func() error {
		if err := s.state.DoString(script); err != nil {
			return err
		}

		retValue = s.state.Get(lua.MultRet)
		return nil
	}
	err = s.run(ctx, opts, f)
	return
}

// Call invokes the given function bound to this sandbox within the sandbox.
func (s *Sandbox) Call(ctx context.Context, opts RunOptions, luaFunction *lua.LFunction, args ...interface{}) (retValue lua.LValue, err error) {
	ctx, endObservation := s.operations.call.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	f := func() error {
		s.state.Push(luaFunction)
		for _, arg := range args {
			s.state.Push(luar.New(s.state, arg))
		}

		if err := s.state.PCall(len(args), lua.MultRet, nil); err != nil {
			return err
		}

		retValue = s.state.Get(lua.MultRet)
		return nil
	}
	err = s.run(ctx, opts, f)
	return
}

type RunOptions struct {
	Timeout time.Duration
}

const DefaultTimeout = time.Millisecond * 200

func (s *Sandbox) run(ctx context.Context, opts RunOptions, f func() error) (err error) {
	s.m.Lock()
	defer s.m.Unlock()

	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	s.state.SetContext(ctx)
	defer s.state.RemoveContext()

	return f()
}
