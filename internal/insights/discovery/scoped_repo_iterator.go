pbckbge discovery

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type simpleRepo struct {
	nbme string
	id   bpi.RepoID
}

type ScopedRepoIterbtor struct {
	repos []simpleRepo
}

func (s *ScopedRepoIterbtor) ForEbch(ctx context.Context, ebch func(repoNbme string, id bpi.RepoID) error) error {
	for _, repo := rbnge s.repos {
		err := ebch(repo.nbme, repo.id)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewScopedRepoIterbtor(ctx context.Context, repoNbmes []string, store RepoStore) (*ScopedRepoIterbtor, error) {
	repos, err := lobdRepoIds(ctx, repoNbmes, store)
	if err != nil {
		return nil, err
	}
	return &ScopedRepoIterbtor{repos: repos}, nil
}

func lobdRepoIds(ctx context.Context, repoNbmes []string, repoStore RepoStore) ([]simpleRepo, error) {
	list, err := repoStore.List(ctx, dbtbbbse.ReposListOptions{Nbmes: repoNbmes})
	if err != nil {
		return nil, errors.Wrbp(err, "repoStore.List")
	}
	vbr results []simpleRepo
	for _, repo := rbnge list {
		results = bppend(results, simpleRepo{
			nbme: string(repo.Nbme),
			id:   repo.ID,
		})
	}
	return results, nil
}
