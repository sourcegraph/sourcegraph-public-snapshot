package events

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
)

func (GraphQLResolver) Events(ctx context.Context, arg *graphqlbackend.EventsArgs) (graphqlbackend.EventConnection, error) {
	var opt dbEventsListOptions
	if arg.Object != nil {
		_, threadID, err := threadlike.UnmarshalID(*arg.Object)
		if err != nil {
			return nil, err
		}
		opt.ObjectThreadID = threadID
	}
	return eventsByOptions(ctx, opt, &arg.ConnectionArgs)
}

func (GraphQLResolver) EventsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.EventConnection, error) {
	var opt dbEventsListOptions
	var err error
	opt.NamespaceUserID, opt.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return eventsByOptions(ctx, opt, arg)
}

func (GraphQLResolver) EventsWithObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.EventConnection, error) {
	return GraphQLResolver{}.Events(ctx, &graphqlbackend.EventsArgs{Object: &object, ConnectionArgs: *arg})
}

func eventsByOptions(ctx context.Context, opt dbEventsListOptions, arg *graphqlutil.ConnectionArgs) (graphqlbackend.EventConnection, error) {
	list, err := dbEvents{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	events := make([]*gqlEvent, len(list))
	for i, a := range list {
		events[i] = newGQLEvent(a)
	}
	return &eventConnection{arg: arg, events: events}, nil
}

type eventConnection struct {
	arg    *graphqlutil.ConnectionArgs
	events []*gqlEvent
}

func (r *eventConnection) Nodes(ctx context.Context) ([]graphqlbackend.Event, error) {
	events := r.events
	if first := r.arg.First; first != nil && len(events) > int(*first) {
		events = events[:int(*first)]
	}

	events2 := make([]graphqlbackend.Event, len(events))
	for i, l := range events {
		events2[i] = l
	}
	return events2, nil
}

func (r *eventConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.events)), nil
}

func (r *eventConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.events)), nil
}
