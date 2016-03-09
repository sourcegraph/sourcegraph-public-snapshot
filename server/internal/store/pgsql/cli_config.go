package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/server/internal/store/fs"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	"src.sourcegraph.com/sourcegraph/store"
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
			Invites:            &invites{},
			RepoPerms:          &repoPerms{},
		}), nil
	})
}
