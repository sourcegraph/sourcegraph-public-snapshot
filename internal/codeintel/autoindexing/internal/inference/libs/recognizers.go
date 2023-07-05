package libs

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Recognizers = recognizerAPI{}

type recognizerAPI struct{}

func (api recognizerAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"path_recognizer": util.WrapLuaFunction(func(state *lua.LState) error {
			recognizer, err := luatypes.RecognizerFromTable(state.CheckTable(1))
			state.Push(luar.New(state, recognizer))
			return err
		}),

		"fallback_recognizer": util.WrapLuaFunction(func(state *lua.LState) error {
			recognizers, err := luatypes.RecognizersFromUserData(state.CheckTable(1))
			state.Push(luar.New(state, luatypes.NewFallback(recognizers)))
			return err
		}),
	}
}
