package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type resolvedBatchSpecWorkspaceResolver struct {
	workspace *service.RepoWorkspace
	store     *store.Store

	repoResolver     *graphqlbackend.RepositoryResolver
	repoResolverOnce sync.Once
}

var _ graphqlbackend.ResolvedBatchSpecWorkspaceResolver = &resolvedBatchSpecWorkspaceResolver{}

func (r *resolvedBatchSpecWorkspaceResolver) OnlyFetchWorkspace() bool {
	return r.workspace.OnlyFetchWorkspace
}

func (r *resolvedBatchSpecWorkspaceResolver) Ignored() bool {
	return r.workspace.Ignored
}

func (r *resolvedBatchSpecWorkspaceResolver) Unsupported() bool {
	return r.workspace.Unsupported
}

func (r *resolvedBatchSpecWorkspaceResolver) Repository() *graphqlbackend.RepositoryResolver {
	return r.computeRepoResolver()
}

func (r *resolvedBatchSpecWorkspaceResolver) Branch(ctx context.Context) *graphqlbackend.GitRefResolver {
	return graphqlbackend.NewGitRefResolver(r.computeRepoResolver(), r.workspace.Branch, graphqlbackend.GitObjectID(r.workspace.Commit))
}

func (r *resolvedBatchSpecWorkspaceResolver) Path() string {
	return r.workspace.Path
}

func (r *resolvedBatchSpecWorkspaceResolver) SearchResultPaths() []string {
	return r.workspace.FileMatches
}

func (r *resolvedBatchSpecWorkspaceResolver) computeRepoResolver() *graphqlbackend.RepositoryResolver {
	r.repoResolverOnce.Do(func() {
		db := r.store.DatabaseDB()
		r.repoResolver = graphqlbackend.NewRepositoryResolver(db, gitserver.NewClient("graphql.batches.resolvedworkspacerepo"), r.workspace.Repo)
	})

	return r.repoResolver
}
