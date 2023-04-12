package libs

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Time = timeAPI{}

type timeAPI struct{}

func (timeAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"shortTimestamp": util.WrapLuaFunction(func(state *lua.LState) error {
			now := time.Now()
			state.Push(luar.New(state, fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())))
			return nil
		}),
	}
}
