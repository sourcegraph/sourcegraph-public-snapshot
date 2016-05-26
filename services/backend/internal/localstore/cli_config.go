package localstore

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"
)

func init() {
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		return store.WithStores(ctx, store.Stores{
			Accounts:           &accounts{},
			Builds:             &builds{},
			BuildLogs:          &buildLogs{},
			Channel:            &channel{},
			Directory:          &directory{},
			ExternalAuthTokens: &externalAuthTokens{},
			GlobalDefs:         &globalDefs{},
			GlobalRefs:         &globalRefs{},
			GlobalDeps:         &globalDeps{},
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
