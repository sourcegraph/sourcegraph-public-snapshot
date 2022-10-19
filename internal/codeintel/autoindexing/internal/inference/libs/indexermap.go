package libs

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var IndexerMap = indexerMapAPI{}

type indexerMapAPI struct{}

var defaultIndexers = map[string]string{
	"clang":      "sourcegraph/lsif-clang",
	"go":         "sourcegraph/lsif-go:latest",
	"java":       "sourcegraph/scip-java",
	"python":     "sourcegraph/scip-python:autoindex",
	"rust":       "sourcegraph/lsif-rust",
	"typescript": "sourcegraph/scip-typescript:autoindex",
}

func (api indexerMapAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"get": util.WrapLuaFunction(func(state *lua.LState) error {
			name := state.CheckString(1)

			if indexer, ok := conf.SiteConfig().CodeIntelAutoIndexingIndexerMap[name]; ok {
				state.Push(luar.New(state, indexer))
				return nil
			}

			if indexer, ok := defaultIndexers[name]; ok {
				state.Push(luar.New(state, indexer))
				return nil
			}

			return errors.Newf("no indexer is registered for %q", name)
		}),
	}
}
