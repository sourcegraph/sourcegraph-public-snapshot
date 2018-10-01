package db

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	AccessTokens MockAccessTokens

	DiscussionThreads         MockDiscussionThreads
	DiscussionComments        MockDiscussionComments
	DiscussionMailReplyTokens MockDiscussionMailReplyTokens

	GlobalDeps MockGlobalDeps
	Pkgs       MockPkgs
	Repos      MockRepos
	Orgs       MockOrgs
	OrgMembers MockOrgMembers
	Settings   MockSettings
	SiteConfig MockSiteConfig
	Users      MockUsers
	UserEmails MockUserEmails

	Phabricator MockPhabricator

	ExternalAccounts MockExternalAccounts

	OrgInvitations MockOrgInvitations
}
