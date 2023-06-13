package gitresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type CachedLocationResolverFactory struct {
	repoStore       database.RepoStore
	gitserverClient gitserver.Client
}

func NewCachedLocationResolverFactory(repoStore database.RepoStore, gitserverClient gitserver.Client) *CachedLocationResolverFactory {
	return &CachedLocationResolverFactory{
		repoStore:       repoStore,
		gitserverClient: gitserverClient,
	}
}

func (f *CachedLocationResolverFactory) Create() *CachedLocationResolver {
	return newCachedLocationResolver(f.repoStore, f.gitserverClient)
}
