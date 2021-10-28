package database

import (
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

// DB is an interface that embeds dbutil.DB, adding methods to
// return specialized stores on top of that interface. In time,
// the expectation is to replace uses of dbutil.DB with database.DB,
// and remove dbutil.DB altogether.
type DB interface {
	dbutil.DB
	AccessTokens() AccessTokenStore
	Namespaces() NamespaceStore
	OrgMembers() OrgMemberStore
	Orgs() OrgStore
	Repos() RepoStore
	SavedSearches() SavedSearchStore
	Settings() SettingsStore
	UserCredentials(encryption.Key) UserCredentialsStore
	UserEmails() UserEmailsStore
	Users() UserStore
}

// NewDB creates a new DB from a dbutil.DB, providing a thin wrapper
// that has constructor methods for the more specialized stores.
func NewDB(inner dbutil.DB) DB {
	return &db{inner}
}

type db struct {
	dbutil.DB
}

func (d *db) AccessTokens() AccessTokenStore {
	return AccessTokens(d.DB)
}

func (d *db) Namespaces() NamespaceStore {
	return Namespaces(d.DB)
}

func (d *db) OrgMembers() OrgMemberStore {
	return OrgMembers(d.DB)
}

func (d *db) Orgs() OrgStore {
	return Orgs(d.DB)
}

func (d *db) Repos() RepoStore {
	return Repos(d.DB)
}

func (d *db) SavedSearches() SavedSearchStore {
	return SavedSearches(d.DB)
}

func (d *db) Settings() SettingsStore {
	return Settings(d.DB)
}

func (d *db) UserCredentials(key encryption.Key) UserCredentialsStore {
	return UserCredentials(d.DB, key)
}

func (d *db) UserEmails() UserEmailsStore {
	return UserEmails(d.DB)
}

func (d *db) Users() UserStore {
	return Users(d.DB)
}
