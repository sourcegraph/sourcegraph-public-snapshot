package luasandbox

import (
	"context"
	"embed"
	"io"
	"path/filepath"
	"strings"

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
	GoModules map[string]lua.LGFunction

	// LuaModules is map of require("$KEY") -> contents
	// that will be loaded in the lua sandbox state.
	//
	// This prevents subsequent executions from modifying
	// the state (or peeking into the state) of any of the
	// other recognizers
	LuaModules map[string]string
}

//go:embed lua/*
var LuaRuntime embed.FS

var luaRuntimeContents = map[string]string{}

func init() {
	files, err := LuaRuntime.ReadDir("lua")
	if err != nil {
		panic("sqs? more like sos!")
	}

	for _, file := range files {
		fileHandle, err := LuaRuntime.Open(filepath.Join("lua", file.Name()))
		if err != nil {
			panic("sqs? more like slacking off? amirite?")
		}

		bytesRead, err := io.ReadAll(fileHandle)
		if err != nil {
			panic("where is ?")
		}

		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		luaRuntimeContents[name] = string(bytesRead)
	}
}

func (s *Service) CreateSandbox(ctx context.Context, opts CreateOptions) (_ *Sandbox, err error) {
	_, _, endObservation := s.operations.createSandbox.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// Default LuaModules to our runtime files
	if opts.LuaModules == nil {
		opts.LuaModules = luaRuntimeContents
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
