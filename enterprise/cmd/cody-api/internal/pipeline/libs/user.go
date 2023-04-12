package libs

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var User = userAPI{}

type userAPI struct{}

type CapabilityPerformer func(ctx context.Context, capability string, data any) (any, error)

func (userAPI) LuaAPI(performCapability CapabilityPerformer) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"perform": util.WrapLuaFunction(func(state *lua.LState) error {
			ctx := state.Context()

			resp, err := performCapability(ctx, state.CheckString(1), state.CheckString(2))
			if err != nil {
				return err
			}

			state.Push(luar.New(state, resp))
			return nil
		}),
	}
}
