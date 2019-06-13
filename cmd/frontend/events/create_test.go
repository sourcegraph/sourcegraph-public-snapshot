package events

import (
	"reflect"
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	resetMocks()
	creationData := CreationData{
		Type:        "t",
		Objects:     Objects{Thread: 1},
		Data:        "d",
		ActorUserID: 2,
		CreatedAt:   time.Now(),
	}
	mocks.events.Create = func(event *dbEvent) (*dbEvent, error) {
		if wantEvent := (&dbEvent{
			Type:        creationData.Type,
			Objects:     creationData.Objects,
			Data:        []byte(`"d"`),
			ActorUserID: 2,
			CreatedAt:   creationData.CreatedAt,
		}); !reflect.DeepEqual(event, wantEvent) {
			t.Errorf("got event %+v, want %+v", event, wantEvent)
		}
		return nil, nil
	}

	if err := CreateEvent(nil, nil, creationData); err != nil {
		t.Fatal(err)
	}
}
