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
	UserOrOrg   string
	OrgID       *int32
	UserID      *int32
}) (*EmptyResponse, error) {
	err := db.SavedSearches.Create(ctx, args.Description, args.Query, args.NotifyOwner, args.NotifySlack, args.UserOrOrg, args.UserID, args.OrgID)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
