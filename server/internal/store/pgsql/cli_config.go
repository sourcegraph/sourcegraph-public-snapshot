package pgsql

import (
	"src.sourcegraph.com/sourcegraph/server/internal/store/fs"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/store/cli"
)

func init() {
	cli.RegisterStores("pgsql", &store.Stores{
		Accounts:            &accounts{},
		Authorizations:      &authorizations{},
		Builds:              &builds{},
		Directory:           &directory{},
		ExternalAuthTokens:  &externalAuthTokens{},
		RepoConfigs:         &repoConfigs{},
		RepoCounters:        &repoCounters{},
		MirroredRepoSSHKeys: &mirroredRepoSSHKeys{},
		Password:            &password{},
		RegisteredClients:   &registeredClients{},
		RepoVCS:             &fs.RepoVCS{},
		Repos:               &repos{},
		Storage:             &storage{},
		RepoStatuses:        &repoStatuses{},
		UserPermissions:     &userPermissions{},
		Users:               &users{},
		Changesets:          &fs.Changesets{},
		Invites:             &invites{},
	})
}
