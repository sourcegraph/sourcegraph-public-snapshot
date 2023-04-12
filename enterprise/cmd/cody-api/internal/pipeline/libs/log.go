package libs

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Log = logAPI{}

type logAPI struct{}

func (logAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"log": util.WrapLuaFunction(func(state *lua.LState) error {
			fmt.Printf("LOG: %v\n", state.CheckAny(1))
			return nil
		}),
	}
}
