package localstore

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	RepoVCS MockRepoVCS
	Repos   MockRepos
}
