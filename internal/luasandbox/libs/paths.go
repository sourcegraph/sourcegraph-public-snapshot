package libs

import (
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Path = pathAPI{}

type pathAPI struct{}

func (api pathAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		// type: (string) -> array[string]
		"ancestors": util.WrapLuaFunction(func(state *lua.LState) error {
			state.Push(luar.New(state, ancestorDirs(state.CheckString(1))))
			return nil
		}),
		// type: (string) -> string
		"basename": util.WrapLuaFunction(func(state *lua.LState) error {
			state.Push(luar.New(state, filepath.Base(state.CheckString(1))))
			return nil
		}),
		// type: (string) -> string
		"dirname": util.WrapLuaFunction(func(state *lua.LState) error {
			state.Push(luar.New(state, dirWithoutDot(state.CheckString(1))))
			return nil
		}),
		// type: (string, string) -> string
		"join": util.WrapLuaFunction(func(state *lua.LState) error {
			state.Push(luar.New(state, filepath.Join(state.CheckString(1), state.CheckString(2))))
			return nil
		}),
	}
}

// dirWithoutDot returns the directory name of the given path. Unlike filepath.Dir,
// this function will return an empty string (instead of a `.`) to indicate an empty
// directory name.
func dirWithoutDot(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}
	if len(dir) > 0 && dir[0] == '/' {
		return dir[1:]
	}

	return dir
}

// ancestorDirs returns all ancestor dirnames of the given path. The last element of
// the returned slice will always be empty (indicating the repository root).
func ancestorDirs(path string) (ancestors []string) {
	dir := dirWithoutDot(path)
	for dir != "" {
		ancestors = append(ancestors, dir)
		dir = dirWithoutDot(dir)
	}

	ancestors = append(ancestors, "")
	return ancestors
}
