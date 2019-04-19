package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type MockSavedSearches struct {
	ListAll func(ctx context.Context) ([]api.SavedQuerySpecAndConfig, error)
	Create  func(ctx context.Context, description, query string, notify, notifySlack bool, ownerKind string, userID, orgID *int32) (*api.ConfigSavedQuery, error)
	Update  func(ctx context.Context, id string, description, query string, notify, notifySlack bool, ownerKind string, userID, orgID *int32) (*api.ConfigSavedQuery, error)
	Delete  func(ctx context.Context, id string) error
}
