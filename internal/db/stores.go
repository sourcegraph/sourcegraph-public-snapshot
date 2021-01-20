package db

var (
	AccessTokens     = &accessTokens{}
	ExternalServices = &ExternalServiceStore{}
	DefaultRepos     = &defaultRepos{}
	Repos            = &RepoStore{}
	Phabricator      = &phabricator{}
	QueryRunnerState = &queryRunnerState{}
	Namespaces       = &namespaces{}
	Orgs             = &orgs{}
	OrgMembers       = &orgMembers{}
	SavedSearches    = &savedSearches{}
	Settings         = &settings{}
	Users            = &users{}
	UserCredentials  = &userCredentials{}
	UserEmails       = &userEmails{}
	EventLogs        = &eventLogs{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	Authz AuthzStore = &authzStore{}

	Secrets = &secrets{}
)
