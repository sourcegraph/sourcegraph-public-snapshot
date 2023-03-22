package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type CachedLocationResolverFactory struct {
	cloneURLToRepoName CloneURLToRepoNameFunc
	repoStore          database.RepoStore
	gitserverClient    gitserver.Client
}

func NewCachedLocationResolverFactory(cloneURLToRepoName CloneURLToRepoNameFunc, repoStore database.RepoStore, gitserverClient gitserver.Client) *CachedLocationResolverFactory {
	return &CachedLocationResolverFactory{
		cloneURLToRepoName: cloneURLToRepoName,
		repoStore:          repoStore,
		gitserverClient:    gitserverClient,
	}
}

func (f *CachedLocationResolverFactory) Create() *CachedLocationResolver {
	return NewCachedLocationResolver(f.cloneURLToRepoName, f.repoStore, f.gitserverClient)
}
