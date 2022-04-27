package inference

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var defaultAPIs = map[string]luasandbox.LuaLib{
	"sg.patterns":    libs.Patterns,
	"sg.recognizers": libs.Recognizers,
}

var defaultModules = (func() map[string]lua.LGFunction {
	modules := make(map[string]lua.LGFunction, len(luasandbox.DefaultModules)+len(defaultAPIs))

	for name, module := range luasandbox.DefaultModules {
		modules[name] = module
	}
	for name, api := range defaultAPIs {
		modules[name] = util.CreateModule(api.LuaAPI())
	}

	return modules
})()
