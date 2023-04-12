package pipeline

import (
	golua "github.com/yuin/gopher-lua"
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/pipeline/libs"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/internal/memo"
)

var defaultAPIs = map[string]luasandbox.LuaLib{
	"embeddings": libs.Embeddings,
	"git":        libs.Git,
	"keywords":   libs.Keywords,
	"llm":        libs.LLM,
	"log":        libs.Log,
	"time":       libs.Time,
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

func makeModules(performCapability libs.CapabilityPerformer) (map[string]golua.LGFunction, error) {
	defaultModules, err := defaultModules.Init()
	if err != nil {
		return nil, err
	}

	requestModules := map[string]golua.LGFunction{
		"user": util.CreateModule(libs.User.LuaAPI(performCapability)),
	}

	return combineMaps(defaultModules, requestModules), nil
}

func combineMaps[K comparable, V any](ms ...map[K]V) map[K]V {
	n := 0
	for _, m := range ms {
		n += len(m)
	}

	m := make(map[K]V, n)
	for _, s := range ms {
		for k, v := range s {
			m[k] = v
		}
	}

	return m
}
