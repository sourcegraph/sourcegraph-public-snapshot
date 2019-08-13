package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type ParticipantConnectionArgs struct {
	graphqlutil.ConnectionArgs
}

type ParticipantEdge interface {
	Actor(context.Context) (Actor, error)
	Reasons() []ParticipantReason
}

type ParticipantReason string

const (
	ParticipantReasonCodeOwner ParticipantReason = "CODE_OWNER"
	ParticipantReasonAssignee                    = "ASSIGNEE"
	ParticipantReasonAuthor                      = "AUTHOR"
)

type hasParticipants interface {
	Participants(context.Context, *ParticipantConnectionArgs) (ParticipantConnection, error)
}

// ParticipantConnection is the interface for the GraphQL type ParticipantConnection.
type ParticipantConnection interface {
	Edges(context.Context) ([]ParticipantEdge, error)
	Nodes(context.Context) ([]Actor, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
