package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoIteratorFromQuery struct {
	scopeQuery string
	repos      []itypes.MinimalRepo
}

func NewRepoIteratorFromQuery(ctx context.Context, query string, executor query.RepoQueryExecutor) (*RepoIteratorFromQuery, error) {
	// 🚨 SECURITY: this context will ensure that this iterator runs a search that can fetch all matching repositories,
	// not just the ones visible for a user context.
	globalCtx := actor.WithInternalActor(ctx)

	repoScopeQuery, err := querybuilder.RepositoryScopeQuery(query)
	if err != nil {
		return nil, errors.Wrap(err, "could not build repository scope query")
	}

	repos, err := executor.ExecuteRepoList(globalCtx, repoScopeQuery.String())
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

// RepoQueryExecutor is the consumer interface for query.RepoQueryExecutor, used for tests.
type RepoQueryExecutor interface {
	ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimalRepo, error)
}
