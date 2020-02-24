package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	// ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
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
	action          *actionResolver
	actionExecution campaigns.ActionExecution
}

func (r *actionExecutionResolver) ID() graphql.ID {
	return marshalActionExecutionID(r.actionExecution.ID)
}

func (r *actionExecutionResolver) Action() graphqlbackend.ActionResolver {
	return r.action
}

func (r *actionExecutionResolver) InvokationReason() campaigns.ActionExecutionInvokationReason {
	return campaigns.ActionExecutionInvokationReasonManual
}

func (r *actionExecutionResolver) Definition() graphqlbackend.ActionDefinitionResolver {
	return &actionDefinitionResolver{}
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

func (r *actionExecutionResolver) CampaignPlan() graphqlbackend.CampaignPlanResolver {
	return nil
}
