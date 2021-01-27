package database

// Global reference to database stores using the global dbconn.Global connection handle.
// Deprecated: Use store constructors instead.
var (
	GlobalAccessTokens                = &AccessTokenStore{}
	GlobalExternalServices            = &ExternalServiceStore{}
	GlobalDefaultRepos                = &DefaultRepoStore{}
	GlobalRepos                       = &RepoStore{}
	GlobalPhabricator                 = &PhabricatorStore{}
	GlobalQueryRunnerState            = &QueryRunnerStateStore{}
	GlobalNamespaces                  = &NamespaceStore{}
	GlobalOrgs                        = &OrgStore{}
	GlobalOrgMembers                  = &OrgMemberStore{}
	GlobalSavedSearches               = &SavedSearchStore{}
	GlobalSettings                    = &SettingStore{}
	GlobalUsers                       = &UserStore{}
	GlobalUserCredentials             = &UserCredentialsStore{}
	GlobalUserEmails                  = &UserEmailsStore{}
	GlobalEventLogs                   = &EventLogStore{}
	GlobalSurveyResponses             = &SurveyResponseStore{}
	GlobalExternalAccounts            = &UserExternalAccountsStore{}
	GlobalOrgInvitations              = &OrgInvitationStore{}
	GlobalAuthz            AuthzStore = &authzStore{}
)
