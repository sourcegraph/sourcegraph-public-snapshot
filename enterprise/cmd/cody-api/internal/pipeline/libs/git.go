package libs

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

var Git = gitAPI{}

type gitAPI struct{}

func (gitAPI) LuaAPI() map[string]lua.LGFunction {
	var (
		client       = gitserver.NewClient()
		authzChecker = authz.DefaultSubRepoPermsChecker
	)

	return map[string]lua.LGFunction{
		"head": util.WrapLuaFunction(func(state *lua.LState) error {
			head, _, err := client.Head(state.Context(), authzChecker, api.RepoName(state.CheckString(1)))
			if err != nil {
				return err
			}

			state.Push(luar.New(state, head))
			return nil
		}),
	}
}
