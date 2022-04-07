package luasandbox

import (
	"context"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	operations *operations
}

func newService(
	obserationContext *observation.Context,
) *Service {
	return &Service{
		operations: newOperations(obserationContext),
	}
}

type RunOptions struct {
	Source  string
	Modules map[string]lua.LGFunction
	Timeout time.Duration
}

const DefaultTimeout = time.Millisecond * 200

func (s *Service) Run(ctx context.Context, opts RunOptions) (err error) {
	ctx, endObservation := s.operations.run.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// Configure execution timeout
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Create a new Lua virtual machine
	state := s.creatNewStateWithModules(ctx, opts.Modules)
	defer state.Close()

	// Execute supplied program
	if err := state.DoString(opts.Source); err != nil {
		return err
	}

	return nil
}

// Modified verison of the libs available from gopher-lua@linit.go:luaLibs
var builtinLibs = []struct {
	libName string
	libFunc lua.LGFunction
}{
	{lua.BaseLibName, lua.OpenBase},
	{lua.ChannelLibName, lua.OpenChannel},
	{lua.CoroutineLibName, lua.OpenCoroutine},
	{lua.DebugLibName, lua.OpenDebug},
	{lua.LoadLibName, lua.OpenPackage},
	{lua.MathLibName, lua.OpenMath},
	{lua.StringLibName, lua.OpenString},
	{lua.TabLibName, lua.OpenTable},

	// Explicitly omitted
	// {lua.IoLibName,lua. OpenIo},
	// {lua.OsLibName, lua.OpenOs},
}

func (s *Service) creatNewStateWithModules(ctx context.Context, modules map[string]lua.LGFunction) *lua.LState {
	state := lua.NewState(lua.Options{
		// Do not open libraries implicitly
		SkipOpenLibs: true,
	})

	for _, lib := range builtinLibs {
		// Load libraries explicitly
		state.Push(state.NewFunction(lib.libFunc))
		state.Push(lua.LString(lib.libName))
		state.Call(1, 0)
	}

	// Preload caller-supplied modules
	for name, loader := range modules {
		state.PreloadModule(name, loader)
	}

	state.SetContext(ctx)
	return state
}
