package events

import (
	"testing"
)

func TestEventsSubscribe(t *testing.T) {
	var miltonWoofEvent EventID = "milton.woof"

	s := newEventServer()
	s.subscribe(miltonWoofEvent, func(payload struct{}) {})

	if len(s.callbacks) != 1 {
		t.Errorf("Failed to subscribe to event")
	}
}

func TestEventsDispatch(t *testing.T) {
	var miltonWoofEvent EventID = "milton.woof"

	called := make(chan int)
	defer close(called)
	callback := func(payload struct{}) { called <- 1 }

	s := newEventServer()
	s.subscribe(miltonWoofEvent, callback)
	s.publish(Event{
		EventID: miltonWoofEvent,
		Payload: struct{}{},
	})

	// If this recieve results in a deadlock error, the callback is not being
	// executed as expected.
	<-called
}

func TestEventsPublishPayload(t *testing.T) {
	var miltonWoofEvent EventID = "milton.woof"
	expectedPayload := 42

	receivePayload := make(chan int)
	defer close(receivePayload)
	callback := func(payload int) { receivePayload <- payload }

	s := newEventServer()
	s.subscribe(miltonWoofEvent, callback)
	s.publish(Event{
		EventID: miltonWoofEvent,
		Payload: expectedPayload,
	})

	if received := <-receivePayload; received != expectedPayload {
		t.Errorf("Expected payload value %d got %d", expectedPayload, received)
	}
}
