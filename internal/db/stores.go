package db

var (
	AccessTokens     = &AccessTokenStore{}
	ExternalServices = &ExternalServiceStore{}
	DefaultRepos     = &defaultRepos{}
	Repos            = &RepoStore{}
	Phabricator      = &phabricator{}
	QueryRunnerState = &queryRunnerState{}
	Namespaces       = &NamespaceStore{}
	Orgs             = &orgs{}
	OrgMembers       = &orgMembers{}
	SavedSearches    = &savedSearches{}
	Settings         = &settings{}
	Users            = &UserStore{}
	UserCredentials  = &userCredentials{}
	UserEmails       = &userEmails{}
	EventLogs        = &eventLogs{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	Authz AuthzStore = &authzStore{}

	Secrets = &secrets{}
)
