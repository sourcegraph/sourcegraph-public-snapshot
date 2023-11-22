package libs

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Patterns = patternAPI{}

type patternAPI struct{}

func (api patternAPI) LuaAPI() map[string]lua.LGFunction {
	newPathPatternCombineConstructor := func(combine func([]*luatypes.PathPattern) *luatypes.PathPattern) func(*lua.LState) error {
		return func(state *lua.LState) error {
			var patterns []*luatypes.PathPattern
			for i := 1; i <= state.GetTop(); i++ {
				additionalPatterns, err := luatypes.PathPatternsFromUserData(state.CheckAny(i))
				if err != nil {
					return err
				}

				patterns = append(patterns, additionalPatterns...)
			}

			state.Push(luar.New(state, combine(patterns)))
			return nil
		}
	}

	return map[string]lua.LGFunction{
		// type: (string, array[string]) -> pattern
		"backdoor": util.WrapLuaFunction(func(state *lua.LState) error {
			glob := state.CheckString(1)
			pathspecTable := state.CheckTable(2)

			pathspecs, err := util.MapSlice(pathspecTable, func(value lua.LValue) (string, error) {
				if s, ok := value.(lua.LString); ok {
					return string(s), nil
				}
				return "", util.NewTypeError("lua.LString", value)
			})
			if err != nil {
				return err
			}

			state.Push(luar.New(state, luatypes.NewPattern(glob, pathspecs)))
			return nil
		}),
		// type: ((pattern | array[pattern])...) -> pattern
		"path_combine": util.WrapLuaFunction(newPathPatternCombineConstructor(luatypes.NewCombinedPattern)),
		// type: ((pattern | array[pattern])...) -> pattern
		"path_exclude": util.WrapLuaFunction(newPathPatternCombineConstructor(luatypes.NewExcludePattern)),
	}
}
