package localstore

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/server/serverctx"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

func init() {
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		return store.WithStores(ctx, store.Stores{
			Accounts:           &accounts{},
			Builds:             &builds{},
			BuildLogs:          &buildLogs{},
			Directory:          &directory{},
			ExternalAuthTokens: &externalAuthTokens{},
			GlobalDefs:         &globalDefs{},
			GlobalRefs:         globalRefsExp,
			RepoConfigs:        &repoConfigs{},
			Password:           &password{},
			RepoVCS:            &repoVCS{},
			Repos:              &repos{},
			RepoStatuses:       &repoStatuses{},
			Users:              &users{},
			RepoPerms:          &repoPerms{},
		}), nil
	})
}
