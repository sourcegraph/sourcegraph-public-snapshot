package libs

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var Indexes = indexesAPI{}

type indexesAPI struct{}

var defaultIndexers = map[string]string{
	"clang":      "sourcegraph/lsif-clang",
	"go":         "sourcegraph/lsif-go",
	"java":       "sourcegraph/scip-java",
	"python":     "sourcegraph/scip-python",
	"rust":       "sourcegraph/lsif-rust",
	"typescript": "sourcegraph/scip-typescript",
	"ruby":       "sourcegraph/scip-ruby",
}

// To update, run `DOCKER_USER=... DOCKER_PASS=... ./update-shas.sh`
var defaultIndexerSHAs = map[string]string{
	"sourcegraph/lsif-clang":      "sha256:5ef2334ac9d58f1f947651812aa8d8ba0ed584913f2429cc9952cb25f94976d8",
	"sourcegraph/lsif-go":         "sha256:cba76f5b3edb5d9af43e1dc59e27ecdb4b8b2fafda6a5d55d7e37def3b502775",
	"sourcegraph/lsif-rust":       "sha256:83cb769788987eb52f21a18b62d51ebb67c9436e1b0d2e99904c70fef424f9d1",
	"sourcegraph/scip-java":       "sha256:eb3996bdc8ab3a56600e7d647bc1ef72f3db8cfffc2026550095a0af7bb762bd",
	"sourcegraph/scip-python":     "sha256:5049c4598d03af542bde5e1254a17fa6d1eb794c1bdd14d0162fb39c604581b4",
	"sourcegraph/scip-typescript": "sha256:37546e04763d6d1853fb6f32285ad630fc0d8671d4f1d2db40c6df272120f2f8",
	"sourcegraph/scip-ruby":       "sha256:1e7538eead787a9a220e54c442eaf10372f3f41d2be2871713e6ec367bd40f81",
}

func DefaultIndexerForLang(language string) (string, bool) {
	indexer, ok := defaultIndexers[language]
	if !ok {
		return "", false
	}

	sha, ok := defaultIndexerSHAs[indexer]
	if !ok {
		panic(fmt.Sprintf("no SHA set for indexer %q", indexer))
	}

	return fmt.Sprintf("%s@%s", indexer, sha), true
}

func (api indexesAPI) LuaAPI() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"get": util.WrapLuaFunction(func(state *lua.LState) error {
			language := state.CheckString(1)

			if indexer, ok := conf.SiteConfig().CodeIntelAutoIndexingIndexerMap[language]; ok {
				state.Push(luar.New(state, indexer))
				return nil
			}

			if indexer, ok := DefaultIndexerForLang(language); ok {
				state.Push(luar.New(state, indexer))
				return nil
			}

			return errors.Newf("no indexer is registered for %q", language)
		}),
	}
}
