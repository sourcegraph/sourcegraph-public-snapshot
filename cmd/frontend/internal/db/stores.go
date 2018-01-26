package db

var (
	GlobalDeps   = &globalDeps{}
	Pkgs         = &pkgs{}
	RepoVCS      = &repoVCS{}
	Repos        = &repos{}
	Phabricator  = &phabricator{}
	OrgRepos     = &orgRepos{}
	Threads      = &threads{}
	Comments     = &comments{}
	UserActivity = &userActivity{}
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
)
