package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// Objects refers to the foreign key relationships that an event may have, to refer to objects
// related to the event.
type Objects struct {
	Thread               int64
	ThreadDiagnosticEdge int64
	Campaign             int64
	Comment              int64
	Rule                 int64
	Repository           int32
	User                 int32
	Organization         int32
	RegistryExtension    int32
}

// CreationData is the data required to create an event (in CreateEvent).
type CreationData struct {
	Type // event type
	Objects
	Data interface{} // JSON-marshaled

	// zero value for these fields means ctx's actor
	ActorUserID           int32
	ExternalActorUsername string
	ExternalActorURL      string

	CreatedAt time.Time // zero value means time.Now() as of CreateEvent call
}

var MockCreateEvent func(CreationData) error

// CreateEvent creates an event in the database.
func CreateEvent(ctx context.Context, tx *sql.Tx, event CreationData) error {
	if MockCreateEvent != nil {
		return MockCreateEvent(event)
	}
	return createEvent(ctx, tx, event, 0)
}

func createEvent(ctx context.Context, tx *sql.Tx, event CreationData, importedFromExternalServiceID int64) error {
	v := &dbEvent{
		Type:                          event.Type,
		ActorUserID:                   event.ActorUserID,
		ExternalActorUsername:         event.ExternalActorUsername,
		ExternalActorURL:              event.ExternalActorURL,
		Objects:                       event.Objects,
		CreatedAt:                     event.CreatedAt,
		ImportedFromExternalServiceID: importedFromExternalServiceID,
	}
	if event.Data != nil {
		var err error
		v.Data, err = json.Marshal(event.Data)
		if err != nil {
			return err
		}
	}
	if v.ActorUserID == 0 && v.ExternalActorUsername == "" && v.ExternalActorURL == "" {
		actor, err := graphqlbackend.CurrentUser(ctx)
		if err != nil {
			return err
		}
		if actor == nil {
			return errors.New("actor required to create event")
		}
		v.ActorUserID = actor.DatabaseID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}

	_, err := dbEvents{}.Create(ctx, tx, v)
	return err
}

var MockImportExternalEvents func(externalServiceID int64, objects Objects, toImport []CreationData) error

// ImportExternalEvents replaces all existing events for the objects from the given external service
// with a new set of events.
func ImportExternalEvents(ctx context.Context, externalServiceID int64, objects Objects, toImport []CreationData) error {
	if MockImportExternalEvents != nil {
		return MockImportExternalEvents(externalServiceID, objects, toImport)
	}

	tx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	if externalServiceID == 0 {
		panic("externalServiceID must be nonzero")
	}
	opt := dbEventsListOptions{
		Objects:                       objects,
		ImportedFromExternalServiceID: externalServiceID,
	}

	// Delete all existing events for the objects from the given external service.
	if err := (dbEvents{}).Delete(ctx, tx, opt); err != nil {
		return err
	}

	// Insert the new events.
	for _, event := range toImport {
		if err := createEvent(ctx, tx, event, externalServiceID); err != nil {
			return err
		}
	}
	return nil
}
