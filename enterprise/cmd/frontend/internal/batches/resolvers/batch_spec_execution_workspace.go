package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.BatchSpecExecutionWorkspaceResolver = &batchSpecExecutionWorkspaceResolver{}

type batchSpecExecutionWorkspaceResolver struct {
	workspacePath string
	repoResolver  *graphqlbackend.RepositoryResolver
}

func (r *batchSpecExecutionWorkspaceResolver) Repository(ctx context.Context) *graphqlbackend.RepositoryResolver {
	return r.repoResolver
}

func (r *batchSpecExecutionWorkspaceResolver) Path() string {
	return r.workspacePath
}
