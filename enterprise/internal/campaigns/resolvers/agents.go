package resolvers

import (
	"context"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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

func (r *agentResolver) RunningJobs(args *graphqlutil.ConnectionArgs) graphqlbackend.ActionJobConnectionResolver {
	runningState := campaigns.ActionJobStateRunning
	return &actionJobConnectionResolver{store: r.store, agentID: r.agent.ID, first: args.First, state: &runningState}
}

// Connection resolver.
type agentConnectionResolver struct {
	first *int32
	// Pass to only retrieve agents in the given state.
	state *campaigns.AgentState

	store *ee.Store

	once       sync.Once
	agents     []*campaigns.Agent
	totalCount int64
	err        error
}

func (r *agentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, totalCount, err := r.compute(ctx)
	// todo: unsafe
	return int32(totalCount), err
}

func (r *agentConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.AgentResolver, error) {
	agents, _, err := r.compute(ctx)
	resolvers := make([]graphqlbackend.AgentResolver, len(agents))
	for i, agent := range agents {
		resolvers[i] = &agentResolver{store: r.store, agent: agent}
	}
	return resolvers, err
}

func (r *agentConnectionResolver) compute(ctx context.Context) ([]*campaigns.Agent, int64, error) {
	r.once.Do(func() {
		var limit = -1
		if r.first != nil {
			limit = int(*r.first)
		}
		r.agents, _, r.err = r.store.ListAgents(ctx, ee.ListAgentsOpts{
			Limit: limit,
			State: r.state,
		})
		if r.err != nil {
			return
		}
		r.totalCount, r.err = r.store.CountAgents(ctx, ee.CountAgentsOpts{
			State: r.state,
		})
	})
	return r.agents, r.totalCount, r.err
}
