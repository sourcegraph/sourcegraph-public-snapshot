package events

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Type is an event type.
type Type string

// ToGraphQLEventFunc is a func called to populate the ToEvent struct from the untyped event data
// queried from the database.
type ToGraphQLEventFunc func(context.Context, graphqlbackend.EventCommon, EventData, *graphqlbackend.ToEvent) error

// Register an event type and its function to convert untyped event data to a GraphQL event type. It
// must be called at init time.
func Register(eventType Type, converter ToGraphQLEventFunc) {
	if _, exists := converters[eventType]; exists {
		panic("event type is already registered: " + eventType)
	}
	converters[eventType] = converter
}

var converters = map[Type]ToGraphQLEventFunc{}

// UnregisteredEventTypeError is an error that occurs when there is no registered converter (to a
// GraphQL event) for an event in the database.
type UnregisteredEventTypeError struct{ Type }

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

	var actor *graphqlbackend.Actor
	switch {
	case v.ActorUserID != 0:
		user, err := graphqlbackend.UserByIDInt32(ctx, int32(v.ActorUserID))
		if err != nil {
			return graphqlbackend.ToEvent{}, err
		}
		actor = &graphqlbackend.Actor{User: user}
	case v.ExternalActorUsername != "" || v.ExternalActorURL != "":
		actor = &graphqlbackend.Actor{
			ExternalActor: &graphqlbackend.ExternalActor{
				Username_: v.ExternalActorUsername,
				URL_:      v.ExternalActorURL,
			},
		}
	}

	var toEvent graphqlbackend.ToEvent
	err := converter(ctx,
		graphqlbackend.EventCommon{
			ID_:        marshalEventID(v.ID),
			Actor_:     actor,
			CreatedAt_: graphqlbackend.DateTime{v.CreatedAt},
		},
		EventData{Objects: v.Objects, Data: v.Data},
		&toEvent,
	)
	return toEvent, err
}
