package pgsql

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/server/internal/store/fs"
	"sourcegraph.com/sourcegraph/sourcegraph/server/serverctx"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

func init() {
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		return store.WithStores(ctx, store.Stores{
			Accounts:           &accounts{},
			Authorizations:     &authorizations{},
			Builds:             &builds{},
			BuildLogs:          &fs.BuildLogs{},
			Directory:          &directory{},
			ExternalAuthTokens: &externalAuthTokens{},
			RepoConfigs:        &repoConfigs{},
			Password:           &password{},
			RegisteredClients:  &registeredClients{},
			RepoVCS:            &fs.RepoVCS{},
			Repos:              &repos{},
			RepoStatuses:       &repoStatuses{},
			Users:              &users{},
			RepoPerms:          &repoPerms{},
		}), nil
	})
}
