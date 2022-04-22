package libs

import (
	"github.com/grafana/regexp"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Patterns = patternAPI{}

type patternAPI struct{}

func (api patternAPI) LuaAPI() map[string]lua.LGFunction {
	newPathPatternConstructor := func(prefix, suffix string) func(*lua.LState) error {
		return func(state *lua.LState) error {
			pattern := regexp.QuoteMeta(state.CheckString(1))
			state.Push(luar.New(state, luatypes.NewPattern(prefix+pattern+suffix)))
			return nil
		}
	}

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
		"path_literal":   util.WrapLuaFunction(newPathPatternConstructor("^", "$")),
		"path_segment":   util.WrapLuaFunction(newPathPatternConstructor("(^|/)", "(/|$)")),
		"path_basename":  util.WrapLuaFunction(newPathPatternConstructor("(^|/)", "$")),
		"path_extension": util.WrapLuaFunction(newPathPatternConstructor("(^|/)[^/]+.", "$")),
		"path_combine":   util.WrapLuaFunction(newPathPatternCombineConstructor(luatypes.NewCombinedPattern)),
		"path_exclude":   util.WrapLuaFunction(newPathPatternCombineConstructor(luatypes.NewExcludePattern)),
	}
}
