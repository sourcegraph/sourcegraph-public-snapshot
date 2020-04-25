package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const agentIDKind = "Agent"

const agentOfflineTimeout = "2m"

func marshalAgentID(id int64) graphql.ID {
	return relay.MarshalID(agentIDKind, id)
}

func unmarshalAgentID(id graphql.ID) (agentID int64, err error) {
	err = relay.UnmarshalSpec(id, &agentID)
	return
}

type agentResolver struct {
	agent *campaigns.Agent

	store *ee.Store
}

func (r *Resolver) AgentByID(ctx context.Context, id graphql.ID) (graphqlbackend.AgentResolver, error) {
	// todo: permissions

	dbId, err := unmarshalAgentID(id)
	if err != nil {
		return nil, err
	}

	return AgentByIDInt64(ctx, r.store, dbId)
}

func AgentByIDInt64(ctx context.Context, store *ee.Store, id int64) (graphqlbackend.AgentResolver, error) {
	agent, err := store.GetAgent(ctx, ee.GetAgentOpts{ID: id})

	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &agentResolver{store: store, agent: agent}, nil
}

func (r *agentResolver) ID() graphql.ID {
	return marshalAgentID(r.agent.ID)
}

func (r *agentResolver) Name() string {
	return r.agent.Name
}

func (r *agentResolver) Specs() string {
	return r.agent.Specs
}

func (r *agentResolver) State() (campaigns.AgentState, error) {
	timeout, err := time.ParseDuration(agentOfflineTimeout)
	if err != nil {
		return campaigns.AgentStateOnline, err
	}
	if time.Since(r.agent.LastSeenAt) > timeout {
		return campaigns.AgentStateOffline, nil
	}
	return campaigns.AgentStateOnline, nil
}

func (r *agentResolver) RunningJobs() graphqlbackend.ActionJobConnectionResolver {
	return &actionJobConnectionResolver{store: r.store, agentID: r.agent.ID}
}
