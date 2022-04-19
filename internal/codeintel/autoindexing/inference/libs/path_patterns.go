package libs

import (
	"github.com/grafana/regexp"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Patterns = patternAPI{}

type patternAPI struct{}

func (api patternAPI) LuaAPI() map[string]lua.LGFunction {
	newPatternConstructor := func(prefix, suffix string) func(*lua.LState) error {
		return func(state *lua.LState) error {
			pattern := regexp.QuoteMeta(state.CheckString(1))
			state.Push(luar.New(state, luatypes.NewPattern(prefix+pattern+suffix)))
			return nil
		}
	}

	newPatternCombineConstructor := func(combine func([]*luatypes.PathPattern) *luatypes.PathPattern) func(*lua.LState) error {
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
		"literal":   util.WrapLuaFunction(newPatternConstructor("^", "$")),
		"segment":   util.WrapLuaFunction(newPatternConstructor("(^|/)", "(/|$)")),
		"basename":  util.WrapLuaFunction(newPatternConstructor("(^|/)", "$")),
		"extension": util.WrapLuaFunction(newPatternConstructor("(^|/)[^/]+.", "$")),
		"combine":   util.WrapLuaFunction(newPatternCombineConstructor(luatypes.NewCombinedPattern)),
		"exclude":   util.WrapLuaFunction(newPatternCombineConstructor(luatypes.NewExcludePattern)),
	}
}
