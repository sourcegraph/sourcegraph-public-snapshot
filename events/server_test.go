package events

import "testing"

func TestEventsSubscribe(t *testing.T) {
	var miltonWoofEvent EventID = "milton.woof"

	s := newEventServer()
	if err := s.subscribe(miltonWoofEvent, func(id EventID, payload struct{}) {}); err != nil {
		t.Fatal(err)
	}

	if len(s.callbacks) != 1 {
		t.Errorf("Failed to subscribe to event")
	}
}

func TestEventsDispatch(t *testing.T) {
	var miltonWoofEvent EventID = "milton.woof"

	called := make(chan int)
	defer close(called)
	callback := func(id EventID, payload struct{}) { called <- 1 }

	s := newEventServer()
	if err := s.subscribe(miltonWoofEvent, callback); err != nil {
		t.Fatal(err)
	}
	s.publish(miltonWoofEvent, struct{}{})

	// If this recieve results in a deadlock error, the callback is not being
	// executed as expected.
	<-called
}

func TestEventsPublishPayload(t *testing.T) {
	var miltonWoofEvent EventID = "milton.woof"
	expectedPayload := 42

	receivePayload := make(chan int)
	defer close(receivePayload)
	callback := func(id EventID, payload int) { receivePayload <- payload }

	s := newEventServer()
	if err := s.subscribe(miltonWoofEvent, callback); err != nil {
		t.Fatal(err)
	}
	s.publish(miltonWoofEvent, expectedPayload)

	if received := <-receivePayload; received != expectedPayload {
		t.Errorf("Expected payload value %d got %d", expectedPayload, received)
	}
}
