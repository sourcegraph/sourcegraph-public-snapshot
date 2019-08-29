package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// TODO!(sqs) document that this is set by enterprise, handle when it's not set, and rethink the
// architecture here.
var ForceRefreshRepositoryThreads func(context.Context, api.RepoID, api.ExternalRepoSpec) error

func (r *schemaResolver) ForceRefreshRepositoryThreads(ctx context.Context, arg *struct{ Repository graphql.ID }) (*RepositoryResolver, error) {
	repo, err := RepositoryByID(ctx, arg.Repository)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): security, also this is only provided by enterprise
	if err := ForceRefreshRepositoryThreads(ctx, repo.repo.ID, repo.repo.ExternalRepo); err != nil {
		return nil, err
	}
	return repo, nil
}
