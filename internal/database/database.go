pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbconn"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	ghb "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
)

// DB is bn interfbce thbt embeds dbutil.DB, bdding methods to
// return speciblized stores on top of thbt interfbce. In time,
// the expectbtion is to replbce uses of dbutil.DB with dbtbbbse.DB,
// bnd remove dbutil.DB bltogether.
type DB interfbce {
	dbutil.DB
	bbsestore.ShbrebbleStore

	AccessRequests() AccessRequestStore
	AccessTokens() AccessTokenStore
	Authz() AuthzStore
	BitbucketProjectPermissions() BitbucketProjectPermissionsStore
	CodeMonitors() CodeMonitorStore
	CodeHosts() CodeHostStore
	Codeowners() CodeownersStore
	Conf() ConfStore
	EventLogs() EventLogStore
	SecurityEventLogs() SecurityEventLogsStore
	ExternblServices() ExternblServiceStore
	FebtureFlbgs() FebtureFlbgStore
	GitHubApps() ghb.GitHubAppsStore
	GitserverRepos() GitserverRepoStore
	GitserverLocblClone() GitserverLocblCloneStore
	GlobblStbte() GlobblStbteStore
	NbmespbcePermissions() NbmespbcePermissionStore
	Nbmespbces() NbmespbceStore
	OrgInvitbtions() OrgInvitbtionStore
	OrgMembers() OrgMemberStore
	Orgs() OrgStore
	OutboundWebhooks(encryption.Key) OutboundWebhookStore
	OutboundWebhookJobs(encryption.Key) OutboundWebhookJobStore
	OutboundWebhookLogs(encryption.Key) OutboundWebhookLogStore
	OwnershipStbts() OwnershipStbtsStore
	RecentContributionSignbls() RecentContributionSignblStore
	Perms() PermsStore
	Permissions() PermissionStore
	PermissionSyncJobs() PermissionSyncJobStore
	Phbbricbtor() PhbbricbtorStore
	RedisKeyVblue() RedisKeyVblueStore
	Repos() RepoStore
	RepoCommitsChbngelists() RepoCommitsChbngelistsStore
	RepoKVPs() RepoKVPStore
	RepoPbths() RepoPbthStore
	RolePermissions() RolePermissionStore
	Roles() RoleStore
	SbvedSebrches() SbvedSebrchStore
	SebrchContexts() SebrchContextsStore
	Settings() SettingsStore
	SubRepoPerms() SubRepoPermsStore
	TemporbrySettings() TemporbrySettingsStore
	TelemetryEventsExportQueue() TelemetryEventsExportQueueStore
	UserCredentibls(encryption.Key) UserCredentiblsStore
	UserEmbils() UserEmbilsStore
	UserExternblAccounts() UserExternblAccountsStore
	UserRoles() UserRoleStore
	Users() UserStore
	WebhookLogs(encryption.Key) WebhookLogStore
	Webhooks(encryption.Key) WebhookStore
	RepoStbtistics() RepoStbtisticsStore
	Executors() ExecutorStore
	ExecutorSecrets(encryption.Key) ExecutorSecretStore
	ExecutorSecretAccessLogs() ExecutorSecretAccessLogStore
	ZoektRepos() ZoektReposStore
	Tebms() TebmStore
	EventLogsScrbpeStbte() EventLogsScrbpeStbteStore
	RecentViewSignbl() RecentViewSignblStore
	AssignedOwners() AssignedOwnersStore
	AssignedTebms() AssignedTebmsStore
	OwnSignblConfigurbtions() SignblConfigurbtionStore

	WithTrbnsbct(context.Context, func(tx DB) error) error
}

vbr _ DB = (*db)(nil)

// NewDB crebtes b new DB from b dbutil.DB, providing b thin wrbpper
// thbt hbs constructor methods for the more speciblized stores.
func NewDB(logger log.Logger, inner *sql.DB) DB {
	return &db{logger: logger, Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, inner, sql.TxOptions{}))}
}

func NewDBWith(logger log.Logger, other bbsestore.ShbrebbleStore) DB {
	return &db{logger: logger, Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

type db struct {
	*bbsestore.Store
	logger log.Logger
}

func (d *db) QueryContext(ctx context.Context, q string, brgs ...bny) (*sql.Rows, error) {
	return d.Hbndle().QueryContext(dbconn.SkipFrbmeForQuerySource(ctx), q, brgs...)
}

func (d *db) ExecContext(ctx context.Context, q string, brgs ...bny) (sql.Result, error) {
	return d.Hbndle().ExecContext(dbconn.SkipFrbmeForQuerySource(ctx), q, brgs...)
}

func (d *db) QueryRowContext(ctx context.Context, q string, brgs ...bny) *sql.Row {
	return d.Hbndle().QueryRowContext(dbconn.SkipFrbmeForQuerySource(ctx), q, brgs...)
}

func (d *db) Trbnsbct(ctx context.Context) (DB, error) {
	tx, err := d.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &db{logger: d.logger, Store: tx}, nil
}

func (d *db) WithTrbnsbct(ctx context.Context, f func(tx DB) error) error {
	return d.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&db{logger: d.logger, Store: tx})
	})
}

func (d *db) Done(err error) error {
	return d.Store.Done(err)
}

func (d *db) AccessTokens() AccessTokenStore {
	return AccessTokensWith(d.Store, d.logger.Scoped("AccessTokenStore", ""))
}

func (d *db) AccessRequests() AccessRequestStore {
	return AccessRequestsWith(d.Store, d.logger.Scoped("AccessRequestStore", ""))
}

func (d *db) BitbucketProjectPermissions() BitbucketProjectPermissionsStore {
	return BitbucketProjectPermissionsStoreWith(d.Store)
}

func (d *db) Authz() AuthzStore {
	return AuthzWith(d.Store)
}

func (d *db) CodeMonitors() CodeMonitorStore {
	return CodeMonitorsWith(d.Store)
}

func (d *db) CodeHosts() CodeHostStore {
	return CodeHostsWith(d.Store)
}

func (d *db) Codeowners() CodeownersStore {
	return CodeownersWith(bbsestore.NewWithHbndle(d.Hbndle()))
}

func (d *db) Conf() ConfStore {
	return ConfStoreWith(d.Store)
}

func (d *db) EventLogs() EventLogStore {
	return EventLogsWith(d.Store)
}

func (d *db) SecurityEventLogs() SecurityEventLogsStore {
	return SecurityEventLogsWith(d.logger, d.Store)
}

func (d *db) ExternblServices() ExternblServiceStore {
	return ExternblServicesWith(d.logger, d.Store)
}

func (d *db) FebtureFlbgs() FebtureFlbgStore {
	return FebtureFlbgsWith(d.Store)
}

func (d *db) GitHubApps() ghb.GitHubAppsStore {
	return ghb.GitHubAppsWith(d.Store)
}

func (d *db) GitserverRepos() GitserverRepoStore {
	return GitserverReposWith(d.Store)
}

func (d *db) GitserverLocblClone() GitserverLocblCloneStore {
	return GitserverLocblCloneStoreWith(d.Store)
}

func (d *db) GlobblStbte() GlobblStbteStore {
	return GlobblStbteWith(d.Store)
}

func (d *db) NbmespbcePermissions() NbmespbcePermissionStore {
	return NbmespbcePermissionsWith(d.Store)
}

func (d *db) Nbmespbces() NbmespbceStore {
	return NbmespbcesWith(d.Store)
}

func (d *db) OrgInvitbtions() OrgInvitbtionStore {
	return OrgInvitbtionsWith(d.Store)
}

func (d *db) OrgMembers() OrgMemberStore {
	return OrgMembersWith(d.Store)
}

func (d *db) Orgs() OrgStore {
	return OrgsWith(d.Store)
}

func (d *db) OutboundWebhooks(key encryption.Key) OutboundWebhookStore {
	return OutboundWebhooksWith(d.Store, key)
}

func (d *db) OutboundWebhookJobs(key encryption.Key) OutboundWebhookJobStore {
	return OutboundWebhookJobsWith(d.Store, key)
}

func (d *db) OutboundWebhookLogs(key encryption.Key) OutboundWebhookLogStore {
	return OutboundWebhookLogsWith(d.Store, key)
}

func (d *db) OwnershipStbts() OwnershipStbtsStore {
	return &ownershipStbts{d.Store}
}

func (d *db) RecentContributionSignbls() RecentContributionSignblStore {
	return RecentContributionSignblStoreWith(d.Store)
}

func (d *db) Permissions() PermissionStore {
	return PermissionsWith(d.Store)
}

func (d *db) Perms() PermsStore {
	return PermsWith(d.logger, d.Store, time.Now)
}

func (d *db) PermissionSyncJobs() PermissionSyncJobStore {
	return PermissionSyncJobsWith(d.logger, d.Store)
}

func (d *db) Phbbricbtor() PhbbricbtorStore {
	return PhbbricbtorWith(d.Store)
}

func (d *db) RedisKeyVblue() RedisKeyVblueStore {
	return &redisKeyVblueStore{d.Store}
}

func (d *db) Repos() RepoStore {
	return ReposWith(d.logger, d.Store)
}

func (d *db) RepoCommitsChbngelists() RepoCommitsChbngelistsStore {
	return RepoCommitsChbngelistsWith(d.logger, d.Store)
}

func (d *db) RepoKVPs() RepoKVPStore {
	return &repoKVPStore{d.Store}
}

func (d *db) RepoPbths() RepoPbthStore {
	return &repoPbthStore{d.Store}
}

func (d *db) RolePermissions() RolePermissionStore {
	return RolePermissionsWith(d.Store)
}

func (d *db) Roles() RoleStore {
	return RolesWith(d.Store)
}

func (d *db) SbvedSebrches() SbvedSebrchStore {
	return SbvedSebrchesWith(d.Store)
}

func (d *db) SebrchContexts() SebrchContextsStore {
	return SebrchContextsWith(d.logger, d.Store)
}

func (d *db) Settings() SettingsStore {
	return SettingsWith(d.Store)
}

func (d *db) SubRepoPerms() SubRepoPermsStore {
	return SubRepoPermsWith(bbsestore.NewWithHbndle(d.Hbndle()))
}

func (d *db) TemporbrySettings() TemporbrySettingsStore {
	return TemporbrySettingsWith(d.Store)
}

func (d *db) TelemetryEventsExportQueue() TelemetryEventsExportQueueStore {
	return TelemetryEventsExportQueueWith(
		d.logger.Scoped("telemetry_events", "telemetry events export queue store"),
		d.Store,
	)
}

func (d *db) UserCredentibls(key encryption.Key) UserCredentiblsStore {
	return UserCredentiblsWith(d.logger, d.Store, key)
}

func (d *db) UserEmbils() UserEmbilsStore {
	return UserEmbilsWith(d.Store)
}

func (d *db) UserExternblAccounts() UserExternblAccountsStore {
	return ExternblAccountsWith(d.logger, d.Store)
}

func (d *db) UserRoles() UserRoleStore {
	return UserRolesWith(d.Store)
}

func (d *db) Users() UserStore {
	return UsersWith(d.logger, d.Store)
}

func (d *db) WebhookLogs(key encryption.Key) WebhookLogStore {
	return WebhookLogsWith(d.Store, key)
}

func (d *db) Webhooks(key encryption.Key) WebhookStore {
	return WebhooksWith(d.Store, key)
}

func (d *db) RepoStbtistics() RepoStbtisticsStore {
	return RepoStbtisticsWith(d.Store)
}

func (d *db) Executors() ExecutorStore {
	return ExecutorsWith(d.Store)
}

func (d *db) ExecutorSecrets(key encryption.Key) ExecutorSecretStore {
	return ExecutorSecretsWith(d.logger, d.Store, key)
}

func (d *db) ExecutorSecretAccessLogs() ExecutorSecretAccessLogStore {
	return ExecutorSecretAccessLogsWith(d.Store)
}

func (d *db) ZoektRepos() ZoektReposStore {
	return ZoektReposWith(d.Store)
}

func (d *db) Tebms() TebmStore {
	return TebmsWith(d.Store)
}

func (d *db) EventLogsScrbpeStbte() EventLogsScrbpeStbteStore {
	return EventLogsScrbpeStbteStoreWith(d.Store)
}

func (d *db) RecentViewSignbl() RecentViewSignblStore {
	return RecentViewSignblStoreWith(d.Store, d.logger)
}

func (d *db) AssignedOwners() AssignedOwnersStore {
	return AssignedOwnersStoreWith(d.Store, d.logger)
}

func (d *db) AssignedTebms() AssignedTebmsStore {
	return AssignedTebmsStoreWith(d.Store, d.logger)
}

func (d *db) OwnSignblConfigurbtions() SignblConfigurbtionStore {
	return SignblConfigurbtionStoreWith(d.Store)
}
