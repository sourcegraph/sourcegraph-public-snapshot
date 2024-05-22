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
}

// To update, run `DOCKER_USER=... DOCKER_PASS=... ./update-shas.sh`
var defaultIndexerSHAs = map[string]string{
	"sourcegraph/scip-go":         "sha256:56414010d8917d6952c051dd5fcc0901fdf5c12031d352cc0b26778f040dddcc",
	"sourcegraph/scip-rust":       "sha256:adf0047fc3050ba4f7be71302b42c74b49901f38fb40916d94ac5fc9181ac078",
	"sourcegraph/scip-java":       "sha256:a2b3828145cd38758a43363f06d786f9e620c97979a9291463c6544f7f17c68f",
	"sourcegraph/scip-python":     "sha256:e3c13f0cadca78098439c541d19a72c21672a3263e22aa706760d941581e068d",
	"sourcegraph/scip-typescript": "sha256:3df8b36a2ad4e073415bfbeaedf38b3cfff3e697614c8f578299f470d140c2c8",
	"sourcegraph/scip-ruby":       "sha256:ef53e5f1450330ddb4a3edce963b7e10d900d44ff1e7de4960680289ac25f319",
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
