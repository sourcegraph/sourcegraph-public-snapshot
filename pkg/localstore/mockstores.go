package localstore

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	GlobalDeps  MockGlobalDeps
	Pkgs        MockPkgs
	RepoVCS     MockRepoVCS
	Repos       MockRepos
	Orgs        MockOrgs
	OrgRepos    MockOrgRepos
	OrgMembers  MockOrgMembers
	Threads     MockThreads
	Comments    MockComments
	Settings    MockSettings
	SharedItems MockSharedItems
	Users       MockUsers
}
