package libs

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Embeddings = embeddingsAPI{}

type embeddingsAPI struct{}

func (embeddingsAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"search": util.WrapLuaFunction(func(state *lua.LState) error {
			if true {
				panic("EMBEDDINGS SEARCH UNIMPL")
			}

			// TODO
			return nil
		}),
	}
}
