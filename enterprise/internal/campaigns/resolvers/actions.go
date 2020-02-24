package resolvers

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const actionIDKind = "Action"

func marshalActionID(id int64) graphql.ID {
	return relay.MarshalID(actionIDKind, id)
}

func unmarshalActionID(id graphql.ID) (actionID int64, err error) {
	err = relay.UnmarshalSpec(id, &actionID)
	return
}

type actionResolver struct {
	store *ee.Store
	once  sync.Once

	// todo: remove
	campaignID *graphql.ID

	action campaigns.Action
}

func (r *actionResolver) ID() graphql.ID {
	return marshalActionID(r.action.ID)
}

func (r *actionResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{}
}

func (r *actionResolver) SavedSearch() *graphqlbackend.SavedSearchResolver {
	return nil
}

func (r *actionResolver) Schedule() *string {
	schedStr := "30 */2 * * *"
	return &schedStr
}

func (r *actionResolver) CancelPreviousScheduledExecution() *bool {
	return nil
}

func (r *actionResolver) Campaign(ctx context.Context) (graphqlbackend.CampaignResolver, error) {
	// todo: much of this is copied from campaignbyid resolver
	if r.campaignID == nil {
		return nil, nil
	}
	campaignID, err := unmarshalCampaignID(*r.campaignID)
	if err != nil {
		return nil, err
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *actionResolver) ActionExecutions() graphqlbackend.ActionExecutionConnectionResolver {
	return &actionExecutionConnectionResolver{store: r.store}
}
