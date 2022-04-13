package luasandbox

import lua "github.com/yuin/gopher-lua"

// CreateModule wraps a map of functions into a lua.LGFunction suitable for
// use in CreateOptions.Modules.
func CreateModule(api map[string]lua.LGFunction) lua.LGFunction {
	return WrapLuaFunction(func(state *lua.LState) error {
		t := state.NewTable()
		state.SetFuncs(t, api)
		state.Push(t)
		return nil
	})
}

// WrapLuaFunction invokes the given callback and returns 2 (raising an error) if the
// returned error is non-nil, and returns 1 (success) otherwise. This wrapper function
// makes no assumptions about how the called function modifies the Lua virtual machine
// state.
func WrapLuaFunction(f func(state *lua.LState) error) func(state *lua.LState) int {
	return func(state *lua.LState) int {
		if err := f(state); err != nil {
			state.RaiseError(err.Error())
			return 2
		}

		return 1
	}
}
