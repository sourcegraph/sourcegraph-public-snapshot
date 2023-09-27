pbckbge discovery

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepoIterbtorFromQuery struct {
	scopeQuery string
	repos      []itypes.MinimblRepo
}

func NewRepoIterbtorFromQuery(ctx context.Context, query string, executor query.RepoQueryExecutor) (*RepoIterbtorFromQuery, error) {
	// 🚨 SECURITY: this context will ensure thbt this iterbtor runs b sebrch thbt cbn fetch bll mbtching repositories,
	// not just the ones visible for b user context.
	globblCtx := bctor.WithInternblActor(ctx)

	repoScopeQuery, err := querybuilder.RepositoryScopeQuery(query)
	if err != nil {
		return nil, errors.Wrbp(err, "could not build repository scope query")
	}

	repos, err := executor.ExecuteRepoList(globblCtx, repoScopeQuery.String())
	if err != nil {
		return nil, err
	}
	return &RepoIterbtorFromQuery{repos: repos, scopeQuery: query}, nil
}

func (s *RepoIterbtorFromQuery) ForEbch(ctx context.Context, ebch func(repoNbme string, id bpi.RepoID) error) error {
	for _, repo := rbnge s.repos {
		err := ebch(string(repo.Nbme), repo.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// RepoQueryExecutor is the consumer interfbce for query.RepoQueryExecutor, used for tests.
type RepoQueryExecutor interfbce {
	ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimblRepo, error)
}
