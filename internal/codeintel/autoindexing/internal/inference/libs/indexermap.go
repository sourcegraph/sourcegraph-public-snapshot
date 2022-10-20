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
	"clang":      "sourcegraph/lsif-clang@sha256:5ef2334ac9d58f1f947651812aa8d8ba0ed584913f2429cc9952cb25f94976d8",
	"go":         "sourcegraph/lsif-go@sha256:253c991fdd8b118afadcfbe6f7a6d03ca91c44fd2860dbe8a9fd69c93c6025f6",
	"java":       "sourcegraph/scip-java@sha256:eb3996bdc8ab3a56600e7d647bc1ef72f3db8cfffc2026550095a0af7bb762bd",
	"python":     "sourcegraph/scip-python@sha256:5049c4598d03af542bde5e1254a17fa6d1eb794c1bdd14d0162fb39c604581b4",
	"rust":       "sourcegraph/lsif-rust@sha256:83cb769788987eb52f21a18b62d51ebb67c9436e1b0d2e99904c70fef424f9d1",
	"typescript": "sourcegraph/scip-typescript@sha256:4184c6a771037d854e23323db2c8a65947029c93401b696f3b5281fff5c3cbe9",
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
