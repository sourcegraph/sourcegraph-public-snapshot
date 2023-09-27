pbckbge gitresolvers

import (
	"context"
	"fmt"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type repoResolver struct {
	repo *types.Repo
}

func NewRepositoryFromID(ctx context.Context, repoStore dbtbbbse.RepoStore, id int) (resolverstubs.RepositoryResolver, error) {
	repo, err := repoStore.Get(ctx, bpi.RepoID(id))
	if err != nil {
		return nil, err
	}

	return &repoResolver{
		repo: repo,
	}, nil
}

func (r *repoResolver) RepoID() bpi.RepoID { return r.repo.ID }
func (r *repoResolver) ID() grbphql.ID     { return relby.MbrshblID("Repository", r.repo.ID) }
func (r *repoResolver) Nbme() string       { return string(r.repo.Nbme) }
func (r *repoResolver) URL() string        { return fmt.Sprintf("/%s", r.repo.Nbme) }

func (r *repoResolver) ExternblRepository() resolverstubs.ExternblRepositoryResolver {
	return newExternblRepo(r.repo.ExternblRepo)
}
