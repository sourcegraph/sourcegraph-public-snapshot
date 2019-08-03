package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Objects refers to the foreign key relationships that an event may have, to refer to objects
// related to the event.
type Objects struct {
	Campaign          int64
	Thread            int64
	Comment           int64
	Rule              int64
	Repository        int32
	User              int32
	Organization      int32
	RegistryExtension int32
}

// CreationData is the data required to create an event (in CreateEvent).
type CreationData struct {
	Type string // event type
	Objects
	Data interface{} // JSON-marshaled

	ActorUserID int32     // zero value means ctx's actor
	CreatedAt   time.Time // zero value means time.Now() as of CreateEvent call
}

// CreateEvent creates an event in the database.
func CreateEvent(ctx context.Context, event CreationData) error {
	v := &dbEvent{
		Type:      event.Type,
		Objects:   event.Objects,
		CreatedAt: event.CreatedAt,
	}
	if event.Data != nil {
		var err error
		v.Data, err = json.Marshal(event.Data)
		if err != nil {
			return err
		}
	}
	if v.ActorUserID == 0 {
		actor, err := graphqlbackend.CurrentUser(ctx)
		if err != nil {
			return err
		}
		if actor == nil {
			return 0, errors.New("actor required to create event")
		}
		v.ActorUserID = actor.DatabaseID()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	_, err := dbEvents{}.Create(ctx, v)
	return err
}
