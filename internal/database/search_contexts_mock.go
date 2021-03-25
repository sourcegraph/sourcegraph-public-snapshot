package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockSearchContexts struct {
	GetSearchContext                    func(ctx context.Context, opts GetSearchContextOptions) (*types.SearchContext, error)
	ListSearchContextsByUserID          func(ctx context.Context, userID int32) ([]*types.SearchContext, error)
	ListInstanceLevelSearchContexts     func(ctx context.Context) ([]*types.SearchContext, error)
	GetSearchContextRepositoryRevisions func(ctx context.Context, searchContextID int64) ([]*types.SearchContextRepositoryRevisions, error)
}
