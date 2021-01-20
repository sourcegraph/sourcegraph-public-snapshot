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
	UserCredentials  = &userCredentials{}
	UserEmails       = &UserEmailsStore{}
	UserPublicRepos  = &UserPublicRepoStore{}
	EventLogs        = &EventLogStore{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	Authz AuthzStore = &authzStore{}

	Secrets = &secrets{}
)
