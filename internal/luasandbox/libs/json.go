package libs

import (
	json "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var JSON = jsonAPI{}

type jsonAPI struct{}

func (api jsonAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"decode": util.WrapSoftFailingLuaFunction(func(state *lua.LState) error {
			value, err := json.Decode(state, []byte(state.CheckString(1)))
			if err != nil {
				return err
			}

			state.Push(value)
			return nil
		}),

		"encode": util.WrapSoftFailingLuaFunction(func(state *lua.LState) error {
			data, err := json.Encode(state.CheckAny(1))
			if err != nil {
				return err
			}

			state.Push(luar.New(state, string(data)))
			return nil
		}),
	}
}
