package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

type batchSpecWorkspaceResolver struct {
	store *store.Store
	node  *service.RepoWorkspace
}

var _ graphqlbackend.BatchSpecWorkspaceResolver = &batchSpecWorkspaceResolver{}

func (r *batchSpecWorkspaceResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.NewRepositoryResolver(r.store.DB(), r.node.Repo), nil
}

func (r *batchSpecWorkspaceResolver) Branch(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	repo, _ := r.Repository(ctx)
	return graphqlbackend.NewGitRefResolver(repo, r.node.Branch, graphqlbackend.GitObjectID(r.node.Commit)), nil
}

func (r *batchSpecWorkspaceResolver) Path() string {
	if r.node.Path == "" {
		return "/"
	}
	return r.node.Path
}

func (r *batchSpecWorkspaceResolver) OnlyFetchWorkspace() bool {
	return r.node.OnlyFetchWorkspace
}

func (r *batchSpecWorkspaceResolver) Steps() []graphqlbackend.BatchSpecWorkspaceStepResolver {
	resolvers := make([]graphqlbackend.BatchSpecWorkspaceStepResolver, 0, len(r.node.Steps))
	for _, step := range r.node.Steps {
		resolvers = append(resolvers, &batchSpecWorkspaceStepResolver{step})
	}
	return resolvers
}

type batchSpecWorkspaceStepResolver struct {
	step batcheslib.Step
}

func (r *batchSpecWorkspaceStepResolver) Command() string {
	return r.step.Run
}

func (r *batchSpecWorkspaceStepResolver) Container() string {
	return r.step.Container
}
