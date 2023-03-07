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
	"rust":       "sourcegraph/scip-rust",
	"typescript": "sourcegraph/scip-typescript",
	"ruby":       "sourcegraph/scip-ruby",
}

// To update, run `DOCKER_USER=... DOCKER_PASS=... ./update-shas.sh`
var defaultIndexerSHAs = map[string]string{
	"sourcegraph/lsif-clang":      "sha256:ea814e5ab5c6e1e6ab4d001e4f4afddcc7b44128edbeeedf1d97da553813a4c8",
	"sourcegraph/lsif-go":         "sha256:2194d2652862966f022b537ed81bccf5a9a535ab763534cb4e98a3083c8a1bc6",
	"sourcegraph/lsif-rust":       "sha256:83cb769788987eb52f21a18b62d51ebb67c9436e1b0d2e99904c70fef424f9d1",
	"sourcegraph/scip-rust":       "sha256:e9c400fd1d3146cd9a3d98f89c6d9a70e0a116618057e0dac452219b1c60b658",
	"sourcegraph/scip-java":       "sha256:964a45ef06b7d914729b1c61b6907b662fc54545b188881c6d25e56fcc8dfb8c",
	"sourcegraph/scip-python":     "sha256:4cb64c4f62cfa611fcb217581073c2831fb9350bbb1c8e855f152cc4b3428a00",
	"sourcegraph/scip-typescript": "sha256:2eb90bf6d52f86608b6a723e14079eafed54121419bfd0129bf61a09e1a0f293",
	"sourcegraph/scip-ruby":       "sha256:e553fee039973cda8726d4c8c13cdbb851f82a6fca5daa15798a595ee4042906",
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
