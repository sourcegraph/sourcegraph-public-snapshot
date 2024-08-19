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
	"go":         "sourcegraph/scip-go",
	"java":       "sourcegraph/scip-java",
	"python":     "sourcegraph/scip-python",
	"rust":       "sourcegraph/scip-rust",
	"typescript": "sourcegraph/scip-typescript",
	"ruby":       "sourcegraph/scip-ruby",
	"dotnet":     "sourcegraph/scip-dotnet",
}

// To update, run `DOCKER_USER=... DOCKER_PASS=... ./update-shas.sh`
var defaultIndexerSHAs = map[string]string{
	"sourcegraph/scip-go":         "sha256:e6ca2d4b55bd1379631d45faab169fc32dc6da2c1939ed11a700261ac4c4d26f",
	"sourcegraph/scip-rust":       "sha256:adf0047fc3050ba4f7be71302b42c74b49901f38fb40916d94ac5fc9181ac078",
	"sourcegraph/scip-java":       "sha256:3de6ba2221880e2ff3a7dcb9045e6c3e86f6079d6c8dc2f913a2ca8427605c69",
	"sourcegraph/scip-python":     "sha256:e3c13f0cadca78098439c541d19a72c21672a3263e22aa706760d941581e068d",
	"sourcegraph/scip-typescript": "sha256:3df8b36a2ad4e073415bfbeaedf38b3cfff3e697614c8f578299f470d140c2c8",
	"sourcegraph/scip-ruby":       "sha256:99b1a1adcb1ec1b6e92956f5817dfc6dfc9940f962b999685954b2ba052e1a7b",
	"sourcegraph/scip-dotnet":     "sha256:1d8a590edfb3834020fceedacac6608811dd31fcba9092426140093876d8d52e",
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
