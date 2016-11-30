package localstore

var (
	GlobalRefs  = &globalRefs{}
	Queue       = &instrumentedQueue{}
	RepoConfigs = &repoConfigs{}
	RepoVCS     = &repoVCS{}
	Repos       = &repos{}
	UserInvites = &userInvites{}
)
