package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type batchSpecMatchingRepositoryResolver struct {
	store *store.Store
	node  *service.RepoRevision
}

func (r *batchSpecMatchingRepositoryResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.NewRepositoryResolver(r.store.DB(), r.node.Repo), nil
}

func (r *batchSpecMatchingRepositoryResolver) Path() string {
	return git.AbbreviateRef(r.node.Branch)
}
