package luasandbox

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/libs"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

type LuaLib interface {
	LuaAPI() map[string]lua.LGFunction
}

var defaultAPIs = map[string]LuaLib{
	"json": libs.JSON,
	"path": libs.Path,
}

var DefaultModules = func() map[string]lua.LGFunction {
	modules := map[string]lua.LGFunction{}
	for name, api := range defaultAPIs {
		modules[name] = util.CreateModule(api.LuaAPI())
	}

	return modules
}()
