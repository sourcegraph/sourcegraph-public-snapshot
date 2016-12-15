package localstore

import srcstore "sourcegraph.com/sourcegraph/srclib/store"

var (
	GlobalDeps           = &globalDeps{}
	DeprecatedGlobalRefs = &deprecatedGlobalRefs{}
	GlobalRefs           = &globalRefs{}
	Graph                srcstore.MultiRepoStoreImporterIndexer
	Queue                = &instrumentedQueue{}
	RepoConfigs          = &repoConfigs{}
	RepoVCS              = &repoVCS{}
	Repos                = &repos{}
	UserInvites          = &userInvites{}
)
