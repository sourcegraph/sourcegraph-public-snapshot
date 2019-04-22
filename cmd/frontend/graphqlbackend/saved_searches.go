package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) CreateSavedSearch(ctx context.Context, args *struct {
	Description string
	Query       string
	NotifyOwner bool
	NotifySlack bool
	OwnerKind   string
	OrgID       *int32
	UserID      *int32
}) (*savedQueryResolver, error) {
	configSavedQuery, err := db.SavedSearches.Create(ctx, &types.SavedSearch{Description: args.Description, Query: args.Query, Notify: args.NotifyOwner, NotifySlack: args.NotifySlack, OwnerKind: args.OwnerKind, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, err
	}

	sq := &savedQueryResolver{
		key:         configSavedQuery.Key,
		description: configSavedQuery.Description,
		query:       configSavedQuery.Query,
		notify:      configSavedQuery.Notify,
		notifySlack: configSavedQuery.NotifySlack,
		ownerKind:   configSavedQuery.OwnerKind,
		userID:      configSavedQuery.UserID,
		orgID:       configSavedQuery.OrgID,
	}
	return sq, nil
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
}) (*savedQueryResolver, error) {
	configSavedQuery, err := db.SavedSearches.Update(ctx, &types.SavedSearch{ID: args.ID, Description: args.Description, Query: args.Query, Notify: args.NotifyOwner, NotifySlack: args.NotifySlack, OwnerKind: args.OwnerKind, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, err
	}
	sq := &savedQueryResolver{
		key:         configSavedQuery.Key,
		description: configSavedQuery.Description,
		query:       configSavedQuery.Query,
		notify:      configSavedQuery.Notify,
		notifySlack: configSavedQuery.NotifySlack,
		ownerKind:   configSavedQuery.OwnerKind,
		userID:      configSavedQuery.UserID,
		orgID:       configSavedQuery.OrgID,
	}
	return sq, nil
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
