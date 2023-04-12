package libs

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Keywords = keywordsAPI{}

type keywordsAPI struct{}

func (keywordsAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"search": util.WrapLuaFunction(func(state *lua.LState) error {
			if true {
				panic("KEYWORD SEARCH UNIMPL")
			}

			// TODO
			return nil
		}),
	}
}
