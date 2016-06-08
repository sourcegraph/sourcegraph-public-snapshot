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
			BuildLogs:          &buildLogs{},
			Builds:             &builds{},
			Channel:            &channel{},
			Directory:          &directory{},
			ExternalAuthTokens: &externalAuthTokens{},
			GlobalDefs:         &globalDefs{},
			GlobalDeps:         &globalDeps{},
			GlobalRefs:         &globalRefs{},
			Password:           &password{},
			RepoConfigs:        &repoConfigs{},
			RepoStatuses:       &repoStatuses{},
			RepoVCS:            &repoVCS{},
			Repos:              &repos{},
			Users:              &users{},
		}), nil
	})
}
