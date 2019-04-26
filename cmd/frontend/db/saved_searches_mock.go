package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSavedSearches struct {
	ListAll func(ctx context.Context) ([]api.SavedQuerySpecAndConfig, error)
	Create  func(ctx context.Context, newSavedSearch *types.SavedSearch) (*types.SavedSearch, error)
	Update  func(ctx context.Context, savedSearch *types.SavedSearch) (*types.SavedSearch, error)
	Delete  func(ctx context.Context, id string) error
}
