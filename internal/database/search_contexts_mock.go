package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockSearchContexts struct {
	GetSearchContext                    func(ctx context.Context, opts GetSearchContextOptions) (*types.SearchContext, error)
	GetSearchContextRepositoryRevisions func(ctx context.Context, searchContextID int64) ([]*types.SearchContextRepositoryRevisions, error)
	ListSearchContexts                  func(ctx context.Context, pageOpts ListSearchContextsPageOptions, opts ListSearchContextsOptions) ([]*types.SearchContext, error)
	CountSearchContexts                 func(ctx context.Context, opts ListSearchContextsOptions) (int32, error)
}
