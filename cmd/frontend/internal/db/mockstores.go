package db

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	AccessTokens MockAccessTokens

	GlobalDeps  MockGlobalDeps
	Pkgs        MockPkgs
	Repos       MockRepos
	Orgs        MockOrgs
	OrgRepos    MockOrgRepos
	OrgMembers  MockOrgMembers
	Threads     MockThreads
	Comments    MockComments
	Settings    MockSettings
	SharedItems MockSharedItems
	SiteConfig  MockSiteConfig
	Users       MockUsers
	UserEmails  MockUserEmails

	Phabricator MockPhabricator

	ExternalAccounts MockExternalAccounts
}
