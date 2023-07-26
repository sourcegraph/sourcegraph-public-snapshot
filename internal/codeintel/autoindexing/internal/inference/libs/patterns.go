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
		"backdoor": util.WrapLuaFunction(func(state *lua.LState) error {
			glob := state.CheckString(1)
			pathspecTable := state.CheckTable(2)

			values, err := util.DecodeSlice(pathspecTable)
			if err != nil {
				return err
			}

			var pathspecs []string
			for _, v := range values {
				pathspecs = append(pathspecs, v.String())
			}

			state.Push(luar.New(state, luatypes.NewPattern(glob, pathspecs)))
			return nil
		}),
		"path_combine": util.WrapLuaFunction(newPathPatternCombineConstructor(luatypes.NewCombinedPattern)),
		"path_exclude": util.WrapLuaFunction(newPathPatternCombineConstructor(luatypes.NewExcludePattern)),
	}
}
