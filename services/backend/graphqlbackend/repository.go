package graphqlbackend

import (
	"context"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

type repositoryResolver struct {
	nodeBase
	repo *sourcegraph.Repo
}

func repositoryByID(ctx context.Context, id graphql.ID) (nodeResolver, error) {
	var repoID int32
	if err := relay.UnmarshalSpec(id, &repoID); err != nil {
		return nil, err
	}
	repo, err := localstore.Repos.Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func (r *repositoryResolver) ToRepository() (*repositoryResolver, bool) {
	return r, true
}

func (r *repositoryResolver) ID() graphql.ID {
	return relay.MarshalID("Repository", r.repo.ID)
}

func (r *repositoryResolver) URI() string {
	return r.repo.URI
}

func (r *repositoryResolver) Commit(ctx context.Context, args *struct{ Rev string }) (*commitResolver, error) {
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
		Rev:  args.Rev,
	})
	if err != nil {
		return nil, err
	}
	return &commitResolver{commit: commitSpec{r.repo.ID, rev.CommitID}}, nil
}

func (r *repositoryResolver) Latest(ctx context.Context) (*commitResolver, error) {
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
	})
	if err != nil {
		return nil, err
	}
	return &commitResolver{commit: commitSpec{r.repo.ID, rev.CommitID}}, nil
}
