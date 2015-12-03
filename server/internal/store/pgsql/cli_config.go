package pgsql

import (
	"src.sourcegraph.com/sourcegraph/server/internal/store/fs"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/store/cli"
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
	Storage:             &Storage{},
	RepoStatuses:        &RepoStatuses{},
	UserPermissions:     &UserPermissions{},
	Users:               &Users{},
	UserKeys:            fs.NewUserKeys(),
	Changesets:          &fs.Changesets{},
	Invites:             &Invites{},
}
