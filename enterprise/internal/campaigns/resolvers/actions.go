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
	store  *ee.Store
	action campaigns.Action
}

func (r *Resolver) ActionByID(ctx context.Context, id graphql.ID) (graphqlbackend.ActionResolver, error) {
	// todo: permissions

	dbId, err := unmarshalActionID(id)
	if err != nil {
		return nil, err
	}

	action, err := r.store.GetAction(ctx, ee.GetActionOpts{
		ID: dbId,
	})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}
	return &actionResolver{store: r.store, action: *action}, nil
}

func (r *actionResolver) ID() graphql.ID {
	return marshalActionID(r.action.ID)
}

func (r *actionResolver) Name() string {
	return r.action.Name
}

func (r *actionResolver) Autoupdate() bool {
	return false
}

func (r *actionResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{envStr: r.action.EnvStr, steps: r.action.Steps}
}

func (r *actionResolver) SavedSearch(ctx context.Context) (*graphqlbackend.SavedSearchResolver, error) {
	if r.action.SavedSearchID == 0 {
		return nil, nil
	}
	return graphqlbackend.SavedSearchByIDInt32(ctx, r.action.SavedSearchID)
}

func (r *actionResolver) Schedule() *string {
	return r.action.Schedule
}

func (r *actionResolver) CancelPreviousScheduledExecution() bool {
	return r.action.CancelPrevious
}

func (r *actionResolver) Campaign(ctx context.Context) (graphqlbackend.CampaignResolver, error) {
	if r.action.CampaignID == 0 {
		return nil, nil
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: r.action.CampaignID})
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *actionResolver) ActionExecutions(args *graphqlbackend.ListActionExecutionsArgs) graphqlbackend.ActionExecutionConnectionResolver {
	return &actionExecutionConnectionResolver{store: r.store, first: args.First, actionID: r.action.ID}
}

// Connection resolver.
type actionConnectionResolver struct {
	once  sync.Once
	store *ee.Store

	first *int32

	actions    []*campaigns.Action
	totalCount int64
	err        error
}

func (r *actionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, totalCount, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	// todo: dangerous
	return int32(totalCount), nil
}

func (r *actionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ActionResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ActionResolver, len(nodes))
	for i, node := range nodes {
		resolvers[i] = &actionResolver{store: r.store, action: *node}
	}
	return resolvers, nil
}

func (r *actionConnectionResolver) compute(ctx context.Context) ([]*campaigns.Action, int64, error) {
	r.once.Do(func() {
		limit := -1
		if r.first != nil {
			limit = int(*r.first)
		}
		r.actions, r.totalCount, r.err = r.store.ListActions(ctx, ee.ListActionsOpts{Limit: limit, Cursor: 0})
	})
	return r.actions, r.totalCount, r.err
}
