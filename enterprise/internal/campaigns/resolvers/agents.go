package resolvers

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const agentIDKind = "Agent"

func marshalAgentID(id int64) graphql.ID {
	return relay.MarshalID(agentIDKind, id)
}

func unmarshalAgentID(id graphql.ID) (agentID int64, err error) {
	err = relay.UnmarshalSpec(id, &agentID)
	return
}

type agentResolver struct {
	// todo

	store *ee.Store
}

func (r *Resolver) AgentByID(ctx context.Context, id graphql.ID) (graphqlbackend.AgentResolver, error) {
	// todo: permissions

	dbId, err := unmarshalAgentID(id)
	if err != nil {
		return nil, err
	}

	// todo: fetch agent from DB and remove this statement
	if dbId == 0 {
		return nil, errors.New("Not implemented")
	}

	return &agentResolver{store: r.store}, nil
}

func (r *agentResolver) ID() graphql.ID {
	return marshalAgentID(123)
}

func (r *agentResolver) Name() string {
	return "agent-sg-dev-123"
}

func (r *agentResolver) Description() string {
	return "macOS 10.15.3, Docker 19.06.03, 8 CPU"
}

func (r *agentResolver) State() campaigns.AgentState {
	return campaigns.AgentStateOnline
}

func (r *agentResolver) RunningJobs() graphqlbackend.ActionJobConnectionResolver {
	// todo: missing agent param
	return &actionJobConnectionResolver{store: r.store}
}
