package luasandbox

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	GoModules map[string]lua.LGFunction

	// LuaModules is map of require("$KEY") -> $VALUE that will be loaded
	// in the lua sandbox state. This prevents subsequent executions from
	// modifying (or peeking into) the state of any other recognizer.
	LuaModules map[string]string
}

func (s *Service) CreateSandbox(ctx context.Context, opts CreateOptions) (_ *Sandbox, err error) {
	_, _, endObservation := s.operations.createSandbox.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// Default LuaModules to our runtime files
	if opts.LuaModules == nil {
		opts.LuaModules = map[string]string{}
	}

	for k, v := range DefaultLuaModules {
		if _, ok := opts.LuaModules[k]; ok {
			return nil, errors.Newf("a Lua module with the name %q already exists", k)
		}

		opts.LuaModules[k] = v
	}

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
	for name, loader := range opts.GoModules {
		state.PreloadModule(name, loader)
	}

	// De-register global functions that could do something unwanted
	for _, name := range globalsToUnset {
		state.SetGlobal(name, lua.LNil)
	}

	// Add our own package loader so that we can do: "require('fun')"
	tbl := state.GetField(state.GetGlobal("package"), "loaders").(*lua.LTable)
	tbl.Insert(1, state.NewFunction(func(laterState *lua.LState) int {
		requireArg := laterState.Get(-1)
		switch luaModString := requireArg.(type) {
		case lua.LString:
			modString := luaModString.String()
			if _, ok := opts.LuaModules[modString]; !ok {
				break
			}

			val, err := state.LoadString(opts.LuaModules[modString])
			if err != nil {
				state.RaiseError(err.Error())
				return 0
			}

			// return the function to tell Lua that we have found
			// an appropriate lua chunk to load
			state.Push(val)
			return 1
		}

		// loaders return nil if they don't do anything
		state.Push(lua.LNil)
		return 1
	}))

	return &Sandbox{
		state:      state,
		operations: s.operations,
	}, nil
}
