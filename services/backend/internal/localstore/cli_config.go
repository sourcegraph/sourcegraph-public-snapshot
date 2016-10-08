package localstore

import srcstore "sourcegraph.com/sourcegraph/srclib/store"

var (
	Defs                 = &defs{} // TODO: remove defs service (replaced by new xlang based globalDefs service)
	GlobalDeps           = &globalDeps{}
	DeprecatedGlobalRefs = &deprecatedGlobalRefs{}
	GlobalRefs           = &globalRefs{}
	Graph                srcstore.MultiRepoStoreImporterIndexer
	Queue                = &instrumentedQueue{}
	RepoConfigs          = &repoConfigs{}
	RepoStatuses         = &repoStatuses{}
	RepoVCS              = &repoVCS{}
	Repos                = &repos{}
	UserInvites          = &userInvites{}
)
