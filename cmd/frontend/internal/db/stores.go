package db

var (
	AccessTokens = &accessTokens{}
	GlobalDeps   = &globalDeps{}
	Pkgs         = &pkgs{}
	Repos        = &repos{}
	Phabricator  = &phabricator{}
	UserActivity = &userActivity{} // DEPRECATED: use package useractivity instead (based on persisted redis cache)
	SavedQueries = &savedQueries{}
	Orgs         = &orgs{}
	OrgTags      = &orgTags{}
	OrgMembers   = &orgMembers{}
	Settings     = &settings{}
	Users        = &users{}
	UserEmails   = &userEmails{}
	UserTags     = &userTags{}
	SiteConfig   = &siteConfig{}
	CertCache    = &certCache{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	RegistryExtensions = &registryExtensions{}
)
