package localstore

import (
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	Defs                 MockDefs
	GlobalDeps           MockGlobalDeps
	DeprecatedGlobalRefs DeprecatedMockGlobalRefs
	Graph                srcstore.MockMultiRepoStore
	Queue                MockQueue
	RepoConfigs          MockRepoConfigs
	RepoStatuses         MockRepoStatuses
	RepoVCS              MockRepoVCS
	Repos                MockRepos
}
