package events

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// EventByID looks up and returns the Event with the given GraphQL ID. If no such Event exists, it
// returns a non-nil error.
func EventByID(ctx context.Context, id graphql.ID) (graphqlbackend.ToEvent, error) {
	dbID, err := unmarshalEventID(id)
	if err != nil {
		return graphqlbackend.ToEvent{}, err
	}
	return EventByDBID(ctx, dbID)
}

// EventByDBID looks up and returns the Event with the given database ID. If no such Event exists,
// it returns a non-nil error.
func EventByDBID(ctx context.Context, dbID int64) (graphqlbackend.ToEvent, error) {
	v, err := dbEvents{}.GetByID(ctx, dbID)
	if err != nil {
		return graphqlbackend.ToEvent{}, err
	}
	return toRegisteredEventType(ctx, v)
}

func marshalEventID(id int64) graphql.ID {
	return relay.MarshalID("Event", id)
}

func unmarshalEventID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}
