package db

var (
	AccessTokens = &accessTokens{}
	GlobalDeps   = &globalDeps{}
	Pkgs         = &pkgs{}
	Repos        = &repos{}
	Phabricator  = &phabricator{}
	OrgRepos     = &orgRepos{}
	Threads      = &threads{}
	Comments     = &comments{}
	UserActivity = &userActivity{} // DEPRECATED: use package useractivity instead (based on persisted redis cache)
	SavedQueries = &savedQueries{}
	SharedItems  = &sharedItems{}
	Orgs         = &orgs{}
	OrgTags      = &orgTags{}
	OrgMembers   = &orgMembers{}
	Settings     = &settings{}
	Users        = &users{}
	UserEmails   = &userEmails{}
	UserTags     = &userTags{}
	SiteConfig   = &siteConfig{}
	CertCache    = &certCache{}
)
