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
func EventByID(ctx context.Context, id graphql.ID) (*gqlEvent, error) {
	dbID, err := unmarshalEventID(id)
	if err != nil {
		return nil, err
	}
	return eventByDBID(ctx, dbID)
}

// EventByDBID looks up and returns the Event with the given database ID. If no such Event exists,
// it returns a non-nil error.
func EventByDBID(ctx context.Context, dbID int64) (*gqlEvent, error) {
	v, err := dbEvents{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return newGQLEvent(v), nil
}

func marshalEventID(id int64) graphql.ID {
	return relay.MarshalID("Event", id)
}

func unmarshalEventID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func toEventCommon(v *dbEvent) graphqlbackend.EventCommon {
	return &graphqlbackend.EventCommon{
		ID:        marshalEventID(v.ID),
		Actor:     graphqlbackend.Actor{}, // TODO!(sqs)
		CreatedAt: graphqlbackend.DateTime{v.CreatedAt},
	}
}
