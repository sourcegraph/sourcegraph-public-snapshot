package luasandbox

import (
	"context"

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

type CreateOptions struct {
	Modules map[string]lua.LGFunction
}

func (s *Service) CreateSandbox(ctx context.Context, opts CreateOptions) (_ *Sandbox, err error) {
	_, _, endObservation := s.operations.createSandbox.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

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
	for name, loader := range opts.Modules {
		state.PreloadModule(name, loader)
	}

	// De-register global functions that could do something unwanted
	for _, name := range globalsToUnset {
		state.SetGlobal(name, lua.LNil)
	}

	return &Sandbox{
		state:      state,
		operations: s.operations,
	}, nil
}
