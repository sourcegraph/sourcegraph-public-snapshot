package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockSearchContexts struct {
	GetSearchContext                    func(ctx context.Context, opts GetSearchContextOptions) (*types.SearchContext, error)
	GetSearchContextRepositoryRevisions func(ctx context.Context, searchContextID int64) ([]*types.SearchContextRepositoryRevisions, error)
	GetAllRevisionsForRepo              func(ctx context.Context, repoID int32) ([]string, error)
}
