package events

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func GetEventConnection(ctx context.Context, arg *graphqlbackend.EventsArgs) (graphqlbackend.EventConnection, error) {
	var opt dbEventsListOptions
	if arg.Since != nil {
		opt.Since = arg.Since.Time
	}
	return eventsByOptions(ctx, opt, &arg.ConnectionArgs)
}

func eventsByOptions(ctx context.Context, opt dbEventsListOptions, arg *graphqlutil.ConnectionArgs) (graphqlbackend.EventConnection, error) {
	list, err := dbEvents{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	events := make([]graphqlbackend.ToEvent, len(list))
	for i, a := range list {
		var err error
		events[i], err = toRegisteredEventType(ctx, a)
		if err != nil {
			return nil, err
		}
	}
	return &eventConnection{arg: arg, events: events}, nil
}

type eventConnection struct {
	arg    *graphqlutil.ConnectionArgs
	events []graphqlbackend.ToEvent
}

func (r *eventConnection) Nodes(ctx context.Context) ([]graphqlbackend.ToEvent, error) {
	events := r.events
	if first := r.arg.First; first != nil && len(events) > int(*first) {
		events = events[:int(*first)]
	}
	return events, nil
}

func (r *eventConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.events)), nil
}

func (r *eventConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.events)), nil
}
