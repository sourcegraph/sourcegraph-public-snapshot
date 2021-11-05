package database

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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
	EventLogs() EventLogStore
	ExternalServices() ExternalServiceStore
	FeatureFlags() FeatureFlagStore
	Namespaces() NamespaceStore
	OrgInvitations() OrgInvitationStore
	OrgMembers() OrgMemberStore
	Orgs() OrgStore
	Phabricator() PhabricatorStore
	Repos() RepoStore
	SavedSearches() SavedSearchStore
	SearchContexts() SearchContextsStore
	Settings() SettingsStore
	SubRepoPerms() SubRepoPermsStore
	TemporarySettings() TemporarySettingsStore
	UserCredentials(encryption.Key) UserCredentialsStore
	UserEmails() UserEmailsStore
	UserExternalAccounts() UserExternalAccountsStore
	UserPublicRepos() UserPublicRepoStore
	Users() UserStore
	WebhookLogs(encryption.Key) WebhookLogStore
}

// NewDB creates a new DB from a dbutil.DB, providing a thin wrapper
// that has constructor methods for the more specialized stores.
func NewDB(inner dbutil.DB) DB {
	return &db{inner}
}

type db struct {
	dbutil.DB
}

var _ DB = (*db)(nil)

func (d *db) AccessTokens() AccessTokenStore {
	return AccessTokens(d.DB)
}

func (d *db) EventLogs() EventLogStore {
	return EventLogs(d.DB)
}

func (d *db) ExternalServices() ExternalServiceStore {
	return ExternalServices(d.DB)
}

func (d *db) FeatureFlags() FeatureFlagStore {
	return FeatureFlags(d.DB)
}

func (d *db) Namespaces() NamespaceStore {
	return Namespaces(d.DB)
}

func (d *db) OrgInvitations() OrgInvitationStore {
	return OrgInvitations(d.DB)
}

func (d *db) OrgMembers() OrgMemberStore {
	return OrgMembers(d.DB)
}

func (d *db) Orgs() OrgStore {
	return Orgs(d.DB)
}

func (d *db) Phabricator() PhabricatorStore {
	return Phabricator(d.DB)
}

func (d *db) Repos() RepoStore {
	return Repos(d.DB)
}

func (d *db) SavedSearches() SavedSearchStore {
	return SavedSearches(d.DB)
}

func (d *db) SearchContexts() SearchContextsStore {
	return SearchContexts(d.DB)
}

func (d *db) Settings() SettingsStore {
	return Settings(d.DB)
}

func (d *db) SubRepoPerms() SubRepoPermsStore {
	return SubRepoPerms(d.DB)
}

func (d *db) TemporarySettings() TemporarySettingsStore {
	return &temporarySettingsStore{Store: basestore.NewWithDB(d.DB, sql.TxOptions{})}
}

func (d *db) UserCredentials(key encryption.Key) UserCredentialsStore {
	return UserCredentials(d.DB, key)
}

func (d *db) UserEmails() UserEmailsStore {
	return UserEmails(d.DB)
}

func (d *db) UserExternalAccounts() UserExternalAccountsStore {
	return ExternalAccounts(d.DB)
}

func (d *db) UserPublicRepos() UserPublicRepoStore {
	return UserPublicRepos(d.DB)
}

func (d *db) Users() UserStore {
	return Users(d.DB)
}

func (d *db) WebhookLogs(key encryption.Key) WebhookLogStore {
	return WebhookLogs(d.DB, key)
}

func (d *db) Unwrap() dbutil.DB {
	// Recursively unwrap in case we ever call `database.NewDB()` with a `database.DB`
	if unwrapper, ok := d.DB.(dbutil.Unwrapper); ok {
		return unwrapper.Unwrap()
	}
	return d.DB
}
