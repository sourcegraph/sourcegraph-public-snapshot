package database

// Global reference to database stores using the global dbconn.Global connection handle.
// Deprecated: Use store constructors instead.
var (
	GlobalExternalServices            = &ExternalServiceStore{}
	GlobalDefaultRepos                = &DefaultRepoStore{}
	GlobalRepos                       = &RepoStore{}
	GlobalPhabricator                 = &PhabricatorStore{}
	GlobalQueryRunnerState            = &QueryRunnerStateStore{}
	GlobalNamespaces                  = &NamespaceStore{}
	GlobalOrgs                        = &OrgStore{}
	GlobalOrgMembers                  = &OrgMemberStore{}
	GlobalSettings                    = &SettingStore{}
	GlobalUsers                       = &UserStore{}
	GlobalUserCredentials             = &UserCredentialsStore{}
	GlobalUserEmails                  = &UserEmailsStore{}
	GlobalEventLogs                   = &EventLogStore{}
	GlobalExternalAccounts            = &UserExternalAccountsStore{}
	GlobalAuthz            AuthzStore = &authzStore{}
)
