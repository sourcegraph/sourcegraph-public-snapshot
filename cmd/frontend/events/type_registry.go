package events

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// ToGraphQLEventFunc is a func called to populate the ToEvent struct from the untyped event data
// queried from the database.
type ToGraphQLEventFunc func(context.Context, graphqlbackend.EventCommon, EventData, *graphqlbackend.ToEvent) error

// Register an event type and its function to convert untyped event data to a GraphQL event type. It
// must be called at init time.
func Register(typeName string, converter ToGraphQLEventFunc) {
	if _, exists := converters[typeName]; exists {
		panic("event type is already registered: " + typeName)
	}
	converters[typeName] = converter
}

var converters = map[string]ToGraphQLEventFunc{}

// UnregisteredEventTypeError is an error that occurs when there is no registered converter (to a
// GraphQL event) for an event in the database.
type UnregisteredEventTypeError struct{ Type string }

func (e *UnregisteredEventTypeError) Error() string {
	return fmt.Sprintf("no converter is registered for event type %q", e.Type)
}

type EventData struct {
	Objects
	Data []byte
}

func toRegisteredEventType(ctx context.Context, v *dbEvent) (graphqlbackend.ToEvent, error) {
	converter, ok := converters[v.Type]
	if !ok {
		return graphqlbackend.ToEvent{}, &UnregisteredEventTypeError{Type: v.Type}
	}

	actorUser, err := graphqlbackend.UserByIDInt32(ctx, v.ActorUserID)
	if err != nil {
		return graphqlbackend.ToEvent{}, err
	}

	var toEvent graphqlbackend.ToEvent
	err = converter(ctx,
		graphqlbackend.EventCommon{
			ID:        marshalEventID(v.ID),
			Actor:     graphqlbackend.Actor{User: actorUser},
			CreatedAt: graphqlbackend.DateTime{v.CreatedAt},
		},
		EventData{Objects: v.Objects, Data: v.Data},
		&toEvent,
	)
	return toEvent, err
}
