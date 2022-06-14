package database

import (
	"context"
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
	basestore.ShareableStore

	AccessTokens() AccessTokenStore
	Authz() AuthzStore
	BitbucketProjectPermissions() BitbucketProjectPermissionsStore
	Conf() ConfStore
	EventLogs() EventLogStore
	SecurityEventLogs() SecurityEventLogsStore
	ExternalServices() ExternalServiceStore
	FeatureFlags() FeatureFlagStore
	GitserverRepos() GitserverRepoStore
	GitserverLocalClone() GitserverLocalCloneStore
	GlobalState() GlobalStateStore
	Namespaces() NamespaceStore
	OrgInvitations() OrgInvitationStore
	OrgMembers() OrgMemberStore
	Orgs() OrgStore
	OrgStats() OrgStatsStore
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

	Transact(context.Context) (DB, error)
	Done(error) error
}

var _ DB = (*db)(nil)

// NewDB creates a new DB from a dbutil.DB, providing a thin wrapper
// that has constructor methods for the more specialized stores.
func NewDB(inner *sql.DB) DB {
	return &db{basestore.NewWithHandle(basestore.NewHandleWithDB(inner, sql.TxOptions{}))}
}

func NewDBWith(other basestore.ShareableStore) DB {
	return &db{basestore.NewWithHandle(other.Handle())}
}

type db struct {
	*basestore.Store
}

func (d *db) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return d.Handle().QueryContext(ctx, q, args...)
}

func (d *db) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return d.Handle().ExecContext(ctx, q, args...)

}

func (d *db) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	return d.Handle().QueryRowContext(ctx, q, args...)
}

func (d *db) Transact(ctx context.Context) (DB, error) {
	tx, err := d.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &db{tx}, nil
}

func (d *db) Done(err error) error {
	return d.Store.Done(err)
}

func (d *db) AccessTokens() AccessTokenStore {
	return AccessTokensWith(d.Store)
}

func (d *db) BitbucketProjectPermissions() BitbucketProjectPermissionsStore {
	return BitbucketProjectPermissionsStoreWith(d.Store)
}

func (d *db) Authz() AuthzStore {
	return AuthzWith(d.Store)
}

func (d *db) Conf() ConfStore {
	return &confStore{Store: basestore.NewWithHandle(d.Handle())}
}

func (d *db) EventLogs() EventLogStore {
	return EventLogsWith(d.Store)
}

func (d *db) SecurityEventLogs() SecurityEventLogsStore {
	return SecurityEventLogsWith(d.Store)
}

func (d *db) ExternalServices() ExternalServiceStore {
	return ExternalServicesWith(d.Store)
}

func (d *db) FeatureFlags() FeatureFlagStore {
	return FeatureFlagsWith(d.Store)
}

func (d *db) GitserverRepos() GitserverRepoStore {
	return GitserverReposWith(d.Store)
}

func (d *db) GitserverLocalClone() GitserverLocalCloneStore {
	return GitserverLocalCloneStoreWith(d.Store)
}

func (d *db) GlobalState() GlobalStateStore {
	return GlobalStateWith(d.Store)
}

func (d *db) Namespaces() NamespaceStore {
	return NamespacesWith(d.Store)
}

func (d *db) OrgInvitations() OrgInvitationStore {
	return OrgInvitationsWith(d.Store)
}

func (d *db) OrgMembers() OrgMemberStore {
	return OrgMembersWith(d.Store)
}

func (d *db) Orgs() OrgStore {
	return OrgsWith(d.Store)
}

func (d *db) OrgStats() OrgStatsStore {
	return OrgStatsWith(d.Store)
}

func (d *db) Phabricator() PhabricatorStore {
	return PhabricatorWith(d.Store)
}

func (d *db) Repos() RepoStore {
	return ReposWith(d.Store)
}

func (d *db) SavedSearches() SavedSearchStore {
	return SavedSearchesWith(d.Store)
}

func (d *db) SearchContexts() SearchContextsStore {
	return SearchContextsWith(d.Store)
}

func (d *db) Settings() SettingsStore {
	return SettingsWith(d.Store)
}

func (d *db) SubRepoPerms() SubRepoPermsStore {
	return SubRepoPermsWith(d.Store)
}

func (d *db) TemporarySettings() TemporarySettingsStore {
	return TemporarySettingsWith(d.Store)
}

func (d *db) UserCredentials(key encryption.Key) UserCredentialsStore {
	return UserCredentialsWith(d.Store, key)
}

func (d *db) UserEmails() UserEmailsStore {
	return UserEmailsWith(d.Store)
}

func (d *db) UserExternalAccounts() UserExternalAccountsStore {
	return ExternalAccountsWith(d.Store)
}

func (d *db) UserPublicRepos() UserPublicRepoStore {
	return UserPublicReposWith(d.Store)
}

func (d *db) Users() UserStore {
	return UsersWith(d.Store)
}

func (d *db) WebhookLogs(key encryption.Key) WebhookLogStore {
	return WebhookLogsWith(d.Store, key)
}

func (d *db) Unwrap() dbutil.DB {
	// Recursively unwrap in case we ever call `database.NewDB()` with a `database.DB`
	if unwrapper, ok := d.Handle().(dbutil.Unwrapper); ok {
		return unwrapper.Unwrap()
	}
	return d.Handle()
}
