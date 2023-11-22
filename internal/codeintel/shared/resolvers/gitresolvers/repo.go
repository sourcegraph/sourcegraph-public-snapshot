package gitresolvers

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type repoResolver struct {
	repo *types.Repo
}

func NewRepositoryFromID(ctx context.Context, repoStore database.RepoStore, id int) (resolverstubs.RepositoryResolver, error) {
	repo, err := repoStore.Get(ctx, api.RepoID(id))
	if err != nil {
		return nil, err
	}

	return &repoResolver{
		repo: repo,
	}, nil
}

func (r *repoResolver) RepoID() api.RepoID { return r.repo.ID }
func (r *repoResolver) ID() graphql.ID     { return relay.MarshalID("Repository", r.repo.ID) }
func (r *repoResolver) Name() string       { return string(r.repo.Name) }
func (r *repoResolver) URL() string        { return fmt.Sprintf("/%s", r.repo.Name) }

func (r *repoResolver) ExternalRepository() resolverstubs.ExternalRepositoryResolver {
	return newExternalRepo(r.repo.ExternalRepo)
}
