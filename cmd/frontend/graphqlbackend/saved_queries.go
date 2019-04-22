package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type savedQueryResolver struct {
	id                                  string
	description                         string
	query                               string
	showOnHomepage, notify, notifySlack bool
	ownerKind                           string
	userID, orgID                       *int32
	slackWebhookURL                     *string
}

func (r savedQueryResolver) ID() string {
	return r.id
}

func (r savedQueryResolver) Notify() bool {
	return r.notify
}

func (r savedQueryResolver) NotifySlack() bool {
	return r.notifySlack
}

func (r savedQueryResolver) Description() string { return r.description }

func (r savedQueryResolver) Query() string { return r.query }

func (r savedQueryResolver) OwnerKind() string        { return r.ownerKind }
func (r savedQueryResolver) UserID() *int32           { return r.userID }
func (r savedQueryResolver) OrgID() *int32            { return r.orgID }
func (r savedQueryResolver) SlackWebhookURL() *string { return r.slackWebhookURL }

func toSavedQueryResolver(entry api.ConfigSavedQuery) *savedQueryResolver {
	return &savedQueryResolver{
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

func (r *schemaResolver) SavedSearches(ctx context.Context) ([]*savedQueryResolver, error) {
	var savedQueries []*savedQueryResolver
	savedSearches, err := db.SavedSearches.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, savedSearch := range savedSearches {
		savedQueries = append(savedQueries, toSavedQueryResolver(savedSearch.Config))
	}

	return savedQueries, nil
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
