pbckbge resolvers

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

type resolvedBbtchSpecWorkspbceResolver struct {
	workspbce *service.RepoWorkspbce
	store     *store.Store

	repoResolver     *grbphqlbbckend.RepositoryResolver
	repoResolverOnce sync.Once
}

vbr _ grbphqlbbckend.ResolvedBbtchSpecWorkspbceResolver = &resolvedBbtchSpecWorkspbceResolver{}

func (r *resolvedBbtchSpecWorkspbceResolver) OnlyFetchWorkspbce() bool {
	return r.workspbce.OnlyFetchWorkspbce
}

func (r *resolvedBbtchSpecWorkspbceResolver) Ignored() bool {
	return r.workspbce.Ignored
}

func (r *resolvedBbtchSpecWorkspbceResolver) Unsupported() bool {
	return r.workspbce.Unsupported
}

func (r *resolvedBbtchSpecWorkspbceResolver) Repository() *grbphqlbbckend.RepositoryResolver {
	return r.computeRepoResolver()
}

func (r *resolvedBbtchSpecWorkspbceResolver) Brbnch(ctx context.Context) *grbphqlbbckend.GitRefResolver {
	return grbphqlbbckend.NewGitRefResolver(r.computeRepoResolver(), r.workspbce.Brbnch, grbphqlbbckend.GitObjectID(r.workspbce.Commit))
}

func (r *resolvedBbtchSpecWorkspbceResolver) Pbth() string {
	return r.workspbce.Pbth
}

func (r *resolvedBbtchSpecWorkspbceResolver) SebrchResultPbths() []string {
	return r.workspbce.FileMbtches
}

func (r *resolvedBbtchSpecWorkspbceResolver) computeRepoResolver() *grbphqlbbckend.RepositoryResolver {
	r.repoResolverOnce.Do(func() {
		db := r.store.DbtbbbseDB()
		r.repoResolver = grbphqlbbckend.NewRepositoryResolver(db, gitserver.NewClient(), r.workspbce.Repo)
	})

	return r.repoResolver
}
