package localstore

import srcstore "sourcegraph.com/sourcegraph/srclib/store"

var (
	Defs         = &defs{}
	GlobalDeps   = &globalDeps{}
	GlobalRefs   = &globalRefs{}
	Graph        srcstore.MultiRepoStoreImporterIndexer
	Queue        = &instrumentedQueue{}
	RepoConfigs  = &repoConfigs{}
	RepoStatuses = &repoStatuses{}
	RepoVCS      = &repoVCS{}
	Repos        = &repos{}
)
