package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func campaignParticipants(ctx context.Context, campaign interface {
	threadsGetter
}) (graphqlbackend.ParticipantConnection, error) {
	threads, err := campaign.getThreads(ctx)
	if err != nil {
		return nil, err
	}

	byActor := map[actor.DBColumns]*participantEdge{}
	getParticipant := func(a graphqlbackend.Actor) *participantEdge {
		key := actor.FromGQL(&a)
		edge, ok := byActor[key]
		if !ok {
			edge = &participantEdge{actor: a}
			byActor[key] = edge
		}
		return edge
	}

	// The campaign's thread assignees.
	for _, thread := range threads {
		assigneeConnection, err := thread.Assignees(ctx, &graphqlutil.ConnectionArgs{})
		if err != nil {
			return nil, err
		}
		assignees, err := assigneeConnection.Nodes(ctx)
		if err != nil {
			return nil, err
		}
		for _, assignee := range assignees {
			edge := getParticipant(assignee)
			edge.reasons = append(edge.reasons, graphqlbackend.ParticipantReasonCodeOwner, graphqlbackend.ParticipantReasonAssignee)
		}
	}

	// The campaign's author.
	var author *graphqlbackend.Actor
	switch campaign := campaign.(type) {
	case *gqlCampaign:
		author, err = campaign.Author(ctx)
	case *gqlCampaignPreview:
		author, err = campaign.Author(ctx)
	}
	if err != nil {
		return nil, err
	}
	if author != nil {
		edge := getParticipant(*author)
		edge.reasons = append(edge.reasons, graphqlbackend.ParticipantReasonAuthor)
	}

	edges := make([]*participantEdge, 0, len(byActor))
	for _, edge := range byActor {
		edges = append(edges, edge)
	}
	return constParticipantConnection(toParticipantEdges(edges)), nil
}

type participantEdge struct {
	actor   graphqlbackend.Actor
	reasons []graphqlbackend.ParticipantReason
}

func (v *participantEdge) Actor(context.Context) (graphqlbackend.Actor, error) { return v.actor, nil }

func (v *participantEdge) Reasons() []graphqlbackend.ParticipantReason { return v.reasons }

func toParticipantEdges(in []*participantEdge) (out []graphqlbackend.ParticipantEdge) {
	out = make([]graphqlbackend.ParticipantEdge, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

type constParticipantConnection []graphqlbackend.ParticipantEdge

func (c constParticipantConnection) Edges(context.Context) ([]graphqlbackend.ParticipantEdge, error) {
	return []graphqlbackend.ParticipantEdge(c), nil
}

func (c constParticipantConnection) Nodes(ctx context.Context) ([]graphqlbackend.Actor, error) {
	actors := make([]graphqlbackend.Actor, len(c))
	for i, e := range c {
		var err error
		actors[i], err = e.Actor(ctx)
		if err != nil {
			return nil, err
		}
	}
	return actors, nil
}

func (c constParticipantConnection) TotalCount(context.Context) (int32, error) {
	return int32(len(c)), nil
}

func (c constParticipantConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
