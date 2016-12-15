package localstore

import (
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	DeprecatedGlobalRefs DeprecatedMockGlobalRefs
	Graph                srcstore.MockMultiRepoStore
	Queue                MockQueue
	RepoConfigs          MockRepoConfigs
	RepoVCS              MockRepoVCS
	Repos                MockRepos
}
