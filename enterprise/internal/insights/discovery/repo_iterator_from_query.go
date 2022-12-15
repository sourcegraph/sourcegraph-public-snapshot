package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/internal/api"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoIteratorFromQuery struct {
	scopeQuery string
	repos      []itypes.MinimalRepo
}

func NewRepoIteratorFromQuery(ctx context.Context, query string, repoQueryExecutor *query.StreamingRepoQueryExecutor) (*RepoIteratorFromQuery, error) {
	repos, err := repoQueryExecutor.ExecuteRepoList(ctx, query)
	if err != nil {
		return nil, err
	}
	return &RepoIteratorFromQuery{repos: repos, scopeQuery: query}, nil
}

func (s *RepoIteratorFromQuery) ForEach(ctx context.Context, each func(repoName string, id api.RepoID) error) error {
	for _, repo := range s.repos {
		err := each(string(repo.Name), repo.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
