package localstore

import srcstore "sourcegraph.com/sourcegraph/srclib/store"

var (
	DeprecatedGlobalRefs = &deprecatedGlobalRefs{}
	GlobalRefs           = &globalRefs{}
	Graph                srcstore.MultiRepoStoreImporterIndexer
	Queue                = &instrumentedQueue{}
	RepoVCS              = &repoVCS{}
	Repos                = &repos{}
	UserInvites          = &userInvites{}
)
