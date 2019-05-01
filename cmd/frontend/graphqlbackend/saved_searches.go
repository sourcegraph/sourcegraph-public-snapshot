package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type savedSearchResolver struct {
	id                  string
	description         string
	query               string
	notify, notifySlack bool
	ownerKind           string
	userID, orgID       *int32
	slackWebhookURL     *string
}

func marshalSavedSearchID(savedSearchID string) graphql.ID {
	return relay.MarshalID("SavedSearch", savedSearchID)
}

func unmarshalSavedSearchID(id graphql.ID) (savedSearchID int32, err error) {
	err = relay.UnmarshalSpec(id, &savedSearchID)
	return
}

func (r savedSearchResolver) ID() graphql.ID {
	return marshalSavedSearchID(r.id)
}

func (r savedSearchResolver) DatabaseID() string {
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
	currentUser, err := CurrentUser(ctx)
	if currentUser == nil {
		return nil, errors.New("No currently authenticated user")
	}
	if err != nil {
		return nil, err
	}
	allSavedSearches, err := db.SavedSearches.ListSavedSearchesByUserID(ctx, currentUser.DatabaseID())
	if err != nil {
		return nil, err
	}
	for _, savedSearch := range allSavedSearches {
		savedSearches = append(savedSearches, toSavedSearchResolver(api.ConfigSavedQuery{
			Key:         string(savedSearch.ID),
			Description: savedSearch.Description,
			Query:       savedSearch.Query,
			Notify:      savedSearch.Notify,
			NotifySlack: savedSearch.NotifySlack,
			OwnerKind:   savedSearch.OwnerKind,
			UserID:      savedSearch.UserID,
			OrgID:       savedSearch.OrgID,
		}))
	}

	return savedSearches, nil
}

func (r *schemaResolver) SendSavedSearchTestNotification(ctx context.Context, args *struct {
	ID string
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins should be able to send test notifications.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	savedSearch, err := db.SavedSearches.GetSavedSearchByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}

	go queryrunnerapi.Client.TestNotification(context.Background(), *savedSearch)
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
	// 🚨 SECURITY: Make sure the current user has permission to create a saved search for the specified user or org.
	if args.UserID != nil {
		if err := backend.CheckSiteAdminOrSameUser(ctx, *args.UserID); err != nil {
			return nil, err
		}
	} else if args.OrgID != nil {
		if err := backend.CheckOrgAccess(ctx, *args.OrgID); err != nil {
			return nil, err
		}
	}

	configSavedQuery, err := db.SavedSearches.Create(ctx, &types.SavedSearch{Description: args.Description, Query: args.Query, Notify: args.NotifyOwner, NotifySlack: args.NotifySlack, OwnerKind: args.OwnerKind, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, err
	}

	sq := &savedSearchResolver{
		id:          configSavedQuery.ID,
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
	// 🚨 SECURITY: Make sure the current user has permission to update a saved search for the specified user or org.
	if args.UserID != nil {
		if err := backend.CheckSiteAdminOrSameUser(ctx, *args.UserID); err != nil {
			return nil, err
		}
	} else if args.OrgID != nil {
		if err := backend.CheckOrgAccess(ctx, *args.OrgID); err != nil {
			return nil, err
		}
	}

	configSavedQuery, err := db.SavedSearches.Update(ctx, &types.SavedSearch{ID: args.ID, Description: args.Description, Query: args.Query, Notify: args.NotifyOwner, NotifySlack: args.NotifySlack, OwnerKind: args.OwnerKind, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, err
	}
	sq := &savedSearchResolver{
		id:          configSavedQuery.ID,
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
	ss, err := db.SavedSearches.GetSavedSearchByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Make sure the current user has permission to delete a saved search for the specified user or org.
	if ss.Config.UserID != nil {
		if err := backend.CheckSiteAdminOrSameUser(ctx, *ss.Config.UserID); err != nil {
			return nil, err
		}
	} else if ss.Config.OrgID != nil {
		if err := backend.CheckOrgAccess(ctx, *ss.Config.OrgID); err != nil {
			return nil, err
		}
	}
	err = db.SavedSearches.Delete(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
