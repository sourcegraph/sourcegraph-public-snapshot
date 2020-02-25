package resolvers

import (
	"context"

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
	store  *ee.Store
	action campaigns.Action
}

func (r *Resolver) ActionByID(ctx context.Context, id graphql.ID) (graphqlbackend.ActionResolver, error) {
	// todo: permissions

	dbId, err := unmarshalActionID(id)
	if err != nil {
		return nil, err
	}

	action, err := r.store.ActionByID(ctx, ee.ActionByIDOpts{
		ID: dbId,
	})
	if err != nil {
		return nil, err
	}
	if action.ID == 0 {
		return nil, nil
	}
	return &actionResolver{store: r.store, action: *action}, nil
}

func (r *actionResolver) ID() graphql.ID {
	return marshalActionID(r.action.ID)
}

func (r *actionResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{envStr: r.action.EnvStr, steps: r.action.Steps}
}

func (r *actionResolver) SavedSearch(ctx context.Context) (*graphqlbackend.SavedSearchResolver, error) {
	if r.action.SavedSearchID == nil {
		return nil, nil
	}
	return graphqlbackend.SavedSearchByIDInt32(ctx, *r.action.SavedSearchID)
}

func (r *actionResolver) Schedule() *string {
	return r.action.Schedule
}

func (r *actionResolver) CancelPreviousScheduledExecution() bool {
	return r.action.CancelPrevious
}

func (r *actionResolver) Campaign(ctx context.Context) (graphqlbackend.CampaignResolver, error) {
	if r.action.CampaignID == nil {
		return nil, nil
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: *r.action.CampaignID})
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *actionResolver) ActionExecutions(args *graphqlbackend.ListActionExecutionsArgs) graphqlbackend.ActionExecutionConnectionResolver {
	return &actionExecutionConnectionResolver{store: r.store, first: args.First, actionID: r.action.ID}
}
