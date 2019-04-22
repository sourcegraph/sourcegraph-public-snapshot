package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type savedSearchResolver struct {
	id                                  string
	description                         string
	query                               string
	showOnHomepage, notify, notifySlack bool
	ownerKind                           string
	userID, orgID                       *int32
	slackWebhookURL                     *string
}

func (r savedSearchResolver) ID() string {
	return r.id
}

func (r savedSearchResolver) Notify() bool {
	return r.notify
}

func (r savedSearchResolver) NotifySlack() bool {
	return r.notifySlack
}

func (r savedSearchResolver) Description() string { return r.description }

func (r savedSearchResolver) Query() string { return r.query }

func (r savedSearchResolver) OwnerKind() string        { return r.ownerKind }
func (r savedSearchResolver) UserID() *int32           { return r.userID }
func (r savedSearchResolver) OrgID() *int32            { return r.orgID }
func (r savedSearchResolver) SlackWebhookURL() *string { return r.slackWebhookURL }

func toSavedSearchResolver(entry api.ConfigSavedQuery) *savedSearchResolver {
	return &savedSearchResolver{
		id:              entry.Key,
		description:     entry.Description,
		query:           entry.Query,
		notify:          entry.Notify,
		notifySlack:     entry.NotifySlack,
		ownerKind:       entry.OwnerKind,
		userID:          entry.UserID,
		orgID:           entry.OrgID,
		slackWebhookURL: entry.SlackWebhookURL,
	}
}

func (r *schemaResolver) SavedSearches(ctx context.Context) ([]*savedSearchResolver, error) {
	var savedSearches []*savedSearchResolver
	allSavedSearches, err := db.SavedSearches.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, savedSearch := range allSavedSearches {
		savedSearches = append(savedSearches, toSavedSearchResolver(savedSearch.Config))
	}

	return savedSearches, nil
}

func (r *schemaResolver) SendSavedSearchTestNotification(ctx context.Context, args *struct {
	ID string
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Look it up to ensure the actor has access to it.
	savedSearch, err := db.SavedSearches.GetSavedSearchByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	var spec api.SavedQueryIDSpec
	if savedSearch.UserID != nil {
		spec = api.SavedQueryIDSpec{Subject: api.SettingsSubject{User: savedSearch.UserID}, Key: savedSearch.Key}
	} else if savedSearch.OrgID != nil {
		spec = api.SavedQueryIDSpec{Subject: api.SettingsSubject{Org: savedSearch.OrgID}, Key: savedSearch.Key}
	}

	go queryrunnerapi.Client.TestNotification(context.Background(), spec)
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) CreateSavedSearch(ctx context.Context, args *struct {
	Description string
	Query       string
	NotifyOwner bool
	NotifySlack bool
	OwnerKind   string
	OrgID       *int32
	UserID      *int32
}) (*savedSearchResolver, error) {
	configSavedQuery, err := db.SavedSearches.Create(ctx, &types.SavedSearch{Description: args.Description, Query: args.Query, Notify: args.NotifyOwner, NotifySlack: args.NotifySlack, OwnerKind: args.OwnerKind, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, err
	}

	sq := &savedSearchResolver{
		id:          configSavedQuery.Key,
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
}) (*savedSearchResolver, error) {
	configSavedQuery, err := db.SavedSearches.Update(ctx, &types.SavedSearch{ID: args.ID, Description: args.Description, Query: args.Query, Notify: args.NotifyOwner, NotifySlack: args.NotifySlack, OwnerKind: args.OwnerKind, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, err
	}
	sq := &savedQueryResolver{
		id:          configSavedQuery.Key,
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
