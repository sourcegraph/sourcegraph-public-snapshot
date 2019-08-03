package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type EventsArgs struct {
	graphqlutil.ConnectionArgs
	Since *DateTime
}

// event is the common interface for event GraphQL types.
type event interface {
	ID() graphql.ID
	Actor() Actor
	CreatedAt() DateTime
}

type EventCommon struct {
	ID        graphql.ID
	Actor     Actor
	CreatedAt DateTime
}

// TODO!(sqs): notneeded because UseStructFields
//
// func (v *EventCommon) ID() graphql.ID      { return v.ID_ }
// func (v *EventCommon) Actor() Actor        { return v.Actor_ }
// func (v *EventCommon) CreatedAt() DateTime { return v.CreatedAt_ }

type ToEvent struct {
	AddThreadToCampaignEvent      *AddRemoveThreadToFromCampaignEvent
	RemoveThreadFromCampaignEvent *AddRemoveThreadToFromCampaignEvent
}

func (v ToEvent) ToAddThreadToCampaignEvent() (*AddRemoveThreadToFromCampaignEvent, bool) {
	return v.AddThreadToCampaignEvent, v.AddThreadToCampaignEvent != nil
}

func (v ToEvent) ToRemoveThreadFromCampaignEvent() (*AddRemoveThreadToFromCampaignEvent, bool) {
	return v.RemoveThreadFromCampaignEvent, v.RemoveThreadFromCampaignEvent != nil
}

// EventConnection is the interface for GraphQL connection types for event nodes.
type EventConnection interface {
	Nodes(context.Context) ([]ToEvent, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
