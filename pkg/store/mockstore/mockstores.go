package mockstore

import (
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

// Stores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type Stores struct {
	Defs         Defs
	GlobalDeps   GlobalDeps
	GlobalRefs   GlobalRefs
	Graph        srcstore.MockMultiRepoStore
	Queue        Queue
	RepoConfigs  RepoConfigs
	RepoStatuses RepoStatuses
	RepoVCS      RepoVCS
	Repos        Repos
}
