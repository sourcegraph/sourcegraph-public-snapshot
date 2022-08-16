package resolvers

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
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

func (r *resolvedBatchSpecWorkspaceResolver) Branch() *graphqlbackend.GitRefResolver {
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
		r.repoResolver = graphqlbackend.NewRepositoryResolver(r.store.DatabaseDB(), r.workspace.Repo)
	})

	return r.repoResolver
}
