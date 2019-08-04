package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// EventConnectionCommonArgs contains the common set of arguments for connections of events.
type EventConnectionCommonArgs struct {
	graphqlutil.ConnectionArgs
	Since *DateTime
}

// event is the common interface for event GraphQL types.
type event interface {
	ID() graphql.ID
	Actor() Actor
	CreatedAt() DateTime
}

// EventCommon is the interface for the GraphQL interface EventCommon.
type EventCommon struct {
	ID_        graphql.ID
	Actor_     Actor
	CreatedAt_ DateTime
}

func (v *EventCommon) ID() graphql.ID      { return v.ID_ }
func (v *EventCommon) Actor() Actor        { return v.Actor_ }
func (v *EventCommon) CreatedAt() DateTime { return v.CreatedAt_ }

type ToEvent struct {
	CreateThreadEvent             *CreateThreadEvent
	AddThreadToCampaignEvent      *AddRemoveThreadToFromCampaignEvent
	RemoveThreadFromCampaignEvent *AddRemoveThreadToFromCampaignEvent
	ReviewEvent                   *ReviewEvent
	ReviewRequestedEvent          *ReviewRequestedEvent
}

func (v ToEvent) ToCreateThreadEvent() (*CreateThreadEvent, bool) {
	return v.CreateThreadEvent, v.CreateThreadEvent != nil
}

func (v ToEvent) ToAddThreadToCampaignEvent() (*AddRemoveThreadToFromCampaignEvent, bool) {
	return v.AddThreadToCampaignEvent, v.AddThreadToCampaignEvent != nil
}

func (v ToEvent) ToRemoveThreadFromCampaignEvent() (*AddRemoveThreadToFromCampaignEvent, bool) {
	return v.RemoveThreadFromCampaignEvent, v.RemoveThreadFromCampaignEvent != nil
}

func (v ToEvent) ToReviewEvent() (*ReviewEvent, bool) {
	return v.ReviewEvent, v.ReviewEvent != nil
}

func (v ToEvent) ToReviewRequestedEvent() (*ReviewRequestedEvent, bool) {
	return v.ReviewRequestedEvent, v.ReviewRequestedEvent != nil
}

// EventConnection is the interface for GraphQL connection types for event nodes.
type EventConnection interface {
	Nodes(context.Context) ([]ToEvent, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
