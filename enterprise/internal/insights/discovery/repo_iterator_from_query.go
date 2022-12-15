package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type RepoIteratorFromQuery struct {
	scopeQuery string
	repos      []simpleRepo
}

func NewRepoIteratorFromQuery(ctx context.Context, query string, repoQueryExecutor *query.StreamingRepoQueryExecutor) (*RepoIteratorFromQuery, error) {
	repos, err := loadRepoIdsFromSearch(ctx, query, repoQueryExecutor)
	if err != nil {
		return nil, err
	}
	return &RepoIteratorFromQuery{repos: repos, scopeQuery: query}, nil
}

func (s *RepoIteratorFromQuery) ForEach(ctx context.Context, each func(repoName string, id api.RepoID) error) error {
	for _, repo := range s.repos {
		err := each(repo.name, repo.id)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadRepoIdsFromSearch(ctx context.Context, query string, executor *query.StreamingRepoQueryExecutor) ([]simpleRepo, error) {
	repos, err := executor.ExecuteRepoList(ctx, query)
	if err != nil {
		return nil, err
	}
	var results []simpleRepo
	for _, repo := range repos {
		results = append(results, simpleRepo{
			name: string(repo.Name),
			id:   repo.ID,
		})
	}
	return results, nil
}
