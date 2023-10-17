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
		// type: ({
		//   "patterns": array[pattern],
		//   "patterns_for_content": array[pattern],
		//   "generate": (registration_api, paths: array[string], contents_by_path: table[string, string]) -> void,
		//   "hints": (registration_api, paths: array[string]) -> void
		// }) -> recognizer
		"path_recognizer": util.WrapLuaFunction(func(state *lua.LState) error {
			recognizer, err := luatypes.RecognizerFromTable(state.CheckTable(1))
			state.Push(luar.New(state, recognizer))
			return err
		}),
		// type: (array[recognizer]) -> recognizer
		"fallback_recognizer": util.WrapLuaFunction(func(state *lua.LState) error {
			recognizers, err := util.MapSlice(state.CheckTable(1), func(value lua.LValue) (*luatypes.Recognizer, error) {
				return util.TypecheckUserData[*luatypes.Recognizer](value, "*Recognizer")
			})
			state.Push(luar.New(state, luatypes.NewFallback(recognizers)))
			return err
		}),
	}
}
