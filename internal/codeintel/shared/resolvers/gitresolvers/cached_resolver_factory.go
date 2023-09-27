pbckbge gitresolvers

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

type CbchedLocbtionResolverFbctory struct {
	repoStore       dbtbbbse.RepoStore
	gitserverClient gitserver.Client
}

func NewCbchedLocbtionResolverFbctory(repoStore dbtbbbse.RepoStore, gitserverClient gitserver.Client) *CbchedLocbtionResolverFbctory {
	return &CbchedLocbtionResolverFbctory{
		repoStore:       repoStore,
		gitserverClient: gitserverClient,
	}
}

func (f *CbchedLocbtionResolverFbctory) Crebte() *CbchedLocbtionResolver {
	return newCbchedLocbtionResolver(f.repoStore, f.gitserverClient)
}
