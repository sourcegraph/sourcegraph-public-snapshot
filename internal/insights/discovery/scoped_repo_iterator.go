package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type simpleRepo struct {
	name string
	id   api.RepoID
}

type ScopedRepoIterator struct {
	repos []simpleRepo
}

func (s *ScopedRepoIterator) ForEach(ctx context.Context, each func(repoName string, id api.RepoID) error) error {
	for _, repo := range s.repos {
		err := each(repo.name, repo.id)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewScopedRepoIterator(ctx context.Context, repoNames []string, store RepoStore) (*ScopedRepoIterator, error) {
	repos, err := loadRepoIds(ctx, repoNames, store)
	if err != nil {
		return nil, err
	}
	return &ScopedRepoIterator{repos: repos}, nil
}

func loadRepoIds(ctx context.Context, repoNames []string, repoStore RepoStore) ([]simpleRepo, error) {
	list, err := repoStore.List(ctx, database.ReposListOptions{Names: repoNames})
	if err != nil {
		return nil, errors.Wrap(err, "repoStore.List")
	}
	var results []simpleRepo
	for _, repo := range list {
		results = append(results, simpleRepo{
			name: string(repo.Name),
			id:   repo.ID,
		})
	}
	return results, nil
}
