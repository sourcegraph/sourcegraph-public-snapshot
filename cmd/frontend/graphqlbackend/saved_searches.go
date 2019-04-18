package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func (r *schemaResolver) CreateSavedSearch(ctx context.Context, args *struct {
	Description string
	Query       string
	NotifyOwner bool
	NotifySlack bool
	OwnerKind   string
	OrgID       *int32
	UserID      *int32
}) (*EmptyResponse, error) {
	err := db.SavedSearches.Create(ctx, args.Description, args.Query, args.NotifyOwner, args.NotifySlack, args.OwnerKind, args.UserID, args.OrgID)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) UpdateSavedSearch(ctx context.Context, args *struct {
	ID          string
	Description string
	Query       string
	NotifyOwner bool
	NotifySlack bool
	OwnerKind   string
	OrgID       *int32
	UserID      *int32
}) (*EmptyResponse, error) {
	err := db.SavedSearches.Update(ctx, args.ID, args.Description, args.Query, args.NotifyOwner, args.NotifySlack, args.OwnerKind, args.UserID, args.OrgID)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) DeleteSavedSearch(ctx context.Context, args *struct {
	ID string
}) (*EmptyResponse, error) {
	err := db.SavedSearches.Delete(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
