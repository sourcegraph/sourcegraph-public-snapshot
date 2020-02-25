package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const actionExecutionIDKind = "ActionExecution"

func marshalActionExecutionID(id int64) graphql.ID {
	return relay.MarshalID(actionExecutionIDKind, id)
}

func unmarshalActionExecutionID(id graphql.ID) (actionExecutionID int64, err error) {
	err = relay.UnmarshalSpec(id, &actionExecutionID)
	return
}

type actionExecutionResolver struct {
	store *ee.Store

	actionExecution campaigns.ActionExecution

	// todo: use this for passing down the action when a parent resolver was an action resolver
	action *campaigns.Action
}

func (r *Resolver) ActionExecutionByID(ctx context.Context, id graphql.ID) (graphqlbackend.ActionExecutionResolver, error) {
	// todo: permissions
	dbId, err := unmarshalActionExecutionID(id)
	if err != nil {
		return nil, err
	}

	actionExecution, err := r.store.ActionExecutionByID(ctx, ee.ActionExecutionByIDOpts{
		ID: dbId,
	})
	if err != nil {
		return nil, err
	}
	if actionExecution.ID == 0 {
		return nil, nil
	}

	return &actionExecutionResolver{store: r.store, actionExecution: *actionExecution}, nil
}

func (r *actionExecutionResolver) ID() graphql.ID {
	return marshalActionExecutionID(r.actionExecution.ID)
}

func (r *actionExecutionResolver) Action(ctx context.Context) (graphqlbackend.ActionResolver, error) {
	if r.action != nil {
		return &actionResolver{store: r.store, action: *r.action}, nil
	}
	action, err := r.store.ActionByID(ctx, ee.ActionByIDOpts{ID: r.actionExecution.ActionID})
	if err != nil {
		return nil, err
	}
	return &actionResolver{store: r.store, action: *action}, nil
}

func (r *actionExecutionResolver) InvokationReason() campaigns.ActionExecutionInvokationReason {
	return r.actionExecution.InvokationReason
}

func (r *actionExecutionResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{steps: r.actionExecution.Steps, envStr: *r.actionExecution.EnvStr}
}

func (r *actionExecutionResolver) ActionWorkspace() *graphqlbackend.GitTreeEntryResolver {
	return nil
}

func (r *actionExecutionResolver) Jobs() graphqlbackend.ActionJobConnectionResolver {
	return &actionJobConnectionResolver{}
}

func (r *actionExecutionResolver) Status() campaigns.BackgroundProcessStatus {
	return campaigns.BackgroundProcessStatus{
		Canceled:      false,
		Total:         int32(100),
		Completed:     int32(10),
		Pending:       int32(90),
		ProcessState:  campaigns.BackgroundProcessStateProcessing,
		ProcessErrors: []string{},
	}
}

func (r *actionExecutionResolver) CampaignPlan(ctx context.Context) (graphqlbackend.CampaignPlanResolver, error) {
	if r.actionExecution.CampaignPlanID == nil {
		return nil, nil
	}
	plan, err := r.store.GetCampaignPlan(ctx, ee.GetCampaignPlanOpts{ID: *r.actionExecution.CampaignPlanID})
	if err != nil {
		return nil, err
	}

	return &campaignPlanResolver{store: r.store, campaignPlan: plan}, nil
}
