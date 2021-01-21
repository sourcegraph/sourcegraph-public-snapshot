package db

var (
	AccessTokens     = &AccessTokenStore{}
	ExternalServices = &ExternalServiceStore{}
	DefaultRepos     = &DefaultRepoStore{}
	Repos            = &RepoStore{}
	Phabricator      = &PhabricatorStore{}
	QueryRunnerState = &QueryRunnerStateStore{}
	Namespaces       = &NamespaceStore{}
	Orgs             = &OrgStore{}
	OrgMembers       = &OrgMemberStore{}
	SavedSearches    = &SavedSearchStore{}
	Settings         = &SettingStore{}
	Users            = &UserStore{}
	UserCredentials  = &UserCredentialsStore{}
	UserEmails       = &UserEmailsStore{}
	UserPublicRepos  = &UserPublicRepoStore{}
	EventLogs        = &EventLogStore{}

	SurveyResponses = &SurveyResponseStore{}

	ExternalAccounts = &UserExternalAccountsStore{}

	OrgInvitations = &OrgInvitationStore{}

	Authz AuthzStore = &authzStore{}
)
