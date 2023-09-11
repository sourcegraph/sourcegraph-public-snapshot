package inference

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/internal/memo"
)

var defaultAPIs = map[string]luasandbox.LuaLib{
	"internal_patterns":    libs.Patterns,
	"internal_recognizers": libs.Recognizers,
	"internal_indexes":     libs.Indexes,
}

var defaultModules = memo.NewMemoizedConstructor(func() (map[string]lua.LGFunction, error) {
	defaultModules, err := luasandbox.DefaultGoModules.Init()
	if err != nil {
		return nil, err
	}

	modules := make(map[string]lua.LGFunction, len(defaultModules)+len(defaultAPIs))
	for name, module := range defaultModules {
		modules[name] = module
	}
	for name, api := range defaultAPIs {
		modules[name] = util.CreateModule(api.LuaAPI())
	}

	return modules, nil
})
