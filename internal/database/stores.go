package database

// Global reference to database stores using the global dbconn.Global connection handle.
// Deprecated: Use store constructors instead.
var (
	GlobalExternalServices            = &ExternalServiceStore{}
	GlobalIndexableRepos              = &IndexableRepoStore{}
	GlobalRepos                       = &RepoStore{}
	GlobalOrgs                        = &OrgStore{}
	GlobalSettings                    = &SettingStore{}
	GlobalUsers                       = &UserStore{}
	GlobalUserCredentials             = &UserCredentialsStore{}
	GlobalUserEmails                  = &UserEmailsStore{}
	GlobalExternalAccounts            = &UserExternalAccountsStore{}
	GlobalAuthz            AuthzStore = &authzStore{}
)
