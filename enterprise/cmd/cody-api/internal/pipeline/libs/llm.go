package libs

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var LLM = llmAPI{}

type llmAPI struct{}

func (llmAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"run": util.WrapLuaFunction(func(state *lua.LState) error {
			if false {
				panic("RUN UNIMPL")
			}

			// TODO
			state.Push(luar.New(state, "Henlo, yes this is Claob."))
			return nil
		}),
	}
}
