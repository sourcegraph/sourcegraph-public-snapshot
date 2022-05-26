package luasandbox

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	// LuaModules is map of require("$KEY") -> contents
	// that will be loaded in the lua sandbox state.
	//
	// This prevents subsequent executions from modifying
	// the state (or peeking into the state) of any of the
	// other recognizers
	LuaModules map[string]string
}

//go:embed lua/*
var luaRuntime embed.FS

var LuaRuntimeContents = map[string]string{}

func init() {
	var err error
	LuaRuntimeContents, err = CreateLuaRuntimeFromFS(luaRuntime, "lua", "")
	if err != nil {
		panic(fmt.Sprintf("error loading lua runtime files: %s", err))
	}
}

func getAllFilepaths(fs embed.FS, path string) (out []string, err error) {
	if len(path) == 0 {
		path = "."
	}

	entries, err := fs.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fp := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			res, err := getAllFilepaths(fs, fp)
			if err != nil {
				return nil, err
			}
			out = append(out, res...)
			continue
		}
		out = append(out, fp)
	}
	return
}

func CreateLuaRuntimeFromFS(runtime embed.FS, dir, prefix string) (map[string]string, error) {
	files, err := getAllFilepaths(runtime, dir)
	if err != nil {
		return nil, err
	}

	// TODO: Handle init.lua
	contents := map[string]string{}
	for _, file := range files {
		fileHandle, err := runtime.Open(file)
		if err != nil {
			return nil, err
		}

		bytesRead, err := io.ReadAll(fileHandle)
		if err != nil {
			return nil, err
		}

		// change "lua/fun.lua" -> "fun.lua"
		file = strings.TrimPrefix(file, dir+string(os.PathSeparator))
		// change "fun.lua" -> "fun"
		file = strings.TrimSuffix(file, filepath.Ext(file))

		parts := filepath.SplitList(file)
		name := strings.Join(parts, ".")
		if prefix != "" {
			name = prefix + "." + name
		}

		contents[name] = string(bytesRead)
	}

	return contents, nil
}

func (s *Service) CreateSandbox(ctx context.Context, opts CreateOptions) (_ *Sandbox, err error) {
	_, _, endObservation := s.operations.createSandbox.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// Default LuaModules to our runtime files
	if opts.LuaModules == nil {
		opts.LuaModules = map[string]string{}
	}

	for k, v := range LuaRuntimeContents {
		if _, ok := opts.LuaModules[k]; ok {
			return nil, errors.Newf("Cannot have lua modules that overwrite each other: %s", k)
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
