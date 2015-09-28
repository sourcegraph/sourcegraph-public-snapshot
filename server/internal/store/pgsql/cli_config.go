package pgsql

import (
	"sourcegraph.com/sourcegraph/sourcegraph/server/internal/store/fs"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/store/cli"
)

func init() {
	cli.RegisterStores("pgsql", Stores)
}

var Stores = &store.Stores{
	Accounts:            &Accounts{},
	Authorizations:      &Authorizations{},
	Builds:              &Builds{},
	Directory:           &Directory{},
	ExternalAuthTokens:  &ExternalAuthTokens{},
	RepoConfigs:         &RepoConfigs{},
	RepoCounters:        &RepoCounters{},
	MirroredRepoSSHKeys: &MirroredRepoSSHKeys{},
	Password:            &Password{},
	RegisteredClients:   &RegisteredClients{},
	RepoVCS:             &fs.RepoVCS{},
	Repos:               &Repos{},
	RepoStatuses:        &RepoStatuses{},
	UserPermissions:     &UserPermissions{},
	Users:               &Users{},
}
