package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func campaignParticipants(ctx context.Context, campaign interface {
	threadsGetter
	eventsGetter
}) (graphqlbackend.ParticipantConnection, error) {
	var edges []*participantEdge
	// TODO!(sqs): hack, dummy data
	users, err := db.Users.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		gqlUser, err := graphqlbackend.UserByIDInt32(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		edges = append(edges, &participantEdge{actor: graphqlbackend.Actor{User: gqlUser}})
	}
	return constParticipantConnection(toParticipantEdges(edges)), nil
}

type participantEdge struct {
	actor graphqlbackend.Actor
}

func (v *participantEdge) Actor(context.Context) (graphqlbackend.Actor, error) { return v.actor, nil }

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
